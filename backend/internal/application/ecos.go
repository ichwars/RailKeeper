package application

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	ecospkg "railkeeper/backend/internal/ecos"
)

const (
	defaultECoSPort           = ecospkg.DefaultPort
	eCoSLiveIdleTimeout       = 20 * time.Minute
	eCoSLiveReadDeadline      = 750 * time.Millisecond
	eCoSLocomotiveListCommand = "queryObjects(10, addr, name, protocol)"
)

type ECoSService struct {
	timeout                  time.Duration
	client                   ecospkg.Client
	liveMu                   sync.Mutex
	liveCancel               context.CancelFunc
	liveStatus               ECoSLiveStatus
	livePulse                time.Time
	liveCount                int
	liveGeneration           uint64
	liveInterruptionNotified bool
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
	ObjectID    int                 `json:"objectId"`
	Name        string              `json:"name,omitempty"`
	Address     int                 `json:"address"`
	Protocol    string              `json:"protocol,omitempty"`
	Profile     string              `json:"profile,omitempty"`
	Functions   []ECoSFunction      `json:"functions,omitempty"`
	Attributes  map[string][]string `json:"attributes,omitempty"`
	DetailError string              `json:"detailError,omitempty"`
}

type ECoSFunction struct {
	Index       int `json:"index"`
	Description int `json:"description,omitempty"`
}

type ecosArgument struct {
	Key    string
	Value  string
	Params []string
}

var eCoSAllowedLocomotiveAttributes = map[string]struct{}{
	"addr":            {},
	"cv":              {},
	"cvlist":          {},
	"cvs":             {},
	"funcdesc":        {},
	"functionmapping": {},
	"name":            {},
	"profile":         {},
	"protocol":        {},
}

func isAllowedECoSLocomotiveAttribute(key string) bool {
	_, allowed := eCoSAllowedLocomotiveAttributes[strings.ToLower(strings.TrimSpace(key))]
	return allowed
}

type ECoSRawProbe struct {
	Host        string              `json:"host"`
	Port        int                 `json:"port"`
	ProbeFields []string            `json:"probeFields"`
	Locomotives []ECoSRawLocomotive `json:"locomotives"`
	RawLines    []string            `json:"rawLines,omitempty"`
	Message     string              `json:"message"`
}

type ECoSLocomotiveSummary struct {
	Host    string `json:"host"`
	Port    int    `json:"port"`
	Count   int    `json:"count"`
	Message string `json:"message"`
}

type ECoSRawLocomotive struct {
	ObjectID          int                   `json:"objectId"`
	Name              string                `json:"name,omitempty"`
	Address           int                   `json:"address,omitempty"`
	Protocol          string                `json:"protocol,omitempty"`
	Profile           string                `json:"profile,omitempty"`
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

type ECoSLiveMonitorState string

const (
	ECoSLiveStopped     ECoSLiveMonitorState = "stopped"
	ECoSLiveRunning     ECoSLiveMonitorState = "running"
	ECoSLiveInterrupted ECoSLiveMonitorState = "interrupted"
)

type ECoSLivePulseSample struct {
	At               string `json:"at"`
	RepliesPerSecond int    `json:"repliesPerSecond"`
}

type ECoSLiveEvent struct {
	At       string `json:"at"`
	Kind     string `json:"kind"`
	Message  string `json:"message"`
	Protocol string `json:"protocol"`
}

type ECoSLiveDiagnosis struct {
	ConnectionState             ECoSLiveMonitorState `json:"connectionState"`
	LastSuccessfulCommunication string               `json:"lastSuccessfulCommunication,omitempty"`
	LastError                   string               `json:"lastError,omitempty"`
	Passive                     bool                 `json:"passive"`
}

type ECoSLiveStatus struct {
	Provider             string                `json:"provider"`
	Connected            bool                  `json:"connected"`
	State                ECoSLiveMonitorState  `json:"state"`
	Host                 string                `json:"host,omitempty"`
	Port                 int                   `json:"port,omitempty"`
	StartedAt            string                `json:"startedAt,omitempty"`
	LastSeenAt           string                `json:"lastSeenAt,omitempty"`
	LastMessage          string                `json:"lastMessage,omitempty"`
	BlocksReceived       int                   `json:"blocksReceived"`
	RepliesReceived      int                   `json:"repliesReceived"`
	EventsReceived       int                   `json:"eventsReceived"`
	SubscriptionCommands []string              `json:"subscriptionCommands,omitempty"`
	PulseSamples         []ECoSLivePulseSample `json:"pulseSamples"`
	RecentEvents         []ECoSLiveEvent       `json:"recentEvents"`
	Diagnosis            ECoSLiveDiagnosis     `json:"diagnosis"`
	Error                string                `json:"error,omitempty"`
	Message              string                `json:"message"`
}

type ECoSLocomotiveSyncInput struct {
	Host     string                    `json:"host"`
	Port     int                       `json:"port"`
	ObjectID int                       `json:"objectId"`
	Desired  ECoSLocomotiveSyncDesired `json:"desired"`
	DryRun   bool                      `json:"dryRun"`
	Confirm  bool                      `json:"confirm"`
}

type ECoSLocomotiveSyncDesired struct {
	Name     string `json:"name,omitempty"`
	Address  int    `json:"address,omitempty"`
	Protocol string `json:"protocol,omitempty"`
}

type ECoSLocomotiveSyncSnapshot struct {
	Name     string `json:"name,omitempty"`
	Address  int    `json:"address,omitempty"`
	Protocol string `json:"protocol,omitempty"`
}

type ECoSLocomotiveSyncChange struct {
	Field   string `json:"field"`
	Current string `json:"current"`
	Desired string `json:"desired"`
}

type ECoSLocomotiveSyncResult struct {
	Host     string                     `json:"host"`
	Port     int                        `json:"port"`
	ObjectID int                        `json:"objectId"`
	DryRun   bool                       `json:"dryRun"`
	Applied  bool                       `json:"applied"`
	Current  ECoSLocomotiveSyncSnapshot `json:"current"`
	Desired  ECoSLocomotiveSyncDesired  `json:"desired"`
	Changes  []ECoSLocomotiveSyncChange `json:"changes"`
	Commands []string                   `json:"commands,omitempty"`
	RawLines []string                   `json:"rawLines,omitempty"`
	Message  string                     `json:"message"`
}

func NewECoSService() *ECoSService {
	timeout := 8 * time.Second
	return &ECoSService{
		timeout: timeout,
		client:  ecospkg.NewClient(timeout),
		liveStatus: ECoSLiveStatus{
			Provider:  "ecos",
			State:     ECoSLiveStopped,
			Diagnosis: ECoSLiveDiagnosis{ConnectionState: ECoSLiveStopped, Passive: true},
			Message:   "Keine ECoS-Live-Verbindung aktiv.",
		},
	}
}

func (s *ECoSService) StartLive(ctx context.Context, input ECoSConnectionInput) (*ECoSLiveStatus, error) {
	return s.StartLiveWithInterruption(ctx, input, nil)
}

func (s *ECoSService) StartLiveWithInterruption(
	ctx context.Context,
	input ECoSConnectionInput,
	onInterrupted func(),
) (*ECoSLiveStatus, error) {
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
	commands := append([]string(nil), eCoSLiveSubscriptionCommands()...)
	nowTime := time.Now().UTC()
	now := nowTime.Format(time.RFC3339)

	s.liveMu.Lock()
	s.stopLiveLocked()
	s.liveGeneration++
	generation := s.liveGeneration
	s.liveInterruptionNotified = false
	s.liveCancel = cancel
	s.liveStatus = ECoSLiveStatus{
		Provider:             "ecos",
		Connected:            true,
		State:                ECoSLiveRunning,
		Host:                 target.Host,
		Port:                 target.Port,
		StartedAt:            now,
		LastSeenAt:           now,
		LastMessage:          "ECoS-Live-Verbindung gestartet.",
		SubscriptionCommands: append([]string(nil), commands...),
		Diagnosis: ECoSLiveDiagnosis{
			ConnectionState: ECoSLiveRunning, LastSuccessfulCommunication: now, Passive: true,
		},
		Message: "ECoS-Live-Verbindung aktiv.",
	}
	s.livePulse = nowTime.Truncate(time.Second)
	s.liveCount = 0
	status := cloneECoSLiveStatus(s.liveStatus)
	s.liveMu.Unlock()

	go s.runECoSLiveSampler(liveCtx, generation)
	go s.runECoSLiveSession(
		liveCtx, cancel, generation, onInterrupted, conn, reader, client, append([]string(nil), commands...),
	)
	return &status, nil
}

func (s *ECoSService) StopLive() ECoSLiveStatus {
	s.liveMu.Lock()
	defer s.liveMu.Unlock()
	s.flushCurrentLivePulseLocked(time.Now().UTC())
	s.stopLiveLocked()
	s.liveGeneration++
	s.liveStatus.Connected = false
	s.liveStatus.State = ECoSLiveStopped
	s.liveStatus.Diagnosis.ConnectionState = ECoSLiveStopped
	s.liveStatus.Diagnosis.LastError = ""
	s.liveStatus.Message = "ECoS-Live-Verbindung beendet."
	s.liveStatus.LastMessage = "Verbindung beendet."
	return cloneECoSLiveStatus(s.liveStatus)
}

func (s *ECoSService) LiveStatus() ECoSLiveStatus {
	return s.liveStatusAt(time.Now().UTC())
}

func (s *ECoSService) liveStatusAt(now time.Time) ECoSLiveStatus {
	s.liveMu.Lock()
	defer s.liveMu.Unlock()
	s.stopIdleLiveLocked(now)
	s.flushLivePulseLocked(now)
	status := cloneECoSLiveStatus(s.liveStatus)
	if status.State == ECoSLiveRunning && !s.livePulse.IsZero() {
		status.PulseSamples = appendBoundedECoSLivePulse(status.PulseSamples, ECoSLivePulseSample{
			At: s.livePulse.Format(time.RFC3339), RepliesPerSecond: s.liveCount,
		})
	}
	return status
}

func (s *ECoSService) stopLiveLocked() {
	if s.liveCancel != nil {
		s.liveCancel()
		s.liveCancel = nil
	}
}

func (s *ECoSService) runECoSLiveSession(
	ctx context.Context,
	cancel context.CancelFunc,
	generation uint64,
	onInterrupted func(),
	conn net.Conn,
	reader *bufio.Reader,
	client ecospkg.Client,
	commands []string,
) {
	defer func() {
		cancel()
		_ = conn.Close()
		s.updateLiveErrorForGenerationWithCallback(
			generation, errors.New("ECoS live connection closed"), onInterrupted,
		)
	}()

	for _, command := range commands {
		if err := client.Send(conn, command); err != nil {
			s.updateLiveErrorForGenerationWithCallback(generation, err, onInterrupted)
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
		_ = conn.SetReadDeadline(time.Now().Add(eCoSLiveReadDeadline))
		line, err := reader.ReadString('\n')
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				if s.stopIdleLiveFromMonitor(time.Now().UTC()) {
					return
				}
				continue
			}
			s.updateLiveErrorForGenerationWithCallback(
				generation, fmt.Errorf("ECoS-Live-Antwort konnte nicht gelesen werden: %w", err), onInterrupted,
			)
			return
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		buffer = append(buffer, line)
		if !ecospkg.HasBlockLine(line) {
			s.updateLiveLineForGeneration(generation, line)
			continue
		}
		blocks, err := ecospkg.ParseBlocks(buffer)
		if err != nil {
			s.updateLiveErrorForGenerationWithCallback(generation, err, onInterrupted)
			return
		}
		s.updateLiveBlocksForGenerationAt(generation, time.Now().UTC(), blocks, line)
		buffer = []string{}
	}
}

func (s *ECoSService) stopIdleLiveFromMonitor(now time.Time) bool {
	s.liveMu.Lock()
	defer s.liveMu.Unlock()
	return s.stopIdleLiveLocked(now)
}

func (s *ECoSService) stopIdleLiveLocked(now time.Time) bool {
	if !s.liveStatus.Connected || s.liveStatus.LastSeenAt == "" {
		return false
	}
	lastSeen, err := time.Parse(time.RFC3339, s.liveStatus.LastSeenAt)
	if err != nil || now.Sub(lastSeen) < eCoSLiveIdleTimeout {
		return false
	}
	s.stopLiveLocked()
	s.flushCurrentLivePulseLocked(now)
	s.liveGeneration++
	s.liveStatus.Connected = false
	s.liveStatus.State = ECoSLiveStopped
	s.liveStatus.Diagnosis.ConnectionState = ECoSLiveStopped
	s.liveStatus.Error = ""
	s.liveStatus.Message = "ECoS-Live-Monitoring nach 20 Minuten ohne Aktivität automatisch beendet."
	s.liveStatus.LastMessage = s.liveStatus.Message
	return true
}

func (s *ECoSService) updateLiveLineForGeneration(generation uint64, line string) {
	_ = line
	s.liveMu.Lock()
	defer s.liveMu.Unlock()
	if generation != s.liveGeneration {
		return
	}
	now := time.Now().UTC().Format(time.RFC3339)
	s.liveStatus.LastMessage = "ECoS-Protokolldaten empfangen."
	s.liveStatus.LastSeenAt = now
	s.liveStatus.Diagnosis.LastSuccessfulCommunication = now
}

func (s *ECoSService) updateLiveBlocksAt(now time.Time, blocks []ecospkg.Block, lastLine string) {
	s.liveMu.Lock()
	generation := s.liveGeneration
	s.liveMu.Unlock()
	s.updateLiveBlocksForGenerationAt(generation, now, blocks, lastLine)
}

func (s *ECoSService) updateLiveBlocksForGenerationAt(
	generation uint64,
	now time.Time,
	blocks []ecospkg.Block,
	lastLine string,
) {
	_ = lastLine
	s.liveMu.Lock()
	defer s.liveMu.Unlock()
	if generation != s.liveGeneration {
		return
	}
	s.liveStatus.BlocksReceived += len(blocks)
	replies := 0
	events := 0
	for _, block := range blocks {
		switch block.Kind {
		case ecospkg.BlockEvent:
			s.liveStatus.EventsReceived++
			events++
			message := "ECoS-Ereignis empfangen."
			if block.ObjectID > 0 {
				message = fmt.Sprintf("ECoS-Ereignis für Objekt %d empfangen.", block.ObjectID)
			}
			s.liveStatus.RecentEvents = appendBoundedECoSLiveEvent(s.liveStatus.RecentEvents, ECoSLiveEvent{
				At: now.Format(time.RFC3339), Kind: "event", Message: message, Protocol: "ECoS",
			})
		case ecospkg.BlockReply:
			s.liveStatus.RepliesReceived++
			replies++
		}
	}
	s.recordLivePulseLocked(now, replies)
	if events > 0 {
		s.liveStatus.LastMessage = "ECoS-Ereignis empfangen."
	} else {
		s.liveStatus.LastMessage = "ECoS-Antwort empfangen."
	}
	s.liveStatus.LastSeenAt = now.Format(time.RFC3339)
	s.liveStatus.Diagnosis.LastSuccessfulCommunication = s.liveStatus.LastSeenAt
	s.liveStatus.Message = "ECoS-Live-Verbindung aktiv."
}

func (s *ECoSService) updateLiveError(_ error) {
	s.liveMu.Lock()
	generation := s.liveGeneration
	s.liveMu.Unlock()
	s.updateLiveErrorForGeneration(generation, nil)
}

func (s *ECoSService) updateLiveErrorForGeneration(generation uint64, err error) {
	s.updateLiveErrorForGenerationWithCallback(generation, err, nil)
}

func (s *ECoSService) updateLiveErrorForGenerationWithCallback(
	generation uint64,
	_ error,
	onInterrupted func(),
) {
	s.liveMu.Lock()
	if generation != s.liveGeneration || s.liveStatus.State == ECoSLiveInterrupted {
		s.liveMu.Unlock()
		return
	}
	s.flushCurrentLivePulseLocked(time.Now().UTC())
	s.liveStatus.Connected = false
	s.liveStatus.State = ECoSLiveInterrupted
	s.liveStatus.Diagnosis.ConnectionState = ECoSLiveInterrupted
	s.liveStatus.Error = "ECoS-Live-Verbindung unterbrochen."
	s.liveStatus.Message = "ECoS-Live-Verbindung ist unterbrochen."
	s.liveStatus.LastMessage = "Die Verbindung wurde unerwartet unterbrochen."
	s.liveStatus.Diagnosis.LastError = "Die Verbindung wurde unerwartet unterbrochen."
	shouldNotify := !s.liveInterruptionNotified
	s.liveInterruptionNotified = true
	s.liveMu.Unlock()
	if shouldNotify && onInterrupted != nil {
		onInterrupted()
	}
}

func (s *ECoSService) recordLivePulseLocked(now time.Time, replies int) {
	if s.livePulse.IsZero() {
		s.livePulse = now.UTC().Truncate(time.Second)
	}
	s.flushLivePulseLocked(now)
	s.liveCount += replies
}

func (s *ECoSService) runECoSLiveSampler(ctx context.Context, generation uint64) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			s.flushLivePulseForGeneration(generation, now)
		}
	}
}

func (s *ECoSService) flushLivePulseForGeneration(generation uint64, now time.Time) {
	s.liveMu.Lock()
	defer s.liveMu.Unlock()
	if generation != s.liveGeneration {
		return
	}
	s.flushLivePulseLocked(now)
}

func (s *ECoSService) flushLivePulseLocked(now time.Time) {
	if s.livePulse.IsZero() {
		return
	}
	cutoff := now.UTC().Truncate(time.Second)
	if cutoff.Sub(s.livePulse) > 61*time.Second {
		s.livePulse = cutoff.Add(-60 * time.Second)
		s.liveCount = 0
	}
	for s.livePulse.Before(cutoff) {
		s.liveStatus.PulseSamples = appendBoundedECoSLivePulse(s.liveStatus.PulseSamples, ECoSLivePulseSample{
			At: s.livePulse.Format(time.RFC3339), RepliesPerSecond: s.liveCount,
		})
		s.livePulse = s.livePulse.Add(time.Second)
		s.liveCount = 0
	}
}

func (s *ECoSService) flushCurrentLivePulseLocked(now time.Time) {
	s.flushLivePulseLocked(now)
	if s.livePulse.IsZero() {
		return
	}
	s.liveStatus.PulseSamples = appendBoundedECoSLivePulse(s.liveStatus.PulseSamples, ECoSLivePulseSample{
		At: s.livePulse.Format(time.RFC3339), RepliesPerSecond: s.liveCount,
	})
	s.livePulse = time.Time{}
	s.liveCount = 0
}

func appendBoundedECoSLivePulse(
	samples []ECoSLivePulseSample,
	sample ECoSLivePulseSample,
) []ECoSLivePulseSample {
	samples = append(samples, sample)
	if len(samples) > 60 {
		samples = append([]ECoSLivePulseSample(nil), samples[len(samples)-60:]...)
	}
	return samples
}

func appendBoundedECoSLiveEvent(events []ECoSLiveEvent, event ECoSLiveEvent) []ECoSLiveEvent {
	events = append(events, event)
	if len(events) > 100 {
		events = append([]ECoSLiveEvent(nil), events[len(events)-100:]...)
	}
	return events
}

func cloneECoSLiveStatus(status ECoSLiveStatus) ECoSLiveStatus {
	status.SubscriptionCommands = append([]string(nil), status.SubscriptionCommands...)
	status.PulseSamples = append([]ECoSLivePulseSample{}, status.PulseSamples...)
	status.RecentEvents = append([]ECoSLiveEvent{}, status.RecentEvents...)
	return status
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
		return result, nil //nolint:nilerr // Connection failures are returned as preview results.
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

func (s *ECoSService) SyncLocomotive(ctx context.Context, input ECoSLocomotiveSyncInput) (*ECoSLocomotiveSyncResult, error) {
	target, err := normalizeECoSInput(ECoSConnectionInput{Host: input.Host, Port: input.Port})
	if err != nil {
		return nil, err
	}
	if input.ObjectID <= 0 {
		return nil, errors.New("ECoS-Objekt-ID ist erforderlich")
	}

	current, err := s.fetchLocomotiveDetails(ctx, target.Host, target.Port, input.ObjectID)
	if err != nil {
		return nil, err
	}
	desired := cleanECoSLocomotiveSyncDesired(input.Desired)
	changes, command, err := buildECoSLocomotiveSyncCommand(input.ObjectID, current, desired)
	if err != nil {
		return nil, err
	}

	result := &ECoSLocomotiveSyncResult{
		Host:     target.Host,
		Port:     target.Port,
		ObjectID: input.ObjectID,
		DryRun:   input.DryRun || !input.Confirm,
		Current: ECoSLocomotiveSyncSnapshot{
			Name:     current.Name,
			Address:  current.Address,
			Protocol: current.Protocol,
		},
		Desired: desired,
		Changes: changes,
		Message: "ECoS-Sync-Vorschau erstellt.",
	}
	if command != "" {
		result.Commands = []string{command}
	}
	if len(changes) == 0 {
		result.Message = "ECoS-Lok ist bereits synchron."
		return result, nil
	}
	if result.DryRun {
		return result, nil
	}

	probes, err := s.exchangeRequestedCommands(ctx, target.Host, target.Port, input.ObjectID, []struct {
		command string
		fields  []string
	}{{command: command, fields: eCoSLocomotiveSyncFields(changes)}})
	if err != nil {
		return nil, err
	}
	if len(probes) == 0 {
		return nil, errors.New("ECoS hat keine Schreibantwort geliefert")
	}
	probe := probes[0]
	result.RawLines = probe.RawLines
	if !probe.OK {
		if probe.Error != "" {
			return nil, fmt.Errorf("%w: ECoS-Schreibantwort fehlt", ErrECoSWriteStateUnknown)
		}
		return nil, fmt.Errorf("ECoS-Schreibbefehl nicht bestätigt: %s", firstNonEmpty(probe.Status, "unbekannter Status"))
	}
	result.Applied = true
	result.Message = "ECoS-Lok wurde geschrieben."
	return result, nil
}

func (s *ECoSService) ProbeLocomotiveRaw(ctx context.Context, input ECoSConnectionInput) (*ECoSRawProbe, error) {
	target, err := normalizeECoSInput(input)
	if err != nil {
		return nil, err
	}
	lines, err := s.exchange(ctx, target.Host, target.Port, eCoSLocomotiveListCommand)
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

func (s *ECoSService) CountLocomotives(ctx context.Context, input ECoSConnectionInput) (*ECoSLocomotiveSummary, error) {
	target, err := normalizeECoSInput(input)
	if err != nil {
		return nil, err
	}
	lines, err := s.exchange(ctx, target.Host, target.Port, eCoSLocomotiveListCommand)
	if err != nil {
		return nil, err
	}
	count := len(parseECoSLocomotives(lines))
	return &ECoSLocomotiveSummary{
		Host:    target.Host,
		Port:    target.Port,
		Count:   count,
		Message: fmt.Sprintf("%d ECoS-Lokdatensätze gefunden.", count),
	}, nil
}

func (s *ECoSService) fetchLocomotiveDetails(ctx context.Context, host string, port int, objectID int) (*ECoSLocomotive, error) {
	command := eCoSLocomotiveDetailCommand(objectID)
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
			command: eCoSLocomotiveDetailCommand(objectID),
			fields:  []string{"profile", "protocol", "name", "addr", "funcdesc"},
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
	}
}

func eCoSLocomotiveDetailCommand(objectID int) string {
	return fmt.Sprintf("get(%d, profile, protocol, name, addr, funcdesc)", objectID)
}

func eCoSRawProbeFields() []string {
	return []string{
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
			key := strings.ToLower(arg.Key)
			if !isAllowedECoSLocomotiveAttribute(key) {
				continue
			}
			attributes[key] = append(attributes[key], cleanECoSValue(arg.Value))
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
		if strings.Contains(normalized, "cv") || normalized == "functionmapping" {
			interesting = append(interesting, key)
		}
	}
	sortStrings(interesting)
	return interesting
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
	if len(source.Functions) > 0 {
		target.Functions = source.Functions
	}
}

func applyECoSArgument(locomotive *ECoSLocomotive, arg ecosArgument) {
	key := strings.ToLower(arg.Key)
	if !isAllowedECoSLocomotiveAttribute(key) {
		return
	}
	if locomotive.Attributes == nil {
		locomotive.Attributes = map[string][]string{}
	}
	locomotive.Attributes[key] = append(locomotive.Attributes[key], cleanECoSValue(arg.Value))
	switch key {
	case "name":
		locomotive.Name = cleanECoSValue(arg.Value)
	case "protocol":
		locomotive.Protocol = cleanECoSValue(arg.Value)
	case "profile":
		locomotive.Profile = cleanECoSValue(arg.Value)
	case "addr":
		locomotive.Address = parseECoSInt(arg.Value)
	case "funcdesc":
		if len(arg.Params) >= 2 {
			index := parseECoSInt(arg.Params[0])
			if index < 0 || index > 31 {
				return
			}
			function := ensureECoSFunction(locomotive, index)
			function.Description = parseECoSInt(arg.Params[1])
		}
	}
}

func ensureECoSFunction(locomotive *ECoSLocomotive, index int) *ECoSFunction {
	for functionIndex := range locomotive.Functions {
		if locomotive.Functions[functionIndex].Index == index {
			return &locomotive.Functions[functionIndex]
		}
	}
	locomotive.Functions = append(locomotive.Functions, ECoSFunction{Index: index})
	position := len(locomotive.Functions) - 1
	for position > 0 && locomotive.Functions[position-1].Index > index {
		locomotive.Functions[position-1], locomotive.Functions[position] =
			locomotive.Functions[position], locomotive.Functions[position-1]
		position--
	}
	return &locomotive.Functions[position]
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

func cleanECoSLocomotiveSyncDesired(input ECoSLocomotiveSyncDesired) ECoSLocomotiveSyncDesired {
	input.Name = strings.TrimSpace(input.Name)
	input.Protocol = strings.TrimSpace(input.Protocol)
	if input.Address < 0 {
		input.Address = 0
	}
	return input
}

func buildECoSLocomotiveSyncCommand(objectID int, current *ECoSLocomotive, desired ECoSLocomotiveSyncDesired) ([]ECoSLocomotiveSyncChange, string, error) {
	if current == nil {
		return nil, "", errors.New("ECoS-Istwerte fehlen")
	}
	changes := []ECoSLocomotiveSyncChange{}
	parts := []string{}
	if desired.Name != "" && desired.Name != current.Name {
		if strings.IndexFunc(desired.Name, unicode.IsControl) >= 0 {
			return nil, "", errors.New("ECoS-Name enthielt unzulässige Zeichen")
		}
		changes = append(changes, ECoSLocomotiveSyncChange{Field: "name", Current: current.Name, Desired: desired.Name})
		parts = append(parts, "name["+quoteECoSString(desired.Name)+"]")
	}
	if desired.Address > 0 && desired.Address != current.Address {
		changes = append(changes, ECoSLocomotiveSyncChange{Field: "address", Current: strconv.Itoa(current.Address), Desired: strconv.Itoa(desired.Address)})
		parts = append(parts, fmt.Sprintf("addr[%d]", desired.Address))
	}
	if desired.Protocol != "" && desired.Protocol != current.Protocol {
		if !isSafeECoSToken(desired.Protocol) {
			return nil, "", errors.New("ECoS-Protokoll enthielt unzulässige Zeichen")
		}
		changes = append(changes, ECoSLocomotiveSyncChange{Field: "protocol", Current: current.Protocol, Desired: desired.Protocol})
		parts = append(parts, "protocol["+desired.Protocol+"]")
	}
	if len(parts) == 0 {
		return changes, "", nil
	}
	return changes, fmt.Sprintf("set(%d, %s)", objectID, strings.Join(parts, ", ")), nil
}

func eCoSLocomotiveSyncFields(changes []ECoSLocomotiveSyncChange) []string {
	fields := make([]string, 0, len(changes))
	for _, change := range changes {
		switch change.Field {
		case "address":
			fields = append(fields, "addr")
		default:
			fields = append(fields, change.Field)
		}
	}
	return fields
}

func quoteECoSString(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, `"`, `\"`)
	return `"` + value + `"`
}

func isSafeECoSToken(value string) bool {
	if value == "" {
		return false
	}
	for _, char := range value {
		if (char >= 'A' && char <= 'Z') || (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '_' || char == '-' || char == '.' || char == ':' {
			continue
		}
		return false
	}
	return true
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
