package application

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	ecospkg "railkeeper/backend/internal/ecos"
)

func (s *ECoSService) TestConnection(ctx context.Context, input ECoSConnectionInput) (*ECoSConnectionResult, error) {
	target, err := normalizeECoSInput(input)
	if err != nil {
		return nil, err
	}
	lines, err := s.exchange(ctx, target.Host, target.Port, "get(1, info, status)")
	result := &ECoSConnectionResult{
		Connected:           false,
		Host:                target.Host,
		Port:                target.Port,
		CompatibilityStatus: ECoSCompatibilityUnreachable,
		MissingFields:       []string{},
		Message:             "ECoS-Verbindung konnte nicht aufgebaut werden.",
	}
	if err != nil {
		result.Message = err.Error()
		return result, nil //nolint:nilerr // Connection failures are returned as preview results.
	}
	fields := parseECoSFields(lines)
	result.Status = fields["status"]
	result.ProtocolVersion = firstNonEmpty(fields["ProtocolVersion"], fields["protocolversion"])
	result.ApplicationVersion = firstNonEmpty(fields["ApplicationVersion"], fields["applicationversion"])
	result.HardwareVersion = firstNonEmpty(fields["HardwareVersion"], fields["hardwareversion"])
	result.RawLines = lines
	result.Fields = fields
	endStatus, endOK := parseECoSEndStatus(lines)
	if endStatus == "" {
		result.CompatibilityStatus = ECoSCompatibilityPartial
		result.Message = "ECoS-Antwort war unvollständig: Abschlussstatus fehlt."
		return result, nil
	}
	result.Fields["endStatus"] = endStatus
	if !endOK {
		result.CompatibilityStatus = ECoSCompatibilityRejected
		result.Message = fmt.Sprintf("ECoS hat den Verbindungstest abgelehnt: %s.", endStatus)
		return result, nil
	}
	result.MissingFields = missingECoSConnectionFields(result)
	if len(result.MissingFields) > 0 {
		result.CompatibilityStatus = ECoSCompatibilityPartial
		result.Message = fmt.Sprintf(
			"ECoS-Antwort war unvollständig: Pflichtfelder fehlen (%s).",
			strings.Join(result.MissingFields, ", "),
		)
		return result, nil
	}
	result.Connected = true
	if result.ApplicationVersion == "4.3.3" {
		result.CompatibilityStatus = ECoSCompatibilityVerified
		result.Message = "ECoS-Verbindung erfolgreich. Firmware 4.3.3 ist geräteverifiziert."
	} else {
		result.CompatibilityStatus = ECoSCompatibilityUnverified
		result.Message = fmt.Sprintf(
			"ECoS-Verbindung erfolgreich. Firmware %s ist noch nicht geräteverifiziert.",
			result.ApplicationVersion,
		)
	}
	return result, nil
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
