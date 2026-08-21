package application

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"sort"
	"strings"
	"time"
)

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

func (s *ExhibitionService) Workspace(ctx context.Context, id string) (ExhibitionWorkspace, error) {
	list, err := s.Get(ctx, id)
	if err != nil {
		return ExhibitionWorkspace{}, err
	}
	conflicts := buildExhibitionConflicts(list, list.Entries)
	if err := s.attachExhibitionExceptions(ctx, id, conflicts); err != nil {
		return ExhibitionWorkspace{}, err
	}
	return buildExhibitionWorkspace(list, conflicts), nil
}

func buildExhibitionWorkspace(list ExhibitionList, conflicts []ExhibitionConflict) ExhibitionWorkspace {
	workspace := ExhibitionWorkspace{
		List:      list,
		Entries:   make([]ExhibitionWorkspaceEntry, 0, len(list.Entries)),
		Conflicts: conflicts,
		DayScopes: exhibitionDayScopes(list),
		Readiness: ExhibitionReadiness{Total: len(list.Entries)},
	}
	owners := map[string]struct{}{}
	entryConflicts := map[string][]ExhibitionConflict{}
	for _, conflict := range conflicts {
		for _, entryID := range conflict.EntryIDs {
			entryConflicts[entryID] = append(entryConflicts[entryID], conflict)
		}
	}
	for _, entry := range list.Entries {
		status := exhibitionEntryStatus(entry, entryConflicts[entry.ID])
		decorated := ExhibitionWorkspaceEntry{ExhibitionEntry: entry, Status: status, ConflictIDs: []string{}}
		for _, conflict := range entryConflicts[entry.ID] {
			decorated.ConflictIDs = append(decorated.ConflictIDs, conflict.ID)
		}
		workspace.Entries = append(workspace.Entries, decorated)
		if entry.Owner != "" {
			owners[strings.ToLower(entry.Owner)] = struct{}{}
		}
		if exhibitionAddressReady(entry) {
			workspace.Readiness.AddressesChecked++
		}
		if strings.TrimSpace(entry.FunctionKeys) != "" {
			workspace.Readiness.FunctionsDocumented++
		}
		if strings.TrimSpace(entry.ImageURL) != "" {
			workspace.Readiness.ImagesPresent++
		}
		if status == ExhibitionEntryStatusReady {
			workspace.Summary.ReadyCount++
		} else {
			workspace.Readiness.Problems++
		}
	}
	workspace.Summary.EntryCount = len(list.Entries)
	workspace.Summary.OwnerCount = len(owners)
	workspace.Summary.ConflictCount = len(conflicts)
	return workspace
}

func exhibitionEntryStatus(entry ExhibitionEntry, conflicts []ExhibitionConflict) ExhibitionEntryStatus {
	if entry.Availability == "unavailable" {
		return ExhibitionEntryStatusUnavailable
	}
	for _, conflict := range conflicts {
		if conflict.Kind == ExhibitionConflictAddress {
			return ExhibitionEntryStatusAddressConflict
		}
	}
	for _, conflict := range conflicts {
		if conflict.Kind == ExhibitionConflictMissing {
			return ExhibitionEntryStatusMissing
		}
	}
	if len(conflicts) > 0 || strings.TrimSpace(entry.FunctionKeys) == "" || strings.TrimSpace(entry.ImageURL) == "" {
		return ExhibitionEntryStatusCheck
	}
	return ExhibitionEntryStatusReady
}

func buildExhibitionConflicts(list ExhibitionList, entries []ExhibitionEntry) []ExhibitionConflict {
	conflicts := make([]ExhibitionConflict, 0)
	dayScopes := exhibitionDayScopes(list)
	vehicleEntries := map[string][]string{}
	for _, entry := range entries {
		missing := exhibitionMissingFields(entry)
		if len(missing) > 0 {
			conflicts = append(conflicts, newExhibitionConflict(
				ExhibitionConflictMissing, []string{entry.ID}, missing, "", "", nil,
			))
		}
		if entry.VehicleID != "" {
			vehicleEntries[entry.VehicleID] = append(vehicleEntries[entry.VehicleID], entry.ID)
		}
	}
	for vehicleID, entryIDs := range vehicleEntries {
		if len(entryIDs) > 1 {
			sort.Strings(entryIDs)
			conflicts = append(conflicts, newExhibitionConflict(
				ExhibitionConflictDuplicateVehicle, entryIDs, []string{"vehicleId"}, "", vehicleID, nil,
			))
		}
	}
	for leftIndex, left := range entries {
		leftAddress := exhibitionAddress(left)
		leftInterface := strings.TrimSpace(left.InterfaceName)
		if leftAddress == "" || leftInterface == "" || left.Analog || left.Availability == "unavailable" {
			continue
		}
		for rightIndex := leftIndex + 1; rightIndex < len(entries); rightIndex++ {
			right := entries[rightIndex]
			if right.Analog || right.Availability == "unavailable" ||
				!strings.EqualFold(leftAddress, exhibitionAddress(right)) ||
				!strings.EqualFold(leftInterface, strings.TrimSpace(right.InterfaceName)) {
				continue
			}
			overlap := overlappingExhibitionDays(left.DayScope, right.DayScope, dayScopes)
			if len(overlap) == 0 {
				continue
			}
			entryIDs := []string{left.ID, right.ID}
			sort.Strings(entryIDs)
			conflicts = append(conflicts, newExhibitionConflict(
				ExhibitionConflictAddress, entryIDs, []string{"decoderNumber"}, leftInterface, leftAddress, overlap,
			))
		}
	}
	sort.Slice(conflicts, func(i, j int) bool { return conflicts[i].ID < conflicts[j].ID })
	return conflicts
}

func newExhibitionConflict(kind string, entryIDs, fields []string, interfaceName, address string, days []string) ExhibitionConflict {
	key := strings.Join([]string{
		kind, strings.Join(entryIDs, ","), strings.Join(fields, ","),
		strings.ToLower(interfaceName), strings.ToLower(address), strings.Join(days, ","),
	}, "|")
	sum := sha256.Sum256([]byte(key))
	return ExhibitionConflict{
		ID: fmt.Sprintf("conflict-%x", sum[:12]), Kind: kind, EntryIDs: entryIDs,
		Fields: fields, InterfaceName: interfaceName, Address: address, DayScopes: days,
	}
}

func exhibitionMissingFields(entry ExhibitionEntry) []string {
	missing := make([]string, 0, 4)
	if strings.TrimSpace(entry.Owner) == "" {
		missing = append(missing, "owner")
	}
	if strings.TrimSpace(entry.DayScope) == "" {
		missing = append(missing, "dayScope")
	}
	if !entry.Analog {
		if exhibitionAddress(entry) == "" {
			missing = append(missing, "decoderNumber")
		}
		if strings.TrimSpace(entry.InterfaceName) == "" {
			missing = append(missing, "interfaceName")
		}
	}
	return missing
}

func exhibitionAddressReady(entry ExhibitionEntry) bool {
	return entry.Analog || (exhibitionAddress(entry) != "" && strings.TrimSpace(entry.InterfaceName) != "")
}

func exhibitionAddress(entry ExhibitionEntry) string {
	if value := strings.TrimSpace(entry.DecoderNumber); value != "" {
		return value
	}
	return strings.TrimSpace(entry.SXAddress)
}

func exhibitionDayScopes(list ExhibitionList) []string {
	start, startErr := time.Parse("2006-01-02", list.Date)
	end, endErr := time.Parse("2006-01-02", list.EndDate)
	if startErr != nil || endErr != nil || end.Before(start) {
		return []string{"day1"}
	}
	days := int(end.Sub(start)/(24*time.Hour)) + 1
	result := make([]string, 0, days)
	for day := 1; day <= days && day <= 31; day++ {
		result = append(result, fmt.Sprintf("day%d", day))
	}
	return result
}

func overlappingExhibitionDays(left, right string, all []string) []string {
	leftDays := exhibitionDaySet(left, all)
	rightDays := exhibitionDaySet(right, all)
	result := make([]string, 0)
	for _, day := range all {
		if leftDays[day] && rightDays[day] {
			result = append(result, day)
		}
	}
	return result
}

func exhibitionDaySet(scope string, all []string) map[string]bool {
	result := map[string]bool{}
	if strings.TrimSpace(scope) == "all" || strings.TrimSpace(scope) == "" {
		for _, day := range all {
			result[day] = true
		}
		return result
	}
	for _, day := range strings.Split(scope, ",") {
		result[strings.TrimSpace(day)] = true
	}
	return result
}

func (s *ExhibitionService) attachExhibitionExceptions(ctx context.Context, listID string, conflicts []ExhibitionConflict) error {
	rows, err := s.db.QueryContext(ctx, `
SELECT conflict_key, reason FROM exhibition_conflict_exceptions WHERE list_id=?
`, listID)
	if err != nil {
		return fmt.Errorf("list exhibition conflict exceptions: %w", err)
	}
	defer func() { _ = rows.Close() }()
	exceptions := map[string]string{}
	for rows.Next() {
		var key, reason string
		if err := rows.Scan(&key, &reason); err != nil {
			return fmt.Errorf("scan exhibition conflict exception: %w", err)
		}
		exceptions[key] = reason
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate exhibition conflict exceptions: %w", err)
	}
	for index := range conflicts {
		if reason, ok := exceptions[conflicts[index].ID]; ok {
			conflicts[index].Excepted = true
			conflicts[index].ExceptionReason = reason
		}
	}
	return nil
}

func (s *ExhibitionService) SetConflictException(
	ctx context.Context,
	listID string,
	conflictID string,
	input ExhibitionConflictExceptionInput,
) (ExhibitionWorkspace, error) {
	reason := strings.TrimSpace(input.Reason)
	if reason == "" {
		return ExhibitionWorkspace{}, ErrExhibitionValidation
	}
	workspace, err := s.Workspace(ctx, listID)
	if err != nil {
		return ExhibitionWorkspace{}, err
	}
	if input.ExpectedRevision > 0 && input.ExpectedRevision != workspace.List.Revision {
		return ExhibitionWorkspace{}, ErrExhibitionStale
	}
	found := false
	for _, conflict := range workspace.Conflicts {
		if conflict.ID == conflictID {
			found = true
			break
		}
	}
	if !found {
		return ExhibitionWorkspace{}, ErrExhibitionNotFound
	}
	now := time.Now().UTC().Format(time.RFC3339)
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return ExhibitionWorkspace{}, fmt.Errorf("begin exhibition exception: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, `
INSERT INTO exhibition_conflict_exceptions(id, list_id, conflict_key, reason, created_at, updated_at)
VALUES(?, ?, ?, ?, ?, ?)
ON CONFLICT(list_id, conflict_key) DO UPDATE SET reason=excluded.reason, updated_at=excluded.updated_at
`, randomID(), listID, conflictID, reason, now, now); err != nil {
		return ExhibitionWorkspace{}, fmt.Errorf("save exhibition conflict exception: %w", err)
	}
	if err := bumpExhibitionRevision(ctx, tx, listID, input.ExpectedRevision, now); err != nil {
		return ExhibitionWorkspace{}, err
	}
	if err := tx.Commit(); err != nil {
		return ExhibitionWorkspace{}, fmt.Errorf("commit exhibition exception: %w", err)
	}
	return s.Workspace(ctx, listID)
}

func (s *ExhibitionService) SetStatus(ctx context.Context, id string, input ExhibitionStatusInput) (ExhibitionList, error) {
	if !validExhibitionStatus(input.Status) {
		return ExhibitionList{}, ErrExhibitionValidation
	}
	workspace, err := s.Workspace(ctx, id)
	if err != nil {
		return ExhibitionList{}, err
	}
	if input.ExpectedRevision > 0 && input.ExpectedRevision != workspace.List.Revision {
		return ExhibitionList{}, ErrExhibitionStale
	}
	if input.Status != workspace.List.Status && !validExhibitionTransition(workspace.List.Status, input.Status) {
		return ExhibitionList{}, ErrExhibitionValidation
	}
	reason := strings.TrimSpace(input.Reason)
	if input.Status == ExhibitionStatusLocked && len(workspace.Conflicts) > 0 {
		if !input.ConfirmConflicts || reason == "" {
			return ExhibitionList{}, ErrExhibitionConflicts
		}
	}
	now := time.Now().UTC().Format(time.RFC3339)
	locked := input.Status == ExhibitionStatusLocked || input.Status == ExhibitionStatusRunning
	lockedAt, completedAt, archivedAt := workspace.List.LockedAt, workspace.List.CompletedAt, workspace.List.ArchivedAt
	if input.Status == ExhibitionStatusLocked && lockedAt == "" {
		lockedAt = now
	}
	if input.Status == ExhibitionStatusCompleted && completedAt == "" {
		completedAt = now
	}
	if input.Status == ExhibitionStatusArchived && archivedAt == "" {
		archivedAt = now
	}
	result, err := s.db.ExecContext(ctx, `
UPDATE exhibition_lists
SET status=?, locked=?, lock_reason=?, locked_at=?, completed_at=?, archived_at=?,
    revision=revision+1, updated_at=?
WHERE id=? AND (?=0 OR revision=?)
`, input.Status, boolToInt(locked), reason, lockedAt, completedAt, archivedAt, now,
		id, input.ExpectedRevision, input.ExpectedRevision)
	if err != nil {
		return ExhibitionList{}, fmt.Errorf("set exhibition status: %w", err)
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return ExhibitionList{}, ErrExhibitionStale
	}
	return s.Get(ctx, id)
}

func validExhibitionTransition(from, to ExhibitionStatus) bool {
	switch from {
	case ExhibitionStatusDraft:
		return to == ExhibitionStatusOpen || to == ExhibitionStatusArchived
	case ExhibitionStatusOpen:
		return to == ExhibitionStatusDraft || to == ExhibitionStatusLocked || to == ExhibitionStatusArchived
	case ExhibitionStatusLocked:
		return to == ExhibitionStatusOpen || to == ExhibitionStatusRunning || to == ExhibitionStatusArchived
	case ExhibitionStatusRunning:
		return to == ExhibitionStatusLocked || to == ExhibitionStatusCompleted
	case ExhibitionStatusCompleted:
		return to == ExhibitionStatusOpen || to == ExhibitionStatusArchived
	case ExhibitionStatusArchived:
		return to == ExhibitionStatusOpen
	default:
		return false
	}
}

type exhibitionRevisionExecer interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}

func bumpExhibitionRevision(ctx context.Context, execer exhibitionRevisionExecer, listID string, expected int, now string) error {
	result, err := execer.ExecContext(ctx, `
UPDATE exhibition_lists SET revision=revision+1, updated_at=? WHERE id=? AND (?=0 OR revision=?)
`, now, listID, expected, expected)
	if err != nil {
		return fmt.Errorf("update exhibition revision: %w", err)
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return ErrExhibitionStale
	}
	return nil
}
