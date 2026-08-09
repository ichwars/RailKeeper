package application

import (
	"context"
	"strings"

	"railkeeper/backend/internal/domain"
)

type AllocationTargetInput struct {
	VehicleID    string `json:"vehicleId,omitempty"`
	LayoutID     string `json:"layoutId,omitempty"`
	LayoutUnitID string `json:"layoutUnitId,omitempty"`
}

type AccessoryTechnicalPlacement struct {
	Placement      string `json:"placement,omitempty"`
	DigitalAddress string `json:"digitalAddress,omitempty"`
	DecoderOutput  string `json:"decoderOutput,omitempty"`
	Connection     string `json:"connection,omitempty"`
	WiringNotes    string `json:"wiringNotes,omitempty"`
}

type AccessoryReservation struct {
	ID         string `json:"id"`
	ProductID  string `json:"productId"`
	AssetID    string `json:"assetId,omitempty"`
	LocationID string `json:"locationId"`
	Quantity   int    `json:"quantity"`
	AllocationTargetInput
	Placement      string                            `json:"placement,omitempty"`
	DigitalAddress string                            `json:"digitalAddress,omitempty"`
	DecoderOutput  string                            `json:"decoderOutput,omitempty"`
	Connection     string                            `json:"connection,omitempty"`
	WiringNotes    string                            `json:"wiringNotes,omitempty"`
	Status         domain.AccessoryReservationStatus `json:"status"`
	Note           string                            `json:"note,omitempty"`
	CreatedBy      string                            `json:"createdBy"`
	CreatedAt      string                            `json:"createdAt"`
	UpdatedAt      string                            `json:"updatedAt"`
}

type CreateAccessoryReservationInput struct {
	ProductID  string `json:"productId"`
	AssetID    string `json:"assetId,omitempty"`
	LocationID string `json:"locationId"`
	Quantity   int    `json:"quantity"`
	AllocationTargetInput
	Placement      string `json:"placement,omitempty"`
	DigitalAddress string `json:"digitalAddress,omitempty"`
	DecoderOutput  string `json:"decoderOutput,omitempty"`
	Connection     string `json:"connection,omitempty"`
	WiringNotes    string `json:"wiringNotes,omitempty"`
	Note           string `json:"note,omitempty"`
}

type AccessoryInstallation struct {
	ID               string `json:"id"`
	ProductID        string `json:"productId"`
	AssetID          string `json:"assetId,omitempty"`
	SourceLocationID string `json:"sourceLocationId"`
	Quantity         int    `json:"quantity"`
	AllocationTargetInput
	Placement          string                             `json:"placement,omitempty"`
	DigitalAddress     string                             `json:"digitalAddress,omitempty"`
	DecoderOutput      string                             `json:"decoderOutput,omitempty"`
	Connection         string                             `json:"connection,omitempty"`
	WiringNotes        string                             `json:"wiringNotes,omitempty"`
	Condition          domain.AccessoryCondition          `json:"condition"`
	InstalledBy        string                             `json:"installedBy"`
	InstalledAt        string                             `json:"installedAt"`
	RemovedBy          string                             `json:"removedBy,omitempty"`
	RemovedAt          string                             `json:"removedAt,omitempty"`
	RemovalDisposition domain.AccessoryRemovalDisposition `json:"removalDisposition,omitempty"`
	Notes              string                             `json:"notes,omitempty"`
	RemovalNotes       string                             `json:"removalNotes,omitempty"`
}

type CreateAccessoryInstallationInput struct {
	ReservationID    string `json:"reservationId,omitempty"`
	ProductID        string `json:"productId"`
	AssetID          string `json:"assetId,omitempty"`
	SourceLocationID string `json:"sourceLocationId"`
	Quantity         int    `json:"quantity"`
	AllocationTargetInput
	Placement      string                    `json:"placement,omitempty"`
	DigitalAddress string                    `json:"digitalAddress,omitempty"`
	DecoderOutput  string                    `json:"decoderOutput,omitempty"`
	Connection     string                    `json:"connection,omitempty"`
	WiringNotes    string                    `json:"wiringNotes,omitempty"`
	Condition      domain.AccessoryCondition `json:"condition"`
	Notes          string                    `json:"notes,omitempty"`
}

type RemoveAccessoryInstallationInput struct {
	Disposition       domain.AccessoryRemovalDisposition `json:"disposition"`
	StorageLocationID string                             `json:"storageLocationId,omitempty"`
	Notes             string                             `json:"notes,omitempty"`
}

type UpdateAccessoryInstallationConditionInput struct {
	Condition domain.AccessoryCondition `json:"condition"`
}

type AccessoryAllocationSummary struct {
	ProductID string `json:"productId"`
	Owned     int    `json:"owned"`
	Stored    int    `json:"stored"`
	Reserved  int    `json:"reserved"`
	Installed int    `json:"installed"`
	Available int    `json:"available"`
	Missing   int    `json:"missing"`
}

type AccessoryUsageEventType string

const (
	AccessoryUsageReservation      AccessoryUsageEventType = "reservation"
	AccessoryUsageInstallation     AccessoryUsageEventType = "installation"
	AccessoryUsageConditionChanged AccessoryUsageEventType = "condition_changed"
	AccessoryUsageRemoval          AccessoryUsageEventType = "removal"
)

type AccessoryUsageEvent struct {
	ID             string                  `json:"id"`
	Type           AccessoryUsageEventType `json:"type"`
	ProductID      string                  `json:"productId"`
	ReservationID  string                  `json:"reservationId,omitempty"`
	InstallationID string                  `json:"installationId,omitempty"`
	AssetID        string                  `json:"assetId,omitempty"`
	LocationID     string                  `json:"locationId,omitempty"`
	Quantity       int                     `json:"quantity"`
	AllocationTargetInput
	AccessoryTechnicalPlacement
	Status             domain.AccessoryReservationStatus  `json:"status,omitempty"`
	PreviousCondition  domain.AccessoryCondition          `json:"previousCondition,omitempty"`
	Condition          domain.AccessoryCondition          `json:"condition,omitempty"`
	RemovalDisposition domain.AccessoryRemovalDisposition `json:"removalDisposition,omitempty"`
	Actor              string                             `json:"actor,omitempty"`
	OccurredAt         string                             `json:"occurredAt"`
}

type AccessoryUsageHistory struct {
	ProductID string                `json:"productId"`
	Events    []AccessoryUsageEvent `json:"events"`
}

type AccessoryAllocationRepository interface {
	ListReservations(context.Context, string) ([]AccessoryReservation, error)
	CreateReservation(context.Context, CreateAccessoryReservationInput, string) (*AccessoryReservation, error)
	CancelReservation(context.Context, string, string) (*AccessoryReservation, error)
	ListInstallations(context.Context, string) ([]AccessoryInstallation, error)
	Install(context.Context, CreateAccessoryInstallationInput, string) (*AccessoryInstallation, error)
	RemoveInstallation(
		context.Context,
		string,
		RemoveAccessoryInstallationInput,
		string,
	) (*AccessoryInstallation, error)
	UpdateInstallationCondition(
		context.Context,
		string,
		UpdateAccessoryInstallationConditionInput,
		string,
	) (*AccessoryInstallation, error)
	GetAllocationSummary(context.Context, string) (*AccessoryAllocationSummary, error)
	GetUsageHistory(context.Context, string) (*AccessoryUsageHistory, error)
}

type AccessoryAllocationService struct {
	repository AccessoryAllocationRepository
}

func NewAccessoryAllocationService(repository AccessoryAllocationRepository) *AccessoryAllocationService {
	return &AccessoryAllocationService{repository: repository}
}

func (s *AccessoryAllocationService) ListReservations(
	ctx context.Context,
	productID string,
) ([]AccessoryReservation, error) {
	return s.repository.ListReservations(ctx, strings.TrimSpace(productID))
}

func (s *AccessoryAllocationService) CreateReservation(
	ctx context.Context,
	input CreateAccessoryReservationInput,
	actor string,
) (*AccessoryReservation, error) {
	input = cleanReservationInput(input)
	if input.ProductID == "" || input.LocationID == "" || input.Quantity <= 0 ||
		!input.valid() || (input.AssetID != "" && input.Quantity != 1) {
		return nil, ErrAccessoryValidation
	}
	return s.repository.CreateReservation(ctx, input, actor)
}

func (s *AccessoryAllocationService) CancelReservation(
	ctx context.Context,
	id string,
	actor string,
) (*AccessoryReservation, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, ErrAccessoryValidation
	}
	return s.repository.CancelReservation(ctx, id, actor)
}

func (s *AccessoryAllocationService) ListInstallations(
	ctx context.Context,
	productID string,
) ([]AccessoryInstallation, error) {
	return s.repository.ListInstallations(ctx, strings.TrimSpace(productID))
}

func (s *AccessoryAllocationService) Install(
	ctx context.Context,
	input CreateAccessoryInstallationInput,
	actor string,
) (*AccessoryInstallation, error) {
	input = cleanInstallationInput(input)
	if input.Condition == "" {
		input.Condition = domain.AccessoryConditionUnknown
	}
	if input.ProductID == "" || input.SourceLocationID == "" || input.Quantity <= 0 ||
		!input.valid() || !input.Condition.Valid() ||
		(input.AssetID != "" && input.Quantity != 1) {
		return nil, ErrAccessoryValidation
	}
	return s.repository.Install(ctx, input, actor)
}

func (s *AccessoryAllocationService) RemoveInstallation(
	ctx context.Context,
	id string,
	input RemoveAccessoryInstallationInput,
	actor string,
) (*AccessoryInstallation, error) {
	id = strings.TrimSpace(id)
	input.StorageLocationID = strings.TrimSpace(input.StorageLocationID)
	input.Notes = strings.TrimSpace(input.Notes)
	stored := input.Disposition == domain.AccessoryRemovalStored
	if id == "" || !input.Disposition.Valid() || stored != (input.StorageLocationID != "") {
		return nil, ErrAccessoryValidation
	}
	return s.repository.RemoveInstallation(ctx, id, input, actor)
}

func (s *AccessoryAllocationService) UpdateInstallationCondition(
	ctx context.Context,
	id string,
	input UpdateAccessoryInstallationConditionInput,
	actor string,
) (*AccessoryInstallation, error) {
	id = strings.TrimSpace(id)
	if id == "" || !input.Condition.Valid() {
		return nil, ErrAccessoryValidation
	}
	return s.repository.UpdateInstallationCondition(ctx, id, input, actor)
}

func (s *AccessoryAllocationService) GetAllocationSummary(
	ctx context.Context,
	productID string,
) (*AccessoryAllocationSummary, error) {
	productID = strings.TrimSpace(productID)
	if productID == "" {
		return nil, ErrAccessoryValidation
	}
	return s.repository.GetAllocationSummary(ctx, productID)
}

func (s *AccessoryAllocationService) GetUsageHistory(
	ctx context.Context,
	productID string,
) (*AccessoryUsageHistory, error) {
	productID = strings.TrimSpace(productID)
	if productID == "" {
		return nil, ErrAccessoryValidation
	}
	return s.repository.GetUsageHistory(ctx, productID)
}

func cleanReservationInput(input CreateAccessoryReservationInput) CreateAccessoryReservationInput {
	input.ProductID = strings.TrimSpace(input.ProductID)
	input.AssetID = strings.TrimSpace(input.AssetID)
	input.LocationID = strings.TrimSpace(input.LocationID)
	input.AllocationTargetInput = input.clean()
	input.Placement = strings.TrimSpace(input.Placement)
	input.DigitalAddress = strings.TrimSpace(input.DigitalAddress)
	input.DecoderOutput = strings.TrimSpace(input.DecoderOutput)
	input.Connection = strings.TrimSpace(input.Connection)
	input.WiringNotes = strings.TrimSpace(input.WiringNotes)
	input.Note = strings.TrimSpace(input.Note)
	return input
}

func cleanInstallationInput(input CreateAccessoryInstallationInput) CreateAccessoryInstallationInput {
	input.ReservationID = strings.TrimSpace(input.ReservationID)
	input.ProductID = strings.TrimSpace(input.ProductID)
	input.AssetID = strings.TrimSpace(input.AssetID)
	input.SourceLocationID = strings.TrimSpace(input.SourceLocationID)
	input.AllocationTargetInput = input.clean()
	input.Placement = strings.TrimSpace(input.Placement)
	input.DigitalAddress = strings.TrimSpace(input.DigitalAddress)
	input.DecoderOutput = strings.TrimSpace(input.DecoderOutput)
	input.Connection = strings.TrimSpace(input.Connection)
	input.WiringNotes = strings.TrimSpace(input.WiringNotes)
	input.Notes = strings.TrimSpace(input.Notes)
	return input
}

func (target AllocationTargetInput) clean() AllocationTargetInput {
	target.VehicleID = strings.TrimSpace(target.VehicleID)
	target.LayoutID = strings.TrimSpace(target.LayoutID)
	target.LayoutUnitID = strings.TrimSpace(target.LayoutUnitID)
	return target
}

func (target AllocationTargetInput) valid() bool {
	count := 0
	for _, id := range []string{target.VehicleID, target.LayoutID, target.LayoutUnitID} {
		if id != "" {
			count++
		}
	}
	return count == 1
}
