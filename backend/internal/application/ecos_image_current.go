package application

import (
	"context"
	"errors"

	ecospkg "railkeeper/backend/internal/ecos"
)

func currentECoSImage(ctx context.Context, web ecosImageWeb, descriptor ECoSImageDescriptor) (dataURL string, missing bool, err error) {
	kind := "system"
	if descriptor.ImageType == "IMAGE_TYPE_USER" {
		// A locomotive can keep referencing a deleted user image. Only a complete
		// inventory reporting that slot free proves absence; HTTP errors do not.
		if err := web.CheckStoredSlot(ctx, descriptor.Index); err != nil {
			if errors.Is(err, ecospkg.ErrImageNotStored) {
				return "", true, nil
			}
			return "", false, err
		}
		kind = "user"
	}
	data, err := web.ReadImage(ctx, kind, descriptor.Index)
	if err != nil {
		return "", false, err
	}
	return imageDataURL(data), false, nil
}

func checkCurrentECoSImage(ctx context.Context, web ecosImageWeb, preview ECoSImagePreview) error {
	dataURL, missing, err := currentECoSImage(ctx, web, preview.Current)
	if err != nil {
		return err
	}
	if missing != preview.CurrentImageMissing || dataURL != preview.CurrentImage {
		return ErrDigitalCenterPreviewStale
	}
	return nil
}
