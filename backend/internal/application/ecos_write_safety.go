package application

import (
	"context"
	"errors"
)

var ErrECoSWriteStateUnknown = errors.New("ECoS write state is unknown")

func (s *ECoSService) ListLocomotives(
	ctx context.Context,
	input ECoSConnectionInput,
) ([]ECoSLocomotive, error) {
	target, err := normalizeECoSInput(input)
	if err != nil {
		return nil, err
	}
	lines, err := s.exchange(ctx, target.Host, target.Port, eCoSLocomotiveListCommand)
	if err != nil {
		return nil, err
	}
	return parseECoSLocomotives(lines), nil
}

func (s *ECoSService) ReadLocomotive(
	ctx context.Context,
	input ECoSConnectionInput,
	objectID int,
) (ECoSLocomotive, error) {
	target, err := normalizeECoSInput(input)
	if err != nil {
		return ECoSLocomotive{}, err
	}
	if objectID < 1 || objectID > maxDigitalCenterObjectID {
		return ECoSLocomotive{}, ErrDigitalCenterDeviceOutput
	}
	locomotive, err := s.fetchLocomotiveDetails(ctx, target.Host, target.Port, objectID)
	if err != nil {
		return ECoSLocomotive{}, err
	}
	return *locomotive, nil
}
