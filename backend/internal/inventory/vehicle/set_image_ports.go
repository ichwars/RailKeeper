package vehicle

import (
	"context"

	"railkeeper/backend/internal/credentials"
)

type SetImageSelection struct {
	Mode          VehicleSetMainImageMode
	MemberImageID string
	Dedicated     VehicleSetImage
}
type SetImageCandidate struct {
	SetID                          string
	Mode                           VehicleSetMainImageMode
	MemberImageID, DedicatedBlobID string
	Image                          *VehicleImage
}
type PreviousSetImage struct{ BlobID, ThumbnailBlobID, CreatedAt string }
type SetImageLookup interface {
	Selection(context.Context, string) (SetImageSelection, error)
	MemberImage(context.Context, string, string) (*VehicleImage, error)
	AutomaticImage(context.Context, string) (*VehicleImage, error)
}
type SetImageMutation interface {
	Exists(context.Context, string) (bool, error)
	MemberBelongs(context.Context, string, string) (bool, error)
	DedicatedBlob(context.Context, string) (string, error)
	PreviousForUpload(context.Context, string) (PreviousSetImage, error)
	PreviousForDelete(context.Context, string) (PreviousSetImage, error)
	UpdateSelection(context.Context, string, VehicleSetMainImageInput, string) error
	StoreDedicated(context.Context, string, VehicleSetImageInput, string, string) error
	DeleteDedicated(context.Context, string, string) error
	Audit(context.Context, string, string, string, string, string) error
	Commit() error
	Rollback() error
}
type SetImageRepository interface {
	SetImageLookup
	DedicatedImage(context.Context, string) (VehicleSetImage, error)
	Candidates(context.Context, []string) ([]SetImageCandidate, error)
	BeginSetImages(context.Context) (SetImageMutation, error)
}
type SetImageService struct {
	repository SetImageRepository
	readSet    func(context.Context, string) (*VehicleSet, error)
}

func NewSetImageService(repository SetImageRepository, readSet func(context.Context, string) (*VehicleSet, error)) *SetImageService {
	return &SetImageService{repository: repository, readSet: readSet}
}
func (s *SetImageService) Resolve(ctx context.Context, id string) (VehicleSetMainImageMode, string, *VehicleSetImage, *VehicleSetMainImage, error) {
	return ResolveSetImage(ctx, s.repository, id)
}
func auditSetImage(ctx context.Context, tx SetImageMutation, actor, setID, action, now string) error {
	return tx.Audit(ctx, credentials.RandomID(), actor, setID, action, now)
}
