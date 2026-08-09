package infrastructure

import (
	"errors"
	"strings"
	"testing"
)

type failingAccessoryFilterRows struct {
	err error
}

func (rows *failingAccessoryFilterRows) Next() bool        { return false }
func (rows *failingAccessoryFilterRows) Scan(...any) error { return nil }
func (rows *failingAccessoryFilterRows) Err() error        { return rows.err }
func (rows *failingAccessoryFilterRows) Close() error      { return nil }

func TestAppendAccessoryStringFilterOptionsWrapsIterationError(t *testing.T) {
	want := errors.New("iteration failed")
	values := []string{}
	err := appendAccessoryStringFilterOptions(
		&failingAccessoryFilterRows{err: want}, "accessory manufacturer filters", &values,
	)
	if !errors.Is(err, want) || !strings.Contains(err.Error(), "iterate accessory manufacturer filters") {
		t.Fatalf("unexpected filter iteration error: %v", err)
	}
}
