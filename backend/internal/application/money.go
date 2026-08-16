package application

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

func parseMoneyCents(value string) (int64, bool) {
	value = strings.TrimSpace(value)
	if strings.HasPrefix(value, "+") {
		value = strings.TrimPrefix(value, "+")
	}
	if value == "" {
		return 0, false
	}
	for _, character := range value {
		if (character < '0' || character > '9') && character != '.' && character != ',' {
			return 0, false
		}
	}
	if strings.HasPrefix(value, ".") || strings.HasPrefix(value, ",") ||
		strings.HasSuffix(value, ".") || strings.HasSuffix(value, ",") {
		return 0, false
	}

	integerDigits, fractionDigits, ok := splitMoneyParts(value)
	if !ok {
		return 0, false
	}
	major, err := strconv.ParseInt(integerDigits, 10, 64)
	if err != nil {
		return 0, false
	}
	fraction := int64(0)
	if fractionDigits != "" {
		if len(fractionDigits) == 1 {
			fractionDigits += "0"
		}
		fraction, err = strconv.ParseInt(fractionDigits, 10, 64)
		if err != nil {
			return 0, false
		}
	}
	if major > (math.MaxInt64-fraction)/100 {
		return 0, false
	}
	return major*100 + fraction, true
}

func splitMoneyParts(value string) (string, string, bool) {
	dots := strings.Count(value, ".")
	commas := strings.Count(value, ",")
	if dots == 0 && commas == 0 {
		return value, "", digitsOnly(value)
	}
	if dots > 0 && commas > 0 {
		decimalSeparator := "."
		groupingSeparator := ","
		if strings.LastIndex(value, ",") > strings.LastIndex(value, ".") {
			decimalSeparator, groupingSeparator = ",", "."
		}
		parts := strings.Split(value, decimalSeparator)
		if len(parts) != 2 || len(parts[1]) < 1 || len(parts[1]) > 2 || !digitsOnly(parts[1]) {
			return "", "", false
		}
		integerDigits, ok := groupedMoneyDigits(parts[0], groupingSeparator)
		return integerDigits, parts[1], ok
	}

	separator := "."
	if commas > 0 {
		separator = ","
	}
	parts := strings.Split(value, separator)
	if len(parts) == 2 && (len(parts[1]) == 1 || len(parts[1]) == 2) &&
		digitsOnly(parts[0]) && digitsOnly(parts[1]) {
		return parts[0], parts[1], true
	}
	integerDigits, ok := groupedMoneyParts(parts)
	return integerDigits, "", ok
}

func groupedMoneyDigits(value, separator string) (string, bool) {
	return groupedMoneyParts(strings.Split(value, separator))
}

func groupedMoneyParts(parts []string) (string, bool) {
	if len(parts) < 2 || len(parts[0]) < 1 || len(parts[0]) > 3 || !digitsOnly(parts[0]) {
		return "", false
	}
	for _, part := range parts[1:] {
		if len(part) != 3 || !digitsOnly(part) {
			return "", false
		}
	}
	return strings.Join(parts, ""), true
}

func digitsOnly(value string) bool {
	if value == "" {
		return false
	}
	for _, character := range value {
		if character < '0' || character > '9' {
			return false
		}
	}
	return true
}

func formatMoneyCents(cents int64) string {
	return fmt.Sprintf("%d.%02d", cents/100, cents%100)
}
