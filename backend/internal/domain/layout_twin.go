package domain

type LayoutTechnicalPositionKind string

const (
	LayoutPositionTurnout  LayoutTechnicalPositionKind = "turnout"
	LayoutPositionSignal   LayoutTechnicalPositionKind = "signal"
	LayoutPositionFeedback LayoutTechnicalPositionKind = "feedback"
	LayoutPositionDecoder  LayoutTechnicalPositionKind = "decoder"
	LayoutPositionLighting LayoutTechnicalPositionKind = "lighting"
	LayoutPositionPower    LayoutTechnicalPositionKind = "power"
	LayoutPositionSensor   LayoutTechnicalPositionKind = "sensor"
	LayoutPositionOther    LayoutTechnicalPositionKind = "other"
)

func (kind LayoutTechnicalPositionKind) Valid() bool {
	switch kind {
	case LayoutPositionTurnout,
		LayoutPositionSignal,
		LayoutPositionFeedback,
		LayoutPositionDecoder,
		LayoutPositionLighting,
		LayoutPositionPower,
		LayoutPositionSensor,
		LayoutPositionOther:
		return true
	default:
		return false
	}
}
