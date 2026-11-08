package exhibition

import (
	"context"
	"strings"
)

func (s *Service) Workspace(ctx context.Context, id string) (ExhibitionWorkspace, error) {
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
func (s *Service) attachExhibitionExceptions(ctx context.Context, listID string, conflicts []ExhibitionConflict) error {
	exceptions, err := s.repository.Exceptions(ctx, listID)
	if err != nil {
		return err
	}
	for index := range conflicts {
		if reason, ok := exceptions[conflicts[index].ID]; ok {
			conflicts[index].Excepted = true
			conflicts[index].ExceptionReason = reason
		}
	}
	return nil
}
