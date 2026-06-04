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
)

type ExhibitionService struct {
	db *sql.DB
}

type ExhibitionList struct {
	ID          string            `json:"id"`
	Designation string            `json:"designation"`
	Date        string            `json:"date"`
	Locked      bool              `json:"locked"`
	EntryCount  int               `json:"entryCount"`
	Entries     []ExhibitionEntry `json:"entries,omitempty"`
	CreatedAt   string            `json:"createdAt"`
	UpdatedAt   string            `json:"updatedAt"`
}

type ExhibitionListInput struct {
	Designation string `json:"designation"`
	Date        string `json:"date"`
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
	SXAddress      string `json:"sxAddress,omitempty"`
	Analog         bool   `json:"analog"`
	FunctionKeys   string `json:"functionKeys,omitempty"`
	Notes          string `json:"notes,omitempty"`
	SortOrder      int    `json:"sortOrder"`
	CreatedAt      string `json:"createdAt"`
	UpdatedAt      string `json:"updatedAt"`
}

type ExhibitionEntryInput struct {
	VehicleID      string `json:"vehicleId"`
	Owner          string `json:"owner"`
	ImageURL       string `json:"imageUrl"`
	LocomotiveName string `json:"locomotiveName"`
	Gattung        string `json:"gattung"`
	Series         string `json:"series"`
	Manufacturer   string `json:"manufacturer"`
	Epoch          string `json:"epoch"`
	RailwayCompany string `json:"railwayCompany"`
	DayScope       string `json:"dayScope"`
	DTDecoder      bool   `json:"dtDecoder"`
	DecoderNumber  string `json:"decoderNumber"`
	DecoderType    string `json:"decoderType"`
	Adapter        string `json:"adapter"`
	SXAddress      string `json:"sxAddress"`
	Analog         bool   `json:"analog"`
	FunctionKeys   string `json:"functionKeys"`
	Notes          string `json:"notes"`
	SortOrder      int    `json:"sortOrder"`
}

func NewExhibitionService(db *sql.DB) *ExhibitionService {
	return &ExhibitionService{db: db}
}

func (s *ExhibitionService) List(ctx context.Context) ([]ExhibitionList, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT l.id, l.designation, l.list_date, l.locked, COUNT(e.id), l.created_at, l.updated_at
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
		if err := rows.Scan(&list.ID, &list.Designation, &list.Date, &locked, &list.EntryCount, &list.CreatedAt, &list.UpdatedAt); err != nil {
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
	designation := strings.TrimSpace(input.Designation)
	date := strings.TrimSpace(input.Date)
	if designation == "" || date == "" {
		return ExhibitionList{}, ErrExhibitionValidation
	}

	id := randomID()
	now := time.Now().UTC().Format(time.RFC3339)
	if _, err := s.db.ExecContext(ctx, `
INSERT INTO exhibition_lists(id, designation, list_date, locked, created_at, updated_at)
VALUES(?, ?, ?, 0, ?, ?)
`, id, designation, date, now, now); err != nil {
		return ExhibitionList{}, fmt.Errorf("create exhibition list: %w", err)
	}
	return s.Get(ctx, id)
}

func (s *ExhibitionService) Update(ctx context.Context, id string, input ExhibitionListInput) (ExhibitionList, error) {
	designation := strings.TrimSpace(input.Designation)
	date := strings.TrimSpace(input.Date)
	if designation == "" || date == "" {
		return ExhibitionList{}, ErrExhibitionValidation
	}
	if _, err := s.getList(ctx, id); err != nil {
		return ExhibitionList{}, err
	}

	now := time.Now().UTC().Format(time.RFC3339)
	if _, err := s.db.ExecContext(ctx, `
UPDATE exhibition_lists SET designation=?, list_date=?, updated_at=? WHERE id=?
`, designation, date, now, id); err != nil {
		return ExhibitionList{}, fmt.Errorf("update exhibition list: %w", err)
	}
	return s.Get(ctx, id)
}

func (s *ExhibitionService) SetLocked(ctx context.Context, id string, locked bool) (ExhibitionList, error) {
	if _, err := s.getList(ctx, id); err != nil {
		return ExhibitionList{}, err
	}
	now := time.Now().UTC().Format(time.RFC3339)
	if _, err := s.db.ExecContext(ctx, `
UPDATE exhibition_lists SET locked=?, updated_at=? WHERE id=?
`, boolToInt(locked), now, id); err != nil {
		return ExhibitionList{}, fmt.Errorf("lock exhibition list: %w", err)
	}
	return s.Get(ctx, id)
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
       dt_decoder, decoder_number, COALESCE(decoder_type, ''), COALESCE(adapter, ''), COALESCE(sx_address, ''), analog,
       function_keys, notes, sort_order, created_at, updated_at
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
	if list.Locked {
		return ExhibitionEntry{}, ErrExhibitionLocked
	}
	input = normalizeExhibitionEntryInput(input)
	if input.Owner == "" || input.LocomotiveName == "" {
		return ExhibitionEntry{}, ErrExhibitionValidation
	}
	if input.SortOrder == 0 {
		input.SortOrder = s.nextEntrySortOrder(ctx, listID)
	}

	id := randomID()
	now := time.Now().UTC().Format(time.RFC3339)
	if _, err := s.db.ExecContext(ctx, `
INSERT INTO exhibition_entries(
  id, list_id, owner, image_url, locomotive_name, vehicle_id,
  gattung, series, manufacturer, epoch, railway_company, day_scope,
  dt_decoder, decoder_number, decoder_type, adapter, sx_address, analog,
  function_keys, notes, sort_order, created_at, updated_at
)
VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`, id, listID, input.Owner, input.ImageURL, input.LocomotiveName, input.VehicleID, input.Gattung, input.Series, input.Manufacturer, input.Epoch, input.RailwayCompany, input.DayScope, boolToInt(input.DTDecoder), input.DecoderNumber, input.DecoderType, input.Adapter, input.SXAddress, boolToInt(input.Analog), input.FunctionKeys, input.Notes, input.SortOrder, now, now); err != nil {
		return ExhibitionEntry{}, fmt.Errorf("create exhibition entry: %w", err)
	}
	return s.getEntry(ctx, listID, id)
}

func (s *ExhibitionService) UpdateEntry(ctx context.Context, listID, entryID string, input ExhibitionEntryInput) (ExhibitionEntry, error) {
	list, err := s.getList(ctx, listID)
	if err != nil {
		return ExhibitionEntry{}, err
	}
	if list.Locked {
		return ExhibitionEntry{}, ErrExhibitionLocked
	}
	if _, err := s.getEntry(ctx, listID, entryID); err != nil {
		return ExhibitionEntry{}, err
	}
	input = normalizeExhibitionEntryInput(input)
	if input.Owner == "" || input.LocomotiveName == "" {
		return ExhibitionEntry{}, ErrExhibitionValidation
	}
	now := time.Now().UTC().Format(time.RFC3339)
	if _, err := s.db.ExecContext(ctx, `
UPDATE exhibition_entries
SET owner=?, image_url=?, locomotive_name=?,
    vehicle_id=?, gattung=?, series=?, manufacturer=?, epoch=?, railway_company=?, day_scope=?,
    dt_decoder=?, decoder_number=?, decoder_type=?, adapter=?, sx_address=?, analog=?,
    function_keys=?, notes=?, sort_order=?, updated_at=?
WHERE id=? AND list_id=?
`, input.Owner, input.ImageURL, input.LocomotiveName, input.VehicleID, input.Gattung, input.Series, input.Manufacturer, input.Epoch, input.RailwayCompany, input.DayScope, boolToInt(input.DTDecoder), input.DecoderNumber, input.DecoderType, input.Adapter, input.SXAddress, boolToInt(input.Analog), input.FunctionKeys, input.Notes, input.SortOrder, now, entryID, listID); err != nil {
		return ExhibitionEntry{}, fmt.Errorf("update exhibition entry: %w", err)
	}
	return s.getEntry(ctx, listID, entryID)
}

func (s *ExhibitionService) DeleteEntry(ctx context.Context, listID, entryID string) error {
	list, err := s.getList(ctx, listID)
	if err != nil {
		return err
	}
	if list.Locked {
		return ErrExhibitionLocked
	}
	result, err := s.db.ExecContext(ctx, `DELETE FROM exhibition_entries WHERE id=? AND list_id=?`, entryID, listID)
	if err != nil {
		return fmt.Errorf("delete exhibition entry: %w", err)
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return ErrExhibitionNotFound
	}
	return nil
}

func (s *ExhibitionService) getList(ctx context.Context, id string) (ExhibitionList, error) {
	var list ExhibitionList
	var locked int
	if err := s.db.QueryRowContext(ctx, `
SELECT id, designation, list_date, locked, created_at, updated_at
FROM exhibition_lists WHERE id=?
`, id).Scan(&list.ID, &list.Designation, &list.Date, &locked, &list.CreatedAt, &list.UpdatedAt); err != nil {
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
       dt_decoder, decoder_number, COALESCE(decoder_type, ''), COALESCE(adapter, ''), COALESCE(sx_address, ''), analog,
       function_keys, notes, sort_order, created_at, updated_at
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
		&entry.SXAddress,
		&analog,
		&functionKeys,
		&notes,
		&entry.SortOrder,
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
	input.SXAddress = strings.TrimSpace(input.SXAddress)
	input.FunctionKeys = strings.TrimSpace(input.FunctionKeys)
	input.Notes = strings.TrimSpace(input.Notes)
	return input
}

func normalizeExhibitionDayScope(value string) string {
	allowed := map[string]bool{
		"day1": true,
		"day2": true,
		"day3": true,
		"day4": true,
	}
	raw := strings.Split(strings.TrimSpace(value), ",")
	seen := map[string]bool{}
	for _, part := range raw {
		scope := strings.TrimSpace(part)
		if scope == "all" {
			return "all"
		}
		if allowed[scope] {
			seen[scope] = true
		}
	}
	if len(seen) == 0 || len(seen) == len(allowed) {
		return "all"
	}
	ordered := []string{"day1", "day2", "day3", "day4"}
	selected := make([]string, 0, len(seen))
	for _, scope := range ordered {
		if seen[scope] {
			selected = append(selected, scope)
		}
	}
	return strings.Join(selected, ",")
}
