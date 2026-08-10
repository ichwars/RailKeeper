package domain

import "math"

type FlexTrackSuggestionInput struct {
	EndXMM              float64
	EndYMM              float64
	EndDirectionDegrees float64
	MaximumLengthMM     float64
	RadiusLimitMM       float64
}

type FlexTrackSuggestion struct {
	Path             FlexTrackPath          `json:"path"`
	Effective        EffectiveTrackGeometry `json:"-"`
	LengthExceeded   bool                   `json:"lengthExceeded"`
	RadiusBelowLimit bool                   `json:"radiusBelowLimit"`
	Applicable       bool                   `json:"applicable"`
}

type flexSuggestionCandidate struct {
	Suggestion         FlexTrackSuggestion
	RadiusShortfallMM  float64
	CurvatureVariation float64
}

func SuggestFlexTrackPath(input FlexTrackSuggestionInput) FlexTrackSuggestion {
	if !validFlexSuggestionInput(input) {
		return FlexTrackSuggestion{}
	}
	input.EndDirectionDegrees = NormalizeTrackRotation(input.EndDirectionDegrees)
	chord := math.Hypot(input.EndXMM, input.EndYMM)
	base := FlexTrackPath{
		SchemaVersion: 1, EndXMM: input.EndXMM, EndYMM: input.EndYMM,
		EndDirectionDegrees: input.EndDirectionDegrees,
		StartHandleMM:       chord / 3, EndHandleMM: chord / 3,
	}
	if math.Abs(input.EndYMM) <= 1e-9 &&
		trackAngleDifference(input.EndDirectionDegrees, 0) <= 1e-9 {
		return evaluateFlexSuggestion(base, input)
	}

	step := chord * 0.1
	var best flexSuggestionCandidate
	found := false
	for startIndex := 1; startIndex <= 15; startIndex++ {
		for endIndex := 1; endIndex <= 15; endIndex++ {
			path := base
			path.StartHandleMM = float64(startIndex) * step
			path.EndHandleMM = float64(endIndex) * step
			candidate, valid := scoreFlexSuggestion(path, input)
			if valid && (!found || betterFlexSuggestion(candidate, best)) {
				best, found = candidate, true
			}
		}
	}
	if !found {
		return evaluateFlexSuggestion(base, input)
	}
	for refinement := 0; refinement < 3; refinement++ {
		step /= 5
		centerStart := best.Suggestion.Path.StartHandleMM
		centerEnd := best.Suggestion.Path.EndHandleMM
		for startOffset := -5; startOffset <= 5; startOffset++ {
			for endOffset := -5; endOffset <= 5; endOffset++ {
				path := base
				path.StartHandleMM = centerStart + float64(startOffset)*step
				path.EndHandleMM = centerEnd + float64(endOffset)*step
				candidate, valid := scoreFlexSuggestion(path, input)
				if valid && betterFlexSuggestion(candidate, best) {
					best = candidate
				}
			}
		}
	}
	return best.Suggestion
}

func validFlexSuggestionInput(input FlexTrackSuggestionInput) bool {
	for _, value := range []float64{
		input.EndXMM, input.EndYMM, input.EndDirectionDegrees,
		input.MaximumLengthMM, input.RadiusLimitMM,
	} {
		if !finiteTrackNumber(value) {
			return false
		}
	}
	return math.Hypot(input.EndXMM, input.EndYMM) > 1e-9 &&
		input.MaximumLengthMM > 0 && input.RadiusLimitMM > 0
}

func scoreFlexSuggestion(path FlexTrackPath, input FlexTrackSuggestionInput) (flexSuggestionCandidate, bool) {
	suggestion := evaluateFlexSuggestion(path, input)
	if !suggestion.Applicable {
		return flexSuggestionCandidate{}, false
	}
	shortfall := 0.0
	if suggestion.Effective.MinimumRadiusMM != nil {
		shortfall = math.Max(0, input.RadiusLimitMM-*suggestion.Effective.MinimumRadiusMM)
	}
	control, _, err := validatedFlexControlPoints(path)
	if err != nil {
		return flexSuggestionCandidate{}, false
	}
	return flexSuggestionCandidate{
		Suggestion: suggestion, RadiusShortfallMM: shortfall,
		CurvatureVariation: flexCurvatureVariation(control),
	}, true
}

func evaluateFlexSuggestion(path FlexTrackPath, input FlexTrackSuggestionInput) FlexTrackSuggestion {
	effective, err := BuildFlexTrackGeometry(path)
	if err != nil {
		return FlexTrackSuggestion{Path: path}
	}
	lengthExceeded := effective.LengthMM > input.MaximumLengthMM+1e-9
	radiusBelow := effective.MinimumRadiusMM != nil &&
		*effective.MinimumRadiusMM+1e-9 < input.RadiusLimitMM
	return FlexTrackSuggestion{
		Path: path, Effective: effective, LengthExceeded: lengthExceeded,
		RadiusBelowLimit: radiusBelow, Applicable: !lengthExceeded,
	}
}

func betterFlexSuggestion(candidate, current flexSuggestionCandidate) bool {
	values := [][2]float64{
		{candidate.RadiusShortfallMM, current.RadiusShortfallMM},
		{candidate.CurvatureVariation, current.CurvatureVariation},
		{candidate.Suggestion.Effective.LengthMM, current.Suggestion.Effective.LengthMM},
		{candidate.Suggestion.Path.StartHandleMM, current.Suggestion.Path.StartHandleMM},
		{candidate.Suggestion.Path.EndHandleMM, current.Suggestion.Path.EndHandleMM},
	}
	for _, pair := range values {
		if pair[0] < pair[1]-1e-9 {
			return true
		}
		if pair[0] > pair[1]+1e-9 {
			return false
		}
	}
	return false
}

func flexCurvatureVariation(control [4]TrackPoint) float64 {
	previous := flexCurvatureAt(control, 0)
	variation := 0.0
	for index := 1; index <= 32; index++ {
		current := flexCurvatureAt(control, float64(index)/32)
		variation += math.Abs(current - previous)
		previous = current
	}
	return variation
}

func flexCurvatureAt(control [4]TrackPoint, parameter float64) float64 {
	oneMinus := 1 - parameter
	firstX := 3 * (oneMinus*oneMinus*(control[1].XMM-control[0].XMM) +
		2*oneMinus*parameter*(control[2].XMM-control[1].XMM) +
		parameter*parameter*(control[3].XMM-control[2].XMM))
	firstY := 3 * (oneMinus*oneMinus*(control[1].YMM-control[0].YMM) +
		2*oneMinus*parameter*(control[2].YMM-control[1].YMM) +
		parameter*parameter*(control[3].YMM-control[2].YMM))
	secondX := 6 * (oneMinus*(control[2].XMM-2*control[1].XMM+control[0].XMM) +
		parameter*(control[3].XMM-2*control[2].XMM+control[1].XMM))
	secondY := 6 * (oneMinus*(control[2].YMM-2*control[1].YMM+control[0].YMM) +
		parameter*(control[3].YMM-2*control[2].YMM+control[1].YMM))
	speed := math.Hypot(firstX, firstY)
	if speed <= 1e-12 {
		return 0
	}
	return math.Abs(firstX*secondY-firstY*secondX) / (speed * speed * speed)
}
