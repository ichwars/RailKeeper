package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"testing"

	"railkeeper/backend/internal/buildinfo"
)

type closedVersionWriter struct{}

func (closedVersionWriter) Write([]byte) (int, error) { return 0, io.ErrClosedPipe }

func TestVersionCommandReportsReleaseAndSourceSeparately(t *testing.T) {
	var output bytes.Buffer
	info := buildinfo.Info{Commit: "0123456789012345678901234567890123456789", ShortCommit: "012345678901", Status: "development", GoVersion: "go-test", Platform: "windows/amd64"}
	if err := writeVersion(&output, "0.1.22", info); err != nil {
		t.Fatal(err)
	}
	var result struct {
		Version string         `json:"version"`
		Build   buildinfo.Info `json:"build"`
	}
	if err := json.Unmarshal(output.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.Version != "0.1.22" || result.Build != info {
		t.Fatal("version command lost candidate identity")
	}
	if err := writeVersion(closedVersionWriter{}, "0.1.22", info); !errors.Is(err, io.ErrClosedPipe) {
		t.Fatal("output failure was reported as success")
	}
}
