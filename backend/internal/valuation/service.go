package valuation

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"sort"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

const recordedTimeFormat = "2006-01-02T15:04:05.000000000Z"

type Service struct {
	store Store
	now   func() time.Time
}

func New(store Store, now func() time.Time) *Service {
	if now == nil {
		now = time.Now
	}
	return &Service{store: store, now: now}
}

func ValidDate(value string) bool {
	date, err := time.Parse(time.DateOnly, value)
	return err == nil && date.Year() > 0 && date.Format(time.DateOnly) == value
}

func validID(value string) bool {
	if len(value) != 32 || strings.ToLower(value) != value {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}

func Normalize(input Input) (Input, error) {
	input.Source = strings.TrimSpace(input.Source)
	input.Notes = strings.TrimSpace(input.Notes)
	input.Currency = strings.ToUpper(strings.TrimSpace(input.Currency))
	units, ok := ParseAmount(input.Amount)
	if !ok || !validID(input.ID) || (input.ReplacesID != "" && !validID(input.ReplacesID)) ||
		input.ReplacesID == input.ID || (input.Withdrawn && input.ReplacesID == "") || !ValidDate(input.ValuedOn) {
		return Input{}, ErrValidation
	}
	if len(input.Currency) != 3 || strings.IndexFunc(input.Currency, func(r rune) bool { return r < 'A' || r > 'Z' }) >= 0 {
		return Input{}, ErrValidation
	}
	for _, field := range []struct {
		value string
		limit int
	}{{input.Source, 1000}, {input.Notes, 4000}} {
		if !utf8.ValidString(field.value) || utf8.RuneCountInString(field.value) > field.limit ||
			strings.IndexFunc(field.value, func(r rune) bool { return unicode.IsControl(r) && r != '\n' && r != '\t' }) >= 0 {
			return Input{}, ErrValidation
		}
	}
	if input.Source == "" {
		return Input{}, ErrValidation
	}
	input.Amount = formatAmount(big.NewInt(units))
	return input, nil
}

func ValidateRecord(record Record) error {
	input, err := Normalize(record.Input)
	recorded, timeErr := time.Parse(recordedTimeFormat, record.RecordedAt)
	if err != nil || input != record.Input || record.VehicleID == "" || timeErr != nil ||
		recorded.Year() < 1 || recorded.Format(recordedTimeFormat) != record.RecordedAt ||
		(record.ReplacedByID != "" && (!validID(record.ReplacedByID) || record.ReplacedByID == record.ID)) {
		return ErrValidation
	}
	return nil
}

func (service *Service) Record(ctx context.Context, vehicleID string, input Input, actor string) (Record, error) {
	input, err := Normalize(input)
	if err != nil || vehicleID == "" {
		return Record{}, ErrValidation
	}
	record := Record{Input: input, VehicleID: vehicleID, RecordedAt: service.now().UTC().Format(recordedTimeFormat)}
	return service.store.Save(ctx, record, actor)
}

func (service *Service) History(ctx context.Context, vehicleID string, offset int) (History, error) {
	if offset < 0 || offset > 1_000_000 || vehicleID == "" {
		return History{}, ErrValidation
	}
	return service.store.History(ctx, vehicleID, offset)
}

type ExportDocument struct {
	Format     string   `json:"format"`
	Version    int      `json:"version"`
	VehicleID  string   `json:"vehicleId"`
	ExportedAt string   `json:"exportedAt"`
	Entries    []Record `json:"entries"`
}

func (service *Service) Export(ctx context.Context, vehicleID string) (ExportDocument, error) {
	if vehicleID == "" {
		return ExportDocument{}, ErrValidation
	}
	history, err := service.store.Export(ctx, vehicleID)
	if err != nil {
		return ExportDocument{}, err
	}
	return ExportDocument{Format: "railkeeper-valuations", Version: 1, VehicleID: vehicleID,
		ExportedAt: service.now().UTC().Format(recordedTimeFormat), Entries: history.Entries}, nil
}

func (service *Service) Summary(ctx context.Context, asOf string) (Summary, error) {
	if asOf == "" {
		asOf = service.now().UTC().Format(time.DateOnly)
	}
	if !ValidDate(asOf) {
		return Summary{}, ErrValidation
	}
	snapshot, err := service.store.Snapshot(ctx, asOf)
	if err != nil {
		return Summary{}, err
	}
	return Summarize(snapshot, asOf)
}

// Summarize chooses one current leaf per vehicle at the valuation date. A later
// recorded correction represents present knowledge, never a fabricated old quote.
func Summarize(snapshot Snapshot, asOf string) (Summary, error) {
	if !ValidDate(asOf) || snapshot.VehicleCount < 0 {
		return Summary{}, ErrValidation
	}
	latest := map[string]Record{}
	for _, entry := range snapshot.Entries {
		if err := ValidateRecord(entry); err != nil {
			return Summary{}, fmt.Errorf("stored valuation: %w", err)
		}
		if entry.ReplacedByID != "" || entry.Withdrawn || entry.ValuedOn > asOf {
			continue
		}
		old, exists := latest[entry.VehicleID]
		if !exists || entry.ValuedOn > old.ValuedOn || (entry.ValuedOn == old.ValuedOn &&
			(entry.RecordedAt > old.RecordedAt || (entry.RecordedAt == old.RecordedAt && entry.ID > old.ID))) {
			latest[entry.VehicleID] = entry
		}
	}
	if len(latest) > snapshot.VehicleCount {
		return Summary{}, fmt.Errorf("valuation coverage: %w", ErrValidation)
	}
	sums, counts := map[string]*big.Int{}, map[string]int{}
	for _, entry := range latest {
		units, _ := ParseAmount(entry.Amount) // Validated above.
		if sums[entry.Currency] == nil {
			sums[entry.Currency] = new(big.Int)
		}
		sums[entry.Currency].Add(sums[entry.Currency], big.NewInt(units))
		counts[entry.Currency]++
	}
	result := Summary{AsOf: asOf, VehicleCount: snapshot.VehicleCount, ValuedVehicles: len(latest),
		MissingVehicles: snapshot.VehicleCount - len(latest), Totals: []CurrencyTotal{}}
	for currency, units := range sums {
		result.Totals = append(result.Totals, CurrencyTotal{Currency: currency, Amount: formatAmount(units), Vehicles: counts[currency]})
	}
	sort.Slice(result.Totals, func(i, j int) bool { return result.Totals[i].Currency < result.Totals[j].Currency })
	return result, nil
}
