package application

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrExhibitionValidation = errors.New("exhibition validation failed")
	ErrExhibitionNotFound   = errors.New("exhibition not found")
	ErrExhibitionLocked     = errors.New("exhibition list locked")
	ErrExhibitionStale      = errors.New("exhibition revision is stale")
	ErrExhibitionConflicts  = errors.New("exhibition has unresolved conflicts")
)

type ExhibitionStatus string

const (
	ExhibitionStatusDraft     ExhibitionStatus = "draft"
	ExhibitionStatusOpen      ExhibitionStatus = "open"
	ExhibitionStatusLocked    ExhibitionStatus = "locked"
	ExhibitionStatusRunning   ExhibitionStatus = "running"
	ExhibitionStatusCompleted ExhibitionStatus = "completed"
	ExhibitionStatusArchived  ExhibitionStatus = "archived"
)

type ExhibitionService struct {
	db *sql.DB
}

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

func NewExhibitionService(db *sql.DB) *ExhibitionService {
	return &ExhibitionService{db: db}
}

func (s *ExhibitionService) List(ctx context.Context) ([]ExhibitionList, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT l.id, l.designation, l.list_date, l.end_date, l.location, l.description,
       l.organization_notes, l.status, l.revision, l.locked, l.lock_reason,
       l.locked_at, l.completed_at, l.archived_at, COUNT(e.id), l.created_at, l.updated_at
FROM exhibition_lists l
LEFT JOIN exhibition_entries e ON e.list_id = l.id
GROUP BY l.id
ORDER BY l.list_date DESC, l.designation COLLATE NOCASE
`)
	if err != nil {
		return nil, fmt.Errorf("list exhibition lists: %w", err)
	}
	defer func() { _ = rows.Close() }()

	lists := []ExhibitionList{}
	for rows.Next() {
		var list ExhibitionList
		var locked int
		if err := rows.Scan(
			&list.ID, &list.Designation, &list.Date, &list.EndDate, &list.Location, &list.Description,
			&list.OrganizationNotes, &list.Status, &list.Revision, &locked, &list.LockReason,
			&list.LockedAt, &list.CompletedAt, &list.ArchivedAt, &list.EntryCount,
			&list.CreatedAt, &list.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan exhibition list: %w", err)
		}
		list.Locked = locked == 1
		lists = append(lists, list)
	}
	return lists, rows.Err()
}

func (s *ExhibitionService) Get(ctx context.Context, id string) (ExhibitionList, error) {
	list, err := s.getList(ctx, id)
	if err != nil {
		return ExhibitionList{}, err
	}
	entries, err := s.ListEntries(ctx, id)
	if err != nil {
		return ExhibitionList{}, err
	}
	list.Entries = entries
	list.EntryCount = len(entries)
	return list, nil
}

func (s *ExhibitionService) Create(ctx context.Context, input ExhibitionListInput) (ExhibitionList, error) {
	input, err := normalizeExhibitionListInput(input)
	if err != nil {
		return ExhibitionList{}, ErrExhibitionValidation
	}
	if input.Status != ExhibitionStatusDraft && input.Status != ExhibitionStatusOpen {
		return ExhibitionList{}, ErrExhibitionValidation
	}

	id := randomID()
	now := time.Now().UTC().Format(time.RFC3339)
	if _, err := s.db.ExecContext(ctx, `
INSERT INTO exhibition_lists(
  id, designation, list_date, end_date, location, description, organization_notes,
  status, revision, locked, created_at, updated_at
)
VALUES(?, ?, ?, ?, ?, ?, ?, ?, 1, 0, ?, ?)
`, id, input.Designation, input.Date, input.EndDate, input.Location, input.Description,
		input.OrganizationNotes, input.Status, now, now); err != nil {
		return ExhibitionList{}, fmt.Errorf("create exhibition list: %w", err)
	}
	return s.Get(ctx, id)
}

func (s *ExhibitionService) Update(ctx context.Context, id string, input ExhibitionListInput) (ExhibitionList, error) {
	input, err := normalizeExhibitionListInput(input)
	if err != nil {
		return ExhibitionList{}, ErrExhibitionValidation
	}
	current, err := s.getList(ctx, id)
	if err != nil {
		return ExhibitionList{}, err
	}
	if !exhibitionAllowsPlanningChanges(current) {
		return ExhibitionList{}, ErrExhibitionLocked
	}
	if input.ExpectedRevision > 0 && input.ExpectedRevision != current.Revision {
		return ExhibitionList{}, ErrExhibitionStale
	}
	input.Status = current.Status

	now := time.Now().UTC().Format(time.RFC3339)
	result, err := s.db.ExecContext(ctx, `
UPDATE exhibition_lists
SET designation=?, list_date=?, end_date=?, location=?, description=?, organization_notes=?,
    status=?, revision=revision+1, updated_at=?
WHERE id=? AND (?=0 OR revision=?)
`, input.Designation, input.Date, input.EndDate, input.Location, input.Description,
		input.OrganizationNotes, input.Status, now, id, input.ExpectedRevision, input.ExpectedRevision)
	if err != nil {
		return ExhibitionList{}, fmt.Errorf("update exhibition list: %w", err)
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return ExhibitionList{}, ErrExhibitionStale
	}
	return s.Get(ctx, id)
}

func (s *ExhibitionService) SetLocked(ctx context.Context, id string, locked bool) (ExhibitionList, error) {
	status := ExhibitionStatusOpen
	if locked {
		status = ExhibitionStatusLocked
	}
	return s.SetStatus(ctx, id, ExhibitionStatusInput{Status: status, ConfirmConflicts: true, Reason: "Legacy lock action"})
}

func (s *ExhibitionService) Delete(ctx context.Context, id string) error {
	result, err := s.db.ExecContext(ctx, `DELETE FROM exhibition_lists WHERE id=?`, id)
	if err != nil {
		return fmt.Errorf("delete exhibition list: %w", err)
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return ErrExhibitionNotFound
	}
	return nil
}

func (s *ExhibitionService) ListEntries(ctx context.Context, listID string) ([]ExhibitionEntry, error) {
	if _, err := s.getList(ctx, listID); err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `
SELECT id, list_id, owner, image_url, locomotive_name,
       COALESCE(vehicle_id, ''),
       COALESCE(gattung, ''), COALESCE(series, ''), COALESCE(manufacturer, ''), COALESCE(epoch, ''), COALESCE(railway_company, ''),
       COALESCE(day_scope, 'all'),
       dt_decoder, decoder_number, COALESCE(decoder_type, ''), COALESCE(adapter, ''),
       COALESCE(interface_name, ''), COALESCE(sx_address, ''), analog,
       function_keys, notes, sort_order, availability, revision, created_at, updated_at
FROM exhibition_entries
WHERE list_id=?
ORDER BY sort_order, locomotive_name COLLATE NOCASE, owner COLLATE NOCASE
`, listID)
	if err != nil {
		return nil, fmt.Errorf("list exhibition entries: %w", err)
	}
	defer func() { _ = rows.Close() }()

	entries := []ExhibitionEntry{}
	for rows.Next() {
		entry, err := scanExhibitionEntry(rows)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return entries, rows.Err()
}

func (s *ExhibitionService) CreateEntry(ctx context.Context, listID string, input ExhibitionEntryInput) (ExhibitionEntry, error) {
	list, err := s.getList(ctx, listID)
	if err != nil {
		return ExhibitionEntry{}, err
	}
	if !exhibitionAllowsPlanningChanges(list) {
		return ExhibitionEntry{}, ErrExhibitionLocked
	}
	input = normalizeExhibitionEntryInput(input)
	if input.LocomotiveName == "" {
		return ExhibitionEntry{}, ErrExhibitionValidation
	}
	if input.SortOrder == 0 {
		input.SortOrder = s.nextEntrySortOrder(ctx, listID)
	}

	id := randomID()
	now := time.Now().UTC().Format(time.RFC3339)
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return ExhibitionEntry{}, fmt.Errorf("begin exhibition entry create: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, `
INSERT INTO exhibition_entries(
  id, list_id, owner, image_url, locomotive_name, vehicle_id,
  gattung, series, manufacturer, epoch, railway_company, day_scope,
  dt_decoder, decoder_number, decoder_type, adapter, interface_name, sx_address, analog,
  function_keys, notes, sort_order, availability, revision, created_at, updated_at
)
VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 1, ?, ?)
`, id, listID, input.Owner, input.ImageURL, input.LocomotiveName, input.VehicleID, input.Gattung,
		input.Series, input.Manufacturer, input.Epoch, input.RailwayCompany, input.DayScope,
		boolToInt(input.DTDecoder), input.DecoderNumber, input.DecoderType, input.Adapter,
		input.InterfaceName, input.SXAddress, boolToInt(input.Analog), input.FunctionKeys,
		input.Notes, input.SortOrder, input.Availability, now, now); err != nil {
		return ExhibitionEntry{}, fmt.Errorf("create exhibition entry: %w", err)
	}
	result, err := tx.ExecContext(ctx, `
UPDATE exhibition_lists SET revision=revision+1, updated_at=? WHERE id=? AND locked=0
`, now, listID)
	if err != nil {
		return ExhibitionEntry{}, fmt.Errorf("touch exhibition list: %w", err)
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return ExhibitionEntry{}, ErrExhibitionLocked
	}
	if err := tx.Commit(); err != nil {
		return ExhibitionEntry{}, fmt.Errorf("commit exhibition entry create: %w", err)
	}
	return s.getEntry(ctx, listID, id)
}

func (s *ExhibitionService) UpdateEntry(ctx context.Context, listID, entryID string, input ExhibitionEntryInput) (ExhibitionEntry, error) {
	list, err := s.getList(ctx, listID)
	if err != nil {
		return ExhibitionEntry{}, err
	}
	if !exhibitionAllowsPlanningChanges(list) {
		return ExhibitionEntry{}, ErrExhibitionLocked
	}
	if _, err := s.getEntry(ctx, listID, entryID); err != nil {
		return ExhibitionEntry{}, err
	}
	input = normalizeExhibitionEntryInput(input)
	if input.LocomotiveName == "" {
		return ExhibitionEntry{}, ErrExhibitionValidation
	}
	current, err := s.getEntry(ctx, listID, entryID)
	if err != nil {
		return ExhibitionEntry{}, err
	}
	if input.ExpectedRevision > 0 && input.ExpectedRevision != current.Revision {
		return ExhibitionEntry{}, ErrExhibitionStale
	}
	now := time.Now().UTC().Format(time.RFC3339)
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return ExhibitionEntry{}, fmt.Errorf("begin exhibition entry update: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	result, err := tx.ExecContext(ctx, `
UPDATE exhibition_entries
SET owner=?, image_url=?, locomotive_name=?,
    vehicle_id=?, gattung=?, series=?, manufacturer=?, epoch=?, railway_company=?, day_scope=?,
    dt_decoder=?, decoder_number=?, decoder_type=?, adapter=?, interface_name=?, sx_address=?, analog=?,
    function_keys=?, notes=?, sort_order=?, availability=?, revision=revision+1, updated_at=?
WHERE id=? AND list_id=? AND (?=0 OR revision=?)
`, input.Owner, input.ImageURL, input.LocomotiveName, input.VehicleID, input.Gattung, input.Series,
		input.Manufacturer, input.Epoch, input.RailwayCompany, input.DayScope, boolToInt(input.DTDecoder),
		input.DecoderNumber, input.DecoderType, input.Adapter, input.InterfaceName, input.SXAddress,
		boolToInt(input.Analog), input.FunctionKeys, input.Notes, input.SortOrder, input.Availability,
		now, entryID, listID, input.ExpectedRevision, input.ExpectedRevision)
	if err != nil {
		return ExhibitionEntry{}, fmt.Errorf("update exhibition entry: %w", err)
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return ExhibitionEntry{}, ErrExhibitionStale
	}
	result, err = tx.ExecContext(ctx, `
UPDATE exhibition_lists SET revision=revision+1, updated_at=? WHERE id=? AND locked=0
`, now, listID)
	if err != nil {
		return ExhibitionEntry{}, fmt.Errorf("touch exhibition list: %w", err)
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return ExhibitionEntry{}, ErrExhibitionLocked
	}
	if err := tx.Commit(); err != nil {
		return ExhibitionEntry{}, fmt.Errorf("commit exhibition entry update: %w", err)
	}
	return s.getEntry(ctx, listID, entryID)
}

func (s *ExhibitionService) DeleteEntry(ctx context.Context, listID, entryID string) error {
	list, err := s.getList(ctx, listID)
	if err != nil {
		return err
	}
	if !exhibitionAllowsPlanningChanges(list) {
		return ErrExhibitionLocked
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin exhibition entry delete: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	result, err := tx.ExecContext(ctx, `DELETE FROM exhibition_entries WHERE id=? AND list_id=?`, entryID, listID)
	if err != nil {
		return fmt.Errorf("delete exhibition entry: %w", err)
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return ErrExhibitionNotFound
	}
	now := time.Now().UTC().Format(time.RFC3339)
	result, err = tx.ExecContext(ctx, `
UPDATE exhibition_lists SET revision=revision+1, updated_at=? WHERE id=? AND locked=0
`, now, listID)
	if err != nil {
		return fmt.Errorf("touch exhibition list: %w", err)
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return ErrExhibitionLocked
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit exhibition entry delete: %w", err)
	}
	return nil
}

func (s *ExhibitionService) getList(ctx context.Context, id string) (ExhibitionList, error) {
	var list ExhibitionList
	var locked int
	if err := s.db.QueryRowContext(ctx, `
SELECT id, designation, list_date, end_date, location, description, organization_notes,
       status, revision, locked, lock_reason, locked_at, completed_at, archived_at,
       created_at, updated_at
FROM exhibition_lists WHERE id=?
`, id).Scan(
		&list.ID, &list.Designation, &list.Date, &list.EndDate, &list.Location,
		&list.Description, &list.OrganizationNotes, &list.Status, &list.Revision, &locked,
		&list.LockReason, &list.LockedAt, &list.CompletedAt, &list.ArchivedAt,
		&list.CreatedAt, &list.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ExhibitionList{}, ErrExhibitionNotFound
		}
		return ExhibitionList{}, fmt.Errorf("get exhibition list: %w", err)
	}
	list.Locked = locked == 1
	return list, nil
}

func (s *ExhibitionService) getEntry(ctx context.Context, listID, entryID string) (ExhibitionEntry, error) {
	row := s.db.QueryRowContext(ctx, `
SELECT id, list_id, owner, image_url, locomotive_name,
       COALESCE(vehicle_id, ''),
       COALESCE(gattung, ''), COALESCE(series, ''), COALESCE(manufacturer, ''), COALESCE(epoch, ''), COALESCE(railway_company, ''),
       COALESCE(day_scope, 'all'),
       dt_decoder, decoder_number, COALESCE(decoder_type, ''), COALESCE(adapter, ''),
       COALESCE(interface_name, ''), COALESCE(sx_address, ''), analog,
       function_keys, notes, sort_order, availability, revision, created_at, updated_at
FROM exhibition_entries
WHERE id=? AND list_id=?
`, entryID, listID)
	entry, err := scanExhibitionEntry(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ExhibitionEntry{}, ErrExhibitionNotFound
		}
		return ExhibitionEntry{}, err
	}
	return entry, nil
}

func (s *ExhibitionService) nextEntrySortOrder(ctx context.Context, listID string) int {
	var next sql.NullInt64
	_ = s.db.QueryRowContext(ctx, `SELECT COALESCE(MAX(sort_order), 0) + 10 FROM exhibition_entries WHERE list_id=?`, listID).Scan(&next)
	if next.Valid && next.Int64 > 0 {
		return int(next.Int64)
	}
	return 10
}

type exhibitionEntryScanner interface {
	Scan(dest ...any) error
}

func scanExhibitionEntry(row exhibitionEntryScanner) (ExhibitionEntry, error) {
	var entry ExhibitionEntry
	var imageURL, decoderNumber, functionKeys, notes sql.NullString
	var dtDecoder int
	var analog int
	if err := row.Scan(
		&entry.ID,
		&entry.ListID,
		&entry.Owner,
		&imageURL,
		&entry.LocomotiveName,
		&entry.VehicleID,
		&entry.Gattung,
		&entry.Series,
		&entry.Manufacturer,
		&entry.Epoch,
		&entry.RailwayCompany,
		&entry.DayScope,
		&dtDecoder,
		&decoderNumber,
		&entry.DecoderType,
		&entry.Adapter,
		&entry.InterfaceName,
		&entry.SXAddress,
		&analog,
		&functionKeys,
		&notes,
		&entry.SortOrder,
		&entry.Availability,
		&entry.Revision,
		&entry.CreatedAt,
		&entry.UpdatedAt,
	); err != nil {
		return ExhibitionEntry{}, fmt.Errorf("scan exhibition entry: %w", err)
	}
	entry.ImageURL = imageURL.String
	entry.DTDecoder = dtDecoder == 1
	entry.DecoderNumber = decoderNumber.String
	entry.Analog = analog == 1
	entry.FunctionKeys = functionKeys.String
	entry.Notes = notes.String
	return entry, nil
}

func normalizeExhibitionEntryInput(input ExhibitionEntryInput) ExhibitionEntryInput {
	input.VehicleID = strings.TrimSpace(input.VehicleID)
	input.Owner = strings.TrimSpace(input.Owner)
	input.ImageURL = strings.TrimSpace(input.ImageURL)
	input.LocomotiveName = strings.TrimSpace(input.LocomotiveName)
	input.Gattung = strings.TrimSpace(input.Gattung)
	input.Series = strings.TrimSpace(input.Series)
	input.Manufacturer = strings.TrimSpace(input.Manufacturer)
	input.Epoch = strings.TrimSpace(input.Epoch)
	input.RailwayCompany = strings.TrimSpace(input.RailwayCompany)
	input.DayScope = normalizeExhibitionDayScope(input.DayScope)
	input.DecoderNumber = strings.TrimSpace(input.DecoderNumber)
	input.DecoderType = strings.TrimSpace(input.DecoderType)
	input.Adapter = strings.TrimSpace(input.Adapter)
	input.InterfaceName = strings.TrimSpace(input.InterfaceName)
	if input.InterfaceName == "" {
		input.InterfaceName = input.Adapter
	}
	input.SXAddress = strings.TrimSpace(input.SXAddress)
	input.Availability = strings.TrimSpace(input.Availability)
	if input.Availability == "" {
		input.Availability = "available"
	}
	if input.Availability != "available" && input.Availability != "unavailable" {
		input.Availability = "available"
	}
	input.FunctionKeys = strings.TrimSpace(input.FunctionKeys)
	input.Notes = strings.TrimSpace(input.Notes)
	return input
}

func normalizeExhibitionDayScope(value string) string {
	raw := strings.Split(strings.TrimSpace(value), ",")
	seen := map[string]bool{}
	for _, part := range raw {
		scope := strings.TrimSpace(part)
		if scope == "all" {
			return "all"
		}
		if isExhibitionDayScope(scope) {
			seen[scope] = true
		}
	}
	if len(seen) == 0 {
		return "all"
	}
	selected := make([]string, 0, len(seen))
	for day := 1; day <= 31; day++ {
		scope := fmt.Sprintf("day%d", day)
		if seen[scope] {
			selected = append(selected, scope)
		}
	}
	return strings.Join(selected, ",")
}

func isExhibitionDayScope(scope string) bool {
	if !strings.HasPrefix(scope, "day") {
		return false
	}
	var day int
	if _, err := fmt.Sscanf(scope, "day%d", &day); err != nil {
		return false
	}
	return day >= 1 && day <= 31 && scope == fmt.Sprintf("day%d", day)
}

func normalizeExhibitionListInput(input ExhibitionListInput) (ExhibitionListInput, error) {
	input.Designation = strings.TrimSpace(input.Designation)
	input.Date = strings.TrimSpace(input.Date)
	input.EndDate = strings.TrimSpace(input.EndDate)
	input.Location = strings.TrimSpace(input.Location)
	input.Description = strings.TrimSpace(input.Description)
	input.OrganizationNotes = strings.TrimSpace(input.OrganizationNotes)
	if input.EndDate == "" {
		input.EndDate = input.Date
	}
	if input.Status == "" {
		input.Status = ExhibitionStatusOpen
	}
	if input.Designation == "" || !validExhibitionStatus(input.Status) {
		return ExhibitionListInput{}, ErrExhibitionValidation
	}
	start, err := time.Parse("2006-01-02", input.Date)
	if err != nil {
		return ExhibitionListInput{}, ErrExhibitionValidation
	}
	end, err := time.Parse("2006-01-02", input.EndDate)
	if err != nil || end.Before(start) || end.Sub(start) > 30*24*time.Hour {
		return ExhibitionListInput{}, ErrExhibitionValidation
	}
	return input, nil
}

func validExhibitionStatus(status ExhibitionStatus) bool {
	switch status {
	case ExhibitionStatusDraft, ExhibitionStatusOpen, ExhibitionStatusLocked,
		ExhibitionStatusRunning, ExhibitionStatusCompleted, ExhibitionStatusArchived:
		return true
	default:
		return false
	}
}

func exhibitionAllowsPlanningChanges(list ExhibitionList) bool {
	return !list.Locked && (list.Status == ExhibitionStatusDraft || list.Status == ExhibitionStatusOpen)
}
