package exhibition

import (
	"fmt"
	"strings"
	"time"
)

func exhibitionDayScopes(list ExhibitionList) []string {
	start, startErr := time.Parse("2006-01-02", list.Date)
	end, endErr := time.Parse("2006-01-02", list.EndDate)
	if startErr != nil || endErr != nil || end.Before(start) {
		return []string{"day1"}
	}
	days := int(end.Sub(start)/(24*time.Hour)) + 1
	result := make([]string, 0, days)
	for day := 1; day <= days && day <= 31; day++ {
		result = append(result, fmt.Sprintf("day%d", day))
	}
	return result
}
func overlappingExhibitionDays(left, right string, all []string) []string {
	leftDays := exhibitionDaySet(left, all)
	rightDays := exhibitionDaySet(right, all)
	result := make([]string, 0)
	for _, day := range all {
		if leftDays[day] && rightDays[day] {
			result = append(result, day)
		}
	}
	return result
}
func exhibitionDaySet(scope string, all []string) map[string]bool {
	result := map[string]bool{}
	if strings.TrimSpace(scope) == "all" || strings.TrimSpace(scope) == "" {
		for _, day := range all {
			result[day] = true
		}
		return result
	}
	for _, day := range strings.Split(scope, ",") {
		result[strings.TrimSpace(day)] = true
	}
	return result
}
