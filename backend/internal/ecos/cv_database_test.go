package ecos

import (
	"fmt"
	"html"
	"reflect"
	"strings"
	"testing"
)

func cvDatabaseXML(name string, entries ...string) string {
	rows := []string{"CV|Dec|Hex"}
	rows = append(rows, entries...)
	var body strings.Builder
	for _, row := range rows {
		body.WriteString("<Row>")
		for _, cell := range strings.Split(row, "|") {
			fmt.Fprintf(&body, `<Cell ss:StyleID="Default"><Data ss:Type="String">%s</Data></Cell>`, html.EscapeString(cell))
		}
		body.WriteString("</Row>")
	}
	return fmt.Sprintf(`<Workbook xmlns="%s" xmlns:ss="%s"><Worksheet ss:Name="%s"><Table ss:ExpandedColumnCount="3" ss:ExpandedRowCount="%d">%s</Table></Worksheet></Workbook>`,
		spreadsheetNS, spreadsheetNS, html.EscapeString("Loco "+name+" cvs"), len(rows), body.String())
}

func TestCVDatabaseRetainsAllPagesZeroAndUnknown(t *testing.T) {
	data := cvDatabaseXML("Test & Lok", "257|7|07", "0:257|0|00", "257:257|144|90", "4099:257|20|14", "4104:257|1|01")
	values, err := parseCVDatabase([]byte(data), "Test & Lok")
	if err != nil || len(values) != 5 {
		t.Fatalf("%+v %v", values, err)
	}
	if values[0].IndexCV31 != nil || values[0].IndexCV32 != nil || values[1].Value != 0 || *values[1].IndexCV31 != 0 ||
		*values[1].IndexCV32 != 0 || *values[2].IndexCV31 != 1 || *values[2].IndexCV32 != 1 ||
		*values[3].IndexCV31 != 16 || *values[3].IndexCV32 != 3 || *values[4].IndexCV32 != 8 {
		t.Fatalf("index identity lost: %+v", values)
	}
}

func TestCVDatabaseRejectsAmbiguousTruncatedOrWrongDocument(t *testing.T) {
	valid := cvDatabaseXML("Test", "4099:257|20|14")
	for name, data := range map[string]string{
		"wrong locomotive": cvDatabaseXML("Another", "1|3|03"),
		"truncated":        valid[:len(valid)-1],
		"duplicate":        cvDatabaseXML("Test", "4099:257|20|14", "04099:0257|20|14"),
		"wrong hex":        cvDatabaseXML("Test", "1|3|04"),
		"invalid byte":     cvDatabaseXML("Test", "1|256|100"),
		"invalid page":     cvDatabaseXML("Test", "65536:257|0|00"),
		"nonindexed range": cvDatabaseXML("Test", "4099:1|0|00"),
		"missing cell":     cvDatabaseXML("Test", "1|3"),
		"invalid CV":       cvDatabaseXML("Test", "1025|0|00"),
		"signed address":   cvDatabaseXML("Test", "+1|0|00"),
		"second root":      valid + valid,
		"wrong namespace":  strings.ReplaceAll(valid, spreadsheetNS, "urn:other"),
		"row count":        strings.Replace(valid, `ExpandedRowCount="2"`, `ExpandedRowCount="3"`, 1),
		"sparse cell":      strings.Replace(valid, `ss:StyleID="Default"`, `ss:Index="2"`, 1),
		"formula":          strings.Replace(valid, `ss:StyleID="Default"`, `ss:Formula="=3"`, 1),
		"DTD":              `<!DOCTYPE Workbook SYSTEM "file:///etc/passwd">` + valid,
	} {
		t.Run(name, func(t *testing.T) {
			if values, err := parseCVDatabase([]byte(data), "Test"); err == nil || values != nil {
				t.Fatalf("accepted malformed export: %+v %v", values, err)
			}
		})
	}
}

func TestCVDatabaseAllowsEmptyAndMoreThan1024Entries(t *testing.T) {
	values, err := parseCVDatabase([]byte(cvDatabaseXML("Empty")), "Empty")
	if err != nil || len(values) != 0 {
		t.Fatal(values, err)
	}
	entries := []string{}
	for page := 4096; page < 4102; page++ {
		for number := 257; number <= 512; number++ {
			entries = append(entries, fmt.Sprintf("%d:%d|0|00", page, number))
		}
	}
	values, err = parseCVDatabase([]byte(cvDatabaseXML("Test", entries...)), "Test")
	if err != nil || len(values) != 1536 {
		t.Fatalf("%d %v", len(values), err)
	}
	if reflect.DeepEqual(values[0], values[256]) {
		t.Fatal("same CV on distinct pages collapsed")
	}
}
