package application

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	ecospkg "railkeeper/backend/internal/ecos"
)

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
	s.liveOnInterrupted = onInterrupted
	s.liveCancel = cancel
	done := make(chan struct{})
	s.liveDone = done
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
		liveCtx, cancel, generation, onInterrupted, conn, reader, client,
		append([]string(nil), commands...), done,
	)
	return &status, nil
}

func (s *ECoSService) StopLive() ECoSLiveStatus {
	s.liveMu.Lock()
	defer s.liveMu.Unlock()
	s.flushCurrentLivePulseLocked(time.Now().UTC())
	s.stopLiveLocked()
	s.liveOnInterrupted = nil
	s.liveGeneration++
	s.liveStatus.Connected = false
	s.liveStatus.State = ECoSLiveStopped
	s.liveStatus.Diagnosis.ConnectionState = ECoSLiveStopped
	s.liveStatus.Diagnosis.LastError = ""
	s.liveStatus.Message = "ECoS-Live-Verbindung beendet."
	s.liveStatus.LastMessage = "Verbindung beendet."
	return cloneECoSLiveStatus(s.liveStatus)
}

func (s *ECoSService) PauseLive(ctx context.Context) (ECoSLiveStatus, error) {
	s.liveMu.Lock()
	done := s.liveDone
	s.stopLiveLocked()
	s.liveGeneration++
	s.liveStatus.Connected = false
	s.liveStatus.State = ECoSLiveStopped
	s.liveStatus.Diagnosis.ConnectionState = ECoSLiveStopped
	s.liveStatus.Message = "ECoS-Live-Verbindung für einen exklusiven Vorgang pausiert."
	status := cloneECoSLiveStatus(s.liveStatus)
	s.liveMu.Unlock()
	if done == nil {
		return status, nil
	}
	select {
	case <-done:
		return status, nil
	case <-ctx.Done():
		return status, fmt.Errorf("ECoS-Live-Verbindung pausieren: %w", ctx.Err())
	}
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
	done chan struct{},
) {
	defer func() {
		cancel()
		_ = conn.Close()
		s.updateLiveErrorForGenerationWithCallback(
			generation, errors.New("ECoS live connection closed"), onInterrupted,
		)
		close(done)
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
