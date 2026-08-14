package domain

import (
	"errors"
	"math"
	"strings"
	"unicode/utf8"
)

type FreePlanObjectKind string

const (
	FreePlanRectangle FreePlanObjectKind = "rectangle"
	FreePlanEllipse   FreePlanObjectKind = "ellipse"
	FreePlanLine      FreePlanObjectKind = "line"
	FreePlanLabel     FreePlanObjectKind = "label"
)

func (kind FreePlanObjectKind) Valid() bool {
	return kind == FreePlanRectangle || kind == FreePlanEllipse || kind == FreePlanLine || kind == FreePlanLabel
}

type FreePlanObjectCategory string

const (
	FreePlanStructure  FreePlanObjectCategory = "structure"
	FreePlanPlatform   FreePlanObjectCategory = "platform"
	FreePlanScenery    FreePlanObjectCategory = "scenery"
	FreePlanAnnotation FreePlanObjectCategory = "annotation"
)

func (category FreePlanObjectCategory) Valid() bool {
	return category == FreePlanStructure || category == FreePlanPlatform ||
		category == FreePlanScenery || category == FreePlanAnnotation
}

var ErrInvalidFreePlanObjectShape = errors.New("invalid free plan object shape")

type FreePlanObjectShape struct {
	SchemaVersion int                `json:"schemaVersion"`
	Kind          FreePlanObjectKind `json:"kind"`
	WidthMM       *float64           `json:"widthMm,omitempty"`
	HeightMM      *float64           `json:"heightMm,omitempty"`
	EndXMM        *float64           `json:"endXMm,omitempty"`
	EndYMM        *float64           `json:"endYMm,omitempty"`
	Text          string             `json:"text,omitempty"`
	FontSizeMM    *float64           `json:"fontSizeMm,omitempty"`
}

type PlanFreeObject struct {
	ID              string                 `json:"id"`
	LineageID       string                 `json:"lineageId"`
	RevisionID      string                 `json:"revisionId"`
	Name            string                 `json:"name"`
	Category        FreePlanObjectCategory `json:"category"`
	PositionXMM     float64                `json:"positionXMm"`
	PositionYMM     float64                `json:"positionYMm"`
	RotationDegrees float64                `json:"rotationDegrees"`
	Shape           FreePlanObjectShape    `json:"shape"`
	Version         int                    `json:"version"`
	CreatedAt       string                 `json:"createdAt"`
	UpdatedAt       string                 `json:"updatedAt"`
}

func ValidateFreePlanObjectShape(shape FreePlanObjectShape) error {
	if shape.SchemaVersion != 1 || !shape.Kind.Valid() {
		return ErrInvalidFreePlanObjectShape
	}
	switch shape.Kind {
	case FreePlanRectangle, FreePlanEllipse:
		if !positiveFreeMeasure(shape.WidthMM) || !positiveFreeMeasure(shape.HeightMM) ||
			shape.EndXMM != nil || shape.EndYMM != nil || shape.Text != "" || shape.FontSizeMM != nil {
			return ErrInvalidFreePlanObjectShape
		}
	case FreePlanLine:
		if !finiteFreeMeasure(shape.EndXMM) || !finiteFreeMeasure(shape.EndYMM) ||
			math.Hypot(*shape.EndXMM, *shape.EndYMM) <= 0 || shape.WidthMM != nil ||
			shape.HeightMM != nil || shape.Text != "" || shape.FontSizeMM != nil {
			return ErrInvalidFreePlanObjectShape
		}
	case FreePlanLabel:
		text := strings.TrimSpace(shape.Text)
		if utf8.RuneCountInString(text) < 1 || utf8.RuneCountInString(text) > 120 ||
			!finiteFreeMeasure(shape.FontSizeMM) || *shape.FontSizeMM < 2 || *shape.FontSizeMM > 50 ||
			shape.WidthMM != nil || shape.HeightMM != nil || shape.EndXMM != nil || shape.EndYMM != nil {
			return ErrInvalidFreePlanObjectShape
		}
	}
	return nil
}

func finiteFreeMeasure(value *float64) bool {
	return value != nil && !math.IsNaN(*value) && !math.IsInf(*value, 0)
}

func positiveFreeMeasure(value *float64) bool {
	return finiteFreeMeasure(value) && *value > 0
}
