package application

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

func TestDigitalCenterServiceTestZ21Connection(t *testing.T) {
	conn, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen udp: %v", err)
	}
	defer func() { _ = conn.Close() }()

	received := make(chan []byte, 1)
	go func() {
		buffer := make([]byte, 128)
		count, address, err := conn.ReadFrom(buffer)
		if err != nil {
			return
		}
		received <- append([]byte(nil), buffer[:count]...)
		response := make([]byte, 8)
		binary.LittleEndian.PutUint16(response[0:2], 8)
		binary.LittleEndian.PutUint16(response[2:4], z21LANGetSerialNumber)
		binary.LittleEndian.PutUint32(response[4:8], 123456)
		_, _ = conn.WriteTo(response, address)
	}()

	host, port := splitDigitalCenterTestAddress(t, conn.LocalAddr().String())
	service := NewDigitalCenterService()
	result, err := service.TestZ21Connection(context.Background(), DigitalCenterConnectionInput{Host: host, Port: port})
	if err != nil {
		t.Fatalf("test z21 failed: %v", err)
	}
	if !result.Connected || result.Provider != "z21" || result.Fields["serialNumber"] != "123456" {
		t.Fatalf("unexpected result: %#v", result)
	}
	command := <-received
	if len(command) != 4 || binary.LittleEndian.Uint16(command[2:4]) != z21LANGetSerialNumber {
		t.Fatalf("unexpected z21 command: %#v", command)
	}
}

func TestDigitalCenterServiceTestIntellibox3ConnectionUsesZ21CompatibleUDP(t *testing.T) {
	conn, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen udp: %v", err)
	}
	defer func() { _ = conn.Close() }()

	received := make(chan []byte, 1)
	go func() {
		buffer := make([]byte, 128)
		count, address, err := conn.ReadFrom(buffer)
		if err != nil {
			return
		}
		received <- append([]byte(nil), buffer[:count]...)
		response := make([]byte, 8)
		binary.LittleEndian.PutUint16(response[0:2], 8)
		binary.LittleEndian.PutUint16(response[2:4], z21LANGetSerialNumber)
		binary.LittleEndian.PutUint32(response[4:8], 987654)
		_, _ = conn.WriteTo(response, address)
	}()

	host, port := splitDigitalCenterTestAddress(t, conn.LocalAddr().String())
	service := NewDigitalCenterService()
	result, err := service.TestIntellibox3Connection(context.Background(), DigitalCenterConnectionInput{Host: host, Port: port})
	if err != nil {
		t.Fatalf("test intellibox 3 failed: %v", err)
	}
	if !result.Connected || result.Provider != "intellibox3" || result.Fields["serialNumber"] != "987654" {
		t.Fatalf("unexpected result: %#v", result)
	}
	command := <-received
	if len(command) != 4 || binary.LittleEndian.Uint16(command[2:4]) != z21LANGetSerialNumber {
		t.Fatalf("unexpected intellibox 3 command: %#v", command)
	}
}

func TestDigitalCenterServiceProbeZ21ConnectionReadsDiagnosticFields(t *testing.T) {
	conn, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen udp: %v", err)
	}
	defer func() { _ = conn.Close() }()

	received := make(chan uint16, 3)
	go func() {
		buffer := make([]byte, 128)
		for i := 0; i < 3; i++ {
			count, address, err := conn.ReadFrom(buffer)
			if err != nil {
				return
			}
			header := binary.LittleEndian.Uint16(buffer[2:4])
			received <- header
			switch header {
			case z21LANGetSerialNumber:
				_, _ = conn.WriteTo(z21TestResponse(header, []byte{0x40, 0xE2, 0x01, 0x00}), address)
			case z21LANGetCode:
				_, _ = conn.WriteTo(z21TestResponse(header, []byte{0x12}), address)
			case z21LANGetHWInfo:
				_, _ = conn.WriteTo(z21TestResponse(header, []byte{0x34, 0x12, 0x00, 0x00, 0x78, 0x56, 0x00, 0x00}), address)
			default:
				t.Errorf("unexpected probe header from %d bytes: 0x%04X", count, header)
			}
		}
	}()

	host, port := splitDigitalCenterTestAddress(t, conn.LocalAddr().String())
	service := NewDigitalCenterService()
	result, err := service.ProbeZ21Connection(context.Background(), DigitalCenterConnectionInput{Host: host, Port: port})
	if err != nil {
		t.Fatalf("probe z21 failed: %v", err)
	}
	if !result.Connected || result.Provider != "z21" || len(result.Commands) != 3 {
		t.Fatalf("unexpected probe result: %#v", result)
	}
	if result.Fields["serialNumber"] != "123456" || result.Fields["centralCode"] != "18 (0x12)" || result.Fields["hardwareTypeRaw"] != "0x00001234" || result.Fields["firmwareVersionRaw"] != "0x00005678" {
		t.Fatalf("unexpected probe fields: %#v", result.Fields)
	}
	for _, expected := range []uint16{z21LANGetSerialNumber, z21LANGetCode, z21LANGetHWInfo} {
		if actual := <-received; actual != expected {
			t.Fatalf("expected header 0x%04X, got 0x%04X", expected, actual)
		}
	}
}

func TestDigitalCenterServiceProbeIntellibox3ConnectionUsesZ21CompatibleDiagnostics(t *testing.T) {
	conn, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen udp: %v", err)
	}
	defer func() { _ = conn.Close() }()

	go func() {
		buffer := make([]byte, 128)
		for i := 0; i < 3; i++ {
			_, address, err := conn.ReadFrom(buffer)
			if err != nil {
				return
			}
			header := binary.LittleEndian.Uint16(buffer[2:4])
			_, _ = conn.WriteTo(z21TestResponse(header, []byte{0x01, 0x00, 0x00, 0x00}), address)
		}
	}()

	host, port := splitDigitalCenterTestAddress(t, conn.LocalAddr().String())
	service := NewDigitalCenterService()
	result, err := service.ProbeIntellibox3Connection(context.Background(), DigitalCenterConnectionInput{Host: host, Port: port})
	if err != nil {
		t.Fatalf("probe intellibox 3 failed: %v", err)
	}
	if !result.Connected || result.Provider != "intellibox3" || len(result.Commands) != 3 {
		t.Fatalf("unexpected probe result: %#v", result)
	}
}

func TestDigitalCenterServiceTestCS3Connection(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/app/api/version" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"version": "2.5.2",
			"name":    "CS3",
		})
	}))
	defer server.Close()

	host, port := splitDigitalCenterTestURL(t, server.URL)
	service := NewDigitalCenterService()
	result, err := service.TestCS3Connection(context.Background(), DigitalCenterConnectionInput{Host: host, Port: port})
	if err != nil {
		t.Fatalf("test cs3 failed: %v", err)
	}
	if !result.Connected || result.Provider != "cs3" || result.Fields["version"] != "2.5.2" {
		t.Fatalf("unexpected result: %#v", result)
	}
}

func z21TestResponse(header uint16, payload []byte) []byte {
	response := make([]byte, 4+len(payload))
	binary.LittleEndian.PutUint16(response[0:2], uint16(len(response)))
	binary.LittleEndian.PutUint16(response[2:4], header)
	copy(response[4:], payload)
	return response
}

func TestNormalizeDigitalCenterInputAcceptsURL(t *testing.T) {
	input, err := normalizeDigitalCenterInput(DigitalCenterConnectionInput{Host: "http://192.168.1.50:8080"}, 80)
	if err != nil {
		t.Fatalf("normalize failed: %v", err)
	}
	if input.Host != "192.168.1.50" || input.Port != 8080 {
		t.Fatalf("unexpected input: %#v", input)
	}
}

func splitDigitalCenterTestURL(t *testing.T, rawURL string) (string, int) {
	t.Helper()
	return splitDigitalCenterTestAddress(t, strings.TrimPrefix(rawURL, "http://"))
}

func splitDigitalCenterTestAddress(t *testing.T, address string) (string, int) {
	t.Helper()
	host, portText, err := net.SplitHostPort(address)
	if err != nil {
		t.Fatalf("split host port: %v", err)
	}
	port, err := strconv.Atoi(portText)
	if err != nil {
		t.Fatalf("parse port: %v", err)
	}
	return host, port
}
