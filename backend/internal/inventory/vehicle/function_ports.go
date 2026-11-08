package vehicle

import "context"

type FunctionWriter interface {
	StoreFunction(context.Context, FunctionWrite) error
}
type FunctionMutation interface {
	FunctionWriter
	FindFunction(context.Context, string, string) (VehicleFunction, bool, error)
	DeleteDraftFunction(context.Context, string, string) error
}
type FunctionRepository interface {
	FunctionWriter
	FindFunction(context.Context, string, string) (VehicleFunction, bool, error)
	Functions(context.Context, string) ([]VehicleFunction, error)
	DeleteFunction(context.Context, string, string) error
}
type FunctionService struct {
	repository FunctionRepository
	parent     func(context.Context, string) error
}

func NewFunctionService(repository FunctionRepository, parent func(context.Context, string) error) *FunctionService {
	return &FunctionService{repository: repository, parent: parent}
}
