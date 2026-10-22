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

// Control belongs to the TCP connection. Keep it open across HTTP upload and
// verification; a separate preflight connection would release the reservation.
type ECoSImageControl interface {
	Assign(context.Context, int) error
	Close() error
}

type ecosImageControl struct {
	service   *ECoSService
	conn      net.Conn
	reader    *bufio.Reader
	objectID  int
	stop      func() bool
	closed    bool
	attempted bool
}

func (s *ECoSService) AcquireImageControl(ctx context.Context, input ECoSConnectionInput, objectID int) (ECoSImageControl, error) {
	if objectID < 1 || objectID > maxDigitalCenterObjectID {
		return nil, ErrDigitalCenterDeviceOutput
	}
	target, err := ecospkg.NormalizeTarget(input.Host, input.Port)
	if err != nil {
		return nil, err
	}
	conn, reader, err := s.client.Dial(ctx, target)
	if err != nil {
		return nil, err
	}
	control := &ecosImageControl{service: s, conn: conn, reader: reader, objectID: objectID}
	control.stop = context.AfterFunc(ctx, func() { _ = conn.Close() })
	if err := control.exchange(ctx, fmt.Sprintf("request(%d, control)", objectID)); err != nil {
		control.stop()
		_ = conn.Close()
		return nil, err
	}
	return control, nil
}

func (control *ecosImageControl) exchange(ctx context.Context, command string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := control.service.client.Send(control.conn, command); err != nil {
		return err
	}
	timeout := control.service.timeout
	if deadline, ok := ctx.Deadline(); ok {
		timeout = min(timeout, time.Until(deadline))
	}
	if timeout <= 0 {
		return context.DeadlineExceeded
	}
	lines, err := ecospkg.ReadReply(control.conn, control.reader, timeout)
	if err != nil {
		return err
	}
	if len(lines) < 2 || lines[0] != "<REPLY "+command+">" {
		return ErrDigitalCenterDeviceOutput
	}
	if status, ok := parseECoSEndStatus(lines); !ok {
		if strings.HasPrefix(command, "request(") && strings.HasPrefix(status, "25 ") {
			return ErrECoSImageControlBusy
		}
		return ErrDigitalCenterDeviceWrite
	}
	return nil
}

func (control *ecosImageControl) Assign(ctx context.Context, index int) error {
	if control.closed || control.attempted || index < 0 || index > 499 {
		return ErrDigitalCenterDeviceOutput
	}
	control.attempted = true
	if err := control.exchange(ctx, fmt.Sprintf("set(%d, locodesc[IMAGE_TYPE_USER,%d])", control.objectID, index)); err != nil {
		return fmt.Errorf("%w: %v", ErrECoSWriteStateUnknown, err)
	}
	return nil
}

func (control *ecosImageControl) Close() error {
	if control.closed {
		return nil
	}
	control.closed = true
	control.stop()
	defer func() { _ = control.conn.Close() }()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return control.exchange(ctx, fmt.Sprintf("release(%d, control)", control.objectID))
}
