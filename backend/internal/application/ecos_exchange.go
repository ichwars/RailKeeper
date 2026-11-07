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

	ecospkg "railkeeper/backend/internal/ecos"
)

func (s *ECoSService) exchange(ctx context.Context, host string, port int, command string) ([]string, error) {
	return s.eCoSClient().Exchange(ctx, ecospkg.Target{Host: host, Port: port}, command)
}

func (s *ECoSService) exchangeRequestedCommands(ctx context.Context, host string, port int, objectID int, commands []struct {
	command string
	fields  []string
}) ([]ECoSRawCommandProbe, error) {
	return s.exchangeRegisteredCommands(ctx, host, port, objectID, "view", "view", false, commands)
}

func (s *ECoSService) exchangeControlledCommands(ctx context.Context, host string, port int, objectID int, commands []struct {
	command string
	fields  []string
}) ([]ECoSRawCommandProbe, error) {
	return s.exchangeRegisteredCommands(ctx, host, port, objectID,
		"view, control, force", "view, control", true, commands)
}

func (s *ECoSService) exchangeRegisteredCommands(
	ctx context.Context,
	host string,
	port int,
	objectID int,
	requestOptions string,
	releaseOptions string,
	stopOnFailure bool,
	commands []struct {
		command string
		fields  []string
	},
) ([]ECoSRawCommandProbe, error) {
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
	requestCommand := fmt.Sprintf("request(%d, %s)", objectID, requestOptions)
	if _, err := fmt.Fprintf(conn, "%s\r\n", requestCommand); err != nil {
		return nil, fmt.Errorf("ECoS-Registrierung konnte nicht angefordert werden: %w", err)
	}
	requestReply, err := readECoSReply(conn, reader, timeout)
	if err != nil {
		return nil, fmt.Errorf("ECoS-Registrierungsantwort konnte nicht gelesen werden: %w", err)
	}
	if status, ok := parseECoSEndStatus(requestReply); !ok {
		return nil, fmt.Errorf("ECoS-Registrierung nicht bestätigt: %s",
			firstNonEmpty(status, "fehlender Status"))
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
		if stopOnFailure && !probe.OK {
			break
		}
	}
	_, _ = fmt.Fprintf(conn, "release(%d, %s)\r\n", objectID, releaseOptions)
	_, _ = readECoSReply(conn, reader, replyTimeout)
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
	if status, ok := parseECoSEndStatus(lines); !ok {
		return nil, fmt.Errorf("ECoS-Lok-Detailantwort nicht vollständig bestätigt: %s",
			firstNonEmpty(status, "fehlender Status"))
	}
	return lines, nil
}
