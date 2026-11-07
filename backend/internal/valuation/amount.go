// Package valuation owns dated manual values, independently of HTTP and storage.
package valuation

import (
	"math/big"
	"strconv"
	"strings"
)

const amountScale int64 = 1_000_000

// ParseAmount accepts an ungrouped decimal with at most six fractional digits.
// Both separators mean a decimal point; grouping is never interpreted.
// Fixed units preserve precision without assuming that every currency has cents.
func ParseAmount(value string) (int64, bool) {
	value = strings.ReplaceAll(strings.TrimSpace(value), ",", ".")
	parts := strings.Split(value, ".")
	if len(parts) > 2 || len(parts[0]) == 0 || len(parts[0]) > 12 {
		return 0, false
	}
	for _, part := range parts {
		if part == "" {
			return 0, false
		}
		for _, char := range part {
			if char < '0' || char > '9' {
				return 0, false
			}
		}
	}
	major, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, false
	}
	var fraction int64
	if len(parts) == 2 {
		if len(parts[1]) > 6 {
			return 0, false
		}
		fraction, err = strconv.ParseInt(parts[1]+strings.Repeat("0", 6-len(parts[1])), 10, 64)
		if err != nil {
			return 0, false
		}
	}
	return major*amountScale + fraction, true
}

func formatAmount(units *big.Int) string {
	major, fraction := new(big.Int), new(big.Int)
	major.QuoRem(units, big.NewInt(amountScale), fraction)
	digits := fraction.String()
	digits = strings.Repeat("0", 6-len(digits)) + digits
	digits = strings.TrimRight(digits, "0")
	if len(digits) < 2 {
		digits += strings.Repeat("0", 2-len(digits))
	}
	return major.String() + "." + digits
}
