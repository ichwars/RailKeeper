package valuation

import (
	"context"
	"errors"
)

var (
	ErrValidation = errors.New("valuation invalid")
	ErrNotFound   = errors.New("valuation or vehicle not found")
	ErrConflict   = errors.New("valuation has already been corrected or request identity changed")
)

const HistoryPageSize = 100

type Input struct {
	ID         string `json:"id"`
	Amount     string `json:"amount"`
	Currency   string `json:"currency"`
	ValuedOn   string `json:"valuedOn"`
	Source     string `json:"source"`
	Notes      string `json:"notes"`
	ReplacesID string `json:"replacesId"`
	Withdrawn  bool   `json:"withdrawn"`
}

// Corrections are new immutable rows, retaining the original recording time/data.
// The unique predecessor is also the concurrency token; corrections cannot fork.
type Record struct {
	Input
	VehicleID    string `json:"vehicleId"`
	RecordedAt   string `json:"recordedAt"`
	ReplacedByID string `json:"replacedById"`
}

func (record Record) SameContent(other Record) bool {
	return record.Input == other.Input && record.VehicleID == other.VehicleID
}

type History struct {
	Entries []Record `json:"entries"`
	Total   int      `json:"total"`
}

type Snapshot struct {
	VehicleCount int
	Entries      []Record
}

type CurrencyTotal struct {
	Currency string `json:"currency"`
	Amount   string `json:"amount"`
	Vehicles int    `json:"vehicles"`
}

type Summary struct {
	AsOf            string          `json:"asOf"`
	VehicleCount    int             `json:"vehicleCount"`
	ValuedVehicles  int             `json:"valuedVehicles"`
	MissingVehicles int             `json:"missingVehicles"`
	Totals          []CurrencyTotal `json:"totals"`
}

// Save owns one transaction for the predecessor check, immutable row and audit.
// The adapter returns the original record for an identical replay of its ID.
type Store interface {
	History(context.Context, string, int) (History, error)
	Export(context.Context, string) (History, error)
	Save(context.Context, Record, string) (Record, error)
	Snapshot(context.Context, string) (Snapshot, error)
}
