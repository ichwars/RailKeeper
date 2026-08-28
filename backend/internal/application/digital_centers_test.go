package application

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"
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
	if result.Fields["transport"] != "z21_udp" || result.Fields["loconetTcpStatus"] != "planned" {
		t.Fatalf("unexpected Intellibox 3 transport fields: %#v", result.Fields)
	}
	command := <-received
	if len(command) != 4 || binary.LittleEndian.Uint16(command[2:4]) != z21LANGetSerialNumber {
		t.Fatalf("unexpected intellibox 3 command: %#v", command)
	}
}

func TestDigitalCenterServiceTestIntellibox3ConnectionSkipsAsyncDatasets(t *testing.T) {
	conn, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen udp: %v", err)
	}
	defer func() { _ = conn.Close() }()

	go func() {
		buffer := make([]byte, 128)
		_, address, readErr := conn.ReadFrom(buffer)
		if readErr != nil {
			return
		}
		asyncCode := z21TestResponse(z21LANGetCode, []byte{0x00})
		serial := z21TestResponse(z21LANGetSerialNumber, []byte{0x40, 0xE2, 0x01, 0x00})
		_, _ = conn.WriteTo(append(asyncCode, serial...), address)
	}()

	host, port := splitDigitalCenterTestAddress(t, conn.LocalAddr().String())
	result, err := NewDigitalCenterService().TestIntellibox3Connection(
		context.Background(), DigitalCenterConnectionInput{Host: host, Port: port},
	)
	if err != nil {
		t.Fatalf("test intellibox 3 failed: %v", err)
	}
	if !result.Connected || result.Fields["serialNumber"] != "123456" {
		t.Fatalf("combined asynchronous response was not matched safely: %#v", result)
	}
}

func TestDigitalCenterServiceTestIntellibox3ConnectionRejectsWrongResponseHeader(t *testing.T) {
	conn, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen udp: %v", err)
	}
	defer func() { _ = conn.Close() }()

	go func() {
		buffer := make([]byte, 128)
		_, address, readErr := conn.ReadFrom(buffer)
		if readErr == nil {
			_, _ = conn.WriteTo(z21TestResponse(z21LANGetCode, []byte{0x00}), address)
		}
	}()

	host, port := splitDigitalCenterTestAddress(t, conn.LocalAddr().String())
	service := NewDigitalCenterService()
	service.timeout = 50 * time.Millisecond
	result, err := service.TestIntellibox3Connection(
		context.Background(), DigitalCenterConnectionInput{Host: host, Port: port},
	)
	if err != nil {
		t.Fatalf("test intellibox 3 failed: %v", err)
	}
	if result.Connected || !strings.Contains(result.Message, "passende Z21-Antwort") {
		t.Fatalf("wrong response header must remain disconnected with actionable message: %#v", result)
	}
}

func TestDigitalCenterServiceTestIntellibox3ConnectionRejectsInvalidSerialPayloadLength(t *testing.T) {
	conn, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen udp: %v", err)
	}
	defer func() { _ = conn.Close() }()

	go func() {
		buffer := make([]byte, 128)
		_, address, readErr := conn.ReadFrom(buffer)
		if readErr == nil {
			_, _ = conn.WriteTo(z21TestResponse(z21LANGetSerialNumber, []byte{0x01, 0x02, 0x03}), address)
		}
	}()

	host, port := splitDigitalCenterTestAddress(t, conn.LocalAddr().String())
	result, err := NewDigitalCenterService().TestIntellibox3Connection(
		context.Background(), DigitalCenterConnectionInput{Host: host, Port: port},
	)
	if err != nil {
		t.Fatalf("test intellibox 3 failed: %v", err)
	}
	if result.Connected || !strings.Contains(result.Message, "Nutzdatenlänge 4") {
		t.Fatalf("invalid serial payload must remain disconnected with actionable message: %#v", result)
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
			payload := map[uint16][]byte{
				z21LANGetSerialNumber: {0x01, 0x00, 0x00, 0x00},
				z21LANGetCode:         {0x01},
				z21LANGetHWInfo:       {0x00, 0x02, 0x00, 0x00, 0x20, 0x01, 0x00, 0x00},
			}[header]
			_, _ = conn.WriteTo(z21TestResponse(header, payload), address)
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
	if result.Fields["transport"] != "z21_udp" || result.Fields["loconetTcpStatus"] != "planned" {
		t.Fatalf("unexpected Intellibox 3 probe transport fields: %#v", result.Fields)
	}
}

func TestParseZ21ProtocolV113Fixtures(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata", "z21_protocol_v1_13.json"))
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		DatagramHex     string   `json:"datagramHex"`
		ExpectedHeaders []uint16 `json:"expectedHeaders"`
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	datagram, err := hex.DecodeString(fixture.DatagramHex)
	if err != nil {
		t.Fatal(err)
	}
	packets, err := parseZ21Packets(datagram)
	if err != nil {
		t.Fatal(err)
	}
	if len(packets) != len(fixture.ExpectedHeaders) {
		t.Fatalf("packet count = %d, want %d", len(packets), len(fixture.ExpectedHeaders))
	}
	for index, packet := range packets {
		if packet.Header != fixture.ExpectedHeaders[index] {
			t.Fatalf("packet %d header = 0x%04X, want 0x%04X", index, packet.Header, fixture.ExpectedHeaders[index])
		}
	}
}

func TestParseZ21LocoInfoProtocolV113Fixture(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata", "z21_loco_info_protocol_v1_13.json"))
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		HardwareCapture bool   `json:"hardwareCapture"`
		Address         int    `json:"address"`
		RequestHex      string `json:"requestHex"`
		ResponseHex     string `json:"responseHex"`
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.HardwareCapture {
		t.Fatal("protocol-derived fixture must not claim to be a hardware capture")
	}
	request, err := hex.DecodeString(fixture.RequestHex)
	if err != nil {
		t.Fatal(err)
	}
	wantRequest, err := z21LocoInfoCommand(fixture.Address)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(request, wantRequest) {
		t.Fatalf("request = % X, want % X", request, wantRequest)
	}
	response, err := hex.DecodeString(fixture.ResponseHex)
	if err != nil {
		t.Fatal(err)
	}
	packets, err := parseZ21Packets(response)
	if err != nil {
		t.Fatal(err)
	}
	locomotive, err := parseZ21LocoInfo(packets[0].Payload)
	if err != nil {
		t.Fatal(err)
	}
	if locomotive.Address != fixture.Address || locomotive.Name != "" || locomotive.Protocol != "" {
		t.Fatalf("locomotive = %#v, want address-only fixture", locomotive)
	}
}

func TestZ21LocoInfoCommandUsesDocumentedAddressEncodingAndXOR(t *testing.T) {
	command, err := z21LocoInfoCommand(2000)
	if err != nil {
		t.Fatal(err)
	}
	want := []byte{0x09, 0x00, 0x40, 0x00, 0xE3, 0xF0, 0xC7, 0xD0, 0x04}
	if !reflect.DeepEqual(command, want) {
		t.Fatalf("command = % X, want % X", command, want)
	}
}

func TestDigitalCenterServiceReadsOnlyMatchingZ21LocoInfoAddress(t *testing.T) {
	conn, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen udp: %v", err)
	}
	defer func() { _ = conn.Close() }()

	received := make(chan []byte, 1)
	go func() {
		buffer := make([]byte, 128)
		count, address, readErr := conn.ReadFrom(buffer)
		if readErr != nil {
			return
		}
		received <- append([]byte(nil), buffer[:count]...)
		unexpectedType := z21XBusTestResponse([]byte{0x61, 0x01})
		wrongAddress := z21LocoInfoTestResponse(4, 0x04, 0x7F)
		matching := z21LocoInfoTestResponse(3, 0x00, 0x01)
		_, _ = conn.WriteTo(append(append(unexpectedType, wrongAddress...), matching...), address)
	}()

	host, port := splitDigitalCenterTestAddress(t, conn.LocalAddr().String())
	locomotives, err := NewDigitalCenterService().ReadZ21Locomotives(
		context.Background(),
		DigitalCenterConnectionInput{Host: host, Port: port},
		[]int{3},
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(locomotives) != 1 || locomotives[0].ObjectID != 3 || locomotives[0].Address != 3 ||
		locomotives[0].Name != "" || locomotives[0].Protocol != "" || locomotives[0].DetailError == "" {
		t.Fatalf("locomotives = %#v, want address-only read result", locomotives)
	}
	command := <-received
	if len(command) != 9 || command[4] != 0xE3 || command[5] != 0xF0 {
		t.Fatalf("request = % X, want LAN_X_GET_LOCO_INFO only", command)
	}
}

func TestDigitalCenterServiceRejectsInvalidZ21LocoInfoChecksum(t *testing.T) {
	conn, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen udp: %v", err)
	}
	defer func() { _ = conn.Close() }()

	go func() {
		buffer := make([]byte, 128)
		_, address, readErr := conn.ReadFrom(buffer)
		if readErr == nil {
			response := z21LocoInfoTestResponse(3, 0x00, 0x01)
			response[len(response)-1] ^= 0xFF
			_, _ = conn.WriteTo(response, address)
		}
	}()

	host, port := splitDigitalCenterTestAddress(t, conn.LocalAddr().String())
	_, err = NewDigitalCenterService().ReadZ21Locomotives(
		context.Background(), DigitalCenterConnectionInput{Host: host, Port: port}, []int{3},
	)
	if err == nil || !strings.Contains(err.Error(), "XOR") {
		t.Fatalf("error = %v, want distinct checksum error", err)
	}
}

func TestDigitalCenterServiceDistinguishesZ21LocoInfoFailures(t *testing.T) {
	tests := []struct {
		name       string
		response   []byte
		wantText   string
		noResponse bool
	}{
		{name: "timeout", wantText: "timeout", noResponse: true},
		{name: "truncated", response: z21XBusTestResponse([]byte{0xEF, 0x00, 0x03}), wantText: "unvollständig"},
		{name: "invalid length", response: z21XBusTestResponse([]byte{0xEF, 0x00, 0x03, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}), wantText: "ungültige Nutzdatenlänge"},
		{name: "unexpected type", response: z21XBusTestResponse([]byte{0x61, 0x01}), wantText: "unerwarteter Z21-X-BUS-Nachrichtentyp"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			conn, listenErr := net.ListenPacket("udp", "127.0.0.1:0")
			if listenErr != nil {
				t.Fatalf("listen udp: %v", listenErr)
			}
			defer func() { _ = conn.Close() }()
			go func() {
				buffer := make([]byte, 128)
				_, address, readErr := conn.ReadFrom(buffer)
				if readErr == nil && !test.noResponse {
					_, _ = conn.WriteTo(test.response, address)
				}
			}()

			host, port := splitDigitalCenterTestAddress(t, conn.LocalAddr().String())
			service := NewDigitalCenterService()
			service.timeout = 50 * time.Millisecond
			_, readErr := service.ReadZ21Locomotives(
				context.Background(), DigitalCenterConnectionInput{Host: host, Port: port}, []int{3},
			)
			if readErr == nil || !strings.Contains(strings.ToLower(readErr.Error()), strings.ToLower(test.wantText)) {
				t.Fatalf("error = %v, want %q", readErr, test.wantText)
			}
		})
	}
}

func TestDigitalCenterServiceRejectsUnboundedZ21ReadScope(t *testing.T) {
	addresses := make([]int, z21MaximumReadAddresses+1)
	for index := range addresses {
		addresses[index] = index + 1
	}
	_, err := NewDigitalCenterService().ReadZ21Locomotives(
		context.Background(), DigitalCenterConnectionInput{Host: "127.0.0.1", Port: 21105}, addresses,
	)
	if err == nil || !strings.Contains(err.Error(), "begrenzt") {
		t.Fatalf("error = %v, want bounded read-scope error", err)
	}
}

func TestDigitalCenterServiceTestCS3Connection(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/app/api/locos" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]map[string]any{{
			"name": "BR 218", "uid": "0xc003", "address": 3, "dectyp": "dcc",
		}})
	}))
	defer server.Close()

	service, input := newCS3TestService(t, server)
	result, err := service.TestCS3Connection(context.Background(), input)
	if err != nil {
		t.Fatalf("test cs3 failed: %v", err)
	}
	if !result.Connected || result.Provider != "cs3" || result.Fields["apiGeneration"] != "2.6+" ||
		result.Fields["locomotiveCount"] != "1" {
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

func z21XBusTestResponse(payload []byte) []byte {
	checksum := byte(0)
	for _, value := range payload {
		checksum ^= value
	}
	return z21TestResponse(z21LANXHeader, append(append([]byte(nil), payload...), checksum))
}

func z21LocoInfoTestResponse(address int, format, speed byte) []byte {
	addressMSB := byte(address >> 8)
	if address >= 128 {
		addressMSB |= 0xC0
	}
	return z21XBusTestResponse([]byte{
		0xEF, addressMSB, byte(address), format, speed, 0x10, 0x20, 0x40, 0x80,
	})
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
