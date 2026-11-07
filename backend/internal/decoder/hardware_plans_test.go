package decoder

import (
	"errors"
	"strings"
	"testing"
)

func stringRef(value string) *string { return &value }

func hardwareContext() *DecoderConfigurationContext {
	return &DecoderConfigurationContext{VehicleID: "vehicle", VehicleName: "Lok", InventoryNumber: "L42",
		Instances: []DecoderInstance{{ID: "legacy-decoder", Label: "Alt", CreatedAt: "original"}}}
}

func TestDecoderInstanceKeepsLegacyIdentityAndOriginalRecording(t *testing.T) {
	before := hardwareContext()
	input := DecoderInstanceInput{ID: "legacy-decoder", Label: " Neu ", Notes: " Bemerkung "}
	item, err := PlanInstance(before, input, "now", false)
	if err != nil || item.ID != input.ID || item.Label != "Neu" || item.Notes != "Bemerkung" || item.CreatedAt != "original" || before.Instances[0].Label != "Alt" {
		t.Fatalf("legacy instance update changed identity or original state: %+v, %v", item, err)
	}
	if _, err := PlanInstance(before, input, "now", true); !errors.Is(err, ErrValidation) {
		t.Fatal("new instances accepted a legacy identifier")
	}
	input.ID = strings.Repeat("a", 32)
	before.Instances = append(before.Instances, DecoderInstance{ID: input.ID})
	if _, err := PlanInstance(before, input, "now", true); !errors.Is(err, ErrConflict) {
		t.Fatal("duplicate physical decoder identity accepted")
	}
	input.ID = "missing"
	if _, err := PlanInstance(before, input, "now", false); !errors.Is(err, ErrNotFound) {
		t.Fatal("update invented a missing decoder")
	}
}

func TestInstallationAndRemovalDoNotInventPhysicalDates(t *testing.T) {
	before := hardwareContext()
	input := DecoderInstallationInput{ID: strings.Repeat("b", 32), DecoderID: "legacy-decoder", Notes: " Montage "}
	installed, err := PlanInstallation(before, input, "recorded-now")
	if err != nil || installed.InstalledAt != nil || installed.RemovedAt != nil || installed.RecordedAt != "recorded-now" || installed.VehicleID == nil || *installed.VehicleID != before.VehicleID {
		t.Fatalf("installation invented a date or lost vehicle identity: %+v, %v", installed, err)
	}
	before.Installations = []DecoderInstallation{installed}
	if _, err := PlanInstallation(before, DecoderInstallationInput{ID: strings.Repeat("c", 32), DecoderID: input.DecoderID}, "now"); !errors.Is(err, ErrConflict) {
		t.Fatal("simultaneous installation of the same physical decoder accepted")
	}
	removed, err := PlanRemoval(before, installed.ID, DecoderRemovalInput{Note: " Ausbau "}, "later-recorded")
	if err != nil || removed.State != "removed" || removed.RemovedAt != nil || removed.Notes != "Montage\nAusbau" || removed.VehicleID == nil || *removed.VehicleID != before.VehicleID || removed.RecordedAt != installed.RecordedAt || before.Installations[0].State != "installed" {
		t.Fatalf("removal lost history, dates or vehicle link: %+v, %v", removed, err)
	}
	before.Installations[0] = removed
	if _, err := PlanRemoval(before, removed.ID, DecoderRemovalInput{}, "now"); !errors.Is(err, ErrConflict) {
		t.Fatal("second removal accepted")
	}
	corrected, err := PlanInstallationDates(before, removed.ID, DecoderInstallationDatesInput{InstalledAt: stringRef("2026-01-01"), RemovedAt: stringRef("2026-02-02")}, "corrected")
	if err != nil || corrected.State != "removed" || corrected.RecordedAt != "recorded-now" {
		t.Fatal("historical date correction changed physical state or recording time")
	}
}

func TestHardwarePlansRejectForeignChronologicalAndOversizedChanges(t *testing.T) {
	before := hardwareContext()
	before.Installations = []DecoderInstallation{{ID: "installation", VehicleReferenceID: before.VehicleID, State: "installed", InstalledAt: stringRef("2026-02-01"), Notes: strings.Repeat("n", 4000)}}
	if _, err := PlanRemoval(before, "installation", DecoderRemovalInput{RemovedAt: stringRef("2026-01-01")}, "now"); !errors.Is(err, ErrValidation) {
		t.Fatal("removal before installation accepted")
	}
	if _, err := PlanRemoval(before, "installation", DecoderRemovalInput{Note: "more"}, "now"); !errors.Is(err, ErrValidation) {
		t.Fatal("combined note limit ignored")
	}
	if _, err := PlanInstallationDates(before, "installation", DecoderInstallationDatesInput{RemovedAt: stringRef("2026-03-01")}, "now"); !errors.Is(err, ErrValidation) {
		t.Fatal("date correction removed an installed decoder")
	}
	before.Installations[0].VehicleReferenceID = "foreign"
	if _, err := PlanRemoval(before, "installation", DecoderRemovalInput{}, "now"); !errors.Is(err, ErrNotFound) {
		t.Fatal("foreign installation modified")
	}
	if ValidInstallationInput(DecoderInstallationInput{ID: strings.Repeat("b", 32), DecoderID: "legacy-decoder", Notes: strings.Repeat(" ", 4001)}) {
		t.Fatal("raw note limit changed by trimming")
	}
}

func TestWorkspaceAssignmentPreservesPrechangeEvidenceAndOriginalProfile(t *testing.T) {
	before := hardwareContext()
	before.Workspaces = []DecoderCVWorkspace{{ID: "workspace", VehicleID: before.VehicleID, Profile: "  Profil ä  ", UpdatedAt: "old"}}
	before.Installations = []DecoderInstallation{{ID: "installation", VehicleID: stringRef(before.VehicleID), State: "installed"}}
	input := DecoderCVAssignmentInput{SnapshotID: strings.Repeat("d", 32), InstallationID: stringRef("installation")}
	plan, err := PlanWorkspaceAssignment(before, "workspace", input, "now")
	if err != nil || plan.Before.InstallationID != nil || plan.After.InstallationID == nil || plan.Before.Profile != "  Profil ä  " || plan.After.Profile != plan.Before.Profile || before.Workspaces[0].InstallationID != nil || plan.After.UpdatedAt != "now" {
		t.Fatalf("assignment lost the prechange partial state: %+v, %v", plan, err)
	}
	before.Workspaces[0] = plan.After
	if _, err := PlanWorkspaceAssignment(before, "workspace", input, "now"); !errors.Is(err, ErrConflict) {
		t.Fatal("unchanged assignment accepted")
	}
	input.InstallationID = nil
	if plan, err := PlanWorkspaceAssignment(before, "workspace", input, "now"); err != nil || plan.Before.InstallationID == nil || plan.After.InstallationID != nil {
		t.Fatal("explicit unassignment lost old hardware reference")
	}
	before.Workspaces[0].InstallationID = nil
	input.InstallationID = stringRef("installation")
	before.Installations[0].State = "removed"
	if _, err := PlanWorkspaceAssignment(before, "workspace", input, "now"); !errors.Is(err, ErrValidation) {
		t.Fatal("removed hardware assigned")
	}
	if _, err := PlanWorkspace(before, DecoderCVWorkspaceInput{ID: strings.Repeat("e", 32), Profile: " Profile "}, "now", true); !errors.Is(err, ErrConflict) {
		t.Fatal("global workspace collision ignored")
	}
}
