package api

import (
	"bytes"
	"encoding/json"
	"fmt"
)

func previewCVJSON(data []byte) ([]cvValuePreview, error) {
	var rows []cvValuePreview
	if bytes.HasPrefix(bytes.TrimSpace(data), []byte("[")) {
		if err := json.Unmarshal(data, &rows); err != nil {
			return nil, err
		}
	} else {
		var document struct {
			Values []cvValuePreview `json:"cvValues"`
		}
		if err := json.Unmarshal(data, &document); err != nil {
			return nil, err
		}
		rows = document.Values
	}
	if len(rows) > maxDecoderCVImportRows {
		return nil, fmt.Errorf("maximum %d CV values", maxDecoderCVImportRows)
	}
	return rows, nil
}
