package application

import (
	"bufio"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"mime"
	"net"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	ecospkg "railkeeper/backend/internal/ecos"
)

const defaultECoSPort = ecospkg.DefaultPort

type ECoSService struct {
	timeout    time.Duration
	client     ecospkg.Client
	liveMu     sync.Mutex
	liveCancel context.CancelFunc
	liveStatus ECoSLiveStatus
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
	ImageCandidates   []ECoSImageCandidate  `json:"imageCandidates,omitempty"`
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

type ECoSImageCandidate struct {
	Key          string `json:"key"`
	Value        string `json:"value"`
	Kind         string `json:"kind"`
	PreviewURL   string `json:"previewUrl,omitempty"`
	MimeType     string `json:"mimeType,omitempty"`
	Transferable bool   `json:"transferable"`
}

type ECoSLiveStatus struct {
	Provider             string   `json:"provider"`
	Connected            bool     `json:"connected"`
	Host                 string   `json:"host,omitempty"`
	Port                 int      `json:"port,omitempty"`
	StartedAt            string   `json:"startedAt,omitempty"`
	LastSeenAt           string   `json:"lastSeenAt,omitempty"`
	LastMessage          string   `json:"lastMessage,omitempty"`
	BlocksReceived       int      `json:"blocksReceived"`
	RepliesReceived      int      `json:"repliesReceived"`
	EventsReceived       int      `json:"eventsReceived"`
	SubscriptionCommands []string `json:"subscriptionCommands,omitempty"`
	Error                string   `json:"error,omitempty"`
	Message              string   `json:"message"`
}

func NewECoSService() *ECoSService {
	timeout := 8 * time.Second
	return &ECoSService{
		timeout: timeout,
		client:  ecospkg.NewClient(timeout),
		liveStatus: ECoSLiveStatus{
			Provider: "ecos",
			Message:  "Keine ECoS-Live-Verbindung aktiv.",
		},
	}
}

func (s *ECoSService) StartLive(ctx context.Context, input ECoSConnectionInput) (*ECoSLiveStatus, error) {
	target, err := normalizeECoSInput(input)
	if err != nil {
		return nil, err
	}
	client := s.eCoSClient()
	conn, reader, err := client.Dial(ctx, ecospkg.Target{Host: target.Host, Port: target.Port})
	if err != nil {
		return nil, err
	}

	liveCtx, cancel := context.WithCancel(context.Background())
	commands := eCoSLiveSubscriptionCommands()
	now := time.Now().UTC().Format(time.RFC3339)

	s.liveMu.Lock()
	s.stopLiveLocked()
	s.liveCancel = cancel
	s.liveStatus = ECoSLiveStatus{
		Provider:             "ecos",
		Connected:            true,
		Host:                 target.Host,
		Port:                 target.Port,
		StartedAt:            now,
		LastSeenAt:           now,
		LastMessage:          "ECoS-Live-Verbindung gestartet.",
		SubscriptionCommands: commands,
		Message:              "ECoS-Live-Verbindung aktiv.",
	}
	status := s.liveStatus
	s.liveMu.Unlock()

	go s.runECoSLiveSession(liveCtx, conn, reader, client, commands)
	return &status, nil
}

func (s *ECoSService) StopLive() ECoSLiveStatus {
	s.liveMu.Lock()
	defer s.liveMu.Unlock()
	s.stopLiveLocked()
	s.liveStatus.Connected = false
	s.liveStatus.Message = "ECoS-Live-Verbindung beendet."
	s.liveStatus.LastMessage = "Verbindung beendet."
	s.liveStatus.LastSeenAt = time.Now().UTC().Format(time.RFC3339)
	return s.liveStatus
}

func (s *ECoSService) LiveStatus() ECoSLiveStatus {
	s.liveMu.Lock()
	defer s.liveMu.Unlock()
	return s.liveStatus
}

func (s *ECoSService) stopLiveLocked() {
	if s.liveCancel != nil {
		s.liveCancel()
		s.liveCancel = nil
	}
}

func (s *ECoSService) runECoSLiveSession(ctx context.Context, conn net.Conn, reader *bufio.Reader, client ecospkg.Client, commands []string) {
	defer func() {
		_ = conn.Close()
		s.liveMu.Lock()
		if s.liveStatus.Connected {
			s.liveStatus.Connected = false
			if s.liveStatus.Error == "" {
				s.liveStatus.Message = "ECoS-Live-Verbindung wurde getrennt."
			}
		}
		s.liveMu.Unlock()
	}()

	for _, command := range commands {
		if err := client.Send(conn, command); err != nil {
			s.updateLiveError(err)
			return
		}
	}

	buffer := []string{}
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		_ = conn.SetReadDeadline(time.Now().Add(750 * time.Millisecond))
		line, err := reader.ReadString('\n')
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}
			s.updateLiveError(fmt.Errorf("ECoS-Live-Antwort konnte nicht gelesen werden: %w", err))
			return
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		buffer = append(buffer, line)
		if !ecospkg.HasBlockLine(line) {
			s.updateLiveLine(line)
			continue
		}
		blocks, err := ecospkg.ParseBlocks(buffer)
		if err != nil {
			s.updateLiveError(err)
			buffer = []string{}
			continue
		}
		s.updateLiveBlocks(blocks, line)
		buffer = []string{}
	}
}

func (s *ECoSService) updateLiveLine(line string) {
	s.liveMu.Lock()
	defer s.liveMu.Unlock()
	s.liveStatus.LastMessage = line
	s.liveStatus.LastSeenAt = time.Now().UTC().Format(time.RFC3339)
}

func (s *ECoSService) updateLiveBlocks(blocks []ecospkg.Block, lastLine string) {
	s.liveMu.Lock()
	defer s.liveMu.Unlock()
	s.liveStatus.BlocksReceived += len(blocks)
	for _, block := range blocks {
		switch block.Kind {
		case ecospkg.BlockEvent:
			s.liveStatus.EventsReceived++
		case ecospkg.BlockReply:
			s.liveStatus.RepliesReceived++
		}
	}
	s.liveStatus.LastMessage = lastLine
	s.liveStatus.LastSeenAt = time.Now().UTC().Format(time.RFC3339)
	s.liveStatus.Message = "ECoS-Live-Verbindung aktiv."
}

func (s *ECoSService) updateLiveError(err error) {
	s.liveMu.Lock()
	defer s.liveMu.Unlock()
	s.liveStatus.Connected = false
	s.liveStatus.Error = err.Error()
	s.liveStatus.Message = "ECoS-Live-Verbindung ist unterbrochen."
	s.liveStatus.LastMessage = err.Error()
	s.liveStatus.LastSeenAt = time.Now().UTC().Format(time.RFC3339)
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
		raw.ImageCandidates = parseECoSImageCandidates(raw.Attributes)
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
	for _, field := range eCoSTargetedCVProbeFields() {
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
	return s.eCoSClient().Exchange(ctx, ecospkg.Target{Host: host, Port: port}, command)
}

func (s *ECoSService) exchangeRequestedCommands(ctx context.Context, host string, port int, objectID int, commands []struct {
	command string
	fields  []string
}) ([]ECoSRawCommandProbe, error) {
	timeout := s.timeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout+time.Duration(len(commands))*1200*time.Millisecond)
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
	replyTimeout := 1200 * time.Millisecond
	for _, item := range commands {
		if err := ctx.Err(); err != nil {
			probes = append(probes, ECoSRawCommandProbe{
				Command: item.command,
				Fields:  item.fields,
				OK:      false,
				Error:   err.Error(),
			})
			break
		}
		_ = conn.SetWriteDeadline(time.Now().Add(replyTimeout))
		if _, err := fmt.Fprintf(conn, "%s\r\n", strings.TrimSpace(item.command)); err != nil {
			probes = append(probes, ECoSRawCommandProbe{
				Command: item.command,
				Fields:  item.fields,
				OK:      false,
				Error:   err.Error(),
			})
			continue
		}
		lines, err := readECoSReply(conn, reader, replyTimeout)
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
	return ecospkg.ReadReply(conn, reader, timeout)
}

func normalizeECoSInput(input ECoSConnectionInput) (ECoSConnectionInput, error) {
	target, err := ecospkg.NormalizeTarget(input.Host, input.Port)
	if err != nil {
		return ECoSConnectionInput{}, err
	}
	return ECoSConnectionInput{Host: target.Host, Port: target.Port}, nil
}

func (s *ECoSService) eCoSClient() ecospkg.Client {
	if s.client.Timeout <= 0 {
		timeout := s.timeout
		if timeout <= 0 {
			timeout = 5 * time.Second
		}
		s.client = ecospkg.NewClient(timeout)
	}
	return s.client
}

func eCoSLiveSubscriptionCommands() []string {
	return []string{
		"request(1, view)",
		"get(1, info, status)",
		"request(10, view)",
		"queryObjects(10, addr, name, protocol)",
		"request(11, view)",
		"queryObjects(11, addr, protocol, type, addrext, mode, symbol, name1, name2, name3, switching)",
		"queryObjects(26, ports)",
		"request(26, view)",
	}
}

func eCoSRawProbeFields() []string {
	return []string{
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

func eCoSTargetedCVProbeFields() []string {
	return []string{
		"cv[1:8]",
		"cv[7]",
		"cv[8]",
		"cv[7:8]",
		"cv[17]",
		"cv[18]",
		"cv[29]",
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

func parseECoSImageCandidates(attributes map[string][]string) []ECoSImageCandidate {
	candidates := []ECoSImageCandidate{}
	seen := map[string]bool{}
	for _, key := range sortedECoSImageAttributeKeys(attributes) {
		for _, value := range attributes[key] {
			candidate := classifyECoSImageCandidate(key, value)
			if candidate.Value == "" {
				continue
			}
			seenKey := strings.ToLower(candidate.Key + "::" + candidate.Value)
			if seen[seenKey] {
				continue
			}
			seen[seenKey] = true
			candidates = append(candidates, candidate)
		}
	}
	return candidates
}

func sortedECoSImageAttributeKeys(attributes map[string][]string) []string {
	keys := []string{}
	for key := range attributes {
		if isECoSImageAttributeKey(key) {
			keys = append(keys, key)
		}
	}
	sortStrings(keys)
	return keys
}

func isECoSImageAttributeKey(key string) bool {
	normalized := strings.ToLower(strings.TrimSpace(key))
	return normalized == "icon" ||
		normalized == "image" ||
		normalized == "picture" ||
		normalized == "pic" ||
		normalized == "userimage" ||
		strings.Contains(normalized, "image") ||
		strings.Contains(normalized, "picture")
}

func classifyECoSImageCandidate(key string, value string) ECoSImageCandidate {
	value = cleanECoSValue(value)
	candidate := ECoSImageCandidate{
		Key:   strings.TrimSpace(key),
		Value: value,
		Kind:  "reference",
	}
	if value == "" {
		return candidate
	}
	lower := strings.ToLower(value)
	switch {
	case strings.HasPrefix(lower, "data:image/"):
		candidate.Kind = "data"
		candidate.PreviewURL = value
		candidate.MimeType = mimeFromDataURL(value)
		candidate.Transferable = true
	case strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://"):
		candidate.Kind = "url"
		candidate.PreviewURL = value
		candidate.MimeType = mime.TypeByExtension(filepath.Ext(value))
		candidate.Transferable = true
	case isIntegerText(value):
		candidate.Kind = "id"
	case strings.HasPrefix(strings.TrimSpace(value), "<svg"):
		encoded := base64.StdEncoding.EncodeToString([]byte(value))
		candidate.Kind = "data"
		candidate.PreviewURL = "data:image/svg+xml;base64," + encoded
		candidate.MimeType = "image/svg+xml"
		candidate.Transferable = true
	default:
		previewURL, mimeType := imageDataURLFromBase64(value)
		if previewURL != "" {
			candidate.Kind = "base64"
			candidate.PreviewURL = previewURL
			candidate.MimeType = mimeType
			candidate.Transferable = true
		}
	}
	return candidate
}

func mimeFromDataURL(value string) string {
	withoutPrefix := strings.TrimPrefix(value, "data:")
	end := strings.Index(withoutPrefix, ";")
	if end < 0 {
		end = strings.Index(withoutPrefix, ",")
	}
	if end < 0 {
		return ""
	}
	return withoutPrefix[:end]
}

func imageDataURLFromBase64(value string) (string, string) {
	compact := strings.NewReplacer("\r", "", "\n", "", " ", "", "\t", "").Replace(strings.TrimSpace(value))
	if len(compact) < 64 {
		return "", ""
	}
	decoded, err := base64.StdEncoding.DecodeString(compact)
	if err != nil {
		decoded, err = base64.RawStdEncoding.DecodeString(compact)
	}
	if err != nil || len(decoded) == 0 {
		return "", ""
	}
	mimeType := http.DetectContentType(decoded)
	if !strings.HasPrefix(mimeType, "image/") {
		return "", ""
	}
	return "data:" + mimeType + ";base64," + compact, mimeType
}

func isIntegerText(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" {
		return false
	}
	for _, char := range value {
		if char < '0' || char > '9' {
			return false
		}
	}
	return true
}

func parseECoSCVValues(attributes map[string][]string) []ECoSCVValue {
	parsed := map[int]int{}
	add := func(number int, value int) {
		if number <= 0 {
			return
		}
		parsed[number] = value
	}
	for _, key := range sortedECoSCVAttributeKeys(attributes) {
		normalizedKey := strings.ToLower(strings.TrimSpace(key))
		keyNumber := parseECoSCVKeyNumber(normalizedKey)
		values := attributes[key]
		for _, value := range values {
			parts := splitECoSParams(value)
			if len(parts) >= 2 {
				for index := 0; index+1 < len(parts); index += 2 {
					add(parseECoSInt(parts[index]), parseECoSInt(parts[index+1]))
				}
				continue
			}
			if keyNumber > 0 && len(parts) == 1 {
				add(keyNumber, parseECoSInt(parts[0]))
				continue
			}
			number, cvValue, ok := parseECoSCVTextPair(value)
			if ok {
				add(number, cvValue)
			}
		}
	}
	cvs := make([]ECoSCVValue, 0, len(parsed))
	for number, value := range parsed {
		cvs = append(cvs, ECoSCVValue{Number: number, Value: value})
	}
	sortECoSCVValues(cvs)
	return cvs
}

func sortedECoSCVAttributeKeys(attributes map[string][]string) []string {
	primary := []string{}
	specific := []string{}
	for key := range attributes {
		normalizedKey := strings.ToLower(strings.TrimSpace(key))
		if !isECoSCVAttributeKey(normalizedKey) {
			continue
		}
		if parseECoSCVKeyNumber(normalizedKey) > 0 {
			specific = append(specific, key)
			continue
		}
		primary = append(primary, key)
	}
	sortStrings(primary)
	sortStrings(specific)
	return append(primary, specific...)
}

func isECoSCVAttributeKey(key string) bool {
	return key == "cv" || key == "cvs" || key == "cvlist" || parseECoSCVKeyNumber(key) > 0
}

func parseECoSCVKeyNumber(key string) int {
	if !strings.HasPrefix(key, "cv") {
		return 0
	}
	suffix := strings.Trim(strings.TrimPrefix(key, "cv"), " _-[]")
	if suffix == "" {
		return 0
	}
	number, err := strconv.Atoi(suffix)
	if err != nil {
		return 0
	}
	return number
}

func parseECoSCVTextPair(value string) (int, int, bool) {
	value = cleanECoSValue(value)
	for _, separator := range []string{"=", ":"} {
		parts := strings.SplitN(value, separator, 2)
		if len(parts) != 2 {
			continue
		}
		number := parseECoSInt(parts[0])
		cvValue := parseECoSInt(parts[1])
		if number > 0 {
			return number, cvValue, true
		}
	}
	return 0, 0, false
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
