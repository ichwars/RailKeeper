package main

import (
	"encoding/json"
	"io"

	"railkeeper/backend/internal/buildinfo"
)

func writeVersion(writer io.Writer, version string, build buildinfo.Info) error {
	return json.NewEncoder(writer).Encode(struct {
		Version string         `json:"version"`
		Build   buildinfo.Info `json:"build"`
	}{version, build})
}
