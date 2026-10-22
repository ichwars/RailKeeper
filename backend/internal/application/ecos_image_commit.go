package application

import (
	"context"
	"errors"

	ecospkg "railkeeper/backend/internal/ecos"
)

func completeECoSImageTransfer(writeCtx context.Context, device ecosImageDevice, web ecosImageWeb, grant ecosImageGrant, result *ECoSImageResult) error {
	connection := ECoSConnectionInput{Host: web.Address(), Port: grant.center.Port}
	preview := grant.preview
	descriptor := preview.Current
	var locomotive ECoSLocomotive
	var err error
	if preview.ExistingImage && !preview.CurrentImageMissing && descriptor.ImageType == "IMAGE_TYPE_USER" && descriptor.Index == preview.Slot {
		result.AssignmentVerified = true
		result.Status = "verified"
		return nil
	}
	control, err := device.AcquireImageControl(writeCtx, connection, preview.Target.ObjectID)
	if err != nil {
		if errors.Is(err, ErrECoSImageControlBusy) {
			result.Status = "control_busy"
		}
		return nil
	}
	defer func() { _ = control.Close() }()
	// Lock acquisition itself can race with edits made after preview validation.
	locomotive, err = device.ReadLocomotive(writeCtx, connection, preview.Target.ObjectID)
	if err != nil {
		return nil
	}
	descriptor, err = device.ReadImageDescriptor(writeCtx, connection, preview.Target.ObjectID)
	if err != nil || imageTarget(locomotive) != preview.Target || descriptor != preview.Current {
		return nil
	}
	if err := checkCurrentECoSImage(writeCtx, web, preview); err != nil {
		return nil
	}
	if !preview.ExistingImage {
		result.Status = "upload_unknown"
		if err := web.Upload(writeCtx, preview.Slot, preview.Description, imageUploadKind(descriptor.LocomotiveType), grant.data); err != nil {
			return nil
		}
	}
	result.Status = "image_unverified"
	result.ImageVerified = false
	if err := web.CheckStoredSlot(writeCtx, preview.Slot); err != nil {
		if errors.Is(err, ecospkg.ErrImageNotStored) {
			result.Status = "image_not_stored"
		}
		return nil
	}
	actual, err := web.ReadImage(writeCtx, "user", preview.Slot)
	if err != nil || !ecospkg.VerifyImage(grant.data, actual) {
		return nil
	}
	result.ImageVerified = true
	result.Status = "assignment_unknown"
	// An upload and assignment cannot be atomic. Re-read the locomotive once more
	// before assignment; preserve a concurrent change made from the ECoS handset.
	locomotive, err = device.ReadLocomotive(writeCtx, connection, preview.Target.ObjectID)
	if err != nil {
		return nil
	}
	descriptor, err = device.ReadImageDescriptor(writeCtx, connection, preview.Target.ObjectID)
	if err != nil || imageTarget(locomotive) != preview.Target || descriptor != preview.Current {
		return nil
	}
	if descriptor.ImageType == "IMAGE_TYPE_USER" && descriptor.Index == preview.Slot {
		result.AssignmentVerified = true
		result.Status = "verified"
		return nil
	}
	if err := control.Assign(writeCtx, preview.Slot); err != nil {
		return nil
	}
	verified, err := device.ReadImageDescriptor(writeCtx, connection, preview.Target.ObjectID)
	if err == nil && verified.ImageType == "IMAGE_TYPE_USER" && verified.Index == preview.Slot && verified.LocomotiveType == preview.Current.LocomotiveType {
		result.AssignmentVerified = true
		result.Status = "verified"
	}
	return nil
}
