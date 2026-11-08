package application

import (
	"database/sql"

	"railkeeper/backend/internal/exhibition"
	"railkeeper/backend/internal/infrastructure/exhibitiondata"
)

type ExhibitionService = exhibition.Service
type ExhibitionStatus = exhibition.ExhibitionStatus
type ExhibitionList = exhibition.ExhibitionList
type ExhibitionListInput = exhibition.ExhibitionListInput
type ExhibitionEntry = exhibition.ExhibitionEntry
type ExhibitionEntryInput = exhibition.ExhibitionEntryInput
type ExhibitionEntryStatus = exhibition.ExhibitionEntryStatus
type ExhibitionSummary = exhibition.ExhibitionSummary
type ExhibitionReadiness = exhibition.ExhibitionReadiness
type ExhibitionConflict = exhibition.ExhibitionConflict
type ExhibitionWorkspaceEntry = exhibition.ExhibitionWorkspaceEntry
type ExhibitionWorkspace = exhibition.ExhibitionWorkspace
type ExhibitionStatusInput = exhibition.ExhibitionStatusInput
type ExhibitionConflictExceptionInput = exhibition.ExhibitionConflictExceptionInput

const (
	ExhibitionStatusDraft                = exhibition.ExhibitionStatusDraft
	ExhibitionStatusOpen                 = exhibition.ExhibitionStatusOpen
	ExhibitionStatusLocked               = exhibition.ExhibitionStatusLocked
	ExhibitionStatusRunning              = exhibition.ExhibitionStatusRunning
	ExhibitionStatusCompleted            = exhibition.ExhibitionStatusCompleted
	ExhibitionStatusArchived             = exhibition.ExhibitionStatusArchived
	ExhibitionEntryStatusReady           = exhibition.ExhibitionEntryStatusReady
	ExhibitionEntryStatusAddressConflict = exhibition.ExhibitionEntryStatusAddressConflict
	ExhibitionEntryStatusMissing         = exhibition.ExhibitionEntryStatusMissing
	ExhibitionEntryStatusCheck           = exhibition.ExhibitionEntryStatusCheck
	ExhibitionEntryStatusUnavailable     = exhibition.ExhibitionEntryStatusUnavailable
	ExhibitionConflictAddress            = exhibition.ExhibitionConflictAddress
	ExhibitionConflictMissing            = exhibition.ExhibitionConflictMissing
	ExhibitionConflictDuplicateVehicle   = exhibition.ExhibitionConflictDuplicateVehicle
)

var (
	ErrExhibitionValidation = exhibition.ErrExhibitionValidation
	ErrExhibitionNotFound   = exhibition.ErrExhibitionNotFound
	ErrExhibitionLocked     = exhibition.ErrExhibitionLocked
	ErrExhibitionStale      = exhibition.ErrExhibitionStale
	ErrExhibitionConflicts  = exhibition.ErrExhibitionConflicts
)

func NewExhibitionService(db *sql.DB) *ExhibitionService {
	return exhibition.New(exhibitiondata.New(db))
}
