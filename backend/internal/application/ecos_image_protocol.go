package application

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
)

var ErrECoSImageControlBusy = errors.New("ECoS locomotive control is held by another controller")

// ESU PC interface specification, third edition, section 7.11 (locodesc).
// Keep image commands separate from the read-only locomotive/CV import commands.
type ECoSImageDescriptor struct {
	LocomotiveType string `json:"locomotiveType"`
	ImageType      string `json:"imageType"`
	Index          int    `json:"index"`
}

var ecosImageDescriptorLine = regexp.MustCompile(`^(\d+) locodesc\[(LOCO_TYPE_(?:E|DIESEL|STEAM|MISC)),(IMAGE_TYPE_(?:INT|USER)),(\d+)\]$`)

func (s *ECoSService) ReadImageDescriptor(ctx context.Context, input ECoSConnectionInput, objectID int) (ECoSImageDescriptor, error) {
	if objectID < 1 || objectID > maxDigitalCenterObjectID {
		return ECoSImageDescriptor{}, ErrDigitalCenterDeviceOutput
	}
	command := fmt.Sprintf("get(%d, locodesc)", objectID)
	lines, err := s.exchange(ctx, input.Host, input.Port, command)
	if err != nil {
		return ECoSImageDescriptor{}, err
	}
	return parseECoSImageDescriptor(lines, objectID)
}

func parseECoSImageDescriptor(lines []string, objectID int) (ECoSImageDescriptor, error) {
	if _, ok := parseECoSEndStatus(lines); !ok {
		return ECoSImageDescriptor{}, ErrDigitalCenterDeviceOutput
	}
	var found *ECoSImageDescriptor
	for _, line := range lines {
		match := ecosImageDescriptorLine.FindStringSubmatch(line)
		if match == nil {
			continue
		}
		id, _ := strconv.Atoi(match[1])
		index, err := strconv.Atoi(match[4])
		if err != nil || id != objectID || index > 65535 || found != nil {
			return ECoSImageDescriptor{}, ErrDigitalCenterDeviceOutput
		}
		found = &ECoSImageDescriptor{LocomotiveType: match[2], ImageType: match[3], Index: index}
	}
	if found == nil {
		return ECoSImageDescriptor{}, ErrDigitalCenterDeviceOutput
	}
	return *found, nil
}

// Preserve the workspace interruption callback across the exclusive image write.
func (s *ECoSService) ResumePausedLive(ctx context.Context, input ECoSConnectionInput) (*ECoSLiveStatus, error) {
	s.liveMu.Lock()
	onInterrupted := s.liveOnInterrupted
	s.liveMu.Unlock()
	return s.StartLiveWithInterruption(ctx, input, onInterrupted)
}
