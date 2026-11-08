package vehicle

import "context"

type ExternalMappingRepository interface {
	Mapping(context.Context, string, string) (*VehicleExternalMap, error)
	FindMapping(context.Context, string, string) (*VehicleExternalMap, bool, error)
	Mappings(context.Context, string) ([]VehicleExternalMap, error)
	WriteMapping(context.Context, ExternalMappingWrite) (*VehicleExternalMap, error)
	ConsolidateMappings(context.Context, string, string, VehicleExternalMapInput, string) error
	RebindMapping(context.Context, string, string, VehicleExternalMapInput, string) error
}
type ExternalMappingService struct {
	repository ExternalMappingRepository
	parent     func(context.Context, string) error
}

func NewExternalMappingService(repository ExternalMappingRepository, parent func(context.Context, string) error) *ExternalMappingService {
	return &ExternalMappingService{repository: repository, parent: parent}
}

type ExternalMappingWrite struct {
	ID        string
	VehicleID string
	Input     VehicleExternalMapInput
	At        string
}
