package application

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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
	timeout time.Duration
	client  *http.Client
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

func NewDigitalCenterService() *DigitalCenterService {
	timeout := 4 * time.Second
	return &DigitalCenterService{
		timeout: timeout,
		client:  &http.Client{Timeout: timeout},
	}
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
		return result, nil
	}
	header, payload, err := parseZ21Packet(response)
	if err != nil {
		result.Message = err.Error()
		return result, nil
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
	for _, path := range []string{"/app/api/version", "/"} {
		status, fields, err := s.fetchCS3HTTP(ctx, target, path)
		if err != nil {
			result.Message = fmt.Sprintf("CS3 nicht erreichbar: %v", err)
			continue
		}
		result.Connected = true
		result.Status = status
		result.Message = "CS3-Verbindung erfolgreich."
		for key, value := range fields {
			result.Fields[key] = value
		}
		return result, nil
	}
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

func (s *DigitalCenterService) fetchCS3HTTP(ctx context.Context, target DigitalCenterConnectionInput, requestPath string) (string, map[string]string, error) {
	endpoint := url.URL{
		Scheme: "http",
		Host:   net.JoinHostPort(target.Host, strconv.Itoa(target.Port)),
		Path:   requestPath,
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return "", nil, err
	}
	request.Header.Set("Accept", "application/json,text/plain,*/*")
	request.Header.Set("User-Agent", "RailKeeper")
	response, err := s.httpClient().Do(request)
	if err != nil {
		return "", nil, err
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode < 200 || response.StatusCode >= 500 {
		return response.Status, nil, fmt.Errorf("HTTP %s", response.Status)
	}
	fields := map[string]string{}
	data, _ := io.ReadAll(io.LimitReader(response.Body, 64*1024))
	var document map[string]any
	if json.Unmarshal(data, &document) == nil {
		for key, value := range document {
			switch typed := value.(type) {
			case string:
				fields[key] = typed
			case float64:
				fields[key] = strconv.FormatFloat(typed, 'f', -1, 64)
			case bool:
				fields[key] = strconv.FormatBool(typed)
			}
		}
	}
	return response.Status, fields, nil
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
		return DigitalCenterConnectionInput{}, errors.New("Port muss zwischen 1 und 65535 liegen")
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
