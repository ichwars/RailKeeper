package application

import (
	"errors"
	"strings"
	"testing"
)

func TestECoSImageDoesNotUploadUnlessControlIsReserved(t *testing.T) {
	for _, mode := range []string{"control_busy", "connection_failed", "target_changed"} {
		t.Run(mode, func(t *testing.T) {
			s, d, w, _, data := imageTransferFixture(t)
			p := previewImage(t, s, data)
			want, releases := "upload_blocked", 0
			switch mode {
			case "control_busy":
				d.controlError = ErrECoSImageControlBusy
				want = "control_busy"
			case "connection_failed":
				d.controlError = errors.New("connection lost")
			case "target_changed":
				d.afterAcquire = func() { d.loco.Address++ }
				releases = 1
			}
			result, err := s.ConfirmECoSImage(t.Context(), "vehicle-1", ECoSImageConfirmInput{Token: p.Token, Confirm: true}, "admin")
			if err != nil || result.Status != want || w.uploads != 0 || d.assigns != 0 || result.ImageVerified || result.AssignmentVerified || !result.AuditRecorded || d.controlled || d.releases != releases {
				t.Fatalf("result=%+v err=%v uploads=%d assignments=%d releases=%d", result, err, w.uploads, d.assigns, d.releases)
			}
			entries := s.auth.(*digitalCenterWriteAuditStub).entries
			if len(entries) != 2 || !strings.Contains(entries[1].details, `"status":"`+want+`"`) {
				t.Fatal("blocked transfer missing audit")
			}
			if _, err := s.ConfirmECoSImage(t.Context(), "vehicle-1", ECoSImageConfirmInput{Token: p.Token, Confirm: true}, "admin"); !errors.Is(err, ErrDigitalCenterGrantMismatch) {
				t.Fatal("replayed blocked confirmation")
			}
		})
	}
}

func TestECoSImageKeepsControlAcrossUploadAndReusesPartialUploadFromNewPreview(t *testing.T) {
	s, d, w, _, data := imageTransferFixture(t)
	p := previewImage(t, s, data)
	w.afterUpload = func() {
		if !d.controlled || d.releases != 0 || d.assigns != 0 {
			t.Error("upload outside exclusive control")
		}
		w.existing = true
	}
	d.assignError = ErrECoSWriteStateUnknown
	result, err := s.ConfirmECoSImage(t.Context(), "vehicle-1", ECoSImageConfirmInput{Token: p.Token, Confirm: true}, "admin")
	if err != nil || result.Status != "assignment_unknown" || w.uploads != 1 || d.controlled || d.releases != 1 {
		t.Fatal(result, err)
	}
	// A new preview has no client-side slot hint, just as after closing the dialog.
	p = previewImage(t, s, data)
	if !p.ExistingImage || p.Slot != 7 || w.uploads != 1 {
		t.Fatal("did not reuse existing image", p)
	}
	d.assignError = nil
	result, err = s.ConfirmECoSImage(t.Context(), "vehicle-1", ECoSImageConfirmInput{Token: p.Token, Confirm: true}, "admin")
	if err != nil || result.Status != "verified" || w.uploads != 1 || d.acquires != 2 || d.releases != 2 || d.controlled {
		t.Fatal(result, err)
	}
}

func TestECoSImageReleasesControlAfterEveryPostUploadFailure(t *testing.T) {
	for _, mode := range []string{"upload", "verification", "assignment"} {
		t.Run(mode, func(t *testing.T) {
			s, d, w, _, data := imageTransferFixture(t)
			p := previewImage(t, s, data)
			switch mode {
			case "upload":
				w.uploadError = errors.New("lost reply")
			case "verification":
				w.badImage = true
			case "assignment":
				d.assignError = ErrECoSWriteStateUnknown
			}
			_, err := s.ConfirmECoSImage(t.Context(), "vehicle-1", ECoSImageConfirmInput{Token: p.Token, Confirm: true}, "admin")
			if err != nil || d.acquires != 1 || d.releases != 1 || d.controlled {
				t.Fatal("leaked control", err, d)
			}
		})
	}
}
