package api

import (
	"bytes"
	"encoding/csv"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

func previewCVValues(filename string, data []byte) []cvValuePreview {
	text, ok := decoderPreviewText(filename, data)
	if !ok {
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(text), "<") {
		return previewCVValuesXML(text)
	}
	return previewCVValuesTable(text)
}

func previewFunctionMappings(filename string, data []byte) []functionMappingPreview {
	text, ok := decoderPreviewText(filename, data)
	if !ok {
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(text), "<") {
		return previewFunctionMappingsXML(text)
	}
	return previewFunctionMappingsTable(text)
}

func decoderPreviewText(filename string, data []byte) (string, bool) {
	extension := strings.ToLower(filepath.Ext(filename))
	switch extension {
	case ".csv", ".txt", ".xml", ".json", ".z21":
	default:
		return "", false
	}
	if len(data) == 0 || len(data) > 2*1024*1024 || bytes.IndexByte(data, 0) >= 0 {
		return "", false
	}
	return strings.TrimPrefix(string(data), "\ufeff"), true
}

func previewCVValuesTable(text string) []cvValuePreview {
	records := readDelimitedRecords(text)
	if len(records) > 0 {
		if rows := cvValuesFromRecords(records); len(rows) > 0 {
			return rows
		}
	}
	rows := []cvValuePreview{}
	for _, match := range cvTextPattern.FindAllStringSubmatch(text, -1) {
		cvNumber, cvOK := boundedAtoi(match[1], 1, 1024)
		value, valueOK := boundedAtoi(match[2], 0, 255)
		if !cvOK || !valueOK {
			continue
		}
		rows = append(rows, cvValuePreview{CVNumber: cvNumber, Value: value})
		if len(rows) >= maxDecoderCVImportRows {
			break
		}
	}
	return rows
}

func cvValuesFromRecords(records [][]string) []cvValuePreview {
	if len(records) == 0 {
		return nil
	}
	header := normalizedHeader(records[0])
	start := 1
	cvIndex := firstHeaderIndex(header, "cv", "cvnumber", "cvnummer", "cvnr", "cvid")
	valueIndex := firstHeaderIndex(header, "value", "wert", "cvvalue")
	descriptionIndex := firstHeaderIndex(header, "description", "beschreibung", "name", "bezeichnung")
	categoryIndex := firstHeaderIndex(header, "category", "kategorie")
	profileIndex := firstHeaderIndex(header, "decoderprofile", "decoderprofil")
	protocolIndex := firstHeaderIndex(header, "protocol", "protokoll")
	highIndex := firstHeaderIndex(header, "indexcv31", "cv31")
	lowIndex := firstHeaderIndex(header, "indexcv32", "cv32")
	sourceLineIndex := firstHeaderIndex(header, "sourceline", "originalzeile")
	exportIndex := firstHeaderIndex(header, "sourceexportedat", "exportdatum")
	if cvIndex < 0 || valueIndex < 0 {
		start = 0
		cvIndex = 0
		valueIndex = 1
		descriptionIndex = 2
		categoryIndex = 3
		profileIndex = 4
		highIndex = 5
		lowIndex = 6
	}

	rows := []cvValuePreview{}
	for _, record := range records[start:] {
		cvNumber, cvOK := cellInt(record, cvIndex, 1, 1024)
		value, valueOK := cellInt(record, valueIndex, 0, 255)
		if !cvOK || !valueOK {
			continue
		}
		var high, low *int
		if text := cell(record, highIndex); text != "" {
			v, err := strconv.Atoi(text)
			if err != nil {
				return nil
			}
			high = &v
		}
		if text := cell(record, lowIndex); text != "" {
			v, err := strconv.Atoi(text)
			if err != nil {
				return nil
			}
			low = &v
		}
		var sourceLine *int
		if cell(record, sourceLineIndex) != "" {
			n, ok := cellInt(record, sourceLineIndex, 1, 10000000)
			if !ok {
				return nil
			}
			sourceLine = &n
		}
		rows = append(rows, cvValuePreview{
			SourceLine: sourceLine, SourceExportedAt: cell(record, exportIndex),
			DecoderProfile: cell(record, profileIndex), Protocol: cell(record, protocolIndex),
			IndexCV31: high, IndexCV32: low,
			CVNumber:    cvNumber,
			Value:       value,
			Description: cell(record, descriptionIndex),
			Category:    cell(record, categoryIndex),
		})
		if len(rows) >= maxDecoderCVImportRows {
			break
		}
	}
	return rows
}

func previewFunctionMappingsTable(text string) []functionMappingPreview {
	records := readDelimitedRecords(text)
	if len(records) > 0 {
		if rows := functionMappingsFromRecords(records); len(rows) > 0 {
			return rows
		}
	}
	rows := []functionMappingPreview{}
	for _, match := range functionTextPattern.FindAllStringSubmatch(text, -1) {
		number, ok := boundedAtoi(match[1], 0, 31)
		name := strings.TrimSpace(match[2])
		if !ok || name == "" {
			continue
		}
		rows = append(rows, functionMappingPreview{
			FunctionKey:  fmt.Sprintf("F%d", number),
			Name:         name,
			FunctionType: inferFunctionType(name),
		})
		if len(rows) >= maxDecoderPreviewRows {
			break
		}
	}
	return rows
}

func functionMappingsFromRecords(records [][]string) []functionMappingPreview {
	if len(records) == 0 {
		return nil
	}
	header := normalizedHeader(records[0])
	start := 1
	keyIndex := firstHeaderIndex(header, "functionkey", "funktion", "taste", "key", "f")
	nameIndex := firstHeaderIndex(header, "name", "funktionsname", "beschreibung", "function")
	typeIndex := firstHeaderIndex(header, "functiontype", "typ", "type")
	hasHeader := keyIndex >= 0 && nameIndex >= 0
	if keyIndex < 0 || nameIndex < 0 {
		start = 0
		keyIndex = 0
		nameIndex = 1
		typeIndex = 2
	}

	rows := []functionMappingPreview{}
	for _, record := range records[start:] {
		if !hasHeader && !strings.Contains(strings.ToUpper(cell(record, keyIndex)), "F") {
			continue
		}
		key := normalizeFunctionPreviewKey(cell(record, keyIndex))
		name := cell(record, nameIndex)
		if key == "" || name == "" {
			continue
		}
		functionType := cell(record, typeIndex)
		if functionType == "" {
			functionType = inferFunctionType(name)
		}
		rows = append(rows, functionMappingPreview{
			FunctionKey:  key,
			Name:         name,
			FunctionType: functionType,
		})
		if len(rows) >= maxDecoderPreviewRows {
			break
		}
	}
	return rows
}

func readDelimitedRecords(text string) [][]string {
	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	for _, separator := range []rune{';', ',', '\t'} {
		reader := csv.NewReader(strings.NewReader(strings.Join(lines, "\n")))
		reader.Comma = separator
		reader.FieldsPerRecord = -1
		reader.TrimLeadingSpace = true
		records, err := reader.ReadAll()
		if err == nil && hasMultiColumnRecord(records) {
			return records
		}
	}
	return nil
}

func hasMultiColumnRecord(records [][]string) bool {
	for _, record := range records {
		if len(record) > 1 {
			return true
		}
	}
	return false
}

func previewCVValuesXML(text string) []cvValuePreview {
	decoder := xml.NewDecoder(strings.NewReader(text))
	rows := []cvValuePreview{}
	for {
		token, err := decoder.Token()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil
		}
		start, ok := token.(xml.StartElement)
		if !ok {
			continue
		}
		attrs := xmlAttrs(start)
		cvNumber, cvOK := attrBoundedInt(attrs, 1, 1024, "cv", "number", "nr", "id")
		value, valueOK := attrBoundedInt(attrs, 0, 255, "value", "wert")
		if !cvOK || !valueOK {
			continue
		}
		rows = append(rows, cvValuePreview{
			CVNumber:    cvNumber,
			Value:       value,
			Description: firstAttr(attrs, "description", "beschreibung", "name"),
			Category:    firstAttr(attrs, "category", "kategorie"),
		})
		if len(rows) >= maxDecoderCVImportRows {
			break
		}
	}
	return rows
}

func previewFunctionMappingsXML(text string) []functionMappingPreview {
	decoder := xml.NewDecoder(strings.NewReader(text))
	rows := []functionMappingPreview{}
	for {
		token, err := decoder.Token()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil
		}
		start, ok := token.(xml.StartElement)
		if !ok {
			continue
		}
		attrs := xmlAttrs(start)
		key := normalizeFunctionPreviewKey(firstAttr(attrs, "functionkey", "key", "taste", "function", "id"))
		name := firstAttr(attrs, "name", "description", "beschreibung", "label")
		if key == "" || name == "" {
			continue
		}
		functionType := firstAttr(attrs, "functiontype", "type", "typ")
		if functionType == "" {
			functionType = inferFunctionType(name)
		}
		rows = append(rows, functionMappingPreview{
			FunctionKey:  key,
			Name:         name,
			FunctionType: functionType,
		})
		if len(rows) >= maxDecoderPreviewRows {
			break
		}
	}
	return rows
}

func xmlAttrs(start xml.StartElement) map[string]string {
	attrs := map[string]string{}
	for _, attr := range start.Attr {
		attrs[normalizeHeaderName(attr.Name.Local)] = strings.TrimSpace(attr.Value)
	}
	return attrs
}

func normalizedHeader(record []string) []string {
	out := make([]string, len(record))
	for index, value := range record {
		out[index] = normalizeHeaderName(value)
	}
	return out
}

func normalizeHeaderName(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	replacer := strings.NewReplacer("ä", "ae", "ö", "oe", "ü", "ue", "ß", "ss")
	value = replacer.Replace(value)
	return regexp.MustCompile(`[^a-z0-9]`).ReplaceAllString(value, "")
}

func firstHeaderIndex(header []string, names ...string) int {
	for _, name := range names {
		needle := normalizeHeaderName(name)
		for index, value := range header {
			if value == needle {
				return index
			}
		}
	}
	return -1
}

func cell(record []string, index int) string {
	if index < 0 || index >= len(record) {
		return ""
	}
	return strings.TrimSpace(record[index])
}

func cellInt(record []string, index, min, max int) (int, bool) {
	return boundedAtoi(cell(record, index), min, max)
}

func boundedAtoi(value string, min, max int) (int, bool) {
	value = strings.TrimSpace(strings.TrimPrefix(strings.ToUpper(value), "CV"))
	value = strings.TrimPrefix(value, "F")
	number, err := strconv.Atoi(value)
	if err != nil || number < min || number > max {
		return 0, false
	}
	return number, true
}

func attrBoundedInt(attrs map[string]string, min, max int, names ...string) (int, bool) {
	for _, name := range names {
		if value := firstAttr(attrs, name); value != "" {
			if number, ok := boundedAtoi(value, min, max); ok {
				return number, true
			}
		}
	}
	return 0, false
}

func firstAttr(attrs map[string]string, names ...string) string {
	for _, name := range names {
		if value := strings.TrimSpace(attrs[normalizeHeaderName(name)]); value != "" {
			return value
		}
	}
	return ""
}

func normalizeFunctionPreviewKey(value string) string {
	value = strings.TrimSpace(strings.ToUpper(value))
	if strings.HasPrefix(value, "F") {
		if number, ok := boundedAtoi(value, 0, 31); ok {
			return fmt.Sprintf("F%d", number)
		}
	}
	if number, ok := boundedAtoi(value, 0, 31); ok {
		return fmt.Sprintf("F%d", number)
	}
	return ""
}

func inferFunctionType(name string) string {
	lower := strings.ToLower(name)
	switch {
	case strings.Contains(lower, "licht") || strings.Contains(lower, "light"):
		return "licht"
	case strings.Contains(lower, "sound") || strings.Contains(lower, "horn") || strings.Contains(lower, "pfiff"):
		return "sound"
	case strings.Contains(lower, "rauch"):
		return "rauch"
	case strings.Contains(lower, "kuppl"):
		return "kupplung"
	default:
		return "standard"
	}
}
