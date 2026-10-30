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

type ECoSFunctionValue struct {
	Index       int  `json:"index"`
	Description int  `json:"description"`
	Momentary   bool `json:"momentary"`
}

type ECoSFunctionSnapshot struct {
	ObjectID  int                 `json:"objectId"`
	Host      string              `json:"host"`
	ReadAt    string              `json:"readAt"`
	Functions []ECoSFunctionValue `json:"functions"`
}

// ESU PC interface, edition 3, section 7.11: funcexists distinguishes hidden
// (-1) and unavailable (-2) keys; the optional literal "moment" is a key mode.
var ecosFunctionLine = regexp.MustCompile(`^(\d+) funcexists\[\s*(\d+)\s*,\s*(-?\d+)\s*(,\s*moment\s*)?\]$`)

func (s *ECoSService) ReadFunctions(ctx context.Context, input ECoSConnectionInput, objectID int) ([]ECoSFunctionValue, error) {
	if objectID < 1 || objectID > maxDigitalCenterObjectID {
		return nil, ErrDigitalCenterDeviceOutput
	}
	target, err := normalizeECoSInput(input)
	if err != nil {
		return nil, err
	}
	client := s.eCoSClient()
	conn, reader, err := client.Dial(ctx, ecospkg.Target{Host: target.Host, Port: target.Port})
	if err != nil {
		return nil, err
	}
	defer func() { _ = conn.Close() }()
	stop := context.AfterFunc(ctx, func() { _ = conn.Close() })
	defer stop()
	exchange := func(command string) ([]string, error) {
		if err := client.Send(conn, command); err != nil {
			return nil, err
		}
		return readECoSFunctionReply(conn, reader, command)
	}
	if _, err := exchange(fmt.Sprintf("request(%d, view)", objectID)); err != nil {
		return nil, err
	}
	values := []ECoSFunctionValue{}
	for index := 0; index < 32; index++ {
		lines, err := exchange(fmt.Sprintf("get(%d, funcexists[%d])", objectID, index))
		if err != nil {
			return nil, err
		}
		if len(lines) != 3 {
			return nil, ErrDigitalCenterDeviceOutput
		}
		match := ecosFunctionLine.FindStringSubmatch(lines[1])
		if match == nil {
			return nil, ErrDigitalCenterDeviceOutput
		}
		id, idErr := strconv.Atoi(match[1])
		key, keyErr := strconv.Atoi(match[2])
		// Keep opaque station descriptors, including codes outside our symbol catalog.
		// The signed 32-bit parse bounds input without guessing icon bits or meanings.
		code, codeErr := strconv.ParseInt(match[3], 10, 32)
		if idErr != nil || keyErr != nil || codeErr != nil || id != objectID || key != index || code < -2 {
			return nil, ErrDigitalCenterDeviceOutput
		}
		if code >= 0 {
			values = append(values, ECoSFunctionValue{Index: index, Description: int(code), Momentary: match[4] != ""})
		}
	}
	if _, err := exchange(fmt.Sprintf("release(%d, view)", objectID)); err != nil {
		return nil, err
	}
	return values, nil
}

func readECoSFunctionReply(conn net.Conn, reader *bufio.Reader, command string) ([]string, error) {
	if err := conn.SetReadDeadline(time.Now().Add(1200 * time.Millisecond)); err != nil {
		return nil, err
	}
	// View events may precede a reply. Bound both events and individual lines;
	// only the exact requested reply with a successful end marker is usable.
	for block := 0; block < 8; block++ {
		lines := []string{}
		for line := 0; line < 64; line++ {
			part, err := reader.ReadSlice('\n')
			if err != nil {
				return nil, err
			}
			value := strings.TrimSpace(string(part))
			if value == "" {
				continue
			}
			lines = append(lines, value)
			if strings.HasPrefix(value, "<END ") {
				break
			}
		}
		if len(lines) >= 2 && strings.HasPrefix(lines[0], "<EVENT ") && lines[len(lines)-1] == "<END 0 (OK)>" {
			continue
		}
		if len(lines) < 2 || lines[0] != "<REPLY "+command+">" || lines[len(lines)-1] != "<END 0 (OK)>" {
			return nil, ErrDigitalCenterDeviceOutput
		}
		return lines, nil
	}
	return nil, ErrDigitalCenterDeviceOutput
}
