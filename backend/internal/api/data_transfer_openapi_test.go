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

func TestOpenAPIDataTransferImportDocumentsRuntimeErrorsAndRevision(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "..", "openapi", "railkeeper.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	contract := string(data)
	paths := []struct {
		path      string
		method    string
		responses []string
	}{
		{"/data-transfer/jobs/import", "post", []string{"403", "404"}},
		{"/data-transfer/jobs/{id}/upload", "post", []string{"403", "404", "409"}},
		{"/data-transfer/jobs/{id}/issues/{issueID}", "put", []string{"403", "404", "409"}},
		{"/data-transfer/jobs/{id}/cancel", "post", []string{"403", "404", "409"}},
	}
	for _, path := range paths {
		operation := openAPIIndentedBlock(t,
			openAPIIndentedBlock(t, contract, path.path, 2), path.method, 4)
		for _, response := range path.responses {
			if !strings.Contains(operation, "        \""+response+"\":") {
				t.Fatalf("%s %s is missing response %s: %s", path.method, path.path, response, operation)
			}
		}
	}
	job := openAPIIndentedBlock(t, contract, "DataTransferJob", 4)
	if !strings.Contains(job, "revision:") || !strings.Contains(job, "packageVersion, revision,") {
		t.Fatalf("data transfer job does not expose required revision: %s", job)
	}
}
