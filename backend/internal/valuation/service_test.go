package valuation

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"testing"
	"time"
)

func example(id int, vehicle, date, amount, currency string) Record {
	return Record{Input: Input{ID: fmt.Sprintf("%032x", id), Amount: amount, Currency: currency,
		ValuedOn: date, Source: "Manual estimate"}, VehicleID: vehicle, RecordedAt: "2026-11-07T12:00:00.000000000Z"}
}

func TestExactAmount(t *testing.T) {
	for _, pair := range [][2]string{{"0", "0.00"}, {" 0001,000001 ", "1.000001"}, {"100.5", "100.50"},
		{"12.345", "12.345"}, {"999999999999.999999", "999999999999.999999"}} {
		units, ok := ParseAmount(pair[0])
		if !ok || formatAmount(big.NewInt(units)) != pair[1] {
			t.Fatalf("%q: %d, %v", pair[0], units, ok)
		}
	}
	for _, value := range []string{"", "-1", "+1", "1e3", ".1", "1.", "1,000.00", "1 000", "1.1234567",
		"1000000000000", "١٢", "1\x002"} {
		if _, ok := ParseAmount(value); ok {
			t.Errorf("accepted invalid amount %q", value)
		}
	}
}

func TestInputAndRecordingValidation(t *testing.T) {
	base := example(1, "vehicle", "2024-02-29", "10.00", "EUR")
	input := base.Input
	input.Amount, input.Currency, input.Source = " 0010,00 ", " eur ", " Local estimate "
	got, err := Normalize(input)
	if err != nil || got.Amount != "10.00" || got.Currency != "EUR" || got.Source != "Local estimate" {
		t.Fatalf("normalization: %+v, %v", got, err)
	}
	for _, mutate := range []func(*Input){
		func(i *Input) { i.ID = "ABC" }, func(i *Input) { i.ReplacesID = i.ID },
		func(i *Input) { i.ReplacesID = "invalid" }, func(i *Input) { i.Withdrawn = true },
		func(i *Input) { i.ValuedOn = "2026-02-29" }, func(i *Input) { i.Currency = "EU1" },
		func(i *Input) { i.Currency = "EURO" }, func(i *Input) { i.Source = "   " },
		func(i *Input) { i.Source = strings.Repeat("ä", 1001) }, func(i *Input) { i.Notes = "bad\x00note" },
		func(i *Input) { i.Notes = strings.Repeat("x", 4001) }, func(i *Input) { i.Notes = string([]byte{255}) },
	} {
		input := base.Input
		mutate(&input)
		if _, err := Normalize(input); !errors.Is(err, ErrValidation) {
			t.Fatalf("accepted invalid input: %+v", input)
		}
	}
	if ValidateRecord(base) != nil {
		t.Fatal("valid record rejected")
	}
	for _, mutate := range []func(*Record){
		func(r *Record) { r.RecordedAt = "2026-11-07" }, func(r *Record) { r.Amount = "10" },
		func(r *Record) { r.VehicleID = "" }, func(r *Record) { r.ReplacedByID = r.ID },
		func(r *Record) { r.ReplacedByID = "bad" },
	} {
		record := base
		mutate(&record)
		if !errors.Is(ValidateRecord(record), ErrValidation) {
			t.Fatal("invalid record accepted")
		}
	}
	for _, date := range []string{"0000-01-01", "2026-13-01", "2026-1-01", "2026-04-31"} {
		if ValidDate(date) {
			t.Errorf("invalid date %s", date)
		}
	}
}

func TestSummaryDatesCorrectionsCoverageAndCurrencies(t *testing.T) {
	first := example(1, "a", "2026-01-01", "100.00", "EUR")
	corrected := example(2, "a", "2026-01-01", "125.123456", "EUR")
	first.ReplacedByID, corrected.ReplacesID = corrected.ID, first.ID
	future := example(3, "a", "2027-01-01", "900.00", "USD")
	b := example(4, "b", "2026-02-01", "50.00", "CHF")
	withdrawn := example(5, "c", "2026-02-01", "300.00", "EUR")
	withdrawn.Withdrawn, withdrawn.ReplacesID = true, fmt.Sprintf("%032x", 6)
	snapshot := Snapshot{VehicleCount: 4, Entries: []Record{first, corrected, future, b, withdrawn}}
	summary, err := Summarize(snapshot, "2026-03-01")
	if err != nil || summary.ValuedVehicles != 2 || summary.MissingVehicles != 2 || len(summary.Totals) != 2 ||
		summary.Totals[0].Currency != "CHF" || summary.Totals[0].Amount != "50.00" || summary.Totals[1].Amount != "125.123456" {
		t.Fatalf("summary: %+v, %v", summary, err)
	}
	before, err := Summarize(snapshot, "2025-12-31")
	if err != nil || before.ValuedVehicles != 0 || len(before.Totals) != 0 {
		t.Fatalf("fabricated past value: %+v", before)
	}
	later, err := Summarize(snapshot, "2027-01-01")
	if err != nil || later.Totals[1].Currency != "USD" || later.Totals[1].Amount != "900.00" {
		t.Fatalf("currency change: %+v", later)
	}
}

func TestSummaryTieAndNoSumOverflow(t *testing.T) {
	var entries []Record
	for i := range 20 {
		entries = append(entries, example(i+1, fmt.Sprint(i), "2026-01-01", "999999999999.999999", "EUR"))
	}
	got, err := Summarize(Snapshot{VehicleCount: 20, Entries: entries}, "2026-01-01")
	if err != nil || got.Totals[0].Amount != "19999999999999.99998" {
		t.Fatalf("exact large sum: %+v %v", got, err)
	}
	later := entries[0]
	later.ID, later.Amount, later.RecordedAt = fmt.Sprintf("%032x", 100), "1.00", "2026-11-07T12:00:00.000000001Z"
	got, err = Summarize(Snapshot{VehicleCount: 1, Entries: []Record{later, entries[0]}}, "2026-01-01")
	if err != nil || got.Totals[0].Amount != "1.00" {
		t.Fatal("same-date order incorrect")
	}
	if _, err := Summarize(Snapshot{Entries: entries}, "2026-01-01"); !errors.Is(err, ErrValidation) {
		t.Fatal("bad coverage accepted")
	}
	if _, err := Summarize(Snapshot{}, "bad"); !errors.Is(err, ErrValidation) {
		t.Fatal("bad date accepted")
	}
}

type memoryStore struct {
	saved Record
	actor string
	err   error
}

func (s *memoryStore) Export(ctx context.Context, id string) (History, error) {
	return s.History(ctx, id, 0)
}

func (s *memoryStore) Save(_ context.Context, r Record, actor string) (Record, error) {
	s.saved, s.actor = r, actor
	return r, s.err
}
func (s *memoryStore) History(context.Context, string, int) (History, error) {
	return History{Entries: []Record{s.saved}, Total: 1}, s.err
}
func (s *memoryStore) Snapshot(context.Context, string) (Snapshot, error) {
	return Snapshot{VehicleCount: 1, Entries: []Record{s.saved}}, s.err
}

func TestServiceRecordingClockAndStoreFailures(t *testing.T) {
	store := &memoryStore{}
	now := time.Date(2026, 11, 7, 23, 5, 0, 0, time.FixedZone("local", 3600))
	service := New(store, func() time.Time { return now })
	r, err := service.Record(t.Context(), "a", example(1, "a", "2026-10-01", "1.00", "EUR").Input, "actor")
	if err != nil || r.RecordedAt != "2026-11-07T22:05:00.000000000Z" || store.actor != "actor" {
		t.Fatal("recording context lost")
	}
	if !r.SameContent(store.saved) {
		t.Fatal("same request differs")
	}
	if _, err := service.History(t.Context(), "a", 0); err != nil {
		t.Fatal(err)
	}
	if result, err := service.Summary(t.Context(), ""); err != nil || result.AsOf != "2026-11-07" {
		t.Fatalf("clock: %+v %v", result, err)
	}
	for _, offset := range []int{-1, 1_000_001} {
		if _, err := service.History(t.Context(), "a", offset); !errors.Is(err, ErrValidation) {
			t.Fatal("offset accepted")
		}
	}
	if _, err := service.Record(t.Context(), "", r.Input, "actor"); !errors.Is(err, ErrValidation) {
		t.Fatal("empty vehicle accepted")
	}
	store.err = errors.New("offline")
	if _, err := service.Record(t.Context(), "a", r.Input, "actor"); !errors.Is(err, store.err) {
		t.Fatal("store error lost")
	}
	if _, err := service.Summary(t.Context(), "2026-11-07"); !errors.Is(err, store.err) {
		t.Fatal("snapshot error lost")
	}
	if _, err := service.Summary(t.Context(), "invalid"); !errors.Is(err, ErrValidation) {
		t.Fatal("invalid summary date accepted")
	}
}

func FuzzAmountRoundTrip(f *testing.F) {
	for _, input := range []string{"0", "12,345", "999999999999.999999", "1e3", "1,000.00"} {
		f.Add(input)
	}
	f.Fuzz(func(t *testing.T, input string) {
		units, ok := ParseAmount(input)
		if !ok {
			return
		}
		next, accepted := ParseAmount(formatAmount(big.NewInt(units)))
		if units < 0 || !accepted || next != units {
			t.Fatalf("unstable amount: %q", input)
		}
	})
}
