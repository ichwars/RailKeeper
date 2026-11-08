package exhibition

import (
	"context"
	"errors"

	"railkeeper/backend/internal/domain"
)

var (
	ErrExhibitionValidation = errors.New("exhibition validation failed")
	ErrExhibitionNotFound   = errors.New("exhibition not found")
	ErrExhibitionLocked     = errors.New("exhibition list locked")
	ErrExhibitionStale      = errors.New("exhibition revision is stale")
	ErrExhibitionConflicts  = errors.New("exhibition has unresolved conflicts")
)

type ExhibitionStatus = domain.ExhibitionStatus

const (
	ExhibitionStatusDraft     = domain.ExhibitionStatusDraft
	ExhibitionStatusOpen      = domain.ExhibitionStatusOpen
	ExhibitionStatusLocked    = domain.ExhibitionStatusLocked
	ExhibitionStatusRunning   = domain.ExhibitionStatusRunning
	ExhibitionStatusCompleted = domain.ExhibitionStatusCompleted
	ExhibitionStatusArchived  = domain.ExhibitionStatusArchived
)

type ExhibitionList struct {
	ID                string            `json:"id"`
	Designation       string            `json:"designation"`
	Date              string            `json:"date"`
	EndDate           string            `json:"endDate"`
	Location          string            `json:"location,omitempty"`
	Description       string            `json:"description,omitempty"`
	OrganizationNotes string            `json:"organizationNotes,omitempty"`
	Status            ExhibitionStatus  `json:"status"`
	Revision          int               `json:"revision"`
	Locked            bool              `json:"locked"`
	LockReason        string            `json:"lockReason,omitempty"`
	LockedAt          string            `json:"lockedAt,omitempty"`
	CompletedAt       string            `json:"completedAt,omitempty"`
	ArchivedAt        string            `json:"archivedAt,omitempty"`
	EntryCount        int               `json:"entryCount"`
	Entries           []ExhibitionEntry `json:"entries,omitempty"`
	CreatedAt         string            `json:"createdAt"`
	UpdatedAt         string            `json:"updatedAt"`
}

type ExhibitionListInput struct {
	Designation       string           `json:"designation"`
	Date              string           `json:"date"`
	EndDate           string           `json:"endDate"`
	Location          string           `json:"location"`
	Description       string           `json:"description"`
	OrganizationNotes string           `json:"organizationNotes"`
	Status            ExhibitionStatus `json:"status"`
	ExpectedRevision  int              `json:"expectedRevision,omitempty"`
}

type ExhibitionEntry struct {
	ID             string `json:"id"`
	ListID         string `json:"listId"`
	VehicleID      string `json:"vehicleId,omitempty"`
	Owner          string `json:"owner"`
	ImageURL       string `json:"imageUrl,omitempty"`
	LocomotiveName string `json:"locomotiveName"`
	Gattung        string `json:"gattung,omitempty"`
	Series         string `json:"series,omitempty"`
	Manufacturer   string `json:"manufacturer,omitempty"`
	Epoch          string `json:"epoch,omitempty"`
	RailwayCompany string `json:"railwayCompany,omitempty"`
	DayScope       string `json:"dayScope"`
	DTDecoder      bool   `json:"dtDecoder"`
	DecoderNumber  string `json:"decoderNumber,omitempty"`
	DecoderType    string `json:"decoderType,omitempty"`
	Adapter        string `json:"adapter,omitempty"`
	InterfaceName  string `json:"interfaceName,omitempty"`
	SXAddress      string `json:"sxAddress,omitempty"`
	Analog         bool   `json:"analog"`
	Availability   string `json:"availability"`
	Status         string `json:"status,omitempty"`
	Revision       int    `json:"revision"`
	FunctionKeys   string `json:"functionKeys,omitempty"`
	Notes          string `json:"notes,omitempty"`
	SortOrder      int    `json:"sortOrder"`
	CreatedAt      string `json:"createdAt"`
	UpdatedAt      string `json:"updatedAt"`
}

type ExhibitionEntryInput struct {
	VehicleID        string `json:"vehicleId"`
	Owner            string `json:"owner"`
	ImageURL         string `json:"imageUrl"`
	LocomotiveName   string `json:"locomotiveName"`
	Gattung          string `json:"gattung"`
	Series           string `json:"series"`
	Manufacturer     string `json:"manufacturer"`
	Epoch            string `json:"epoch"`
	RailwayCompany   string `json:"railwayCompany"`
	DayScope         string `json:"dayScope"`
	DTDecoder        bool   `json:"dtDecoder"`
	DecoderNumber    string `json:"decoderNumber"`
	DecoderType      string `json:"decoderType"`
	Adapter          string `json:"adapter"`
	InterfaceName    string `json:"interfaceName"`
	SXAddress        string `json:"sxAddress"`
	Analog           bool   `json:"analog"`
	Availability     string `json:"availability"`
	ExpectedRevision int    `json:"expectedRevision,omitempty"`
	FunctionKeys     string `json:"functionKeys"`
	Notes            string `json:"notes"`
	SortOrder        int    `json:"sortOrder"`
}

type ExhibitionEntryStatus string

const (
	ExhibitionEntryStatusReady           ExhibitionEntryStatus = "ready"
	ExhibitionEntryStatusAddressConflict ExhibitionEntryStatus = "addressConflict"
	ExhibitionEntryStatusMissing         ExhibitionEntryStatus = "missing"
	ExhibitionEntryStatusCheck           ExhibitionEntryStatus = "check"
	ExhibitionEntryStatusUnavailable     ExhibitionEntryStatus = "unavailable"
)

const (
	ExhibitionConflictAddress          = "address"
	ExhibitionConflictMissing          = "missing"
	ExhibitionConflictDuplicateVehicle = "duplicateVehicle"
)

type ExhibitionSummary struct {
	EntryCount    int `json:"entryCount"`
	OwnerCount    int `json:"ownerCount"`
	ConflictCount int `json:"conflictCount"`
	ReadyCount    int `json:"readyCount"`
}

type ExhibitionReadiness struct {
	Total               int `json:"total"`
	AddressesChecked    int `json:"addressesChecked"`
	FunctionsDocumented int `json:"functionsDocumented"`
	ImagesPresent       int `json:"imagesPresent"`
	Problems            int `json:"problems"`
}

type ExhibitionConflict struct {
	ID              string   `json:"id"`
	Kind            string   `json:"kind"`
	EntryIDs        []string `json:"entryIds"`
	Fields          []string `json:"fields,omitempty"`
	InterfaceName   string   `json:"interfaceName,omitempty"`
	Address         string   `json:"address,omitempty"`
	DayScopes       []string `json:"dayScopes,omitempty"`
	Excepted        bool     `json:"excepted"`
	ExceptionReason string   `json:"exceptionReason,omitempty"`
}

type ExhibitionWorkspaceEntry struct {
	ExhibitionEntry
	Status      ExhibitionEntryStatus `json:"status"`
	ConflictIDs []string              `json:"conflictIds"`
}

type ExhibitionWorkspace struct {
	List      ExhibitionList             `json:"list"`
	Summary   ExhibitionSummary          `json:"summary"`
	Readiness ExhibitionReadiness        `json:"readiness"`
	Entries   []ExhibitionWorkspaceEntry `json:"entries"`
	Conflicts []ExhibitionConflict       `json:"conflicts"`
	DayScopes []string                   `json:"dayScopes"`
}

type ExhibitionStatusInput struct {
	Status           ExhibitionStatus `json:"status"`
	ExpectedRevision int              `json:"expectedRevision,omitempty"`
	ConfirmConflicts bool             `json:"confirmConflicts,omitempty"`
	Reason           string           `json:"reason,omitempty"`
}

type ExhibitionConflictExceptionInput struct {
	Reason           string `json:"reason"`
	ExpectedRevision int    `json:"expectedRevision,omitempty"`
}

// Repository preserves the existing autocommit list operations and caller-owned entry transactions.
type Repository interface {
	Lists(context.Context) ([]ExhibitionList, error)
	List(context.Context, string) (ExhibitionList, error)
	Entries(context.Context, string) ([]ExhibitionEntry, error)
	Entry(context.Context, string, string) (ExhibitionEntry, error)
	NextSort(context.Context, string) int
	Exceptions(context.Context, string) (map[string]string, error)
	InsertList(context.Context, ListWrite) error
	UpdateList(context.Context, ListWrite) (bool, error)
	DeleteList(context.Context, string) (bool, error)
	SetStatus(context.Context, StatusChange) (bool, error)
	Begin(context.Context) (Mutation, error)
}
type Mutation interface {
	InsertEntry(context.Context, EntryWrite) error
	UpdateEntry(context.Context, EntryWrite) (bool, error)
	DeleteEntry(context.Context, string, string) (bool, error)
	TouchUnlockedList(context.Context, string, string) (bool, error)
	UpsertException(context.Context, ExceptionWrite) error
	BumpRevision(context.Context, string, int, string) (bool, error)
	Commit() error
	Rollback() error
}
type ListWrite struct {
	ID, At string
	Input  ExhibitionListInput
}
type EntryWrite struct {
	ID, ListID, At string
	Input          ExhibitionEntryInput
}
type ExceptionWrite struct{ ID, ListID, ConflictID, Reason, At string }
type StatusChange struct {
	ID, Reason, LockedAt, CompletedAt, ArchivedAt, At string
	Status                                            ExhibitionStatus
	Locked                                            bool
	ExpectedRevision                                  int
}
type Service struct{ repository Repository }

func New(repository Repository) *Service { return &Service{repository: repository} }
