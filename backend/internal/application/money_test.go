package application

import (
	"math"
	"testing"
)

func TestParseMoneyCents(t *testing.T) {
	tests := []struct {
		input string
		cents int64
		ok    bool
	}{
		{"", 0, false},
		{"0", 0, true},
		{"129.90", 12990, true},
		{"129,90", 12990, true},
		{"1.299,90", 129990, true},
		{"1,299.90", 129990, true},
		{"1.299", 129900, true},
		{"-1.00", 0, false},
		{"12.345", 1234500, true},
		{"12.3456", 0, false},
		{"abc", 0, false},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			cents, ok := parseMoneyCents(test.input)
			if cents != test.cents || ok != test.ok {
				t.Fatalf("parseMoneyCents(%q) = (%d, %t), want (%d, %t)",
					test.input, cents, ok, test.cents, test.ok)
			}
		})
	}
}

func TestFormatMoneyCents(t *testing.T) {
	if got := formatMoneyCents(0); got != "0.00" {
		t.Fatalf("formatMoneyCents(0) = %q", got)
	}
	if got := formatMoneyCents(129990); got != "1299.90" {
		t.Fatalf("formatMoneyCents(129990) = %q", got)
	}
}

func TestMoneyArithmeticRejectsOverflow(t *testing.T) {
	if _, ok := parseMoneyCents("92233720368547758.08"); ok {
		t.Fatal("expected parser overflow to be rejected")
	}
	if _, err := checkedMoneyMultiply(math.MaxInt64, 2); err == nil {
		t.Fatal("expected multiplication overflow")
	}
	if _, err := checkedMoneyAdd(math.MaxInt64, 1); err == nil {
		t.Fatal("expected addition overflow")
	}
}
