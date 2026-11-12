package infrastructure_test

import (
	"math"
	"slices"
	"testing"

	"railkeeper/backend/internal/application"
	"railkeeper/backend/internal/infrastructure"
)

func TestDigitalCenterWorkItemPagePreservesSearchOrderAndTotals(t *testing.T) {
	ctx := t.Context()
	repo := infrastructure.NewDigitalCenterWorkspaceRepository(testDB(t))
	newSession := func() string {
		t.Helper()
		session, err := repo.CreateSession(ctx, application.DigitalCenterReadSession{
			Provider: "ecos", State: application.DigitalCenterSessionReady,
		})
		if err != nil {
			t.Fatal(err)
		}
		return session.ID
	}
	session := newSession()
	items := []application.DigitalCenterWorkItem{
		{ID: "second", CenterObjectID: "2", Name: "br 18", CompareStatus: application.DigitalCompareDeviation},
		{ID: "first", CenterObjectID: "10", Name: "BR 18", CompareStatus: application.DigitalCompareOK},
		{ID: "accent", CenterObjectID: "4", Name: "KÖF %_A", VehicleID: "fahrzeug-Ä", Protocol: "mFx", Address: 521,
			CompareStatus: application.DigitalCompareDeviation, Center: map[string]any{"description": "payload only"}},
		{ID: "brand", CenterObjectID: "5", Name: "Märklin", CompareStatus: application.DigitalCompareDeviation},
		{ID: "nul", CenterObjectID: "6", Name: "NUL\x00Ä", CompareStatus: application.DigitalCompareOK},
		{ID: "object", CenterObjectID: "KELVIN-K", Name: "Zug", CompareStatus: application.DigitalCompareNew},
	}
	if err := repo.ReplaceWorkItems(ctx, session, items); err != nil {
		t.Fatal(err)
	}
	otherSession := newSession()
	if err := repo.ReplaceWorkItems(ctx, otherSession, []application.DigitalCenterWorkItem{{
		CenterObjectID: "4", Name: "KÖF %_A", CompareStatus: application.DigitalCompareDeviation,
	}}); err != nil {
		t.Fatal(err)
	}
	service := application.NewDigitalCenterWorkspaceService(repo, nil, nil, nil, nil, nil)
	for _, test := range []struct {
		name       string
		filter     application.DigitalCenterWorkItemFilter
		ids        []string
		total      int
		totalPages int
	}{
		{"ties", application.DigitalCenterWorkItemFilter{PageSize: 2}, []string{"first", "second"}, 6, 3},
		{"second page", application.DigitalCenterWorkItemFilter{Page: 2, PageSize: 2}, []string{"accent", "brand"}, 6, 3},
		{"last page", application.DigitalCenterWorkItemFilter{Page: 2, PageSize: 4}, []string{"nul", "object"}, 6, 2},
		{"search and status", application.DigitalCenterWorkItemFilter{Query: "  kÖf  ", CompareStatus: application.DigitalCompareDeviation}, []string{"accent"}, 1, 1},
		{"literal percent underscore", application.DigitalCenterWorkItemFilter{Query: "%_"}, []string{"accent"}, 1, 1},
		{"vehicle id", application.DigitalCenterWorkItemFilter{Query: "FAHRZEUG-ä"}, []string{"accent"}, 1, 1},
		{"protocol", application.DigitalCenterWorkItemFilter{Query: "MFX"}, []string{"accent"}, 1, 1},
		{"address", application.DigitalCenterWorkItemFilter{Query: "521"}, []string{"accent"}, 1, 1},
		{"object id", application.DigitalCenterWorkItemFilter{Query: "kelvin-k"}, []string{"object"}, 1, 1},
		{"embedded NUL", application.DigitalCenterWorkItemFilter{Query: "\x00ä"}, []string{"nul"}, 1, 1},
		{"status only", application.DigitalCenterWorkItemFilter{CompareStatus: application.DigitalCompareDeviation, PageSize: 2}, []string{"second", "accent"}, 3, 2},
		{"search excludes JSON", application.DigitalCenterWorkItemFilter{Query: "payload only"}, nil, 0, 0},
		{"excessive page", application.DigitalCenterWorkItemFilter{Page: math.MaxInt, PageSize: 100}, nil, 6, 1},
	} {
		t.Run(test.name, func(t *testing.T) {
			page, err := service.ListWorkItems(ctx, session, test.filter)
			if err != nil {
				t.Fatal(err)
			}
			var ids []string
			for _, item := range page.Items {
				ids = append(ids, item.ID)
				if item.SessionID != session {
					t.Fatalf("item %q from session %q", item.ID, item.SessionID)
				}
			}
			if !slices.Equal(ids, test.ids) || (test.ids == nil && page.Items != nil) || page.Total != test.total ||
				page.TotalPages != test.totalPages || page.SessionTotal != len(items) || page.Page != max(1, test.filter.Page) {
				t.Fatalf("page=%#v ids=%v, want ids=%v total=%d pages=%d", page, ids, test.ids, test.total, test.totalPages)
			}
		})
	}
}
