package application

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"strings"
	"testing"
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
	if len(locomotive.CVs) != 3 || locomotive.CVs[2].Number != 29 || locomotive.CVs[2].Value != 38 {
		t.Fatalf("expected structured CV values: %#v", locomotive.CVs)
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
