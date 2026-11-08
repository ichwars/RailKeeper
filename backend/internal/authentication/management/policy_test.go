package management

import "testing"

func TestRetiredPlannerHasNoImplicitAccess(t *testing.T) {
	if HasRole([]string{"Planner"}, "Viewer") {
		t.Fatal("Unmigrated Planner must not have read access")
	}
	if HasRole([]string{"Planner"}, "Editor") {
		t.Fatal("Planner must not inherit Editor write access")
	}
	if HasRole([]string{"Editor"}, "Planner") {
		t.Fatal("Editor must not inherit Planner write access")
	}
}
