package numbering

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

type counterStub struct {
	schemes  map[string]Scheme
	lookups  []string
	advanced int
	next     int
	failure  error
}

func (s *counterStub) Active(_ context.Context, category string) (*Scheme, error) {
	s.lookups = append(s.lookups, category)
	if scheme, ok := s.schemes[category]; ok {
		return &scheme, nil
	}
	return nil, ErrNotFound
}
func (s *counterStub) Advance(_ context.Context, category string, next int) error {
	if s.failure != nil {
		return s.failure
	}
	s.advanced++
	s.next = next
	return nil
}

func TestReservationKeepsRecordedSchemeAndSkipsConflicts(t *testing.T) {
	counter := &counterStub{schemes: map[string]Scheme{
		"Custom": {ID: "recorded", Category: "Custom", Prefix: "LEGACY", NextNumber: 17, Padding: 4},
	}}
	attempts := 0
	value, err := Reserve(t.Context(), counter, " Custom ", "Fahrzeug", func(candidate string) error {
		attempts++
		if candidate == "LEGACY-0019" {
			return nil
		}
		return ErrConflict
	})
	if err != nil || value != "LEGACY-0019" || attempts != 3 || counter.advanced != 1 || counter.next != 20 ||
		!reflect.DeepEqual(counter.lookups, []string{"Custom"}) {
		t.Fatal("reservation changed numbering identity or consumed extra numbers", value, err, counter)
	}
}

func TestReservationKeepsLegacyFallbackOrder(t *testing.T) {
	for _, test := range []struct {
		category, fallback, found string
		lookups                   []string
	}{
		{"Custom", "Lokomotive", "Lokomotive", []string{"Custom", "Lokomotive"}},
		{"Custom", "Lokomotive", "Fahrzeug", []string{"Custom", "Lokomotive", "Fahrzeug"}},
		{"Custom", "", "Fahrzeug", []string{"Custom"}},
		{"Fahrzeug", "Fahrzeug", "", []string{"Fahrzeug"}},
	} {
		t.Run(test.category+"/"+test.fallback+"/"+test.found, func(t *testing.T) {
			counter := &counterStub{schemes: map[string]Scheme{test.found: {Category: test.found, Prefix: "RK", NextNumber: 1, Padding: 6}}}
			value, err := Reserve(t.Context(), counter, test.category, test.fallback, func(string) error { return nil })
			found := len(test.lookups) > 0 && test.lookups[len(test.lookups)-1] == test.found
			if found && (err != nil || value != "RK-000001") || !found && (!errors.Is(err, ErrNotFound) || value != "") ||
				!reflect.DeepEqual(counter.lookups, test.lookups) {
				t.Fatal("fallback order changed", value, err, counter.lookups)
			}
		})
	}
}

func TestFailedReservationDoesNotAcknowledgeOrAdvance(t *testing.T) {
	for _, stage := range []string{"availability", "advance", "exhausted"} {
		t.Run(stage, func(t *testing.T) {
			sentinel := errors.New("synthetic counter failure")
			counter := &counterStub{schemes: map[string]Scheme{"Fahrzeug": {Category: "Fahrzeug", Prefix: "RK", NextNumber: 1, Padding: 6}}}
			if stage == "advance" {
				counter.failure = sentinel
			}
			attempts := 0
			value, err := Reserve(t.Context(), counter, "Fahrzeug", "", func(string) error {
				attempts++
				if stage == "availability" {
					return sentinel
				}
				if stage == "exhausted" {
					return ErrConflict
				}
				return nil
			})
			if value != "" || err == nil || counter.advanced != 0 || counter.next != 0 {
				t.Fatal("failed reservation was acknowledged or consumed a counter", value, err, counter)
			}
			if stage == "exhausted" && attempts != 500 || stage != "exhausted" && !errors.Is(err, sentinel) {
				t.Fatal("failure cause or bounded search changed", attempts, err)
			}
		})
	}
}
