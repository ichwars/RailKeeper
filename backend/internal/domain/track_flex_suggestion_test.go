package domain

import (
	"reflect"
	"testing"
)

func TestFlexSuggestionIsDeterministicAndUsesNaturalStraightHandles(t *testing.T) {
	input := FlexTrackSuggestionInput{
		EndXMM: 664, MaximumLengthMM: 664, RadiusLimitMM: 543,
	}
	first := SuggestFlexTrackPath(input)
	second := SuggestFlexTrackPath(input)
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("suggestion is not deterministic:\nfirst=%#v\nsecond=%#v", first, second)
	}
	if first.Path.EndXMM != 664 || first.Path.EndYMM != 0 ||
		first.Path.EndDirectionDegrees != 0 || first.Path.StartHandleMM != 664.0/3 ||
		first.Path.EndHandleMM != 664.0/3 || first.Effective.LengthMM != 664 ||
		first.LengthExceeded || first.RadiusBelowLimit || !first.Applicable {
		t.Fatalf("unexpected straight suggestion: %#v", first)
	}
}

func TestFlexSuggestionDistinguishesOverlengthAndRadiusWarning(t *testing.T) {
	overlength := SuggestFlexTrackPath(FlexTrackSuggestionInput{
		EndXMM: 700, MaximumLengthMM: 664, RadiusLimitMM: 543,
	})
	if !overlength.LengthExceeded || overlength.Applicable {
		t.Fatalf("overlength suggestion must not be applicable: %#v", overlength)
	}

	tight := SuggestFlexTrackPath(FlexTrackSuggestionInput{
		EndXMM: 350, EndYMM: 180, EndDirectionDegrees: 45,
		MaximumLengthMM: 664, RadiusLimitMM: 700,
	})
	if tight.LengthExceeded || !tight.RadiusBelowLimit || !tight.Applicable ||
		tight.Effective.MinimumRadiusMM == nil || *tight.Effective.MinimumRadiusMM >= 700 {
		t.Fatalf("tight suggestion must remain applicable with warning: %#v", tight)
	}
}
