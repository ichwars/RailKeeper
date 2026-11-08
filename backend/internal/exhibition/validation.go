package exhibition

import (
	"fmt"
	"strings"
	"time"
)

func normalizeExhibitionEntryInput(input ExhibitionEntryInput) ExhibitionEntryInput {
	input.VehicleID = strings.TrimSpace(input.VehicleID)
	input.Owner = strings.TrimSpace(input.Owner)
	input.ImageURL = strings.TrimSpace(input.ImageURL)
	input.LocomotiveName = strings.TrimSpace(input.LocomotiveName)
	input.Gattung = strings.TrimSpace(input.Gattung)
	input.Series = strings.TrimSpace(input.Series)
	input.Manufacturer = strings.TrimSpace(input.Manufacturer)
	input.Epoch = strings.TrimSpace(input.Epoch)
	input.RailwayCompany = strings.TrimSpace(input.RailwayCompany)
	input.DayScope = normalizeExhibitionDayScope(input.DayScope)
	input.DecoderNumber = strings.TrimSpace(input.DecoderNumber)
	input.DecoderType = strings.TrimSpace(input.DecoderType)
	input.Adapter = strings.TrimSpace(input.Adapter)
	input.InterfaceName = strings.TrimSpace(input.InterfaceName)
	if input.InterfaceName == "" {
		input.InterfaceName = input.Adapter
	}
	input.SXAddress = strings.TrimSpace(input.SXAddress)
	input.Availability = strings.TrimSpace(input.Availability)
	if input.Availability == "" {
		input.Availability = "available"
	}
	if input.Availability != "available" && input.Availability != "unavailable" {
		input.Availability = "available"
	}
	input.FunctionKeys = strings.TrimSpace(input.FunctionKeys)
	input.Notes = strings.TrimSpace(input.Notes)
	return input
}
func normalizeExhibitionDayScope(value string) string {
	raw := strings.Split(strings.TrimSpace(value), ",")
	seen := map[string]bool{}
	for _, part := range raw {
		scope := strings.TrimSpace(part)
		if scope == "all" {
			return "all"
		}
		if isExhibitionDayScope(scope) {
			seen[scope] = true
		}
	}
	if len(seen) == 0 {
		return "all"
	}
	selected := make([]string, 0, len(seen))
	for day := 1; day <= 31; day++ {
		scope := fmt.Sprintf("day%d", day)
		if seen[scope] {
			selected = append(selected, scope)
		}
	}
	return strings.Join(selected, ",")
}
func isExhibitionDayScope(scope string) bool {
	if !strings.HasPrefix(scope, "day") {
		return false
	}
	var day int
	if _, err := fmt.Sscanf(scope, "day%d", &day); err != nil {
		return false
	}
	return day >= 1 && day <= 31 && scope == fmt.Sprintf("day%d", day)
}
func normalizeExhibitionListInput(input ExhibitionListInput) (ExhibitionListInput, error) {
	input.Designation = strings.TrimSpace(input.Designation)
	input.Date = strings.TrimSpace(input.Date)
	input.EndDate = strings.TrimSpace(input.EndDate)
	input.Location = strings.TrimSpace(input.Location)
	input.Description = strings.TrimSpace(input.Description)
	input.OrganizationNotes = strings.TrimSpace(input.OrganizationNotes)
	if input.EndDate == "" {
		input.EndDate = input.Date
	}
	if input.Status == "" {
		input.Status = ExhibitionStatusOpen
	}
	if input.Designation == "" || !validExhibitionStatus(input.Status) {
		return ExhibitionListInput{}, ErrExhibitionValidation
	}
	start, err := time.Parse("2006-01-02", input.Date)
	if err != nil {
		return ExhibitionListInput{}, ErrExhibitionValidation
	}
	end, err := time.Parse("2006-01-02", input.EndDate)
	if err != nil || end.Before(start) || end.Sub(start) > 30*24*time.Hour {
		return ExhibitionListInput{}, ErrExhibitionValidation
	}
	return input, nil
}
func validExhibitionStatus(status ExhibitionStatus) bool {
	switch status {
	case ExhibitionStatusDraft, ExhibitionStatusOpen, ExhibitionStatusLocked,
		ExhibitionStatusRunning, ExhibitionStatusCompleted, ExhibitionStatusArchived:
		return true
	default:
		return false
	}
}
func exhibitionAllowsPlanningChanges(list ExhibitionList) bool {
	return !list.Locked && (list.Status == ExhibitionStatusDraft || list.Status == ExhibitionStatusOpen)
}
