package application

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

const defaultECoSPort = 15471

type ECoSService struct {
	timeout time.Duration
}

type ECoSConnectionInput struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

type ECoSConnectionResult struct {
	Connected          bool              `json:"connected"`
	Host               string            `json:"host"`
	Port               int               `json:"port"`
	Status             string            `json:"status,omitempty"`
	ProtocolVersion    string            `json:"protocolVersion,omitempty"`
	ApplicationVersion string            `json:"applicationVersion,omitempty"`
	HardwareVersion    string            `json:"hardwareVersion,omitempty"`
	Message            string            `json:"message"`
	RawLines           []string          `json:"rawLines,omitempty"`
	Fields             map[string]string `json:"fields,omitempty"`
}

type ECoSLocomotive struct {
	ObjectID          int                 `json:"objectId"`
	Name              string              `json:"name,omitempty"`
	Address           int                 `json:"address"`
	Protocol          string              `json:"protocol,omitempty"`
	Profile           string              `json:"profile,omitempty"`
	Speed             int                 `json:"speed"`
	SpeedStep         int                 `json:"speedStep"`
	Direction         int                 `json:"direction"`
	FunctionSet       string              `json:"functionSet,omitempty"`
	NumberOfFunctions int                 `json:"numberOfFunctions,omitempty"`
	Functions         []ECoSFunction      `json:"functions,omitempty"`
	Attributes        map[string][]string `json:"attributes,omitempty"`
	DetailError       string              `json:"detailError,omitempty"`
}

type ECoSFunction struct {
	Index       int  `json:"index"`
	Active      bool `json:"active"`
	Description int  `json:"description,omitempty"`
}

type ecosArgument struct {
	Key    string
	Value  string
	Params []string
}

type ECoSRawProbe struct {
	Host        string              `json:"host"`
	Port        int                 `json:"port"`
	ProbeFields []string            `json:"probeFields"`
	Locomotives []ECoSRawLocomotive `json:"locomotives"`
	RawLines    []string            `json:"rawLines,omitempty"`
	Message     string              `json:"message"`
}

type ECoSRawLocomotive struct {
	ObjectID          int                   `json:"objectId"`
	Name              string                `json:"name,omitempty"`
	Address           int                   `json:"address,omitempty"`
	Protocol          string                `json:"protocol,omitempty"`
	Profile           string                `json:"profile,omitempty"`
	Speed             int                   `json:"speed"`
	SpeedStep         int                   `json:"speedStep"`
	Direction         int                   `json:"direction"`
	FunctionSet       string                `json:"functionSet,omitempty"`
	NumberOfFunctions int                   `json:"numberOfFunctions,omitempty"`
	Functions         []ECoSFunction        `json:"functions,omitempty"`
	CVs               []ECoSCVValue         `json:"cvs,omitempty"`
	Attributes        map[string][]string   `json:"attributes,omitempty"`
	SupportedFields   []string              `json:"supportedFields,omitempty"`
	MissingFields     []string              `json:"missingFields,omitempty"`
	InterestingFields []string              `json:"interestingFields,omitempty"`
	Probes            []ECoSRawCommandProbe `json:"probes,omitempty"`
	DetailError       string                `json:"detailError,omitempty"`
}

type ECoSRawCommandProbe struct {
	Command    string              `json:"command"`
	Fields     []string            `json:"fields"`
	OK         bool                `json:"ok"`
	Status     string              `json:"status,omitempty"`
	Error      string              `json:"error,omitempty"`
	RawLines   []string            `json:"rawLines,omitempty"`
	Attributes map[string][]string `json:"attributes,omitempty"`
}

type ECoSCVValue struct {
	Number int `json:"number"`
	Value  int `json:"value"`
}

func NewECoSService() *ECoSService {
	return &ECoSService{timeout: 5 * time.Second}
}

func (s *ECoSService) TestConnection(ctx context.Context, input ECoSConnectionInput) (*ECoSConnectionResult, error) {
	target, err := normalizeECoSInput(input)
	if err != nil {
		return nil, err
	}
	lines, err := s.exchange(ctx, target.Host, target.Port, "get(1, info, status)")
	result := &ECoSConnectionResult{
		Connected: false,
		Host:      target.Host,
		Port:      target.Port,
		Message:   "ECoS-Verbindung konnte nicht aufgebaut werden.",
	}
	if err != nil {
		result.Message = err.Error()
		return result, nil
	}
	fields := parseECoSFields(lines)
	result.Connected = true
	result.Status = fields["status"]
	result.ProtocolVersion = firstNonEmpty(fields["ProtocolVersion"], fields["protocolversion"])
	result.ApplicationVersion = firstNonEmpty(fields["ApplicationVersion"], fields["applicationversion"])
	result.HardwareVersion = firstNonEmpty(fields["HardwareVersion"], fields["hardwareversion"])
	result.Message = "ECoS-Verbindung erfolgreich."
	result.RawLines = lines
	result.Fields = fields
	return result, nil
}

func (s *ECoSService) ProbeLocomotiveRaw(ctx context.Context, input ECoSConnectionInput) (*ECoSRawProbe, error) {
	target, err := normalizeECoSInput(input)
	if err != nil {
		return nil, err
	}
	lines, err := s.exchange(ctx, target.Host, target.Port, "queryObjects(10, addr, name, protocol)")
	if err != nil {
		return nil, err
	}
	locomotives := parseECoSLocomotives(lines)
	probeFields := eCoSRawProbeFields()
	rawLocomotives := make([]ECoSRawLocomotive, 0, len(locomotives))
	for _, locomotive := range locomotives {
		raw := ECoSRawLocomotive{
			ObjectID:   locomotive.ObjectID,
			Name:       locomotive.Name,
			Address:    locomotive.Address,
			Protocol:   locomotive.Protocol,
			Attributes: map[string][]string{},
		}
		probes, err := s.fetchLocomotiveRawProbes(ctx, target.Host, target.Port, locomotive.ObjectID)
		if err != nil {
			raw.DetailError = err.Error()
			rawLocomotives = append(rawLocomotives, raw)
			continue
		}
		raw.Probes = probes
		supported := map[string]bool{}
		missing := map[string]bool{}
		for _, probe := range probes {
			for _, detail := range parseECoSLocomotives(probe.RawLines) {
				if detail.ObjectID == raw.ObjectID {
					mergeECoSRawLocomotive(&raw, detail)
				}
			}
			for key, values := range probe.Attributes {
				raw.Attributes[key] = append(raw.Attributes[key], values...)
				supported[strings.ToLower(key)] = true
			}
			for _, field := range probe.Fields {
				field = strings.ToLower(field)
				if !supported[field] {
					missing[field] = true
				}
			}
		}
		raw.SupportedFields = sortedECoSFieldNames(supported)
		raw.MissingFields = sortedMissingECoSFieldNames(missing, supported)
		raw.CVs = parseECoSCVValues(raw.Attributes)
		raw.InterestingFields = interestingECoSFields(raw.Attributes)
		rawLocomotives = append(rawLocomotives, raw)
	}
	return &ECoSRawProbe{
		Host:        target.Host,
		Port:        target.Port,
		ProbeFields: probeFields,
		Locomotives: rawLocomotives,
		RawLines:    lines,
		Message:     fmt.Sprintf("%d ECoS-Lokomotiven roh geprüft.", len(rawLocomotives)),
	}, nil
}

func (s *ECoSService) fetchLocomotiveDetails(ctx context.Context, host string, port int, objectID int) (*ECoSLocomotive, error) {
	command := fmt.Sprintf("get(%d, speed, speedstep, profile, protocol, name, addr, dir, funcset, funcdesc)", objectID)
	lines, err := s.exchangeRequestedGet(ctx, host, port, objectID, command)
	if err != nil {
		return nil, err
	}
	locomotives := parseECoSLocomotives(lines)
	if len(locomotives) == 0 {
		return nil, errors.New("keine Detaildaten gelesen")
	}
	return &locomotives[0], nil
}

func (s *ECoSService) fetchLocomotiveRawProbes(ctx context.Context, host string, port int, objectID int) ([]ECoSRawCommandProbe, error) {
	commands := []struct {
		command string
		fields  []string
	}{
		{
			command: fmt.Sprintf("get(%d, speed, speedstep, profile, protocol, name, addr, dir, funcset, funcdesc)", objectID),
			fields:  []string{"speed", "speedstep", "profile", "protocol", "name", "addr", "dir", "funcset", "funcdesc"},
		},
	}
	for _, field := range eCoSRawProbeFields() {
		commands = append(commands, struct {
			command string
			fields  []string
		}{
			command: fmt.Sprintf("get(%d, %s)", objectID, field),
			fields:  []string{field},
		})
	}
	return s.exchangeRequestedCommands(ctx, host, port, objectID, commands)
}

func (s *ECoSService) exchange(ctx context.Context, host string, port int, command string) ([]string, error) {
	timeout := s.timeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	dialer := net.Dialer{Timeout: timeout}
	conn, err := dialer.DialContext(ctx, "tcp", net.JoinHostPort(host, strconv.Itoa(port)))
	if err != nil {
		return nil, fmt.Errorf("ECoS nicht erreichbar: %w", err)
	}
	defer func() { _ = conn.Close() }()

	if err := conn.SetDeadline(time.Now().Add(timeout)); err != nil {
		return nil, fmt.Errorf("ECoS-Zeitlimit konnte nicht gesetzt werden: %w", err)
	}
	if _, err := fmt.Fprintf(conn, "%s\r\n", strings.TrimSpace(command)); err != nil {
		return nil, fmt.Errorf("ECoS-Kommando konnte nicht gesendet werden: %w", err)
	}

	lines := []string{}
	reader := bufio.NewReader(conn)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() && len(lines) > 0 {
				return lines, nil
			}
			return nil, fmt.Errorf("ECoS-Antwort konnte nicht gelesen werden: %w", err)
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		lines = append(lines, line)
		if strings.HasPrefix(line, "<END") {
			break
		}
		if len(lines) > 0 {
			_ = conn.SetReadDeadline(time.Now().Add(250 * time.Millisecond))
		}
	}
	if len(lines) == 0 {
		return nil, errors.New("ECoS hat keine Antwort geliefert")
	}
	return lines, nil
}

func (s *ECoSService) exchangeRequestedCommands(ctx context.Context, host string, port int, objectID int, commands []struct {
	command string
	fields  []string
}) ([]ECoSRawCommandProbe, error) {
	timeout := s.timeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	dialer := net.Dialer{Timeout: timeout}
	conn, err := dialer.DialContext(ctx, "tcp", net.JoinHostPort(host, strconv.Itoa(port)))
	if err != nil {
		return nil, fmt.Errorf("ECoS nicht erreichbar: %w", err)
	}
	defer func() { _ = conn.Close() }()

	if err := conn.SetDeadline(time.Now().Add(timeout)); err != nil {
		return nil, fmt.Errorf("ECoS-Zeitlimit konnte nicht gesetzt werden: %w", err)
	}
	reader := bufio.NewReader(conn)
	requestCommand := fmt.Sprintf("request(%d, view)", objectID)
	if _, err := fmt.Fprintf(conn, "%s\r\n", requestCommand); err != nil {
		return nil, fmt.Errorf("ECoS-View konnte nicht angefordert werden: %w", err)
	}
	if _, err := readECoSReply(conn, reader, timeout); err != nil {
		return nil, fmt.Errorf("ECoS-Viewantwort konnte nicht gelesen werden: %w", err)
	}

	probes := make([]ECoSRawCommandProbe, 0, len(commands))
	for _, item := range commands {
		if _, err := fmt.Fprintf(conn, "%s\r\n", strings.TrimSpace(item.command)); err != nil {
			probes = append(probes, ECoSRawCommandProbe{
				Command: item.command,
				Fields:  item.fields,
				OK:      false,
				Error:   err.Error(),
			})
			continue
		}
		lines, err := readECoSReply(conn, reader, timeout)
		status, ok := parseECoSEndStatus(lines)
		probe := ECoSRawCommandProbe{
			Command:    item.command,
			Fields:     item.fields,
			OK:         ok,
			Status:     status,
			RawLines:   lines,
			Attributes: parseECoSAttributes(lines),
		}
		if err != nil {
			probe.OK = false
			probe.Error = err.Error()
		}
		probes = append(probes, probe)
	}
	_, _ = fmt.Fprintf(conn, "release(%d, view)\r\n", objectID)
	return probes, nil
}

func (s *ECoSService) exchangeRequestedGet(ctx context.Context, host string, port int, objectID int, getCommand string) ([]string, error) {
	timeout := s.timeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	dialer := net.Dialer{Timeout: timeout}
	conn, err := dialer.DialContext(ctx, "tcp", net.JoinHostPort(host, strconv.Itoa(port)))
	if err != nil {
		return nil, fmt.Errorf("ECoS nicht erreichbar: %w", err)
	}
	defer func() { _ = conn.Close() }()

	if err := conn.SetDeadline(time.Now().Add(timeout)); err != nil {
		return nil, fmt.Errorf("ECoS-Zeitlimit konnte nicht gesetzt werden: %w", err)
	}
	if _, err := fmt.Fprintf(conn, "request(%d, view)\r\n", objectID); err != nil {
		return nil, fmt.Errorf("ECoS-View konnte nicht angefordert werden: %w", err)
	}
	if _, err := fmt.Fprintf(conn, "%s\r\n", strings.TrimSpace(getCommand)); err != nil {
		return nil, fmt.Errorf("ECoS-Detailkommando konnte nicht gesendet werden: %w", err)
	}

	reader := bufio.NewReader(conn)
	lines := []string{}
	inGetReply := false
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() && len(lines) > 0 {
				break
			}
			return nil, fmt.Errorf("ECoS-Detailantwort konnte nicht gelesen werden: %w", err)
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "<REPLY get(") {
			inGetReply = true
			lines = append(lines, line)
			continue
		}
		if !inGetReply {
			continue
		}
		lines = append(lines, line)
		if strings.HasPrefix(line, "<END") {
			break
		}
	}
	_, _ = fmt.Fprintf(conn, "release(%d, view)\r\n", objectID)
	if len(lines) == 0 {
		return nil, errors.New("ECoS hat keine Lok-Detailantwort geliefert")
	}
	return lines, nil
}

func readECoSReply(conn net.Conn, reader *bufio.Reader, timeout time.Duration) ([]string, error) {
	if err := conn.SetReadDeadline(time.Now().Add(timeout)); err != nil {
		return nil, err
	}
	lines := []string{}
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() && len(lines) > 0 {
				return lines, nil
			}
			return lines, err
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		lines = append(lines, line)
		if strings.HasPrefix(line, "<END") {
			return lines, nil
		}
	}
}

func normalizeECoSInput(input ECoSConnectionInput) (ECoSConnectionInput, error) {
	host := strings.TrimSpace(input.Host)
	if host == "" {
		return ECoSConnectionInput{}, errors.New("ECoS-IP oder Hostname fehlt")
	}
	port := input.Port
	if port == 0 {
		port = defaultECoSPort
	}
	if port < 1 || port > 65535 {
		return ECoSConnectionInput{}, errors.New("ECoS-Port muss zwischen 1 und 65535 liegen")
	}
	return ECoSConnectionInput{Host: host, Port: port}, nil
}

func eCoSRawProbeFields() []string {
	return []string{
		"decoder",
		"decodertype",
		"decodername",
		"mfxid",
		"mfxuid",
		"railcom",
		"snifferaddr",
		"longaddr",
		"consist",
		"consistid",
		"category",
		"type",
		"symbol",
		"icon",
		"image",
		"picture",
		"pic",
		"userimage",
		"cv",
		"cvs",
		"cvlist",
		"functionmapping",
	}
}

func parseECoSFields(lines []string) map[string]string {
	fields := map[string]string{}
	for _, line := range lines {
		if strings.HasPrefix(line, "<") {
			continue
		}
		for key, value := range parseECoSArguments(line) {
			fields[key] = value
		}
	}
	return fields
}

func parseECoSAttributes(lines []string) map[string][]string {
	attributes := map[string][]string{}
	for _, line := range lines {
		if strings.HasPrefix(line, "<") {
			continue
		}
		for _, arg := range parseECoSArgumentList(line) {
			attributes[arg.Key] = append(attributes[arg.Key], cleanECoSValue(arg.Value))
		}
	}
	return attributes
}

func parseECoSEndStatus(lines []string) (string, bool) {
	for _, line := range lines {
		if !strings.HasPrefix(line, "<END") {
			continue
		}
		status := strings.Trim(strings.TrimPrefix(line, "<END"), " >")
		return status, strings.Contains(line, "(OK)")
	}
	return "", false
}

func sortedECoSFieldNames(fields map[string]bool) []string {
	names := make([]string, 0, len(fields))
	for field := range fields {
		names = append(names, field)
	}
	sortStrings(names)
	return names
}

func sortedMissingECoSFieldNames(missing map[string]bool, supported map[string]bool) []string {
	names := []string{}
	for field := range missing {
		if !supported[field] {
			names = append(names, field)
		}
	}
	sortStrings(names)
	return names
}

func interestingECoSFields(attributes map[string][]string) []string {
	interesting := []string{}
	for key := range attributes {
		normalized := strings.ToLower(key)
		if strings.Contains(normalized, "cv") || strings.Contains(normalized, "image") || strings.Contains(normalized, "pic") || strings.Contains(normalized, "icon") {
			interesting = append(interesting, key)
		}
	}
	sortStrings(interesting)
	return interesting
}

func parseECoSCVValues(attributes map[string][]string) []ECoSCVValue {
	values := attributes["cv"]
	if len(values) == 0 {
		values = attributes["CV"]
	}
	cvs := make([]ECoSCVValue, 0, len(values))
	seen := map[int]bool{}
	for _, value := range values {
		parts := strings.Split(value, ",")
		if len(parts) < 2 {
			continue
		}
		number := parseECoSInt(parts[0])
		cvValue := parseECoSInt(parts[1])
		if number <= 0 || seen[number] {
			continue
		}
		seen[number] = true
		cvs = append(cvs, ECoSCVValue{Number: number, Value: cvValue})
	}
	sortECoSCVValues(cvs)
	return cvs
}

func sortECoSCVValues(values []ECoSCVValue) {
	for i := 1; i < len(values); i++ {
		value := values[i]
		j := i - 1
		for j >= 0 && values[j].Number > value.Number {
			values[j+1] = values[j]
			j--
		}
		values[j+1] = value
	}
}

func sortStrings(values []string) {
	for i := 1; i < len(values); i++ {
		value := values[i]
		j := i - 1
		for j >= 0 && values[j] > value {
			values[j+1] = values[j]
			j--
		}
		values[j+1] = value
	}
}

func parseECoSLocomotives(lines []string) []ECoSLocomotive {
	locomotives := []ECoSLocomotive{}
	byID := map[int]int{}
	for _, line := range lines {
		if strings.HasPrefix(line, "<") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		objectID, err := strconv.Atoi(fields[0])
		if err != nil || objectID <= 0 {
			continue
		}
		locomotiveIndex, exists := byID[objectID]
		if !exists {
			locomotives = append(locomotives, ECoSLocomotive{
				ObjectID:  objectID,
				Functions: []ECoSFunction{},
			})
			locomotiveIndex = len(locomotives) - 1
			byID[objectID] = locomotiveIndex
		}
		for _, arg := range parseECoSArgumentList(line) {
			applyECoSArgument(&locomotives[locomotiveIndex], arg)
		}
	}
	return locomotives
}

func mergeECoSLocomotive(target *ECoSLocomotive, source *ECoSLocomotive) {
	if source == nil {
		return
	}
	if source.Name != "" {
		target.Name = source.Name
	}
	if source.Address != 0 {
		target.Address = source.Address
	}
	if source.Protocol != "" {
		target.Protocol = source.Protocol
	}
	if source.Profile != "" {
		target.Profile = source.Profile
	}
	target.Speed = source.Speed
	target.SpeedStep = source.SpeedStep
	target.Direction = source.Direction
	if source.FunctionSet != "" {
		target.FunctionSet = source.FunctionSet
	}
	if source.NumberOfFunctions != 0 {
		target.NumberOfFunctions = source.NumberOfFunctions
	}
	if len(source.Functions) > 0 {
		target.Functions = source.Functions
	}
	if len(source.Attributes) > 0 {
		if target.Attributes == nil {
			target.Attributes = map[string][]string{}
		}
		for key, values := range source.Attributes {
			target.Attributes[key] = append(target.Attributes[key], values...)
		}
	}
}

func mergeECoSRawLocomotive(target *ECoSRawLocomotive, source ECoSLocomotive) {
	if source.Name != "" {
		target.Name = source.Name
	}
	if source.Address != 0 {
		target.Address = source.Address
	}
	if source.Protocol != "" {
		target.Protocol = source.Protocol
	}
	if source.Profile != "" {
		target.Profile = source.Profile
	}
	target.Speed = source.Speed
	target.SpeedStep = source.SpeedStep
	target.Direction = source.Direction
	if source.FunctionSet != "" {
		target.FunctionSet = source.FunctionSet
	}
	if source.NumberOfFunctions != 0 {
		target.NumberOfFunctions = source.NumberOfFunctions
	}
	if len(source.Functions) > 0 {
		target.Functions = source.Functions
	}
}

func applyECoSArgument(locomotive *ECoSLocomotive, arg ecosArgument) {
	if locomotive.Attributes == nil {
		locomotive.Attributes = map[string][]string{}
	}
	locomotive.Attributes[arg.Key] = append(locomotive.Attributes[arg.Key], cleanECoSValue(arg.Value))
	switch strings.ToLower(arg.Key) {
	case "name":
		locomotive.Name = cleanECoSValue(arg.Value)
	case "protocol":
		locomotive.Protocol = cleanECoSValue(arg.Value)
	case "profile":
		locomotive.Profile = cleanECoSValue(arg.Value)
	case "addr":
		locomotive.Address = parseECoSInt(arg.Value)
	case "speed":
		locomotive.Speed = parseECoSInt(arg.Value)
	case "speedstep":
		locomotive.SpeedStep = parseECoSInt(arg.Value)
	case "dir":
		locomotive.Direction = parseECoSInt(arg.Value)
	case "funcset":
		locomotive.FunctionSet = cleanECoSValue(arg.Value)
		locomotive.NumberOfFunctions = len(locomotive.FunctionSet)
		for index, state := range locomotive.FunctionSet {
			function := ensureECoSFunction(locomotive, index)
			function.Active = state == '1'
		}
	case "func":
		if len(arg.Params) >= 2 {
			index := parseECoSInt(arg.Params[0])
			function := ensureECoSFunction(locomotive, index)
			function.Active = parseECoSInt(arg.Params[1]) == 1
		}
	case "funcdesc":
		if len(arg.Params) >= 2 {
			index := parseECoSInt(arg.Params[0])
			function := ensureECoSFunction(locomotive, index)
			function.Description = parseECoSInt(arg.Params[1])
		}
	}
}

func ensureECoSFunction(locomotive *ECoSLocomotive, index int) *ECoSFunction {
	for len(locomotive.Functions) <= index {
		locomotive.Functions = append(locomotive.Functions, ECoSFunction{Index: len(locomotive.Functions)})
	}
	return &locomotive.Functions[index]
}

func parseECoSArguments(line string) map[string]string {
	out := map[string]string{}
	for _, arg := range parseECoSArgumentList(line) {
		out[arg.Key] = arg.Value
	}
	return out
}

func parseECoSArgumentList(line string) []ecosArgument {
	out := []ecosArgument{}
	for index := 0; index < len(line); index++ {
		if line[index] != '[' {
			continue
		}
		keyStart := index - 1
		for keyStart >= 0 {
			c := line[keyStart]
			if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '_' {
				keyStart--
				continue
			}
			break
		}
		key := strings.TrimSpace(line[keyStart+1 : index])
		if key == "" {
			continue
		}
		valueStart := index + 1
		valueEnd := valueStart
		inQuote := false
		escaped := false
		for valueEnd < len(line) {
			c := line[valueEnd]
			if escaped {
				escaped = false
				valueEnd++
				continue
			}
			if c == '\\' {
				escaped = true
				valueEnd++
				continue
			}
			if c == '"' {
				inQuote = !inQuote
				valueEnd++
				continue
			}
			if c == ']' && !inQuote {
				break
			}
			valueEnd++
		}
		if valueEnd >= len(line) {
			continue
		}
		value := strings.TrimSpace(line[valueStart:valueEnd])
		out = append(out, ecosArgument{
			Key:    key,
			Value:  value,
			Params: splitECoSParams(value),
		})
		index = valueEnd
	}
	return out
}

func splitECoSParams(value string) []string {
	parts := []string{}
	start := 0
	inQuote := false
	escaped := false
	for index := 0; index < len(value); index++ {
		c := value[index]
		if escaped {
			escaped = false
			continue
		}
		if c == '\\' {
			escaped = true
			continue
		}
		if c == '"' {
			inQuote = !inQuote
			continue
		}
		if c == ',' && !inQuote {
			parts = append(parts, cleanECoSValue(value[start:index]))
			start = index + 1
		}
	}
	parts = append(parts, cleanECoSValue(value[start:]))
	return parts
}

func cleanECoSValue(value string) string {
	value = strings.TrimSpace(value)
	if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
		value = value[1 : len(value)-1]
	}
	value = strings.ReplaceAll(value, `\"`, `"`)
	value = strings.ReplaceAll(value, `\\`, `\`)
	return value
}

func parseECoSInt(value string) int {
	parsed, _ := strconv.Atoi(cleanECoSValue(value))
	return parsed
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
