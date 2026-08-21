package application_test

import (
	"context"
	"errors"
	"testing"

	"railkeeper/backend/internal/application"
)

func TestExhibitionWorkspacePersistsEventMetadataAndRevision(t *testing.T) {
	db := testDB(t)
	service := application.NewExhibitionService(db)
	ctx := context.Background()

	created, err := service.Create(ctx, application.ExhibitionListInput{
		Designation:       "Modellbahntage Köln",
		Date:              "2026-08-22",
		EndDate:           "2026-08-24",
		Location:          "Köln Messe",
		Description:       "Drei öffentliche Fahrtage",
		OrganizationNotes: "Aufbau Freitagabend",
		Status:            application.ExhibitionStatusOpen,
	})
	if err != nil {
		t.Fatal(err)
	}
	if created.EndDate != "2026-08-24" || created.Location != "Köln Messe" ||
		created.Description != "Drei öffentliche Fahrtage" || created.OrganizationNotes != "Aufbau Freitagabend" {
		t.Fatalf("event metadata not persisted: %#v", created)
	}
	if created.Status != application.ExhibitionStatusOpen || created.Revision != 1 || created.Locked {
		t.Fatalf("unexpected lifecycle defaults: %#v", created)
	}

	_, err = service.Create(ctx, application.ExhibitionListInput{
		Designation: "Ungültig", Date: "2026-08-24", EndDate: "2026-08-22",
	})
	if !errors.Is(err, application.ErrExhibitionValidation) {
		t.Fatalf("expected date-range validation, got %v", err)
	}
}

func TestExhibitionWorkspaceDetectsAddressConflictsByDayAndInterface(t *testing.T) {
	db := testDB(t)
	service := application.NewExhibitionService(db)
	ctx := context.Background()
	list, err := service.Create(ctx, application.ExhibitionListInput{
		Designation: "Köln", Date: "2026-08-22", EndDate: "2026-08-24",
		Status: application.ExhibitionStatusOpen,
	})
	if err != nil {
		t.Fatal(err)
	}

	entries := []application.ExhibitionEntryInput{
		{Owner: "Michael", LocomotiveName: "BR 103", DayScope: "day1,day2", DTDecoder: true, DecoderNumber: "103", InterfaceName: "Lokmaus 3", FunctionKeys: "F0", ImageURL: "image-a"},
		{Owner: "Thomas", LocomotiveName: "V 200", DayScope: "day2,day3", DTDecoder: true, DecoderNumber: "103", InterfaceName: "Lokmaus 3", FunctionKeys: "F0", ImageURL: "image-b"},
		{Owner: "Peter", LocomotiveName: "BR 86", DayScope: "day2,day3", DTDecoder: true, DecoderNumber: "103", InterfaceName: "ESU ECoS", FunctionKeys: "F0", ImageURL: "image-c"},
	}
	for _, input := range entries {
		if _, err := service.CreateEntry(ctx, list.ID, input); err != nil {
			t.Fatal(err)
		}
	}

	workspace, err := service.Workspace(ctx, list.ID)
	if err != nil {
		t.Fatal(err)
	}
	if workspace.Summary.EntryCount != 3 || workspace.Summary.OwnerCount != 3 {
		t.Fatalf("unexpected summary: %#v", workspace.Summary)
	}
	if len(workspace.Conflicts) != 1 {
		t.Fatalf("expected one scoped address conflict, got %#v", workspace.Conflicts)
	}
	conflict := workspace.Conflicts[0]
	if conflict.Kind != application.ExhibitionConflictAddress || conflict.InterfaceName != "Lokmaus 3" ||
		len(conflict.DayScopes) != 1 || conflict.DayScopes[0] != "day2" {
		t.Fatalf("unexpected conflict scope: %#v", conflict)
	}
	if workspace.Summary.ConflictCount != 1 || workspace.Summary.ReadyCount != 1 {
		t.Fatalf("unexpected ready/conflict counts: %#v", workspace.Summary)
	}
}

func TestExhibitionWorkspaceReportsMissingDataAndReadiness(t *testing.T) {
	db := testDB(t)
	service := application.NewExhibitionService(db)
	ctx := context.Background()
	list, err := service.Create(ctx, application.ExhibitionListInput{
		Designation: "Köln", Date: "2026-08-22", EndDate: "2026-08-24",
	})
	if err != nil {
		t.Fatal(err)
	}
	entry, err := service.CreateEntry(ctx, list.ID, application.ExhibitionEntryInput{
		LocomotiveName: "Gastlok", DayScope: "day1", DTDecoder: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	workspace, err := service.Workspace(ctx, list.ID)
	if err != nil {
		t.Fatal(err)
	}
	if workspace.Summary.ReadyCount != 0 || workspace.Readiness.Total != 1 ||
		workspace.Readiness.AddressesChecked != 0 || workspace.Readiness.FunctionsDocumented != 0 ||
		workspace.Readiness.ImagesPresent != 0 {
		t.Fatalf("unexpected readiness: %#v", workspace)
	}
	if len(workspace.Entries) != 1 || workspace.Entries[0].ID != entry.ID ||
		workspace.Entries[0].Status != application.ExhibitionEntryStatusMissing {
		t.Fatalf("expected missing-data status, got %#v", workspace.Entries)
	}
}

func TestExhibitionLifecycleRejectsStaleAndUnsafeLock(t *testing.T) {
	db := testDB(t)
	service := application.NewExhibitionService(db)
	ctx := context.Background()
	list, err := service.Create(ctx, application.ExhibitionListInput{
		Designation: "Köln", Date: "2026-08-22", EndDate: "2026-08-24",
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"BR 103", "V 200"} {
		if _, err := service.CreateEntry(ctx, list.ID, application.ExhibitionEntryInput{
			Owner: name, LocomotiveName: name, DayScope: "day1", DTDecoder: true,
			DecoderNumber: "3", InterfaceName: "ECoS",
		}); err != nil {
			t.Fatal(err)
		}
	}
	current, err := service.Get(ctx, list.ID)
	if err != nil {
		t.Fatal(err)
	}
	_, err = service.SetStatus(ctx, list.ID, application.ExhibitionStatusInput{
		Status: application.ExhibitionStatusLocked, ExpectedRevision: current.Revision,
	})
	if !errors.Is(err, application.ErrExhibitionConflicts) {
		t.Fatalf("expected unsafe lock rejection, got %v", err)
	}
	locked, err := service.SetStatus(ctx, list.ID, application.ExhibitionStatusInput{
		Status: application.ExhibitionStatusLocked, ExpectedRevision: current.Revision,
		ConfirmConflicts: true, Reason: "Betrieblich getrennte Fahrabschnitte",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !locked.Locked || locked.Status != application.ExhibitionStatusLocked || locked.Revision <= current.Revision {
		t.Fatalf("expected locked revision, got %#v", locked)
	}
	_, err = service.SetStatus(ctx, list.ID, application.ExhibitionStatusInput{
		Status: application.ExhibitionStatusOpen, ExpectedRevision: current.Revision,
	})
	if !errors.Is(err, application.ErrExhibitionStale) {
		t.Fatalf("expected stale revision, got %v", err)
	}
}

func TestExhibitionConflictExceptionIsPersistentAndVisible(t *testing.T) {
	db := testDB(t)
	service := application.NewExhibitionService(db)
	ctx := context.Background()
	list, err := service.Create(ctx, application.ExhibitionListInput{Designation: "Köln", Date: "2026-08-22"})
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"Lok A", "Lok B"} {
		if _, err := service.CreateEntry(ctx, list.ID, application.ExhibitionEntryInput{
			Owner: name, LocomotiveName: name, DayScope: "all", DTDecoder: true,
			DecoderNumber: "7", InterfaceName: "Z21",
		}); err != nil {
			t.Fatal(err)
		}
	}
	workspace, err := service.Workspace(ctx, list.ID)
	if err != nil || len(workspace.Conflicts) != 1 {
		t.Fatalf("expected conflict: %#v %v", workspace, err)
	}
	updated, err := service.SetConflictException(ctx, list.ID, workspace.Conflicts[0].ID,
		application.ExhibitionConflictExceptionInput{Reason: "Getrennte Boosterbezirke", ExpectedRevision: workspace.List.Revision})
	if err != nil {
		t.Fatal(err)
	}
	if len(updated.Conflicts) != 1 || !updated.Conflicts[0].Excepted ||
		updated.Conflicts[0].ExceptionReason != "Getrennte Boosterbezirke" {
		t.Fatalf("expected visible persisted exception, got %#v", updated.Conflicts)
	}
}

func TestExhibitionLifecycleRejectsInvalidTransition(t *testing.T) {
	db := testDB(t)
	service := application.NewExhibitionService(db)
	list, err := service.Create(t.Context(), application.ExhibitionListInput{
		Designation: "Entwurf", Date: "2026-08-22", Status: application.ExhibitionStatusDraft,
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = service.SetStatus(t.Context(), list.ID, application.ExhibitionStatusInput{
		Status: application.ExhibitionStatusRunning, ExpectedRevision: list.Revision,
	})
	if !errors.Is(err, application.ErrExhibitionValidation) {
		t.Fatalf("expected invalid lifecycle transition, got %v", err)
	}
}

func TestCompletedAndArchivedExhibitionsRejectEntryChanges(t *testing.T) {
	for _, finalStatus := range []application.ExhibitionStatus{
		application.ExhibitionStatusCompleted,
		application.ExhibitionStatusArchived,
	} {
		t.Run(string(finalStatus), func(t *testing.T) {
			db := testDB(t)
			service := application.NewExhibitionService(db)
			list, err := service.Create(t.Context(), application.ExhibitionListInput{
				Designation: "Lifecycle", Date: "2026-08-22", Status: application.ExhibitionStatusOpen,
			})
			if err != nil {
				t.Fatal(err)
			}
			current := list
			for _, status := range []application.ExhibitionStatus{
				application.ExhibitionStatusLocked,
				application.ExhibitionStatusRunning,
				finalStatus,
			} {
				if finalStatus == application.ExhibitionStatusArchived && status == application.ExhibitionStatusRunning {
					continue
				}
				current, err = service.SetStatus(t.Context(), list.ID, application.ExhibitionStatusInput{
					Status: status, ExpectedRevision: current.Revision,
				})
				if err != nil {
					t.Fatal(err)
				}
			}
			_, err = service.CreateEntry(t.Context(), list.ID, application.ExhibitionEntryInput{
				LocomotiveName: "Zu spät",
			})
			if !errors.Is(err, application.ErrExhibitionLocked) {
				t.Fatalf("CreateEntry() error = %v, want locked", err)
			}
		})
	}
}
