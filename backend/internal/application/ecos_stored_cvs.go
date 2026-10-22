package application

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"

	ecospkg "railkeeper/backend/internal/ecos"
)

const ECoSStoredCVLimit = 1024
const ECoSStoredCVBatchSize = 32

type ECoSStoredCVBatch struct {
	Values      []ECoSCVValue `json:"values"`
	Unavailable int           `json:"unavailable"`
}

var ecosStoredCVLine = regexp.MustCompile(`^(\d+) cv\[(\d+),\s*(\d+)\]$`)

// ESU PC interface, edition 3, section 7.11: get(id, cv[nr]) reads the
// locomotive object's memory. It does not read or program a physical decoder.
func (s *ECoSService) ReadStoredCVs(
	ctx context.Context, input ECoSConnectionInput, objectID, startCV int,
) (ECoSStoredCVBatch, error) {
	result := ECoSStoredCVBatch{Values: []ECoSCVValue{}}
	if objectID < 1 || objectID > maxDigitalCenterObjectID || !validCVNumber(startCV) {
		return result, ErrDigitalCenterFilterValidation
	}
	target, err := normalizeECoSInput(input)
	if err != nil {
		return result, err
	}
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	client := s.eCoSClient()
	conn, reader, err := client.Dial(ctx, ecospkg.Target{Host: target.Host, Port: target.Port})
	if err != nil {
		return result, err
	}
	defer func() { _ = conn.Close() }()
	stop := context.AfterFunc(ctx, func() { _ = conn.Close() })
	defer stop()
	for number := startCV; number <= min(startCV+ECoSStoredCVBatchSize-1, ECoSStoredCVLimit); number++ {
		if err := ctx.Err(); err != nil {
			return result, err
		}
		command := fmt.Sprintf("get(%d, cv[%d])", objectID, number)
		if err := client.Send(conn, command); err != nil {
			return result, err
		}
		lines, err := readECoSStoredCVReply(conn, reader, command)
		if err != nil {
			return result, err
		}
		_, ok := parseECoSEndStatus(lines)
		if !ok {
			// Recognized unavailable responses contain no reading. A broken,
			// mismatched or truncated exchange must not advance the cursor.
			status, _ := parseECoSEndStatus(lines)
			code := strings.Fields(status)
			if len(code) == 0 || (code[0] != "11" && code[0] != "12" && code[0] != "22") {
				return result, ErrDigitalCenterDeviceOutput
			}
			result.Unavailable++
			continue
		}
		value, found, err := parseStoredECoSCVReply(lines, objectID, number)
		if err != nil {
			return result, err
		}
		if found {
			result.Values = append(result.Values, ECoSCVValue{Number: number, Value: value})
		} else {
			result.Unavailable++
		}
	}
	return result, nil
}

func readECoSStoredCVReply(conn net.Conn, reader *bufio.Reader, command string) ([]string, error) {
	if err := conn.SetReadDeadline(time.Now().Add(1200 * time.Millisecond)); err != nil {
		return nil, err
	}
	lines := []string{}
	for i := 0; i < 64; i++ {
		part, err := reader.ReadSlice('\n')
		if err != nil {
			return nil, err
		}
		line := strings.TrimSpace(string(part))
		if len(lines) == 0 && line == "" {
			continue
		}
		if len(lines) == 0 && line != "<REPLY "+command+">" {
			return nil, ErrDigitalCenterDeviceOutput
		}
		lines = append(lines, line)
		if strings.HasPrefix(line, "<END ") && strings.HasSuffix(line, ">") {
			return lines, nil
		}
	}
	return nil, ErrDigitalCenterDeviceOutput
}

func parseStoredECoSCVReply(lines []string, objectID, number int) (int, bool, error) {
	value, found := 0, false
	for _, line := range lines[1 : len(lines)-1] {
		match := ecosStoredCVLine.FindStringSubmatch(line)
		if match == nil {
			return 0, false, ErrDigitalCenterDeviceOutput
		}
		id, _ := strconv.Atoi(match[1])
		cv, _ := strconv.Atoi(match[2])
		parsed, err := strconv.Atoi(match[3])
		if err != nil || id != objectID || cv != number || !validCVValue(parsed) || found {
			return 0, false, ErrDigitalCenterDeviceOutput
		}
		value, found = parsed, true
	}
	return value, found, nil
}
