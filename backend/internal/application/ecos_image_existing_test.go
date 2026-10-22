package application

import (
	"errors"
	"strings"
	"testing"

	ecospkg "railkeeper/backend/internal/ecos"
)

func TestECoSImageControlRefusalCanBeRetriedWithoutAnotherUpload(t *testing.T) {
	s, d, w, _, data := imageTransferFixture(t)
	p := previewImage(t, s, data)
	d.assignError = ErrECoSImageControlBusy
	result, err := s.ConfirmECoSImage(t.Context(), "vehicle-1", ECoSImageConfirmInput{Token: p.Token, Confirm: true}, "admin")
	if err != nil || result.Status != "control_busy" || !result.ImageVerified || result.AssignmentVerified || w.uploads != 1 {
		t.Fatalf("result=%+v err=%v uploads=%d", result, err, w.uploads)
	}
	input := ECoSImagePreviewInput{ObjectID: 1001, Description: "BR 218", BMP: data, ExistingSlot: &result.Slot}
	p, err = s.PreviewECoSImage(t.Context(), "vehicle-1", input, "admin")
	if err != nil || !p.ExistingImage || p.Slot != 7 || d.assigns != 1 || w.uploads != 1 {
		t.Fatalf("preview=%+v err=%v", p, err)
	}
	d.assignError = nil
	result, err = s.ConfirmECoSImage(t.Context(), "vehicle-1", ECoSImageConfirmInput{Token: p.Token, Confirm: true}, "admin")
	if err != nil || result.Status != "verified" || !result.AssignmentVerified || d.assigns != 2 || w.uploads != 1 {
		t.Fatalf("result=%+v err=%v uploads=%d assignments=%d", result, err, w.uploads, d.assigns)
	}
	entries := s.auth.(*digitalCenterWriteAuditStub).entries
	if len(entries) != 4 || !strings.Contains(entries[1].details, `"status":"control_busy"`) || !strings.Contains(entries[3].details, `"existingImage":true`) {
		t.Fatal("missing refusal/reuse provenance")
	}
	if _, err := s.ConfirmECoSImage(t.Context(), "vehicle-1", ECoSImageConfirmInput{Token: p.Token, Confirm: true}, "admin"); !errors.Is(err, ErrDigitalCenterGrantMismatch) || w.uploads != 1 || d.assigns != 2 {
		t.Fatal("repeated assignment without fresh consent")
	}
}

func TestECoSExistingImageRequiresFreshContentTargetActorAndAudit(t *testing.T) {
	for _, change := range []string{"changed pixels", "removed", "target", "actor", "audit", "slot zero"} {
		t.Run(change, func(t *testing.T) {
			s, d, w, _, data := imageTransferFixture(t)
			slot := 0
			p, err := s.PreviewECoSImage(t.Context(), "vehicle-1", ECoSImagePreviewInput{ObjectID: 1001, Description: "BR 218", BMP: data, ExistingSlot: &slot}, "admin")
			if err != nil || !p.ExistingImage || p.Slot != 0 {
				t.Fatal(p, err)
			}
			actor := "admin"
			switch change {
			case "changed pixels":
				w.badImage = true
			case "removed":
				w.storedError = ecospkg.ErrImageNotStored
			case "target":
				d.loco.Address++
			case "actor":
				actor = "other"
			case "audit":
				s.auth.(*digitalCenterWriteAuditStub).err = errors.New("disk full")
			}
			result, err := s.ConfirmECoSImage(t.Context(), "vehicle-1", ECoSImageConfirmInput{Token: p.Token, Confirm: true}, actor)
			if change == "slot zero" {
				if err != nil || result.Status != "verified" || d.assigns != 1 {
					t.Fatal(result, err)
				}
			} else if err == nil || d.assigns != 0 {
				t.Fatal("accepted changed/unapproved existing image", result, err)
			}
			if w.uploads != 0 {
				t.Fatal("uploaded during assignment-only flow")
			}
		})
	}
}

func TestECoSExistingImageRejectsInvalidOrUnverifiedSlotsAtPreview(t *testing.T) {
	for _, slot := range []int{-1, 500, 7} {
		s, d, w, _, data := imageTransferFixture(t)
		w.badImage = true
		_, err := s.PreviewECoSImage(t.Context(), "vehicle-1", ECoSImagePreviewInput{ObjectID: 1001, Description: "BR 218", BMP: data, ExistingSlot: &slot}, "admin")
		if err == nil || len(s.imageGrants) != 0 || d.assigns != 0 || w.uploads != 0 {
			t.Fatal("accepted invalid reuse", slot, err)
		}
	}
}

func TestECoSExistingImageAlreadyAssignedDoesNotRequestControl(t *testing.T) {
	s, d, w, _, data := imageTransferFixture(t)
	d.descriptor.ImageType = "IMAGE_TYPE_USER"
	d.descriptor.Index = 1
	d.assignError = ErrECoSImageControlBusy
	p, err := s.PreviewECoSImage(t.Context(), "vehicle-1", ECoSImagePreviewInput{ObjectID: 1001, Description: "BR 218", BMP: data, ExistingSlot: &d.descriptor.Index}, "admin")
	if err != nil {
		t.Fatal(err)
	}
	result, err := s.ConfirmECoSImage(t.Context(), "vehicle-1", ECoSImageConfirmInput{Token: p.Token, Confirm: true}, "admin")
	if err != nil || result.Status != "verified" || !result.AssignmentVerified || w.uploads != 0 || d.assigns != 0 {
		t.Fatal(result, err)
	}
}
