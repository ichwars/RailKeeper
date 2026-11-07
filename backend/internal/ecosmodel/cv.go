package ecosmodel

import (
	"sort"
	"strconv"
	"strings"

	"railkeeper/backend/internal/decoder"
)

func ECoSTargetedCVProbeFields() []string {
	return []string{
		"cv[1:8]",
		"cv[7]",
		"cv[8]",
		"cv[7:8]",
		"cv[17]",
		"cv[18]",
		"cv[29]",
	}
}

func ParseECoSCVValues(attributes map[string][]string) []ECoSCVValue {
	parsed := map[int]int{}
	add := func(number int, value int) {
		if !decoder.ValidCVNumber(number) || !decoder.ValidCVValue(value) {
			return
		}
		parsed[number] = value
	}
	for _, key := range SortedECoSCVAttributeKeys(attributes) {
		normalizedKey := strings.ToLower(strings.TrimSpace(key))
		keyNumber := ParseECoSCVKeyNumber(normalizedKey)
		values := attributes[key]
		for _, value := range values {
			parts := SplitECoSParams(value)
			if len(parts) >= 2 {
				for index := 0; index+1 < len(parts); index += 2 {
					number, numberErr := strconv.Atoi(strings.TrimSpace(parts[index]))
					cvValue, valueErr := strconv.Atoi(strings.TrimSpace(parts[index+1]))
					if numberErr == nil && valueErr == nil {
						add(number, cvValue)
					}
				}
				continue
			}
			if keyNumber > 0 && len(parts) == 1 {
				if cvValue, err := strconv.Atoi(strings.TrimSpace(parts[0])); err == nil {
					add(keyNumber, cvValue)
				}
				continue
			}
			number, cvValue, ok := ParseECoSCVTextPair(value)
			if ok {
				add(number, cvValue)
			}
		}
	}
	cvs := make([]ECoSCVValue, 0, len(parsed))
	for number, value := range parsed {
		cvs = append(cvs, ECoSCVValue{Number: number, Value: value})
	}
	SortECoSCVValues(cvs)
	return cvs
}

func SortedECoSCVAttributeKeys(attributes map[string][]string) []string {
	primary := []string{}
	specific := []string{}
	for key := range attributes {
		normalizedKey := strings.ToLower(strings.TrimSpace(key))
		if !IsECoSCVAttributeKey(normalizedKey) {
			continue
		}
		if ParseECoSCVKeyNumber(normalizedKey) > 0 {
			specific = append(specific, key)
			continue
		}
		primary = append(primary, key)
	}
	SortStrings(primary)
	SortStrings(specific)
	return append(primary, specific...)
}

func IsECoSCVAttributeKey(key string) bool {
	return key == "cv" || key == "cvs" || key == "cvlist" || ParseECoSCVKeyNumber(key) > 0
}

func ParseECoSCVKeyNumber(key string) int {
	if !strings.HasPrefix(key, "cv") {
		return 0
	}
	suffix := strings.Trim(strings.TrimPrefix(key, "cv"), " _-[]")
	if suffix == "" {
		return 0
	}
	number, err := strconv.Atoi(suffix)
	if err != nil {
		return 0
	}
	return number
}

func ParseECoSCVTextPair(value string) (int, int, bool) {
	value = CleanECoSValue(value)
	for _, separator := range []string{"=", ":"} {
		parts := strings.SplitN(value, separator, 2)
		if len(parts) != 2 {
			continue
		}
		number, numberErr := strconv.Atoi(strings.TrimSpace(parts[0]))
		cvValue, valueErr := strconv.Atoi(strings.TrimSpace(parts[1]))
		if numberErr == nil && valueErr == nil && decoder.ValidCVNumber(number) && decoder.ValidCVValue(cvValue) {
			return number, cvValue, true
		}
	}
	return 0, 0, false
}

func SortECoSCVValues(values []ECoSCVValue) {
	sort.SliceStable(values, func(i, j int) bool {
		if values[i].Number != values[j].Number {
			return values[i].Number < values[j].Number
		}
		return ECoSCVPage(values[i]) < ECoSCVPage(values[j])
	})
}

func ECoSCVPage(value ECoSCVValue) int {
	if value.IndexCV31 == nil || value.IndexCV32 == nil {
		return -1
	}
	return *value.IndexCV31*256 + *value.IndexCV32
}
