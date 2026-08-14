package domain

type LayoutKind string

const (
	LayoutKindPrivate LayoutKind = "private"
	LayoutKindClub    LayoutKind = "club"
)

func (kind LayoutKind) Valid() bool {
	return kind == LayoutKindPrivate || kind == LayoutKindClub
}

type LayoutUnitKind string

const (
	LayoutUnitKindBaseboard LayoutUnitKind = "baseboard"
	LayoutUnitKindModule    LayoutUnitKind = "module"
	LayoutUnitKindSegment   LayoutUnitKind = "segment"
	LayoutUnitKindArea      LayoutUnitKind = "area"
)

func (kind LayoutUnitKind) Valid() bool {
	switch kind {
	case LayoutUnitKindBaseboard, LayoutUnitKindModule, LayoutUnitKindSegment, LayoutUnitKindArea:
		return true
	default:
		return false
	}
}

type LayoutUnitPortKind string

const (
	LayoutUnitPortTrack     LayoutUnitPortKind = "track"
	LayoutUnitPortPower     LayoutUnitPortKind = "power"
	LayoutUnitPortDigital   LayoutUnitPortKind = "digital"
	LayoutUnitPortFeedback  LayoutUnitPortKind = "feedback"
	LayoutUnitPortAccessory LayoutUnitPortKind = "accessory"
	LayoutUnitPortOther     LayoutUnitPortKind = "other"
)

func (kind LayoutUnitPortKind) Valid() bool {
	switch kind {
	case LayoutUnitPortTrack, LayoutUnitPortPower, LayoutUnitPortDigital,
		LayoutUnitPortFeedback, LayoutUnitPortAccessory, LayoutUnitPortOther:
		return true
	default:
		return false
	}
}

type PlanRevisionStatus string

const (
	PlanRevisionDraft     PlanRevisionStatus = "draft"
	PlanRevisionReview    PlanRevisionStatus = "review"
	PlanRevisionPublished PlanRevisionStatus = "published"
	PlanRevisionArchived  PlanRevisionStatus = "archived"
)

func (status PlanRevisionStatus) Valid() bool {
	switch status {
	case PlanRevisionDraft, PlanRevisionReview, PlanRevisionPublished, PlanRevisionArchived:
		return true
	default:
		return false
	}
}
