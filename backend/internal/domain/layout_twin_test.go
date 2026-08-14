package domain

import "testing"

func TestLayoutTechnicalPositionKinds(t *testing.T) {
	valid := []LayoutTechnicalPositionKind{
		LayoutPositionTurnout,
		LayoutPositionSignal,
		LayoutPositionFeedback,
		LayoutPositionDecoder,
		LayoutPositionLighting,
		LayoutPositionPower,
		LayoutPositionSensor,
		LayoutPositionOther,
	}
	for _, kind := range valid {
		if !kind.Valid() {
			t.Fatalf("expected %q to be valid", kind)
		}
	}
	if LayoutTechnicalPositionKind("command").Valid() {
		t.Fatal("control commands must not be valid technical positions")
	}
}
