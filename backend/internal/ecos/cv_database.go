package ecos

import (
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

const CVDatabaseLimit = 8192
const spreadsheetNS = "urn:schemas-microsoft-com:office:spreadsheet"

var ErrCVDatabase = errors.New("invalid ECoS CV database export")
var databaseAddress = regexp.MustCompile(`^(?:([0-9]+):)?([0-9]+)$`)

type DatabaseCV struct {
	Number    int
	Value     int
	IndexCV31 *int
	IndexCV32 *int
}

// Reuse the configured-LAN-only, pinned HTTP transport; never follow a supplied
// export URL, redirect or proxy. Reading this export does not program a decoder.
type CVDatabaseReader struct{ web *ImageWeb }

func NewCVDatabaseReader(ctx context.Context, host string) (*CVDatabaseReader, error) {
	web, err := NewImageWeb(ctx, host)
	if err != nil {
		return nil, err
	}
	return &CVDatabaseReader{web: web}, nil
}

func (reader *CVDatabaseReader) Address() string { return reader.web.Address() }

func (reader *CVDatabaseReader) Read(ctx context.Context, objectID int, name string) ([]DatabaseCV, error) {
	if objectID < 1 || objectID > 2_000_000_000 {
		return nil, ErrCVDatabase
	}
	path := fmt.Sprintf("/fcgi/ecos-infolokcv?lang=en&target=xml&oid=%d", objectID)
	data, err := reader.web.request(ctx, http.MethodGet, path, "", nil, 4<<20)
	if err != nil {
		return nil, err
	}
	return parseCVDatabase(data, name)
}

type databaseCell struct {
	Attributes []xml.Attr `xml:",any,attr"`
	Data       []struct {
		Body string `xml:",innerxml"`
	} `xml:"urn:schemas-microsoft-com:office:spreadsheet Data"`
}

func parseCVDatabase(data []byte, name string) ([]DatabaseCV, error) {
	var workbook struct {
		XMLName    xml.Name
		Worksheets []struct {
			Name   string `xml:"urn:schemas-microsoft-com:office:spreadsheet Name,attr"`
			Tables []struct {
				Columns string `xml:"urn:schemas-microsoft-com:office:spreadsheet ExpandedColumnCount,attr"`
				Count   string `xml:"urn:schemas-microsoft-com:office:spreadsheet ExpandedRowCount,attr"`
				Rows    []struct {
					Cells []databaseCell `xml:"urn:schemas-microsoft-com:office:spreadsheet Cell"`
				} `xml:"urn:schemas-microsoft-com:office:spreadsheet Row"`
			} `xml:"urn:schemas-microsoft-com:office:spreadsheet Table"`
		} `xml:"urn:schemas-microsoft-com:office:spreadsheet Worksheet"`
	}
	if len(data) > 4<<20 || bytes.Contains(data, []byte("<!DOCTYPE")) || bytes.Contains(data, []byte("<!ENTITY")) {
		return nil, ErrCVDatabase
	}
	decoder := xml.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(&workbook); err != nil || workbook.XMLName != (xml.Name{Space: spreadsheetNS, Local: "Workbook"}) ||
		len(workbook.Worksheets) != 1 {
		return nil, ErrCVDatabase
	}
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, ErrCVDatabase
		}
		if space, ok := token.(xml.CharData); !ok || strings.TrimSpace(string(space)) != "" {
			return nil, ErrCVDatabase
		}
	}
	worksheet := workbook.Worksheets[0]
	if worksheet.Name != "Loco "+name+" cvs" || len(worksheet.Tables) != 1 {
		return nil, ErrCVDatabase
	}
	table := worksheet.Tables[0]
	count, err := strconv.Atoi(table.Count)
	if err != nil || table.Columns != "3" || count != len(table.Rows) || count < 1 || count > CVDatabaseLimit+1 {
		return nil, ErrCVDatabase
	}
	values := make([]DatabaseCV, 0, count-1)
	seen := map[string]bool{}
	for index, row := range table.Rows {
		if len(row.Cells) != 3 {
			return nil, ErrCVDatabase
		}
		cells := make([]string, 3)
		for column, cell := range row.Cells {
			if len(cell.Data) != 1 {
				return nil, ErrCVDatabase
			}
			for _, attribute := range cell.Attributes {
				if attribute.Name.Local != "StyleID" {
					return nil, ErrCVDatabase
				}
			}
			cells[column] = strings.TrimSpace(cell.Data[0].Body)
		}
		if index == 0 {
			if cells[0] != "CV" || cells[1] != "Dec" || cells[2] != "Hex" {
				return nil, ErrCVDatabase
			}
			continue
		}
		value, key, err := databaseCVAddress(cells[0])
		decimal, decErr := strconv.ParseUint(cells[1], 10, 8)
		hexadecimal, hexErr := strconv.ParseUint(cells[2], 16, 8)
		if err != nil || decErr != nil || hexErr != nil || decimal != hexadecimal || seen[key] {
			return nil, ErrCVDatabase
		}
		seen[key] = true
		value.Value = int(decimal)
		values = append(values, value)
	}
	return values, nil
}

func databaseCVAddress(address string) (DatabaseCV, string, error) {
	value := DatabaseCV{}
	match := databaseAddress.FindStringSubmatch(address)
	if match == nil {
		return value, "", ErrCVDatabase
	}
	number, err := strconv.Atoi(match[2])
	if err != nil || number < 1 || number > 1024 {
		return value, "", ErrCVDatabase
	}
	value.Number = number
	if match[1] == "" {
		return value, strconv.Itoa(number), nil
	}
	page, err := strconv.Atoi(match[1])
	if err != nil || page < 0 || page > 65535 || number < 257 || number > 512 {
		return value, "", ErrCVDatabase
	}
	// ECoS exports a combined 16-bit page. Keep its two register bytes,
	// including page zero; never infer a page from the current CV31/CV32.
	high, low := page/256, page%256
	value.IndexCV31, value.IndexCV32 = &high, &low
	return value, fmt.Sprintf("%d:%d", page, number), nil
}
