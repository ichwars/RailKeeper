package api

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"
	"testing"

	"railkeeper/backend/internal/application"
	"railkeeper/backend/internal/infrastructure"
)

func TestDigitalReadHTTPContractsValidateBeforeNetworkAndUseOnlyReadCommands(t *testing.T) {
	f := newContractFixture(t)
	// The listener belongs to this test. No layout or command station is contacted.
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = listener.Close() })
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go func() {
				defer func() { _ = conn.Close() }()
				scanner := bufio.NewScanner(conn)
				for scanner.Scan() {
					command := strings.TrimSpace(scanner.Text())
					viewLease := (strings.HasPrefix(command, "request(") || strings.HasPrefix(command, "release(")) && strings.HasSuffix(command, ", view)")
					readCommand := strings.HasPrefix(command, "get(") || strings.HasPrefix(command, "queryObjects(") || viewLease
					if !readCommand {
						t.Errorf("unexpected device write: %s", command)
						return
					}
					body := ""
					if strings.HasPrefix(command, "get(1001,") {
						body = "1001 name[\"Original\"] addr[3] protocol[DCC128]\r\n"
					}
					if command == "get(1, info, status)" {
						body = "1 status[GO]\r\n1 ProtocolVersion[0.5]\r\n1 ApplicationVersion[4.3.3]\r\n1 HardwareVersion[2.1]\r\n"
					}
					if _, err := fmt.Fprintf(conn, "<REPLY %s>\r\n%s<END 0 (OK)>\r\n", command, body); err != nil {
						return
					}
				}
			}()
		}
	}()
	host, rawPort, err := net.SplitHostPort(listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	port, err := strconv.Atoi(rawPort)
	if err != nil {
		t.Fatal(err)
	}
	for _, endpoint := range []string{"/api/v1/ecos/test", "/api/v1/ecos/locomotives/count", "/api/v1/ecos/locomotives/raw"} {
		assertProblem(t, f.raw(t, "POST", endpoint, "{"), 400, "invalid_json")
		assertStatus(t, f.call(t, "POST", endpoint, map[string]any{}), map[bool]int{true: 400, false: 502}[strings.HasSuffix(endpoint, "/test")])
		assertStatus(t, f.call(t, "POST", endpoint, map[string]any{"host": host, "port": port}), 200)
	}
	for _, provider := range []string{"z21", "intellibox3", "cs3"} {
		for _, operation := range []string{"test", "probe"} {
			path := "/api/v1/digital-centers/" + provider + "/" + operation
			assertProblem(t, f.raw(t, "POST", path, "{"), 400, "invalid_json")
			assertProblem(t, f.call(t, "POST", path, map[string]any{"host": "", "port": -1}), 400, "digital_center_validation")
		}
	}
	id := f.vehicle(t)
	path := "/api/v1/vehicles/" + id + "/external-mappings"
	mapping := map[string]any{"provider": "ecos", "externalId": "1001", "externalAddress": "3", "externalName": "Original"}
	assertStatus(t, f.call(t, "POST", path, mapping), 200)
	assertStatus(t, f.call(t, "POST", path, mapping), 200)
	assertProblem(t, f.raw(t, "POST", path, "{"), 400, "invalid_json")
	assertProblem(t, f.call(t, "POST", path, map[string]any{}), 400, "external_mapping_validation")
	assertProblem(t, f.call(t, "POST", "/api/v1/vehicles/missing/external-mappings", mapping), 404, "vehicle_not_found")
	another := f.vehicle(t)
	assertProblem(t, f.call(t, "POST", "/api/v1/vehicles/"+another+"/external-mappings", mapping), 409, "external_mapping_conflict")
	// A legacy sync preview must not send a write and must retain its stored mapping.
	syncPath := "/api/v1/digital-centers/ecos/locomotives/sync"
	assertProblem(t, f.raw(t, "POST", syncPath, "{"), 400, "invalid_json")
	assertProblem(t, f.call(t, "POST", syncPath, map[string]any{"vehicleId": another}), 400, "ecos_object_required")
	assertProblem(t, f.call(t, "POST", syncPath, map[string]any{"vehicleId": "missing"}), 404, "vehicle_not_found")
	assertStatus(t, f.call(t, "POST", syncPath, map[string]any{"vehicleId": id, "dryRun": true, "host": host, "port": port}), 200)
}

func TestWorkspaceReadAndPreviewHTTPContractsNeverBypassConfiguration(t *testing.T) {
	f := newContractFixture(t)
	endpoints := []struct {
		method, path string
		body         any
	}{
		{"GET", "/api/v1/vehicles/missing/ecos-functions", nil},
		{"GET", "/api/v1/vehicles/missing/ecos-image/options", nil},
		{"POST", "/api/v1/digital-centers/read-sessions/missing/items/missing/write-preview", map[string]any{}},
	}
	for _, endpoint := range endpoints {
		assertProblem(t, f.call(t, endpoint.method, endpoint.path, endpoint.body), 503, "digital_center_workspace_unavailable")
	}
	workspace := application.NewDigitalCenterWorkspaceService(infrastructure.NewDigitalCenterWorkspaceRepository(f.db), f.config.SettingsService, &apiWorkspaceECoSReader{}, nil, f.vehicles, f.auth)
	f.config.DigitalCenterWorkspace = workspace
	f.router = NewRouter(f.config)
	for _, endpoint := range endpoints {
		assertStatus(t, f.call(t, endpoint.method, endpoint.path, endpoint.body), 404)
	}
	assertProblem(t, f.raw(t, "POST", endpoints[2].path, "{"), 400, "invalid_json")
}
