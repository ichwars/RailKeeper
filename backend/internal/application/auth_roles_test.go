package application

import "testing"

func TestPlannerInheritsViewerReadAccessOnly(t *testing.T) {
	if !hasRole([]string{"Planner"}, "Viewer") {
		t.Fatal("Planner must inherit Viewer read access")
	}
	if hasRole([]string{"Planner"}, "Editor") {
		t.Fatal("Planner must not inherit Editor write access")
	}
	if hasRole([]string{"Editor"}, "Planner") {
		t.Fatal("Editor must not inherit Planner write access")
	}
}
