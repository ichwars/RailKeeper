package application

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestParseECoSLocomotives(t *testing.T) {
	lines := []string{
		"<REPLY queryObjects(10, addr, name, protocol)>",
		`1001 addr[3] name["BR 218"] protocol[DCC128]`,
		`1001 speed[12]`,
		`1001 speedstep[3]`,
		`1001 dir[1]`,
		`1001 funcset[101]`,
		`1001 funcdesc[0,3]`,
		`1001 funcdesc[2,6]`,
		`1002 addr[24] name["V 180"] protocol[MM27]`,
		"<END 0 (OK)>",
	}

	locomotives := parseECoSLocomotives(lines)
	if len(locomotives) != 2 {
		t.Fatalf("expected two locomotives, got %d", len(locomotives))
	}
	if locomotives[0].ObjectID != 1001 || locomotives[0].Name != "BR 218" || locomotives[0].Address != 3 || locomotives[0].Protocol != "DCC128" {
		t.Fatalf("unexpected first locomotive: %#v", locomotives[0])
	}
	if locomotives[0].Speed != 12 || locomotives[0].SpeedStep != 3 || locomotives[0].Direction != 1 {
		t.Fatalf("unexpected movement state: %#v", locomotives[0])
	}
	if len(locomotives[0].Functions) != 3 || !locomotives[0].Functions[0].Active || locomotives[0].Functions[2].Description != 6 {
		t.Fatalf("unexpected functions: %#v", locomotives[0].Functions)
	}
}

func TestECoSServiceTestConnection(t *testing.T) {
	listener := startECoSTestServer(t, func(command string) []string {
		if command != "get(1, info, status)" {
			t.Fatalf("unexpected command: %s", command)
		}
		return []string{
			"<REPLY get(1, info, status)>",
			"1 status[GO]",
			"1 ProtocolVersion[0.5]",
			"1 ApplicationVersion[4.2.2]",
			"1 HardwareVersion[2.1]",
			"<END 0 (OK)>",
		}
	})
	defer func() { _ = listener.Close() }()

	host, port := splitTestAddress(t, listener.Addr().String())
	service := NewECoSService()
	result, err := service.TestConnection(context.Background(), ECoSConnectionInput{Host: host, Port: port})
	if err != nil {
		t.Fatalf("test connection failed: %v", err)
	}
	if !result.Connected || result.Status != "GO" || result.ProtocolVersion != "0.5" {
		t.Fatalf("unexpected result: %#v", result)
	}
}

func TestECoSServiceProbeLocomotiveRaw(t *testing.T) {
	listener := startECoSTestServer(t, func(command string) []string {
		switch {
		case command == "queryObjects(10, addr, name, protocol)":
			return []string{
				"<REPLY queryObjects(10, addr, name, protocol)>",
				`1001 addr[3] name["BR 218"] protocol[DCC128]`,
				"<END 0 (OK)>",
			}
		case command == "request(1001, view)":
			return []string{
				"<REPLY request(1001, view)>",
				"<END 0 (OK)>",
			}
		case command == "get(1001, speed, speedstep, profile, protocol, name, addr, dir, funcset, funcdesc)":
			return []string{
				"<REPLY get(1001, speed, speedstep, profile, protocol, name, addr, dir, funcset, funcdesc)>",
				`1001 speed[0] protocol[DCC128] name["BR 218"] addr[3] funcset[10] funcdesc[0,3]`,
				"<END 0 (OK)>",
			}
		case command == "get(1001, image)":
			return []string{
				"<REPLY get(1001, image)>",
				`1001 image[12]`,
				"<END 0 (OK)>",
			}
		case command == "get(1001, cv)":
			return []string{
				"<REPLY get(1001, cv)>",
				`1001 cv[1,3] cv[2,0] cv[29,38]`,
				"<END 0 (OK)>",
			}
		case command == "get(1001, cv[8])":
			return []string{
				"<REPLY get(1001, cv[8])>",
				`1001 cv[8,151]`,
				"<END 0 (OK)>",
			}
		case command == "release(1001, view)":
			return []string{
				"<REPLY release(1001, view)>",
				"<END 0 (OK)>",
			}
		case strings.HasPrefix(command, "get(1001, "):
			return []string{
				fmt.Sprintf("<REPLY %s>", command),
				"<END 12 (unsupported)>",
			}
		default:
			t.Fatalf("unexpected command: %s", command)
		}
		return nil
	})
	defer func() { _ = listener.Close() }()

	host, port := splitTestAddress(t, listener.Addr().String())
	service := NewECoSService()
	probe, err := service.ProbeLocomotiveRaw(context.Background(), ECoSConnectionInput{Host: host, Port: port})
	if err != nil {
		t.Fatalf("probe failed: %v", err)
	}
	if len(probe.Locomotives) != 1 {
		t.Fatalf("unexpected locomotive count: %#v", probe)
	}
	locomotive := probe.Locomotives[0]
	if locomotive.Attributes["image"][0] != "12" {
		t.Fatalf("expected image field in raw attributes: %#v", locomotive.Attributes)
	}
	if len(locomotive.InterestingFields) == 0 {
		t.Fatalf("expected image to be marked as interesting: %#v", locomotive)
	}
	if len(locomotive.CVs) != 4 || locomotive.CVs[2].Number != 8 || locomotive.CVs[2].Value != 151 {
		t.Fatalf("expected structured CV values: %#v", locomotive.CVs)
	}
}

func TestECoSServiceCountLocomotivesUsesOnlyObjectList(t *testing.T) {
	commands := []string{}
	var mu sync.Mutex
	listener := startECoSTestServer(t, func(command string) []string {
		mu.Lock()
		commands = append(commands, command)
		mu.Unlock()
		if command != "queryObjects(10, addr, name, protocol)" {
			t.Fatalf("unexpected command: %s", command)
		}
		return []string{
			"<REPLY queryObjects(10, addr, name, protocol)>",
			`1001 addr[3] name["BR 218"] protocol[DCC128]`,
			`1002 addr[24] name["V 180"] protocol[MM27]`,
			"<END 0 (OK)>",
		}
	})
	defer func() { _ = listener.Close() }()

	host, port := splitTestAddress(t, listener.Addr().String())
	service := NewECoSService()
	summary, err := service.CountLocomotives(context.Background(), ECoSConnectionInput{Host: host, Port: port})
	if err != nil {
		t.Fatalf("count failed: %v", err)
	}
	if summary.Count != 2 {
		t.Fatalf("unexpected count: %#v", summary)
	}
	mu.Lock()
	defer mu.Unlock()
	if len(commands) != 1 {
		t.Fatalf("expected one lightweight command, got %#v", commands)
	}
}

func TestECoSLiveStatusStopsIdleSession(t *testing.T) {
	service := NewECoSService()
	service.liveStatus = ECoSLiveStatus{
		Provider:   "ecos",
		Connected:  true,
		LastSeenAt: time.Now().UTC().Add(-(eCoSLiveIdleTimeout + time.Minute)).Format(time.RFC3339),
		Message:    "ECoS-Live-Verbindung aktiv.",
	}

	status := service.LiveStatus()
	if status.Connected {
		t.Fatalf("expected idle live session to be stopped: %#v", status)
	}
	if !strings.Contains(status.Message, "automatisch beendet") {
		t.Fatalf("expected automatic timeout message, got %q", status.Message)
	}
}

func TestBuildECoSLocomotiveSyncCommand(t *testing.T) {
	current := &ECoSLocomotive{Name: `BR "Alt"`, Address: 3, Protocol: "DCC128"}
	changes, command, err := buildECoSLocomotiveSyncCommand(1001, current, ECoSLocomotiveSyncDesired{
		Name:     `BR "Neu"`,
		Address:  24,
		Protocol: "MM27",
	})
	if err != nil {
		t.Fatalf("build command failed: %v", err)
	}
	if len(changes) != 3 {
		t.Fatalf("expected three changes, got %#v", changes)
	}
	expected := `set(1001, name["BR \"Neu\""], addr[24], protocol[MM27])`
	if command != expected {
		t.Fatalf("unexpected command:\n%s", command)
	}
}

func TestECoSServiceSyncLocomotiveDryRunDoesNotWrite(t *testing.T) {
	commands := []string{}
	var mu sync.Mutex
	listener := startECoSTestServer(t, func(command string) []string {
		mu.Lock()
		commands = append(commands, command)
		mu.Unlock()
		switch command {
		case "request(1001, view)":
			return []string{
				"<REPLY request(1001, view)>",
				"<END 0 (OK)>",
			}
		case "get(1001, speed, speedstep, profile, protocol, name, addr, dir, funcset, funcdesc)":
			return []string{
				"<REPLY get(1001, speed, speedstep, profile, protocol, name, addr, dir, funcset, funcdesc)>",
				`1001 protocol[DCC128] name["BR 218"] addr[3]`,
				"<END 0 (OK)>",
			}
		case "release(1001, view)":
			return []string{
				"<REPLY release(1001, view)>",
				"<END 0 (OK)>",
			}
		default:
			t.Fatalf("unexpected command: %s", command)
		}
		return nil
	})
	defer func() { _ = listener.Close() }()

	host, port := splitTestAddress(t, listener.Addr().String())
	service := NewECoSService()
	result, err := service.SyncLocomotive(context.Background(), ECoSLocomotiveSyncInput{
		Host:     host,
		Port:     port,
		ObjectID: 1001,
		Desired: ECoSLocomotiveSyncDesired{
			Name:     "BR 218 neu",
			Address:  4,
			Protocol: "DCC128",
		},
		DryRun:  true,
		Confirm: false,
	})
	if err != nil {
		t.Fatalf("sync dry run failed: %v", err)
	}
	if result.Applied || !result.DryRun || len(result.Changes) != 2 {
		t.Fatalf("unexpected dry-run result: %#v", result)
	}
	mu.Lock()
	defer mu.Unlock()
	for _, command := range commands {
		if strings.HasPrefix(command, "set(") {
			t.Fatalf("dry run sent write command: %#v", commands)
		}
	}
}

func TestECoSServiceSyncLocomotiveWritesConfirmed(t *testing.T) {
	var written string
	var mu sync.Mutex
	listener := startECoSTestServer(t, func(command string) []string {
		switch command {
		case "request(1001, view)":
			return []string{
				"<REPLY request(1001, view)>",
				"<END 0 (OK)>",
			}
		case "get(1001, speed, speedstep, profile, protocol, name, addr, dir, funcset, funcdesc)":
			return []string{
				"<REPLY get(1001, speed, speedstep, profile, protocol, name, addr, dir, funcset, funcdesc)>",
				`1001 protocol[DCC128] name["BR 218"] addr[3]`,
				"<END 0 (OK)>",
			}
		case `set(1001, name["BR 218 neu"], addr[4])`:
			mu.Lock()
			written = command
			mu.Unlock()
			return []string{
				`<REPLY set(1001, name["BR 218 neu"], addr[4])>`,
				"<END 0 (OK)>",
			}
		case "release(1001, view)":
			return []string{
				"<REPLY release(1001, view)>",
				"<END 0 (OK)>",
			}
		default:
			t.Fatalf("unexpected command: %s", command)
		}
		return nil
	})
	defer func() { _ = listener.Close() }()

	host, port := splitTestAddress(t, listener.Addr().String())
	service := NewECoSService()
	result, err := service.SyncLocomotive(context.Background(), ECoSLocomotiveSyncInput{
		Host:     host,
		Port:     port,
		ObjectID: 1001,
		Desired: ECoSLocomotiveSyncDesired{
			Name:     "BR 218 neu",
			Address:  4,
			Protocol: "DCC128",
		},
		Confirm: true,
	})
	if err != nil {
		t.Fatalf("sync write failed: %v", err)
	}
	if !result.Applied || result.DryRun {
		t.Fatalf("unexpected write result: %#v", result)
	}
	mu.Lock()
	defer mu.Unlock()
	if written == "" {
		t.Fatal("expected set command to be written")
	}
}

func TestParseECoSCVValuesVariants(t *testing.T) {
	cvs := parseECoSCVValues(map[string][]string{
		"cvs":    {"1,3,2,0"},
		"cvlist": {"7:42"},
		"cv8":    {"151"},
	})
	if len(cvs) != 4 {
		t.Fatalf("expected four CV values, got %#v", cvs)
	}
	expected := map[int]int{1: 3, 2: 0, 7: 42, 8: 151}
	for _, cv := range cvs {
		if expected[cv.Number] != cv.Value {
			t.Fatalf("unexpected CV value: %#v", cvs)
		}
	}
}

func startECoSTestServer(t *testing.T, handler func(command string) []string) net.Listener {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen failed: %v", err)
	}
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go func(conn net.Conn) {
				defer func() { _ = conn.Close() }()
				reader := bufio.NewReader(conn)
				for {
					line, err := reader.ReadString('\n')
					if err != nil {
						return
					}
					for _, responseLine := range handler(strings.TrimSpace(line)) {
						_, _ = fmt.Fprint(conn, responseLine+"\r\n")
					}
				}
			}(conn)
		}
	}()
	return listener
}

func splitTestAddress(t *testing.T, address string) (string, int) {
	t.Helper()
	host, portText, err := net.SplitHostPort(address)
	if err != nil {
		t.Fatalf("split host port: %v", err)
	}
	var port int
	if _, err := fmt.Sscanf(portText, "%d", &port); err != nil {
		t.Fatalf("parse port: %v", err)
	}
	return host, port
}
