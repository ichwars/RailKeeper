package ecos

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

const DefaultPort = 15471

type Target struct {
	Host string
	Port int
}

type Client struct {
	Timeout time.Duration
}

func NewClient(timeout time.Duration) Client {
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	return Client{Timeout: timeout}
}

func NormalizeTarget(host string, port int) (Target, error) {
	host = strings.TrimSpace(host)
	if host == "" {
		return Target{}, errors.New("ECoS-IP oder Hostname fehlt")
	}
	if port == 0 {
		port = DefaultPort
	}
	if port < 1 || port > 65535 {
		return Target{}, errors.New("ECoS-Port muss zwischen 1 und 65535 liegen")
	}
	return Target{Host: host, Port: port}, nil
}

func (c Client) Dial(ctx context.Context, target Target) (net.Conn, *bufio.Reader, error) {
	timeout := c.effectiveTimeout()
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	dialer := net.Dialer{Timeout: timeout}
	conn, err := dialer.DialContext(ctx, "tcp", net.JoinHostPort(target.Host, strconv.Itoa(target.Port)))
	if err != nil {
		return nil, nil, fmt.Errorf("ECoS nicht erreichbar: %w", err)
	}
	return conn, bufio.NewReader(conn), nil
}

func (c Client) Send(conn net.Conn, command string) error {
	command = strings.TrimSpace(command)
	if command == "" {
		return errors.New("ECoS-Kommando fehlt")
	}
	if err := conn.SetWriteDeadline(time.Now().Add(c.effectiveTimeout())); err != nil {
		return fmt.Errorf("ECoS-Zeitlimit konnte nicht gesetzt werden: %w", err)
	}
	if _, err := fmt.Fprintf(conn, "%s\r\n", command); err != nil {
		return fmt.Errorf("ECoS-Kommando konnte nicht gesendet werden: %w", err)
	}
	return nil
}

func (c Client) Exchange(ctx context.Context, target Target, command string) ([]string, error) {
	timeout := c.effectiveTimeout()
	conn, reader, err := c.Dial(ctx, target)
	if err != nil {
		return nil, err
	}
	defer func() { _ = conn.Close() }()

	if err := c.Send(conn, command); err != nil {
		return nil, err
	}
	lines, err := ReadReply(conn, reader, timeout)
	if err != nil {
		return nil, fmt.Errorf("ECoS-Antwort konnte nicht gelesen werden: %w", err)
	}
	if len(lines) == 0 {
		return nil, errors.New("ECoS hat keine Antwort geliefert")
	}
	return lines, nil
}

func ReadReply(conn net.Conn, reader *bufio.Reader, timeout time.Duration) ([]string, error) {
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
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

func (c Client) effectiveTimeout() time.Duration {
	if c.Timeout <= 0 {
		return 5 * time.Second
	}
	return c.Timeout
}
