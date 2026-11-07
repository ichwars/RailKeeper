package application

import (
	"context"
	"fmt"
	"time"

	ecospkg "railkeeper/backend/internal/ecos"
)

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
