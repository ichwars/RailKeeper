package workshop

import (
	"strconv"
	"strings"
	"time"
)

var allowedMaintenanceKinds = map[string]struct{}{
	"Decoder-Einbau":   {},
	"Ersatzteiltausch": {},
	"Reinigung":        {},
	"Reparatur":        {},
	"Schmierung":       {},
	"Superung":         {},
	"Umbau":            {},
	"Wartung":          {},
}
var allowedMaintenanceStatuses = map[string]struct{}{
	"erledigt": {},
	"faellig":  {},
	"geplant":  {},
}
var allowedConditionRatings = map[string]struct{}{
	"gebraucht":          {},
	"gut":                {},
	"neuwertig":          {},
	"reparaturbedürftig": {},
	"sehr gut":           {},
}

func cleanVehicleMaintenanceInput(input VehicleMaintenanceInput) VehicleMaintenanceInput {
	input.Kind = strings.TrimSpace(input.Kind)
	input.Status = normalizeMaintenanceStatus(input.Status)
	input.ConditionRating = strings.TrimSpace(input.ConditionRating)
	input.DueDate = strings.TrimSpace(input.DueDate)
	input.CompletedAt = strings.TrimSpace(input.CompletedAt)
	input.Cost = cleanMaintenanceCost(input.Cost)
	input.Notes = strings.TrimSpace(input.Notes)
	if input.Kind == "" {
		input.Kind = "Wartung"
	}
	if input.Status == "" {
		input.Status = "geplant"
	}
	return input
}
func normalizeMaintenanceStatus(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	switch value {
	case "fällig", "faellig":
		return "faellig"
	case "erledigt":
		return "erledigt"
	case "geplant", "":
		return value
	default:
		return value
	}
}
func cleanMaintenanceCost(value string) string {
	value = strings.TrimSpace(value)
	value = strings.TrimSuffix(value, "€")
	value = strings.TrimSpace(value)
	value = strings.ReplaceAll(value, " ", "")
	return value
}
func cleanVehicleSparePartInput(input VehicleSparePartInput) VehicleSparePartInput {
	input.ArticleNumber = strings.TrimSpace(input.ArticleNumber)
	input.Description = strings.TrimSpace(input.Description)
	input.Price = strings.TrimSpace(input.Price)
	input.URL = strings.TrimSpace(input.URL)
	return input
}
func vehicleSparePartDedupKey(articleNumber, description, link string) string {
	article := strings.ToLower(strings.TrimSpace(articleNumber))
	article = strings.TrimPrefix(article, "et")
	article = strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, article)
	if article != "" {
		return "article:" + article
	}
	link = strings.TrimRight(strings.ToLower(strings.TrimSpace(link)), "/")
	if link != "" {
		return "url:" + link
	}
	description = strings.ToLower(strings.TrimSpace(description))
	if description != "" {
		return "description:" + description
	}
	return ""
}
func isValidVehicleSparePartInput(input VehicleSparePartInput) bool {
	return input.ArticleNumber != "" || input.Description != "" || input.URL != ""
}
func isValidVehicleMaintenanceInput(input VehicleMaintenanceInput) bool {
	if _, ok := allowedMaintenanceKinds[input.Kind]; !ok {
		return false
	}
	if _, ok := allowedMaintenanceStatuses[input.Status]; !ok {
		return false
	}
	if input.ConditionRating != "" {
		if _, ok := allowedConditionRatings[input.ConditionRating]; !ok {
			return false
		}
	}
	return isValidDateOnly(input.DueDate) &&
		isValidDateOnly(input.CompletedAt) &&
		isValidMaintenanceCost(input.Cost) &&
		len(input.Notes) <= 4000
}
func isValidDateOnly(value string) bool {
	if value == "" {
		return true
	}
	parsed, err := time.Parse("2006-01-02", value)
	return err == nil && parsed.Format("2006-01-02") == value
}
func isValidMaintenanceCost(value string) bool {
	if value == "" {
		return true
	}
	normalized := strings.ReplaceAll(value, ",", ".")
	amount, err := strconv.ParseFloat(normalized, 64)
	return err == nil && amount >= 0
}
