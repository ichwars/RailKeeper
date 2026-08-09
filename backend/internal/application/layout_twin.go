package application

import (
	"context"
	"strings"

	"railkeeper/backend/internal/domain"
)

type LayoutTwinStatus string

const (
	LayoutTwinPlanned        LayoutTwinStatus = "planned"
	LayoutTwinReserved       LayoutTwinStatus = "reserved"
	LayoutTwinInstalled      LayoutTwinStatus = "installed"
	LayoutTwinMaintenanceDue LayoutTwinStatus = "maintenance_due"
	LayoutTwinDefective      LayoutTwinStatus = "defective"
)

type LayoutTwinSelection struct {
	ConfigurationID string
	UnitID          string
}

type LayoutTwinPoint struct {
	XMM float64 `json:"xMm"`
	YMM float64 `json:"yMm"`
}

type LayoutTwinBounds struct {
	MinXMM   float64 `json:"minXMm"`
	MinYMM   float64 `json:"minYMm"`
	WidthMM  float64 `json:"widthMm"`
	HeightMM float64 `json:"heightMm"`
}

type LayoutTwinWarning struct {
	Code   string `json:"code"`
	UnitID string `json:"unitId,omitempty"`
}

type LayoutTwinAllocation struct {
	ID                    string                    `json:"id"`
	ProductID             string                    `json:"productId"`
	InventoryNumber       string                    `json:"inventoryNumber"`
	Manufacturer          string                    `json:"manufacturer"`
	ArticleNumber         string                    `json:"articleNumber,omitempty"`
	ProductName           string                    `json:"productName"`
	Quantity              int                       `json:"quantity"`
	ReservationStatus     string                    `json:"reservationStatus,omitempty"`
	InstallationCondition domain.AccessoryCondition `json:"installationCondition,omitempty"`
	Placement             string                    `json:"placement,omitempty"`
	DigitalAddress        string                    `json:"digitalAddress,omitempty"`
	DecoderOutput         string                    `json:"decoderOutput,omitempty"`
	Connection            string                    `json:"connection,omitempty"`
	WiringNotes           string                    `json:"wiringNotes,omitempty"`
	Note                  string                    `json:"note,omitempty"`
}

type LayoutTwinPosition struct {
	ID                   string                             `json:"id"`
	LayoutUnitID         string                             `json:"layoutUnitId"`
	Label                string                             `json:"label"`
	Kind                 domain.LayoutTechnicalPositionKind `json:"kind"`
	LocalXMM             float64                            `json:"localXMm"`
	LocalYMM             float64                            `json:"localYMm"`
	LocalRotationDegrees float64                            `json:"localRotationDegrees"`
	GlobalXMM            float64                            `json:"globalXMm"`
	GlobalYMM            float64                            `json:"globalYMm"`
	RotationDegrees      float64                            `json:"rotationDegrees"`
	ProductID            string                             `json:"productId,omitempty"`
	InventoryNumber      string                             `json:"inventoryNumber,omitempty"`
	Manufacturer         string                             `json:"manufacturer,omitempty"`
	ArticleNumber        string                             `json:"articleNumber,omitempty"`
	ProductName          string                             `json:"productName,omitempty"`
	Description          string                             `json:"description,omitempty"`
	Version              int                                `json:"version"`
	Statuses             []LayoutTwinStatus                 `json:"statuses"`
	Reservations         []LayoutTwinAllocation             `json:"reservations"`
	Installations        []LayoutTwinAllocation             `json:"installations"`
}

type LayoutTwinUnit struct {
	ID              string                `json:"id"`
	Name            string                `json:"name"`
	Kind            domain.LayoutUnitKind `json:"kind"`
	PositionXMM     float64               `json:"positionXMm"`
	PositionYMM     float64               `json:"positionYMm"`
	RotationDegrees float64               `json:"rotationDegrees"`
	Version         int                   `json:"version"`
	LocalOutline    []LayoutTwinPoint     `json:"localOutline"`
	Outline         []LayoutTwinPoint     `json:"outline"`
	Positions       []LayoutTwinPosition  `json:"positions"`
}

type LayoutTwin struct {
	LayoutID          string              `json:"layoutId"`
	ConfigurationID   string              `json:"configurationId,omitempty"`
	ConfigurationName string              `json:"configurationName,omitempty"`
	UnitID            string              `json:"unitId,omitempty"`
	Bounds            LayoutTwinBounds    `json:"bounds"`
	HasGeometry       bool                `json:"hasGeometry"`
	Units             []LayoutTwinUnit    `json:"units"`
	Warnings          []LayoutTwinWarning `json:"warnings"`
}

func (s *LayoutService) GetTwin(
	ctx context.Context,
	layoutID string,
	selection LayoutTwinSelection,
) (*LayoutTwin, error) {
	layoutID = strings.TrimSpace(layoutID)
	selection.ConfigurationID = strings.TrimSpace(selection.ConfigurationID)
	selection.UnitID = strings.TrimSpace(selection.UnitID)
	if layoutID == "" || (selection.ConfigurationID != "" && selection.UnitID != "") {
		return nil, ErrLayoutValidation
	}
	return s.repository.GetTwin(ctx, layoutID, selection)
}
