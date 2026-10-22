package application

import (
	"errors"
	"testing"

	ecospkg "railkeeper/backend/internal/ecos"
)

func TestECoSImagePreviewCanReplaceReferenceToDeletedUserImage(t *testing.T) {
	s, d, w, _, data := imageTransferFixture(t)
	d.descriptor.ImageType = "IMAGE_TYPE_USER"
	d.descriptor.Index = 3
	w.storedBySlot = map[int]error{3: ecospkg.ErrImageNotStored}
	w.readErrors = map[int]error{3: ecospkg.ErrImageRead}
	p := previewImage(t, s, data)
	if !p.CurrentImageMissing || p.CurrentImage != "" || p.Current.Index != 3 || w.imageReads[3] != 0 || w.uploads != 0 || d.acquires != 0 {
		t.Fatal("invalid missing-image preview", p)
	}
	result, err := s.ConfirmECoSImage(t.Context(), "vehicle-1", ECoSImageConfirmInput{Token: p.Token, Confirm: true}, "admin")
	if err != nil || result.Status != "verified" || w.uploads != 1 || d.assigns != 1 || d.descriptor.Index != p.Slot || w.imageReads[3] != 0 {
		t.Fatal(result, err)
	}
}

func TestECoSMissingCurrentImageRequiresInventoryProofAndFreshState(t *testing.T) {
	for _, mode := range []string{"read failed", "inventory failed", "refilled before confirmation", "refilled during control"} {
		t.Run(mode, func(t *testing.T) {
			s, d, w, _, data := imageTransferFixture(t)
			d.descriptor.ImageType = "IMAGE_TYPE_USER"
			d.descriptor.Index = 3
			w.storedBySlot = map[int]error{3: ecospkg.ErrImageNotStored}
			switch mode {
			case "read failed":
				w.storedBySlot[3] = nil
				w.readErrors = map[int]error{3: ecospkg.ErrImageRead}
			case "inventory failed":
				w.storedBySlot[3] = ecospkg.ErrImageInventory
			}
			p, err := s.PreviewECoSImage(t.Context(), "vehicle-1", ECoSImagePreviewInput{ObjectID: 1001, Description: "BR 218", BMP: data}, "admin")
			if mode == "read failed" || mode == "inventory failed" {
				if err == nil || len(s.imageGrants) != 0 || d.acquires != 0 {
					t.Fatal("invented missing image without proof", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if mode == "refilled before confirmation" {
				w.storedBySlot[3] = nil
			} else {
				d.afterAcquire = func() { w.storedBySlot[3] = nil }
			}
			result, err := s.ConfirmECoSImage(t.Context(), "vehicle-1", ECoSImageConfirmInput{Token: p.Token, Confirm: true}, "admin")
			if mode == "refilled before confirmation" && !errors.Is(err, ErrDigitalCenterPreviewStale) {
				t.Fatal(result, err)
			}
			if mode == "refilled during control" && (err != nil || result.Status != "upload_blocked" || d.controlled || d.releases != 1) {
				t.Fatal(result, err)
			}
			if w.uploads != 0 || d.assigns != 0 {
				t.Fatal("wrote stale image state")
			}
		})
	}
}
