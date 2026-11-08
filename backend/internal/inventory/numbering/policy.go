package numbering

import (
	"fmt"
	"strings"
)

func CleanSchemeInput(input SchemeInput) SchemeInput {
	input.Prefix = strings.ToUpper(strings.TrimSpace(input.Prefix))
	input.Prefix = strings.ReplaceAll(input.Prefix, " ", "-")
	if input.Padding == 0 {
		input.Padding = 6
	}
	if input.NextNumber == 0 {
		input.NextNumber = 1
	}
	return input
}

func CleanCategory(category string) string {
	category = strings.TrimSpace(category)
	switch strings.ToLower(category) {
	case "fahrzeug", "lokomotive", "wagen":
		return category
	default:
		return category
	}
}

func CategoryForVehicle(category string) string {
	value := strings.ToLower(strings.TrimSpace(category))
	switch {
	case strings.Contains(value, "lok"):
		return "Lokomotive"
	case strings.Contains(value, "wagen"), strings.Contains(value, "waggon"):
		return "Wagen"
	default:
		return "Fahrzeug"
	}
}

func Format(prefix string, number int, padding int) string {
	if padding < 1 {
		padding = 6
	}
	if number < 1 {
		number = 1
	}
	return fmt.Sprintf("%s-%0*d", prefix, padding, number)
}
