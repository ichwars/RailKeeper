package ecosmodel

import (
	"reflect"
	"strings"
	"testing"
	"unicode"
)

func TestReadProjectionKeepsProfilesAndExcludesRuntimeState(t *testing.T) {
	values := ParseECoSLocomotives([]string{
		`1001 addr[3] name["BR 218"] profile["  Mein Profil  "] protocol[DCC128]`,
		`1001 speed[12] dir[1] funcset[101] image[9] funcdesc[0,3]`,
	})
	if len(values) != 1 || values[0].Name != "BR 218" || values[0].Profile != "  Mein Profil  " {
		t.Fatalf("recorded profile changed: %#v", values)
	}
	for _, key := range []string{"speed", "dir", "funcset", "image"} {
		if IsAllowedECoSLocomotiveAttribute(key) {
			t.Fatalf("runtime field permitted: %s", key)
		}
	}
	if len(values[0].Functions) != 1 || values[0].Functions[0].Index != 0 {
		t.Fatal("static F0 description lost")
	}
}

func TestComparisonPlansRemainReadOnlyValuesAndRejectCommandInjection(t *testing.T) {
	current := ECoSLocomotive{Name: "Before", Address: 3, Protocol: "DCC128"}
	before := current
	for _, desired := range []ECoSLocomotiveSyncDesired{{Name: "bad\nset(1,status[STOP])"}, {Protocol: "DCC128])\nset(1,status[STOP])"}} {
		if _, commands, err := BuildECoSLocomotiveSyncCommands(1001, &current, desired); err == nil || len(commands) != 0 {
			t.Fatalf("unsafe comparison generated commands: %#v, %v", commands, err)
		}
	}
	changes, commands, err := BuildECoSLocomotiveSyncCommands(1001, &current,
		ECoSLocomotiveSyncDesired{Name: `My "quoted" locomotive`, Address: 24, Protocol: "DCC28"})
	if err != nil || len(commands) != 3 || !reflect.DeepEqual(current, before) {
		t.Fatalf("comparison mutated state: %v", err)
	}
	if !reflect.DeepEqual(ECoSLocomotiveSyncFields(changes), []string{"name", "addr", "protocol"}) {
		t.Fatal("comparison fields changed")
	}
}

func TestUnindexedPageRemainsDifferentFromExplicitZero(t *testing.T) {
	zero, one := 0, 1
	values := []ECoSCVValue{{Number: 257, IndexCV31: &zero, IndexCV32: &one},
		{Number: 257, IndexCV31: &zero, IndexCV32: &zero}, {Number: 257}}
	SortECoSCVValues(values)
	if ECoSCVPage(values[0]) != -1 || ECoSCVPage(values[1]) != 0 || ECoSCVPage(values[2]) != 1 {
		t.Fatalf("unknown page became a zero page: %#v", values)
	}
}

func TestTelemetryCopiesDoNotAliasMutableSlices(t *testing.T) {
	original := ECoSLiveStatus{SubscriptionCommands: []string{"original"}, PulseSamples: make([]ECoSLivePulseSample, 1), RecentEvents: make([]ECoSLiveEvent, 1)}
	copy := CloneECoSLiveStatus(original)
	copy.SubscriptionCommands[0] = "changed"
	if original.SubscriptionCommands[0] != "original" || &copy.PulseSamples[0] == &original.PulseSamples[0] ||
		&copy.RecentEvents[0] == &original.RecentEvents[0] {
		t.Fatal("telemetry state escaped its owner")
	}
}

func FuzzQuotedComparisonName(f *testing.F) {
	for _, name := range []string{"BR 218", `My "quoted" locomotive`, `C:\Users\train`, "Profil ä", "bad\ncommand", "]])"} {
		f.Add(name)
	}
	f.Fuzz(func(t *testing.T, name string) {
		if len(name) > 4096 {
			return
		}
		desired := CleanECoSLocomotiveSyncDesired(ECoSLocomotiveSyncDesired{Name: name})
		_, commands, err := BuildECoSLocomotiveSyncCommands(1001, &ECoSLocomotive{}, desired)
		if strings.IndexFunc(desired.Name, unicode.IsControl) >= 0 {
			if err == nil {
				t.Fatal("control characters were accepted")
			}
			return
		}
		if err != nil {
			t.Fatal(err)
		}
		if desired.Name == "" {
			if len(commands) != 0 {
				t.Fatal("empty name changed the device")
			}
			return
		}
		if len(commands) != 1 || CleanECoSValue(ParseECoSArguments(commands[0])["name"]) != desired.Name {
			t.Fatalf("quoted comparison did not round-trip: %q", desired.Name)
		}
	})
}
