package application

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/color"
	"image/png"
	"strings"
	"sync"
	"testing"
	"time"

	"golang.org/x/image/bmp"
	ecospkg "railkeeper/backend/internal/ecos"
)

type imageDeviceStub struct {
	workspaceECoSReaderStub
	loco        ECoSLocomotive
	descriptor  ECoSImageDescriptor
	assigns     int
	assignError error
}

func (d *imageDeviceStub) ListLocomotives(context.Context, ECoSConnectionInput) ([]ECoSLocomotive, error) {
	return []ECoSLocomotive{d.loco}, nil
}
func (d *imageDeviceStub) ReadLocomotive(context.Context, ECoSConnectionInput, int) (ECoSLocomotive, error) {
	return d.loco, nil
}
func (d *imageDeviceStub) ReadImageDescriptor(context.Context, ECoSConnectionInput, int) (ECoSImageDescriptor, error) {
	return d.descriptor, nil
}
func (d *imageDeviceStub) AssignUserImage(_ context.Context, _ ECoSConnectionInput, _ int, index int) error {
	d.assigns++
	if d.assignError != nil {
		return d.assignError
	}
	d.descriptor.ImageType = "IMAGE_TYPE_USER"
	d.descriptor.Index = index
	return nil
}

type imageWebStub struct {
	address     string
	image       []byte
	uploads     int
	uploadError error
	checkError  error
	badImage    bool
	afterUpload func()
}

func (w *imageWebStub) Address() string                            { return w.address }
func (w *imageWebStub) FindFreeSlot(context.Context) (int, error)  { return 7, nil }
func (w *imageWebStub) CheckFreeSlot(context.Context, int) error   { return w.checkError }
func (w *imageWebStub) CheckStoredSlot(context.Context, int) error { return nil }
func (w *imageWebStub) CheckUploadForm(context.Context, int) error { return nil }
func (w *imageWebStub) ReadImage(_ context.Context, kind string, _ int) ([]byte, error) {
	if w.badImage && kind == "user" {
		return []byte("invalid"), nil
	}
	return w.image, nil
}
func (w *imageWebStub) Upload(_ context.Context, index int, description, kind string, data []byte) error {
	w.uploads++
	if w.afterUpload != nil {
		w.afterUpload()
	}
	return w.uploadError
}

func imageTransferFixture(t *testing.T) (*DigitalCenterWorkspaceService, *imageDeviceStub, *imageWebStub, *workspaceSettingsReaderStub, []byte) {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 190, 40))
	for y := 0; y < 40; y++ {
		for x := 0; x < 190; x++ {
			img.SetRGBA(x, y, color.RGBA{182, 182, 182, 255})
		}
	}
	var source, device bytes.Buffer
	if err := bmp.Encode(&source, img); err != nil {
		t.Fatal(err)
	}
	if err := png.Encode(&device, img); err != nil {
		t.Fatal(err)
	}
	if err := ecospkg.ValidateImageBMP(source.Bytes()); err != nil {
		t.Fatal(err)
	}
	d := &imageDeviceStub{loco: ECoSLocomotive{ObjectID: 1001, Name: "BR 218", Address: 3, Protocol: "DCC128"}, descriptor: ECoSImageDescriptor{LocomotiveType: "LOCO_TYPE_DIESEL", ImageType: "IMAGE_TYPE_INT", Index: 118}}
	w := &imageWebStub{address: "192.168.1.2", image: device.Bytes()}
	settings := &workspaceSettingsReaderStub{value: configuredECoSSettings()}
	s := NewDigitalCenterWorkspaceService(nil, settings, d, nil, &workspaceVehicleReaderStub{vehicles: []Vehicle{{ID: "vehicle-1"}}}, &digitalCenterWriteAuditStub{})
	s.imageWebFactory = func(context.Context, string) (ecosImageWeb, error) { return w, nil }
	return s, d, w, settings, source.Bytes()
}

func previewImage(t *testing.T, s *DigitalCenterWorkspaceService, data []byte) ECoSImagePreview {
	t.Helper()
	preview, err := s.PreviewECoSImage(t.Context(), "vehicle-1", ECoSImagePreviewInput{ObjectID: 1001, Description: "BR 218", BMP: data}, "admin")
	if err != nil {
		t.Fatal(err)
	}
	return preview
}

func TestECoSImagePreviewIsReadOnlyAndConfirmationIsSingleUse(t *testing.T) {
	s, d, w, _, data := imageTransferFixture(t)
	p := previewImage(t, s, data)
	data[100] = 0 // Stored bytes must not share the caller's mutable buffer.
	if w.uploads != 0 || d.assigns != 0 || p.Slot != 7 || p.Target.ObjectID != 1001 || !strings.HasPrefix(p.NewImage, "data:image/png;base64,") {
		t.Fatal("invalid/read-write preview")
	}
	if _, err := s.ConfirmECoSImage(t.Context(), "vehicle-1", ECoSImageConfirmInput{Token: p.Token}, "admin"); !errors.Is(err, ErrDigitalCenterConfirmationRequired) {
		t.Fatal(err)
	}
	result, err := s.ConfirmECoSImage(t.Context(), "vehicle-1", ECoSImageConfirmInput{Token: p.Token, Confirm: true}, "admin")
	if err != nil || result.Status != "verified" || !result.ImageVerified || !result.AssignmentVerified || !result.AuditRecorded || w.uploads != 1 || d.assigns != 1 {
		t.Fatalf("result=%+v err=%v uploads=%d assignments=%d", result, err, w.uploads, d.assigns)
	}
	if _, err := s.ConfirmECoSImage(t.Context(), "vehicle-1", ECoSImageConfirmInput{Token: p.Token, Confirm: true}, "admin"); !errors.Is(err, ErrDigitalCenterGrantMismatch) {
		t.Fatal("replayed token", err)
	}
	if w.uploads != 1 || d.assigns != 1 {
		t.Fatal("repeated device write")
	}
	entries := s.auth.(*digitalCenterWriteAuditStub).entries
	if len(entries) != 2 || entries[1].targetType != "vehicle" || !strings.Contains(entries[1].details, p.SHA256) || strings.Contains(entries[1].details, p.Token) {
		t.Fatal("invalid audit provenance")
	}
}

func TestECoSImageConfirmationRejectsChangedTargetsAndGrants(t *testing.T) {
	for _, test := range []struct {
		name           string
		change         func(*DigitalCenterWorkspaceService, *imageDeviceStub, *imageWebStub, *workspaceSettingsReaderStub)
		actor, vehicle string
	}{
		{name: "actor", actor: "other"}, {name: "vehicle", vehicle: "other"},
		{name: "expired", change: func(s *DigitalCenterWorkspaceService, _ *imageDeviceStub, _ *imageWebStub, _ *workspaceSettingsReaderStub) {
			s.now = func() time.Time { return time.Now().Add(time.Hour) }
		}},
		{name: "new preview", change: func(s *DigitalCenterWorkspaceService, _ *imageDeviceStub, _ *imageWebStub, _ *workspaceSettingsReaderStub) {
			_, _, _, _, data := imageTransferFixture(t)
			previewImage(t, s, data)
		}},
		{name: "name", change: func(_ *DigitalCenterWorkspaceService, d *imageDeviceStub, _ *imageWebStub, _ *workspaceSettingsReaderStub) {
			d.loco.Name = "Another"
		}},
		{name: "image", change: func(_ *DigitalCenterWorkspaceService, d *imageDeviceStub, _ *imageWebStub, _ *workspaceSettingsReaderStub) {
			d.descriptor.Index = 119
		}},
		{name: "host", change: func(_ *DigitalCenterWorkspaceService, _ *imageDeviceStub, _ *imageWebStub, s *workspaceSettingsReaderStub) {
			s.value.ECoS.Host = "other.local"
		}},
		{name: "dns", change: func(_ *DigitalCenterWorkspaceService, _ *imageDeviceStub, w *imageWebStub, _ *workspaceSettingsReaderStub) {
			w.address = "192.168.1.3"
		}},
		{name: "occupied", change: func(_ *DigitalCenterWorkspaceService, _ *imageDeviceStub, w *imageWebStub, _ *workspaceSettingsReaderStub) {
			w.checkError = ecospkg.ErrImageSlotOccupied
		}},
		{name: "audit unavailable", change: func(s *DigitalCenterWorkspaceService, _ *imageDeviceStub, _ *imageWebStub, _ *workspaceSettingsReaderStub) {
			s.auth.(*digitalCenterWriteAuditStub).err = errors.New("disk full")
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			s, d, w, settings, data := imageTransferFixture(t)
			p := previewImage(t, s, data)
			if test.change != nil {
				test.change(s, d, w, settings)
			}
			actor, vehicle := test.actor, test.vehicle
			if actor == "" {
				actor = "admin"
			}
			if vehicle == "" {
				vehicle = "vehicle-1"
			}
			if _, err := s.ConfirmECoSImage(t.Context(), vehicle, ECoSImageConfirmInput{Token: p.Token, Confirm: true}, actor); err == nil {
				t.Fatal("accepted stale or invalid confirmation")
			}
			if w.uploads != 0 || d.assigns != 0 {
				t.Fatal("wrote invalid preview")
			}
		})
	}
}

func TestECoSImageReportsPartialResultsWithoutRetryOrUnsafeAssignment(t *testing.T) {
	for _, status := range []string{"upload_unknown", "image_unverified", "assignment_unknown", "concurrent_change"} {
		t.Run(status, func(t *testing.T) {
			s, d, w, _, data := imageTransferFixture(t)
			p := previewImage(t, s, data)
			switch status {
			case "upload_unknown":
				w.uploadError = errors.New("lost reply")
			case "image_unverified":
				w.badImage = true
			case "assignment_unknown":
				d.assignError = ErrECoSWriteStateUnknown
			case "concurrent_change":
				w.afterUpload = func() { d.loco.Name = "Changed" }
			}
			result, err := s.ConfirmECoSImage(t.Context(), "vehicle-1", ECoSImageConfirmInput{Token: p.Token, Confirm: true}, "admin")
			want := status
			if want == "concurrent_change" {
				want = "assignment_unknown"
			}
			if err != nil || result.Status != want || w.uploads != 1 || result.AssignmentVerified {
				t.Fatalf("result=%+v err=%v uploads=%d", result, err, w.uploads)
			}
			if status != "assignment_unknown" && d.assigns != 0 {
				t.Fatal("assigned an unverified/changed target")
			}
		})
	}
}

type imageLiveDeviceStub struct {
	imageDeviceStub
	running                 bool
	pauses, resumes         int
	pauseError, resumeError error
}

func (d *imageLiveDeviceStub) LiveStatus() ECoSLiveStatus {
	return ECoSLiveStatus{State: ECoSLiveRunning, Connected: d.running}
}
func (d *imageLiveDeviceStub) StopLive() ECoSLiveStatus { d.running = false; return d.LiveStatus() }
func (d *imageLiveDeviceStub) PauseLive(context.Context) (ECoSLiveStatus, error) {
	d.pauses++
	if d.pauseError != nil {
		return d.LiveStatus(), d.pauseError
	}
	return d.StopLive(), nil
}
func (d *imageLiveDeviceStub) StartLive(context.Context, ECoSConnectionInput) (*ECoSLiveStatus, error) {
	return nil, errors.New("must preserve interruption callback through ResumePausedLive")
}
func (d *imageLiveDeviceStub) ResumePausedLive(context.Context, ECoSConnectionInput) (*ECoSLiveStatus, error) {
	d.resumes++
	if d.resumeError != nil {
		return nil, d.resumeError
	}
	d.running = true
	status := d.LiveStatus()
	return &status, nil
}

func TestECoSImagePausesAndResumesMonitoringAndReportsResumeFailure(t *testing.T) {
	for _, mode := range []string{"verified", "upload_unknown", "pause_failed", "resume_failed"} {
		t.Run(mode, func(t *testing.T) {
			s, base, w, _, data := imageTransferFixture(t)
			d := &imageLiveDeviceStub{imageDeviceStub: *base, running: true}
			s.ecos = d
			p := previewImage(t, s, data)
			switch mode {
			case "upload_unknown":
				w.uploadError = errors.New("lost reply")
			case "pause_failed":
				d.pauseError = errors.New("busy")
			case "resume_failed":
				d.resumeError = errors.New("offline")
			}
			result, err := s.ConfirmECoSImage(t.Context(), "vehicle-1", ECoSImageConfirmInput{Token: p.Token, Confirm: true}, "admin")
			if mode == "pause_failed" {
				if err == nil || w.uploads != 0 || d.assigns != 0 {
					t.Fatal("wrote without safely pausing")
				}
				return
			}
			if err != nil || d.pauses != 1 || d.resumes != 1 || !result.LiveMonitor.WasRunning || result.LiveMonitor.Restarted != (mode != "resume_failed") {
				t.Fatalf("result=%+v err=%v pauses=%d resumes=%d", result, err, d.pauses, d.resumes)
			}
		})
	}
}

func TestECoSImageConcurrentConfirmationWritesOnlyOnce(t *testing.T) {
	s, d, w, _, data := imageTransferFixture(t)
	preview := previewImage(t, s, data)
	var wg sync.WaitGroup
	results := make(chan error, 2)
	for range 2 {
		wg.Go(func() {
			_, err := s.ConfirmECoSImage(t.Context(), "vehicle-1", ECoSImageConfirmInput{Token: preview.Token, Confirm: true}, "admin")
			results <- err
		})
	}
	wg.Wait()
	close(results)
	success := 0
	for err := range results {
		if err == nil {
			success++
		} else if !errors.Is(err, ErrDigitalCenterGrantMismatch) {
			t.Fatal(err)
		}
	}
	if success != 1 || w.uploads != 1 || d.assigns != 1 {
		t.Fatalf("success=%d uploads=%d assignments=%d", success, w.uploads, d.assigns)
	}
}

func TestECoSImageFinishesVerificationAfterBrowserDisconnect(t *testing.T) {
	s, _, w, _, data := imageTransferFixture(t)
	preview := previewImage(t, s, data)
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	w.afterUpload = cancel
	result, err := s.ConfirmECoSImage(ctx, "vehicle-1", ECoSImageConfirmInput{Token: preview.Token, Confirm: true}, "admin")
	if err != nil || result.Status != "verified" || w.uploads != 1 {
		t.Fatalf("result=%+v err=%v", result, err)
	}
}
