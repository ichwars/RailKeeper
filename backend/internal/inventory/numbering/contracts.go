// Package numbering owns numbering rules without concrete persistence.
package numbering

import (
	"context"
	"errors"
)

var (
	ErrValidation = errors.New("inventory number validation failed")
	ErrConflict   = errors.New("inventory number already exists")
	ErrNotFound   = errors.New("inventory number scheme not found")
)

type Scheme struct {
	ID         string `json:"id"`
	Category   string `json:"category"`
	Prefix     string `json:"prefix"`
	NextNumber int    `json:"nextNumber"`
	Padding    int    `json:"padding"`
	Active     bool   `json:"active"`
	Preview    string `json:"preview"`
	CreatedAt  string `json:"createdAt"`
	UpdatedAt  string `json:"updatedAt"`
}

type SchemeInput struct {
	Prefix     string `json:"prefix"`
	NextNumber int    `json:"nextNumber"`
	Padding    int    `json:"padding"`
	Active     bool   `json:"active"`
}

// InventoryNumberSchemeInput retains the embedded field name used by existing callers.
type InventoryNumberSchemeInput = SchemeInput

type CreateInput struct {
	Category string `json:"category"`
	InventoryNumberSchemeInput
}

type Availability func(string) error
type Schemes interface {
	List(context.Context) ([]Scheme, error)
	Create(context.Context, string, SchemeInput) (*Scheme, error)
	Update(context.Context, string, SchemeInput) (*Scheme, error)
}

// Counter is supplied by the owner of the caller's inventory mutation transaction.
type Counter interface {
	Active(context.Context, string) (*Scheme, error)
	Advance(context.Context, string, int) error
}
