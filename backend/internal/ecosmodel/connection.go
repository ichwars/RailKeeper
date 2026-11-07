package ecosmodel

import (
	"strings"
)

func MissingECoSConnectionFields(result *ECoSConnectionResult) []string {
	missing := []string{}
	for _, field := range []struct {
		name  string
		value string
	}{
		{name: "status", value: result.Status},
		{name: "ProtocolVersion", value: result.ProtocolVersion},
		{name: "ApplicationVersion", value: result.ApplicationVersion},
		{name: "HardwareVersion", value: result.HardwareVersion},
	} {
		if strings.TrimSpace(field.value) == "" {
			missing = append(missing, field.name)
		}
	}
	return missing
}

func SortedECoSFieldNames(fields map[string]bool) []string {
	names := make([]string, 0, len(fields))
	for field := range fields {
		names = append(names, field)
	}
	SortStrings(names)
	return names
}

func SortedMissingECoSFieldNames(missing map[string]bool, supported map[string]bool) []string {
	names := []string{}
	for field := range missing {
		if !supported[field] {
			names = append(names, field)
		}
	}
	SortStrings(names)
	return names
}
