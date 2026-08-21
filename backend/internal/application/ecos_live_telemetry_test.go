package application

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	ecospkg "railkeeper/backend/internal/ecos"
)

func TestECoSLiveTelemetryIsBoundedMaskedAndCopied(t *testing.T) {
	service := NewECoSService()
	service.liveStatus = ECoSLiveStatus{Provider: "ecos", Connected: true, State: ECoSLiveRunning}
	start := time.Date(2026, 8, 21, 8, 0, 0, 0, time.UTC)
	for index := 0; index < 121; index++ {
		at := start.Add(time.Duration(index) * time.Second)
		service.updateLiveBlocksAt(at, []ecospkg.Block{
			{Kind: ecospkg.BlockEvent, ObjectID: index + 1, Header: `<EVENT 192.168.2.151 password["secret"]>`},
			{Kind: ecospkg.BlockReply, Header: `<REPLY get(1, password)>`},
		}, "\x00<REPLY get(1, password)> token=private\r\n")
	}

	status := service.LiveStatus()
	if len(status.PulseSamples) != 60 || len(status.RecentEvents) != 100 {
		t.Fatalf("telemetry bounds exceeded: samples=%d events=%d", len(status.PulseSamples), len(status.RecentEvents))
	}
	for _, event := range status.RecentEvents {
		text := strings.ToLower(event.Message + " " + event.Protocol)
		for _, forbidden := range []string{"<reply", "<event", "password", "secret", "token", "192.168.2.151", "\x00"} {
			if strings.Contains(text, forbidden) {
				t.Fatalf("unmasked event: %#v", event)
			}
		}
	}
	if strings.Contains(strings.ToLower(status.LastMessage), "password") || strings.Contains(status.LastMessage, "<") {
		t.Fatalf("unmasked last message: %q", status.LastMessage)
	}
	if status.BlocksReceived != 242 || status.RepliesReceived != 121 || status.EventsReceived != 121 ||
		status.LastSeenAt != start.Add(120*time.Second).Format(time.RFC3339) {
		t.Fatalf("live counters or last-message time are incorrect: %#v", status)
	}

	status.PulseSamples[0].RepliesPerSecond = 999
	status.RecentEvents[0].Message = "mutated"
	again := service.LiveStatus()
	if again.PulseSamples[0].RepliesPerSecond == 999 || again.RecentEvents[0].Message == "mutated" {
		t.Fatal("live status returned mutable telemetry slices")
	}
}

func TestECoSLiveDisconnectMarksInterruptedWithoutAdvancingLastMessageTime(t *testing.T) {
	service := NewECoSService()
	lastSeen := "2026-08-21T08:00:00Z"
	service.liveStatus = ECoSLiveStatus{
		Provider: "ecos", Connected: true, State: ECoSLiveRunning, LastSeenAt: lastSeen,
	}

	service.updateLiveError(errors.New("dial 192.168.2.151: password=secret\x00"))
	status := service.LiveStatus()

	if status.Connected || status.State != ECoSLiveInterrupted {
		t.Fatalf("disconnect status = %#v, want interrupted", status)
	}
	if status.LastSeenAt != lastSeen {
		t.Fatalf("last message time advanced after disconnect: got %q want %q", status.LastSeenAt, lastSeen)
	}
	for _, text := range []string{status.Error, status.Message, status.LastMessage, status.Diagnosis.LastError} {
		lower := strings.ToLower(text)
		if strings.Contains(lower, "192.168.2.151") || strings.Contains(lower, "password") ||
			strings.Contains(lower, "secret") || strings.Contains(text, "\x00") {
			t.Fatalf("disconnect text leaked private data: %q", text)
		}
	}
}

func TestECoSLiveCommandsRemainPassive(t *testing.T) {
	want := eCoSLiveSubscriptionCommands()
	if len(want) == 0 {
		t.Fatal("expected passive subscription commands")
	}
	joined := strings.ToLower(strings.Join(want, "\n"))
	for _, forbidden := range []string{
		"set(", "create(", "delete(", "speed", "dir", "func", "switch", "accessory", "cv",
	} {
		if strings.Contains(joined, forbidden) {
			t.Fatalf("live monitor emitted active command %q in %#v", forbidden, want)
		}
	}

	emitted := make(chan string, 8)
	listener := startECoSTestServer(t, func(command string) []string {
		emitted <- command
		return []string{fmt.Sprintf("<REPLY %s>", command), "<END 0 (OK)>"}
	})
	defer func() { _ = listener.Close() }()
	host, port := splitTestAddress(t, listener.Addr().String())
	service := NewECoSService()
	started, err := service.StartLive(context.Background(), ECoSConnectionInput{Host: host, Port: port})
	if err != nil {
		t.Fatal(err)
	}
	if !started.Connected || started.State != ECoSLiveRunning || !started.Diagnosis.Passive {
		t.Fatalf("started live status = %#v", started)
	}
	for _, expected := range want {
		select {
		case command := <-emitted:
			if command != expected {
				t.Fatalf("emitted command = %q, want %q", command, expected)
			}
		case <-time.After(time.Second):
			t.Fatalf("timed out waiting for passive command %q", expected)
		}
	}
	select {
	case command := <-emitted:
		t.Fatalf("unexpected extra live command: %q", command)
	case <-time.After(100 * time.Millisecond):
	}
	status := service.StopLive()
	if status.Connected || status.State != ECoSLiveStopped || !status.Diagnosis.Passive {
		t.Fatalf("stopped live status = %#v", status)
	}
}
