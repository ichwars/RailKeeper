package ecosmodel

import (
	"strconv"
	"strings"
)

var eCoSAllowedLocomotiveAttributes = map[string]struct{}{
	"addr":            {},
	"cv":              {},
	"cvlist":          {},
	"cvs":             {},
	"funcdesc":        {},
	"functionmapping": {},
	"name":            {},
	"profile":         {},
	"protocol":        {},
}

func IsAllowedECoSLocomotiveAttribute(key string) bool {
	_, allowed := eCoSAllowedLocomotiveAttributes[strings.ToLower(strings.TrimSpace(key))]
	return allowed
}

func ParseECoSFields(lines []string) map[string]string {
	fields := map[string]string{}
	for _, line := range lines {
		if strings.HasPrefix(line, "<") {
			continue
		}
		for key, value := range ParseECoSArguments(line) {
			fields[key] = value
		}
	}
	return fields
}

func ParseECoSAttributes(lines []string) map[string][]string {
	attributes := map[string][]string{}
	for _, line := range lines {
		if strings.HasPrefix(line, "<") {
			continue
		}
		for _, arg := range ParseECoSArgumentList(line) {
			key := strings.ToLower(arg.Key)
			if !IsAllowedECoSLocomotiveAttribute(key) {
				continue
			}
			attributes[key] = append(attributes[key], CleanECoSValue(arg.Value))
		}
	}
	return attributes
}

func ParseECoSEndStatus(lines []string) (string, bool) {
	for _, line := range lines {
		if !strings.HasPrefix(line, "<END") {
			continue
		}
		status := strings.Trim(strings.TrimPrefix(line, "<END"), " >")
		fields := strings.Fields(status)
		return status, len(fields) > 0 && fields[0] == "0"
	}
	return "", false
}

func SortStrings(values []string) {
	for i := 1; i < len(values); i++ {
		value := values[i]
		j := i - 1
		for j >= 0 && values[j] > value {
			values[j+1] = values[j]
			j--
		}
		values[j+1] = value
	}
}

func ParseECoSArguments(line string) map[string]string {
	out := map[string]string{}
	for _, arg := range ParseECoSArgumentList(line) {
		out[arg.Key] = arg.Value
	}
	return out
}

func ParseECoSArgumentList(line string) []Argument {
	out := []Argument{}
	for index := 0; index < len(line); index++ {
		if line[index] != '[' {
			continue
		}
		keyStart := index - 1
		for keyStart >= 0 {
			c := line[keyStart]
			if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '_' {
				keyStart--
				continue
			}
			break
		}
		key := strings.TrimSpace(line[keyStart+1 : index])
		if key == "" {
			continue
		}
		valueStart := index + 1
		valueEnd := valueStart
		inQuote := false
		escaped := false
		for valueEnd < len(line) {
			c := line[valueEnd]
			if escaped {
				escaped = false
				valueEnd++
				continue
			}
			if c == '\\' {
				escaped = true
				valueEnd++
				continue
			}
			if c == '"' {
				inQuote = !inQuote
				valueEnd++
				continue
			}
			if c == ']' && !inQuote {
				break
			}
			valueEnd++
		}
		if valueEnd >= len(line) {
			continue
		}
		value := strings.TrimSpace(line[valueStart:valueEnd])
		out = append(out, Argument{
			Key:    key,
			Value:  value,
			Params: SplitECoSParams(value),
		})
		index = valueEnd
	}
	return out
}

func SplitECoSParams(value string) []string {
	parts := []string{}
	start := 0
	inQuote := false
	escaped := false
	for index := 0; index < len(value); index++ {
		c := value[index]
		if escaped {
			escaped = false
			continue
		}
		if c == '\\' {
			escaped = true
			continue
		}
		if c == '"' {
			inQuote = !inQuote
			continue
		}
		if c == ',' && !inQuote {
			parts = append(parts, CleanECoSValue(value[start:index]))
			start = index + 1
		}
	}
	parts = append(parts, CleanECoSValue(value[start:]))
	return parts
}

func CleanECoSValue(value string) string {
	value = strings.TrimSpace(value)
	if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
		value = value[1 : len(value)-1]
	}
	value = strings.ReplaceAll(value, `\"`, `"`)
	value = strings.ReplaceAll(value, `\\`, `\`)
	return value
}

func ParseECoSInt(value string) int {
	parsed, _ := strconv.Atoi(CleanECoSValue(value))
	return parsed
}
