package api

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestOpenAPIDataTransferProfileSecurityAndLastUsedAt(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "..", "openapi", "railkeeper.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	contract := string(data)
	profiles := openAPIIndentedBlock(t, contract, "/data-transfer/profiles", 2)
	read := openAPIIndentedBlock(t, profiles, "get", 4)
	if !strings.Contains(read, "security:\n        - sessionCookie: []") {
		t.Fatalf("data-transfer profile read is missing session security: %s", read)
	}
	for _, method := range []string{"post"} {
		write := openAPIIndentedBlock(t, profiles, method, 4)
		if !strings.Contains(write, "security:\n        - sessionCookie: []\n          csrfHeader: []") {
			t.Fatalf("data-transfer profile %s is missing session and CSRF security: %s", method, write)
		}
	}
	profile := openAPIIndentedBlock(t, contract, "DataTransferProfile", 4)
	if strings.Contains(profile, "required: [id, name, direction, format, areas, options, enabled, createdByUserId, lastUsedAt") {
		t.Fatalf("lastUsedAt must be optional for newly created profiles: %s", profile)
	}
}

func TestOpenAPIDataTransferProfileItemWriteSecurity(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "..", "openapi", "railkeeper.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	profile := openAPIIndentedBlock(t, string(data), "/data-transfer/profiles/{id}", 2)
	for _, method := range []string{"put", "delete"} {
		write := openAPIIndentedBlock(t, profile, method, 4)
		if !strings.Contains(write, "security:\n        - sessionCookie: []\n          csrfHeader: []") {
			t.Fatalf("data-transfer profile %s is missing session and CSRF security: %s", method, write)
		}
	}
}
