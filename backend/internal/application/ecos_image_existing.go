package application

import (
	"context"

	ecospkg "railkeeper/backend/internal/ecos"
)

// Reusing a slot never permits replacing its content. Both preview and confirmation
// require the same pixel verification as a new upload, alongside the target grant.
func previewECoSImageSlot(ctx context.Context, web ecosImageWeb, input ECoSImagePreviewInput) (int, error) {
	if input.ExistingSlot != nil {
		return *input.ExistingSlot, verifyExistingECoSImage(ctx, web, *input.ExistingSlot, input.BMP)
	}
	slot, err := web.FindFreeSlot(ctx)
	if err != nil {
		return 0, err
	}
	return slot, web.CheckUploadForm(ctx, slot)
}

func verifyExistingECoSImage(ctx context.Context, web ecosImageWeb, slot int, source []byte) error {
	if err := web.CheckStoredSlot(ctx, slot); err != nil {
		return ErrDigitalCenterPreviewStale
	}
	actual, err := web.ReadImage(ctx, "user", slot)
	if err != nil {
		return err
	}
	if !ecospkg.VerifyImage(source, actual) {
		return ErrDigitalCenterPreviewStale
	}
	return nil
}
