package application

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	defaultZ21Port          = 21105
	defaultCS3Port          = 80
	z21MaximumUDPPayload    = 1472
	z21MaximumDatagrams     = 32
	z21MaximumReadAddresses = 64
	z21LANXHeader           = 0x0040
	z21LANGetSerialNumber   = 0x0010
	z21LANGetCode           = 0x0018
	z21LANGetHWInfo         = 0x001A
)

type DigitalCenterService struct {
	timeout        time.Duration
	client         *http.Client
	cs3Resolver    digitalCenterIPResolver
	cs3DialContext func(context.Context, string, string) (net.Conn, error)
}

type DigitalCenterServiceOption func(*DigitalCenterService)

// WithCS3DialContext injects the trusted socket boundary used after CS3 target validation.
func WithCS3DialContext(
	dialContext func(context.Context, string, string) (net.Conn, error),
) DigitalCenterServiceOption {
	return func(service *DigitalCenterService) {
		if dialContext != nil {
			service.cs3DialContext = dialContext
		}
	}
}

type DigitalCenterConnectionInput struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

type DigitalCenterConnectionResult struct {
	Provider  string            `json:"provider"`
	Connected bool              `json:"connected"`
	Host      string            `json:"host"`
	Port      int               `json:"port"`
	Status    string            `json:"status,omitempty"`
	Message   string            `json:"message"`
	Fields    map[string]string `json:"fields,omitempty"`
}

type DigitalCenterProbeResult struct {
	Provider  string                            `json:"provider"`
	Connected bool                              `json:"connected"`
	Host      string                            `json:"host"`
	Port      int                               `json:"port"`
	Message   string                            `json:"message"`
	Fields    map[string]string                 `json:"fields,omitempty"`
	Commands  []DigitalCenterProbeCommandResult `json:"commands"`
}

type DigitalCenterProbeCommandResult struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	CommandHex  string            `json:"commandHex"`
	Request     string            `json:"request,omitempty"`
	ResponseHex string            `json:"responseHex,omitempty"`
	Header      string            `json:"header,omitempty"`
	PayloadHex  string            `json:"payloadHex,omitempty"`
	OK          bool              `json:"ok"`
	Error       string            `json:"error,omitempty"`
	Fields      map[string]string `json:"fields,omitempty"`
}

type z21ProbeCommand struct {
	name        string
	description string
	header      uint16
}

type z21Packet struct {
	Raw     []byte
	Header  uint16
	Payload []byte
}

func NewDigitalCenterService(options ...DigitalCenterServiceOption) *DigitalCenterService {
	timeout := 4 * time.Second
	dialer := &net.Dialer{Timeout: timeout}
	service := &DigitalCenterService{
		timeout:        timeout,
		client:         &http.Client{Timeout: timeout},
		cs3Resolver:    net.DefaultResolver,
		cs3DialContext: dialer.DialContext,
	}
	for _, option := range options {
		if option != nil {
			option(service)
		}
	}
	return service
}

func (s *DigitalCenterService) TestZ21Connection(ctx context.Context, input DigitalCenterConnectionInput) (*DigitalCenterConnectionResult, error) {
	return s.testZ21CompatibleConnection(ctx, input, "z21", "Z21")
}

func (s *DigitalCenterService) TestIntellibox3Connection(ctx context.Context, input DigitalCenterConnectionInput) (*DigitalCenterConnectionResult, error) {
	return s.testZ21CompatibleConnection(ctx, input, "intellibox3", "Intellibox 3")
}

func (s *DigitalCenterService) ProbeZ21Connection(ctx context.Context, input DigitalCenterConnectionInput) (*DigitalCenterProbeResult, error) {
	return s.probeZ21CompatibleConnection(ctx, input, "z21", "Z21")
}

func (s *DigitalCenterService) ProbeIntellibox3Connection(ctx context.Context, input DigitalCenterConnectionInput) (*DigitalCenterProbeResult, error) {
	return s.probeZ21CompatibleConnection(ctx, input, "intellibox3", "Intellibox 3")
}

// ReadZ21Locomotives polls only decoder addresses already known to RailKeeper.
// LAN_X_GET_LOCO_INFO is a status query, not a roster endpoint. Dynamic driving
// and function state is deliberately discarded at this boundary.
func (s *DigitalCenterService) ReadZ21Locomotives(
	ctx context.Context,
	input DigitalCenterConnectionInput,
	addresses []int,
) ([]ECoSRawLocomotive, error) {
	target, err := normalizeDigitalCenterInput(input, defaultZ21Port)
	if err != nil {
		return nil, err
	}
	addresses, err = normalizeZ21ReadAddresses(addresses)
	if err != nil {
		return nil, err
	}
	if len(addresses) == 0 {
		return nil, fmt.Errorf("%w: keine bekannte Z21-Decoderadresse in RailKeeper", ErrDigitalCenterDeviceOutput)
	}
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	locomotives := make([]ECoSRawLocomotive, 0, len(addresses))
	for _, address := range addresses {
		locomotive, readErr := s.readZ21LocoInfo(ctx, target, address)
		if readErr != nil {
			return nil, fmt.Errorf("Z21-Adresse %d lesen: %w", address, readErr)
		}
		locomotives = append(locomotives, locomotive)
	}
	return locomotives, nil
}

func (s *DigitalCenterService) testZ21CompatibleConnection(ctx context.Context, input DigitalCenterConnectionInput, provider, label string) (*DigitalCenterConnectionResult, error) {
	target, err := normalizeDigitalCenterInput(input, defaultZ21Port)
	if err != nil {
		return nil, err
	}
	result := &DigitalCenterConnectionResult{
		Provider: provider,
		Host:     target.Host,
		Port:     target.Port,
		Message:  fmt.Sprintf("%s nicht erreichbar.", label),
		Fields:   z21TransportFields(provider),
	}
	response, err := s.exchangeZ21UDP(ctx, target, z21SerialNumberCommand())
	if err != nil {
		result.Message = fmt.Sprintf("%s nicht erreichbar: %v", label, err)
		return result, nil //nolint:nilerr // Connection failures are returned as preview results.
	}
	header, payload, err := parseZ21Packet(response)
	if err != nil {
		result.Message = err.Error()
		return result, nil //nolint:nilerr // Protocol failures are returned as preview results.
	}
	if err := validateZ21PayloadLength(header, payload); err != nil {
		result.Message = err.Error()
		return result, nil //nolint:nilerr // Protocol failures are returned as preview results.
	}
	result.Connected = true
	result.Status = fmt.Sprintf("0x%04X", header)
	result.Message = fmt.Sprintf("%s-Verbindung erfolgreich.", label)
	if header == z21LANGetSerialNumber {
		result.Fields["serialNumber"] = strconv.FormatUint(uint64(binary.LittleEndian.Uint32(payload[:4])), 10)
	}
	return result, nil
}

func (s *DigitalCenterService) probeZ21CompatibleConnection(ctx context.Context, input DigitalCenterConnectionInput, provider, label string) (*DigitalCenterProbeResult, error) {
	target, err := normalizeDigitalCenterInput(input, defaultZ21Port)
	if err != nil {
		return nil, err
	}
	result := &DigitalCenterProbeResult{
		Provider: provider,
		Host:     target.Host,
		Port:     target.Port,
		Message:  fmt.Sprintf("%s-Diagnose ohne Antwort.", label),
		Fields:   z21TransportFields(provider),
		Commands: []DigitalCenterProbeCommandResult{},
	}
	for _, probe := range []z21ProbeCommand{
		{name: "LAN_GET_SERIAL_NUMBER", description: "Seriennummer der Zentrale", header: z21LANGetSerialNumber},
		{name: "LAN_GET_CODE", description: "Z21-Zentralencode bzw. Protokollkennung", header: z21LANGetCode},
		{name: "LAN_GET_HWINFO", description: "Hardware- und Firmware-Rohdaten", header: z21LANGetHWInfo},
	} {
		command := z21HeaderCommand(probe.header)
		commandResult := DigitalCenterProbeCommandResult{
			Name:        probe.name,
			Description: probe.description,
			CommandHex:  formatHex(command),
			Fields:      map[string]string{},
		}
		response, err := s.exchangeZ21UDP(ctx, target, command)
		if err != nil {
			commandResult.Error = err.Error()
			result.Commands = append(result.Commands, commandResult)
			continue
		}
		commandResult.ResponseHex = formatHex(response)
		header, payload, err := parseZ21Packet(response)
		if err != nil {
			commandResult.Error = err.Error()
			result.Commands = append(result.Commands, commandResult)
			continue
		}
		if err := validateZ21PayloadLength(header, payload); err != nil {
			commandResult.Error = err.Error()
			result.Commands = append(result.Commands, commandResult)
			continue
		}
		commandResult.OK = true
		commandResult.Header = fmt.Sprintf("0x%04X", header)
		commandResult.PayloadHex = formatHex(payload)
		commandResult.Fields = decodeZ21ProbeFields(header, payload)
		for key, value := range commandResult.Fields {
			result.Fields[key] = value
		}
		result.Connected = true
		result.Commands = append(result.Commands, commandResult)
	}
	if result.Connected {
		result.Message = fmt.Sprintf("%s-Diagnose abgeschlossen.", label)
	} else if len(result.Commands) > 0 {
		result.Message = fmt.Sprintf("%s-Diagnose ohne verwertbare Antwort.", label)
	}
	return result, nil
}

func (s *DigitalCenterService) TestCS3Connection(ctx context.Context, input DigitalCenterConnectionInput) (*DigitalCenterConnectionResult, error) {
	target, err := normalizeDigitalCenterInput(input, defaultCS3Port)
	if err != nil {
		return nil, err
	}
	result := &DigitalCenterConnectionResult{
		Provider: "cs3",
		Host:     target.Host,
		Port:     target.Port,
		Message:  "CS3 nicht erreichbar.",
		Fields:   map[string]string{},
	}
	locomotives, metadata, readErr := s.readCS3Locomotives(ctx, target)
	if readErr != nil {
		result.Message = cs3UserMessage(readErr)
		result.Fields["errorKind"] = string(cs3ErrorKindOf(readErr))
		if metadata.HTTPStatus != "" {
			result.Status = metadata.HTTPStatus
		}
		return result, nil
	}
	result.Connected = true
	result.Status = metadata.HTTPStatus
	result.Message = "CS3-Verbindung erfolgreich. Kompatible read-only Loklisten-API erkannt."
	result.Fields = cs3DiagnosticFields(metadata, len(locomotives))
	return result, nil
}

func (s *DigitalCenterService) exchangeZ21UDP(ctx context.Context, target DigitalCenterConnectionInput, command []byte) ([]byte, error) {
	if len(command) < 4 {
		return nil, errors.New("Z21-Befehl war unvollständig")
	}
	expectedHeader := binary.LittleEndian.Uint16(command[2:4])
	return s.exchangeZ21UDPMatching(ctx, target, command, func(packet z21Packet) (bool, error) {
		if packet.Header != expectedHeader {
			return false, nil
		}
		if err := validateZ21PayloadLength(packet.Header, packet.Payload); err != nil {
			return false, err
		}
		return true, nil
	})
}

func (s *DigitalCenterService) exchangeZ21UDPMatching(
	ctx context.Context,
	target DigitalCenterConnectionInput,
	command []byte,
	matcher func(z21Packet) (bool, error),
) ([]byte, error) {
	if len(command) < 4 {
		return nil, errors.New("Z21-Befehl war unvollständig")
	}
	expectedHeader := binary.LittleEndian.Uint16(command[2:4])
	timeout := s.timeout
	if timeout <= 0 {
		timeout = 4 * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	dialer := net.Dialer{Timeout: timeout}
	conn, err := dialer.DialContext(ctx, "udp", net.JoinHostPort(target.Host, strconv.Itoa(target.Port)))
	if err != nil {
		return nil, err
	}
	defer func() { _ = conn.Close() }()
	if deadline, ok := ctx.Deadline(); ok {
		_ = conn.SetDeadline(deadline)
	} else {
		_ = conn.SetDeadline(time.Now().Add(timeout))
	}
	if _, err := conn.Write(command); err != nil {
		return nil, err
	}
	buffer := make([]byte, z21MaximumUDPPayload+1)
	unmatchedResponse := false
	for datagram := 0; datagram < z21MaximumDatagrams; datagram++ {
		count, readErr := conn.Read(buffer)
		if readErr != nil {
			if unmatchedResponse {
				return nil, fmt.Errorf("keine passende Z21-Antwort für Header 0x%04X erhalten", expectedHeader)
			}
			return nil, readErr
		}
		if count > z21MaximumUDPPayload {
			return nil, fmt.Errorf("Z21-Antwort überschritt die maximale UDP-Nutzdatenlänge von %d Byte", z21MaximumUDPPayload)
		}
		packets, parseErr := parseZ21Packets(buffer[:count])
		if parseErr != nil {
			return nil, parseErr
		}
		for _, packet := range packets {
			matched, matchErr := matcher(packet)
			if matchErr != nil {
				return nil, matchErr
			}
			if !matched {
				unmatchedResponse = true
				continue
			}
			return append([]byte(nil), packet.Raw...), nil
		}
	}
	return nil, fmt.Errorf("keine passende Z21-Antwort für Header 0x%04X nach %d UDP-Datagrammen erhalten", expectedHeader, z21MaximumDatagrams)
}

func (s *DigitalCenterService) readZ21LocoInfo(
	ctx context.Context,
	target DigitalCenterConnectionInput,
	address int,
) (ECoSRawLocomotive, error) {
	command, err := z21LocoInfoCommand(address)
	if err != nil {
		return ECoSRawLocomotive{}, err
	}
	var result ECoSRawLocomotive
	mismatch := ""
	_, err = s.exchangeZ21UDPMatching(ctx, target, command, func(packet z21Packet) (bool, error) {
		if packet.Header != z21LANXHeader {
			return false, nil
		}
		if len(packet.Payload) == 0 {
			return false, errors.New("Z21-Lokantwort war unvollständig: X-Header fehlt")
		}
		if packet.Payload[0] != 0xEF {
			mismatch = fmt.Sprintf("unerwarteter Z21-X-BUS-Nachrichtentyp 0x%02X", packet.Payload[0])
			return false, nil
		}
		locomotive, parseErr := parseZ21LocoInfo(packet.Payload)
		if parseErr != nil {
			return false, parseErr
		}
		if locomotive.Address != address {
			mismatch = fmt.Sprintf("Z21-Lokantwort für unerwartete Adresse %d", locomotive.Address)
			return false, nil
		}
		result = locomotive
		return true, nil
	})
	if err != nil {
		if mismatch != "" {
			return ECoSRawLocomotive{}, fmt.Errorf("%s: %w", mismatch, err)
		}
		return ECoSRawLocomotive{}, err
	}
	return result, nil
}

func (s *DigitalCenterService) httpClient() *http.Client {
	if s.client == nil {
		timeout := s.timeout
		if timeout <= 0 {
			timeout = 4 * time.Second
		}
		s.client = &http.Client{Timeout: timeout}
	}
	return s.client
}

func normalizeDigitalCenterInput(input DigitalCenterConnectionInput, defaultPort int) (DigitalCenterConnectionInput, error) {
	host := strings.TrimSpace(input.Host)
	if parsed, err := url.Parse(host); err == nil && parsed.Hostname() != "" {
		host = parsed.Hostname()
		if input.Port == 0 && parsed.Port() != "" {
			input.Port, _ = strconv.Atoi(parsed.Port())
		}
	}
	if host == "" {
		return DigitalCenterConnectionInput{}, errors.New("IP oder Hostname der Digitalzentrale fehlt")
	}
	port := input.Port
	if port == 0 {
		port = defaultPort
	}
	if port < 1 || port > 65535 {
		return DigitalCenterConnectionInput{}, errors.New("Port muss zwischen 1 und 65535 liegen") //nolint:staticcheck // User-facing German validation text.
	}
	return DigitalCenterConnectionInput{Host: host, Port: port}, nil
}

func z21SerialNumberCommand() []byte {
	return z21HeaderCommand(z21LANGetSerialNumber)
}

func z21HeaderCommand(header uint16) []byte {
	packet := make([]byte, 4)
	binary.LittleEndian.PutUint16(packet[0:2], 4)
	binary.LittleEndian.PutUint16(packet[2:4], header)
	return packet
}

func z21LocoInfoCommand(address int) ([]byte, error) {
	if address < 1 || address > maxDigitalCenterAddress {
		return nil, fmt.Errorf("Z21-Decoderadresse %d liegt außerhalb 1 bis %d", address, maxDigitalCenterAddress)
	}
	addressMSB := byte(address >> 8)
	if address >= 128 {
		addressMSB |= 0xC0
	}
	command := []byte{0x09, 0x00, 0x40, 0x00, 0xE3, 0xF0, addressMSB, byte(address), 0x00}
	for _, value := range command[4 : len(command)-1] {
		command[len(command)-1] ^= value
	}
	return command, nil
}

func normalizeZ21ReadAddresses(addresses []int) ([]int, error) {
	unique := make(map[int]struct{}, len(addresses))
	for _, address := range addresses {
		if address < 1 || address > maxDigitalCenterAddress {
			return nil, fmt.Errorf("Z21-Decoderadresse %d liegt außerhalb 1 bis %d", address, maxDigitalCenterAddress)
		}
		unique[address] = struct{}{}
	}
	if len(unique) > z21MaximumReadAddresses {
		return nil, fmt.Errorf("Z21-Leseumfang ist auf %d eindeutige Decoderadressen begrenzt", z21MaximumReadAddresses)
	}
	result := make([]int, 0, len(unique))
	for address := range unique {
		result = append(result, address)
	}
	sort.Ints(result)
	return result, nil
}

func parseZ21LocoInfo(payload []byte) (ECoSRawLocomotive, error) {
	if len(payload) < 10 {
		return ECoSRawLocomotive{}, fmt.Errorf("Z21-Lokantwort war unvollständig: erwartete mindestens 10 Nutzdatenbytes, erhielt %d", len(payload))
	}
	if len(payload) > 17 {
		return ECoSRawLocomotive{}, fmt.Errorf("Z21-Lokantwort hatte eine ungültige Nutzdatenlänge von %d Byte", len(payload))
	}
	if payload[0] != 0xEF {
		return ECoSRawLocomotive{}, fmt.Errorf("unerwarteter Z21-X-BUS-Nachrichtentyp 0x%02X", payload[0])
	}
	checksum := byte(0)
	for _, value := range payload {
		checksum ^= value
	}
	if checksum != 0 {
		return ECoSRawLocomotive{}, errors.New("Z21-Lokantwort hatte eine ungültige XOR-Prüfsumme")
	}
	address := int(payload[1]&0x3F)<<8 | int(payload[2])
	if address < 1 || address > maxDigitalCenterAddress {
		return ECoSRawLocomotive{}, fmt.Errorf("Z21-Lokantwort enthielt ungültige Decoderadresse %d", address)
	}
	return ECoSRawLocomotive{
		ObjectID: address,
		Address:  address,
		DetailError: "Z21-LAN stellt weder Loknamen noch ein verlässlich eindeutiges Decoderprotokoll bereit; " +
			"Fahrzustände wurden verworfen.",
	}, nil
}

func decodeZ21ProbeFields(header uint16, payload []byte) map[string]string {
	fields := map[string]string{}
	switch header {
	case z21LANGetSerialNumber:
		if len(payload) == 4 {
			fields["serialNumber"] = strconv.FormatUint(uint64(binary.LittleEndian.Uint32(payload[:4])), 10)
		}
	case z21LANGetCode:
		if len(payload) == 1 {
			fields["centralCode"] = fmt.Sprintf("%d (0x%02X)", payload[0], payload[0])
		}
	case z21LANGetHWInfo:
		if len(payload) == 8 {
			fields["hardwareTypeRaw"] = fmt.Sprintf("0x%08X", binary.LittleEndian.Uint32(payload[:4]))
			fields["firmwareVersionRaw"] = fmt.Sprintf("0x%08X", binary.LittleEndian.Uint32(payload[4:8]))
		}
	}
	return fields
}

func z21TransportFields(provider string) map[string]string {
	fields := map[string]string{"transport": "z21_udp"}
	if provider == "intellibox3" {
		fields["z21UdpStatus"] = "available"
		fields["loconetTcpStatus"] = "planned"
	}
	return fields
}

func validateZ21PayloadLength(header uint16, payload []byte) error {
	expected := 0
	switch header {
	case z21LANGetSerialNumber:
		expected = 4
	case z21LANGetCode:
		expected = 1
	case z21LANGetHWInfo:
		expected = 8
	default:
		return nil
	}
	if len(payload) != expected {
		return fmt.Errorf("Z21-Antwort für Header 0x%04X erwartete Nutzdatenlänge %d, erhielt %d", header, expected, len(payload))
	}
	return nil
}

func formatHex(data []byte) string {
	if len(data) == 0 {
		return ""
	}
	return strings.ToUpper(hex.EncodeToString(data))
}

func parseZ21Packet(packet []byte) (uint16, []byte, error) {
	packets, err := parseZ21Packets(packet)
	if err != nil {
		return 0, nil, err
	}
	if len(packets) != 1 {
		return 0, nil, errors.New("Z21-Antwort enthielt unerwartet mehrere Datensätze")
	}
	return packets[0].Header, packets[0].Payload, nil
}

func parseZ21Packets(datagram []byte) ([]z21Packet, error) {
	if len(datagram) > z21MaximumUDPPayload {
		return nil, fmt.Errorf("Z21-Antwort überschritt die maximale UDP-Nutzdatenlänge von %d Byte", z21MaximumUDPPayload)
	}
	if len(datagram) < 4 {
		return nil, errors.New("Z21-Antwort war unvollständig")
	}
	packets := make([]z21Packet, 0, 1)
	for offset := 0; offset < len(datagram); {
		if len(datagram)-offset < 4 {
			return nil, errors.New("Z21-Antwort enthielt unvollständige nachfolgende Daten")
		}
		length := int(binary.LittleEndian.Uint16(datagram[offset : offset+2]))
		if length < 4 || length > len(datagram)-offset {
			return nil, errors.New("Z21-Antwort hatte eine ungültige Datensatzlänge")
		}
		raw := datagram[offset : offset+length]
		packets = append(packets, z21Packet{
			Raw:     raw,
			Header:  binary.LittleEndian.Uint16(raw[2:4]),
			Payload: raw[4:],
		})
		offset += length
	}
	return packets, nil
}
