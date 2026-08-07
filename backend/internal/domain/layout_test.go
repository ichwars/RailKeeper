package domain_test

import (
	"testing"

	"railkeeper/backend/internal/domain"
)

func TestLayoutEnumsAcceptOnlyDefinedValues(t *testing.T) {
	tests := []struct {
		name  string
		valid bool
		want  bool
	}{
		{"private layout", domain.LayoutKindPrivate.Valid(), true},
		{"club layout", domain.LayoutKindClub.Valid(), true},
		{"invalid layout", domain.LayoutKind("public").Valid(), false},
		{"baseboard unit", domain.LayoutUnitKindBaseboard.Valid(), true},
		{"module unit", domain.LayoutUnitKindModule.Valid(), true},
		{"segment unit", domain.LayoutUnitKindSegment.Valid(), true},
		{"area unit", domain.LayoutUnitKindArea.Valid(), true},
		{"invalid unit", domain.LayoutUnitKind("room").Valid(), false},
		{"draft revision", domain.PlanRevisionDraft.Valid(), true},
		{"review revision", domain.PlanRevisionReview.Valid(), true},
		{"published revision", domain.PlanRevisionPublished.Valid(), true},
		{"archived revision", domain.PlanRevisionArchived.Valid(), true},
		{"invalid revision", domain.PlanRevisionStatus("deleted").Valid(), false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.valid != test.want {
				t.Fatalf("Valid() = %t, want %t", test.valid, test.want)
			}
		})
	}
}
