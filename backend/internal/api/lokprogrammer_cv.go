package api

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const maxDecoderCVImportRows = 10000

var (
	lokCVLine         = regexp.MustCompile(`^CV\s+(\d+)\s*=\s*(\d+)\s*$`)
	lokIndexLine      = regexp.MustCompile(`^Index:\s*(\d+)\s*\(CV31=(\d+),\s*CV32=(\d+)\)\s*$`)
	lokCVFormat       = regexp.MustCompile(`(?m)^(?:Decoder:|Index:|CV\s+\d+\s*=)`)
	lokFirmwareSuffix = regexp.MustCompile(`\s+\([0-9]+(?:\.[0-9]+)+\)$`)
)

func parseLokProgrammerCVList(text string) ([]cvValuePreview, string, bool, error) {
	text = strings.TrimPrefix(text, "\ufeff")
	if !lokCVFormat.MatchString(text) {
		return nil, "", false, nil
	}
	values := []cvValuePreview{}
	profile := ""
	decoderHeader := ""
	exportedAt := ""
	var cv31, cv32 *int
	seen := map[string]bool{}
	for lineNumber, raw := range strings.Split(text, "\n") {
		line := strings.TrimSpace(raw)
		switch {
		case line == "", strings.Trim(line, "-") == "", strings.HasPrefix(line, "Filename:"):
			continue
		case strings.HasPrefix(line, "Created:"):
			value := strings.TrimSpace(strings.TrimPrefix(line, "Created:"))
			created, err := time.Parse("1/2/2006 15:04:05", value)
			if err != nil || exportedAt != "" || created.Year() < 1 {
				return nil, profile, true, fmt.Errorf("line %d: invalid or duplicate export date", lineNumber+1)
			}
			exportedAt = created.Format("2006-01-02T15:04:05")
			continue
		case strings.HasPrefix(line, "Decoder:"):
			decoderHeader = strings.TrimSpace(strings.TrimPrefix(line, "Decoder:"))
			profile = lokFirmwareSuffix.ReplaceAllString(decoderHeader, "")
			continue
		case strings.HasPrefix(line, "Index:"):
			match := lokIndexLine.FindStringSubmatch(line)
			if match == nil {
				return nil, profile, true, fmt.Errorf("line %d: invalid index header", lineNumber+1)
			}
			index, ok := boundedAtoi(match[1], 0, 65535)
			high, highOK := boundedAtoi(match[2], 0, 255)
			low, lowOK := boundedAtoi(match[3], 0, 255)
			if !ok || !highOK || !lowOK || index != high*256+low {
				return nil, profile, true, fmt.Errorf("line %d: inconsistent index", lineNumber+1)
			}
			cv31, cv32 = &high, &low
			continue
		}
		match := lokCVLine.FindStringSubmatch(line)
		if match == nil {
			return nil, profile, true, fmt.Errorf("line %d: invalid CV row", lineNumber+1)
		}
		number, numberOK := boundedAtoi(match[1], 1, 1024)
		value, valueOK := boundedAtoi(match[2], 0, 255)
		if !numberOK || !valueOK || (number > 256 && (cv31 == nil || number > 512)) {
			return nil, profile, true, fmt.Errorf("line %d: invalid CV value or missing index", lineNumber+1)
		}
		sourceLine := lineNumber + 1
		row := cvValuePreview{CVNumber: number, Value: value, SourceLine: &sourceLine}
		if number == 7 && decoderHeader != "" {
			row.Description = "LokProgrammer Decoder: " + decoderHeader
		}
		key := strconv.Itoa(number)
		if number > 256 {
			row.IndexCV31, row.IndexCV32 = cv31, cv32
			key += fmt.Sprintf(":%d:%d", *cv31, *cv32)
		}
		if seen[key] || len(values) >= maxDecoderCVImportRows {
			return nil, profile, true, fmt.Errorf("line %d: duplicate CV or too many values", lineNumber+1)
		}
		seen[key] = true
		values = append(values, row)
	}
	for i := range values {
		values[i].SourceExportedAt = exportedAt
	}
	return values, profile, true, nil
}
