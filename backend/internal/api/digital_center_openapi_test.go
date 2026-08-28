package api

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestOpenAPIDocumentsDigitalCenterWorkspaceOperationsAndSecurity(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "..", "openapi", "railkeeper.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	contract := string(data)
	operations := []struct {
		path      string
		method    string
		write     bool
		responses []string
	}{
		{path: "/digital-centers/workspace", method: "get", responses: []string{"403"}},
		{path: "/digital-centers/{provider}/read-sessions", method: "post", write: true,
			responses: []string{"400", "403", "404", "409"}},
		{path: "/digital-centers/read-sessions/{id}", method: "get", responses: []string{"403", "404"}},
		{path: "/digital-centers/read-sessions/{id}/messages", method: "get", responses: []string{"403", "404"}},
		{path: "/digital-centers/read-sessions/{id}/items", method: "get",
			responses: []string{"400", "403", "404"}},
		{path: "/digital-centers/read-sessions/{id}/items/{itemID}", method: "get",
			responses: []string{"403", "404"}},
		{path: "/digital-centers/{provider}/live/status", method: "get",
			responses: []string{"400", "403"}},
		{path: "/digital-centers/{provider}/live/start", method: "post", write: true,
			responses: []string{"400", "403", "404", "409", "502"}},
		{path: "/digital-centers/{provider}/live/stop", method: "post", write: true,
			responses: []string{"400", "403", "404"}},
		{path: "/digital-centers/read-sessions/{id}/items/{itemID}/write-preview", method: "post", write: true,
			responses: []string{"400", "403", "404", "409"}},
		{path: "/digital-centers/read-sessions/{id}/items/{itemID}/write-confirm", method: "post", write: true,
			responses: []string{"400", "403", "404", "409", "502"}},
	}
	for _, operation := range operations {
		block := openAPIIndentedBlock(t, openAPIIndentedBlock(t, contract, operation.path, 2), operation.method, 4)
		security := "security:\n        - sessionCookie: []"
		if operation.write {
			security += "\n          csrfHeader: []"
		}
		if !strings.Contains(block, security) {
			t.Errorf("%s %s security mismatch: %s", operation.method, operation.path, block)
		}
		for _, response := range operation.responses {
			if !strings.Contains(block, "        \""+response+"\":") {
				t.Errorf("%s %s missing %s response: %s", operation.method, operation.path, response, block)
			}
		}
	}
}

func TestOpenAPIDigitalCenterLiveStatusIsConfigurationIndependent(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "..", "openapi", "railkeeper.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	block := openAPIIndentedBlock(t, openAPIIndentedBlock(t, string(data),
		"/digital-centers/{provider}/live/status", 2), "get", 4)
	if strings.Contains(block, "        \"404\":") {
		t.Fatalf("passive live status must not depend on provider configuration: %s", block)
	}
	if !strings.Contains(block, "Provider does not support passive live monitoring") {
		t.Fatalf("passive live status must document capability-unavailable behavior: %s", block)
	}
}

func TestOpenAPIDocumentsCS3ReadOnlyProbeContract(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "..", "openapi", "railkeeper.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	contract := string(data)
	block := openAPIIndentedBlock(t, openAPIIndentedBlock(t, contract,
		"/digital-centers/cs3/probe", 2), "post", 4)
	for _, fragment := range []string{
		"Probe the compatible read-only Märklin CS3 locomotive API",
		"$ref: \"#/components/schemas/DigitalCenterConnectionInput\"",
		"$ref: \"#/components/schemas/DigitalCenterProbeResult\"",
		"security:\n        - sessionCookie: []\n          csrfHeader: []",
		"        \"400\":",
		"        \"403\":",
	} {
		if !strings.Contains(block, fragment) {
			t.Errorf("CS3 probe contract missing %q: %s", fragment, block)
		}
	}
	for schema, fragments := range map[string][]string{
		"DigitalCenterConnectionInput":    {"required: [host, port]", "maximum: 65535"},
		"DigitalCenterProbeResult":        {"commands:", "DigitalCenterProbeCommandResult"},
		"DigitalCenterProbeCommandResult": {"request:", "commandHex:", "read-only HTTP request or binary command"},
	} {
		schemaBlock := openAPIIndentedBlock(t, contract, schema, 4)
		for _, fragment := range fragments {
			if !strings.Contains(schemaBlock, fragment) {
				t.Errorf("%s missing %q: %s", schema, fragment, schemaBlock)
			}
		}
	}
}

func TestOpenAPIDigitalCenterSchemasExposeExactRuntimeEnumsAndTelemetry(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "..", "openapi", "railkeeper.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	contract := string(data)
	expectations := map[string][]string{
		"DigitalCenterSummary":        {"capabilities:", "transports:", "selected:", "active:"},
		"DigitalCenterTransport":      {"enum: [ecos_tcp, z21_udp, loconet_tcp, cs3_http]", "enum: [available, planned]", "capabilities:"},
		"DigitalCenterReadSession":    {"enum: [reading, ready, interrupted, failed]", "readCompletedAt:"},
		"DigitalCenterWorkItem":       {"enum: [ok, deviation, missing, new, conflict]", "center:", "railkeeper:", "conflicts:"},
		"DigitalCenterSessionMessage": {"enum: [info, warning, error]", "nextAction:"},
		"ECoSLiveStatus":              {"enum: [stopped, running, interrupted]", "pulseSamples:", "recentEvents:", "diagnosis:"},
		"ECoSLivePulseSample":         {"repliesPerSecond:", "at:"},
		"ECoSLiveEvent":               {"kind:", "protocol:", "message:"},
		"DigitalCenterWritePreview":   {"enum: [railkeeper_to_center]", "operation:", "changes:", "token:", "expiresAt:"},
		"DigitalCenterWriteConfirmation": {
			"enum: [verified, verification_failed, failed, unknown]", "operation:", "verifiedValues:", "liveMonitor:", "workItem:",
		},
	}
	for schema, fragments := range expectations {
		block := openAPIIndentedBlock(t, contract, schema, 4)
		for _, fragment := range fragments {
			if !strings.Contains(block, fragment) {
				t.Errorf("%s missing %q: %s", schema, fragment, block)
			}
		}
	}
}
