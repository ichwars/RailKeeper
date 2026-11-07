package transfer

import "railkeeper/backend/internal/domain"

type VehicleSetInput = domain.VehicleSetInput
type ExhibitionStatus = domain.ExhibitionStatus

const (
	ExhibitionStatusDraft     = domain.ExhibitionStatusDraft
	ExhibitionStatusOpen      = domain.ExhibitionStatusOpen
	ExhibitionStatusLocked    = domain.ExhibitionStatusLocked
	ExhibitionStatusRunning   = domain.ExhibitionStatusRunning
	ExhibitionStatusCompleted = domain.ExhibitionStatusCompleted
	ExhibitionStatusArchived  = domain.ExhibitionStatusArchived
)
