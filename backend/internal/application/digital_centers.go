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
	"strconv"
	"strings"
	"time"
)

const (
	defaultZ21Port        = 21105
	defaultCS3Port        = 80
	z21LANGetSerialNumber = 0x0010
	z21LANGetCode         = 0x0018
	z21LANGetHWInfo       = 0x001A
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
		Fields:   map[string]string{},
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
	result.Connected = true
	result.Status = fmt.Sprintf("0x%04X", header)
	result.Message = fmt.Sprintf("%s-Verbindung erfolgreich.", label)
	if header == z21LANGetSerialNumber && len(payload) >= 4 {
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
		Fields:   map[string]string{},
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
	buffer := make([]byte, 1024)
	count, err := conn.Read(buffer)
	if err != nil {
		return nil, err
	}
	return buffer[:count], nil
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

func decodeZ21ProbeFields(header uint16, payload []byte) map[string]string {
	fields := map[string]string{}
	switch header {
	case z21LANGetSerialNumber:
		if len(payload) >= 4 {
			fields["serialNumber"] = strconv.FormatUint(uint64(binary.LittleEndian.Uint32(payload[:4])), 10)
		}
	case z21LANGetCode:
		if len(payload) >= 1 {
			fields["centralCode"] = fmt.Sprintf("%d (0x%02X)", payload[0], payload[0])
		}
	case z21LANGetHWInfo:
		if len(payload) >= 4 {
			fields["hardwareTypeRaw"] = fmt.Sprintf("0x%08X", binary.LittleEndian.Uint32(payload[:4]))
		}
		if len(payload) >= 8 {
			fields["firmwareVersionRaw"] = fmt.Sprintf("0x%08X", binary.LittleEndian.Uint32(payload[4:8]))
		}
	}
	return fields
}

func formatHex(data []byte) string {
	if len(data) == 0 {
		return ""
	}
	return strings.ToUpper(hex.EncodeToString(data))
}

func parseZ21Packet(packet []byte) (uint16, []byte, error) {
	if len(packet) < 4 {
		return 0, nil, errors.New("Z21-Antwort war unvollständig")
	}
	length := int(binary.LittleEndian.Uint16(packet[0:2]))
	if length < 4 || length > len(packet) {
		return 0, nil, errors.New("Z21-Antwort hatte eine ungültige Länge")
	}
	header := binary.LittleEndian.Uint16(packet[2:4])
	return header, packet[4:length], nil
}
