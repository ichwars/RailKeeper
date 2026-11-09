package application

import (
	"fmt"
	"railkeeper/backend/internal/ecosmodel"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

func normalizeDigitalCenterRoster(provider string, probe *ECoSRawProbe) ([]ECoSRawLocomotive, int, error) {
	if probe == nil {
		return nil, 0, fmt.Errorf("%w: empty ECoS response", ErrDigitalCenterDeviceOutput)
	}
	if len(probe.Locomotives) > maxDigitalCenterLocomotives {
		return nil, 0, fmt.Errorf("%w: too many locomotives", ErrDigitalCenterDeviceOutput)
	}
	skippedConsists := 0
	normalized := make([]ECoSRawLocomotive, 0, len(probe.Locomotives))
	seen := make(map[int]struct{}, len(probe.Locomotives))
	for _, locomotive := range probe.Locomotives {
		if locomotive.ObjectID < 1 || locomotive.ObjectID > maxDigitalCenterObjectID {
			return nil, 0, fmt.Errorf("%w: invalid object ID", ErrDigitalCenterDeviceOutput)
		}
		if _, found := seen[locomotive.ObjectID]; found {
			return nil, 0, fmt.Errorf("%w: duplicate object ID", ErrDigitalCenterDeviceOutput)
		}
		seen[locomotive.ObjectID] = struct{}{}
		name, err := normalizeDigitalCenterName(locomotive.Name)
		if err != nil {
			return nil, 0, err
		}
		protocol, err := normalizeDigitalCenterProtocol(locomotive.Protocol)
		if err != nil {
			return nil, 0, err
		}
		if provider == "ecos" && strings.EqualFold(strings.TrimSpace(locomotive.Protocol), "MULTI") {
			skippedConsists++
			continue
		}
		if locomotive.Address < 1 || locomotive.Address > maxDigitalCenterAddress {
			return nil, 0, fmt.Errorf("%w: invalid decoder address", ErrDigitalCenterDeviceOutput)
		}
		cvs, err := validateStoredECoSCVs(locomotive.CVs)
		if err != nil {
			return nil, 0, err
		}
		normalized = append(normalized, ECoSRawLocomotive{
			ObjectID: locomotive.ObjectID, Name: name, Address: locomotive.Address,
			Protocol: protocol, DetailError: strings.TrimSpace(locomotive.DetailError),
			CVs: cvs, Profile: locomotive.Profile, CVRead: locomotive.CVRead,
			CVSources: locomotive.CVSources, FunctionSnapshot: locomotive.FunctionSnapshot,
		})
	}
	return normalized, skippedConsists, nil
}

func normalizeDigitalCenterName(value string) (string, error) {
	if !utf8.ValidString(value) {
		return "", fmt.Errorf("%w: locomotive name is not valid UTF-8", ErrDigitalCenterDeviceOutput)
	}
	value = strings.Map(func(character rune) rune {
		if unicode.IsControl(character) || unicode.In(character, unicode.Cf, unicode.Co, unicode.Cs) {
			return ' '
		}
		return character
	}, value)
	value = strings.Join(strings.Fields(value), " ")
	if utf8.RuneCountInString(value) > maxDigitalCenterNameRunes {
		return "", fmt.Errorf("%w: locomotive name is too long", ErrDigitalCenterDeviceOutput)
	}
	return value, nil
}

func normalizeDigitalCenterProtocol(value string) (string, error) {
	normalized, err := ecosmodel.NormalizeProtocol(value, maxDigitalCenterProtocol)
	if err != nil {
		return "", fmt.Errorf("%w: %s", ErrDigitalCenterDeviceOutput, err)
	}
	return normalized, nil
}

func validDigitalCenterCompareStatus(status DigitalCenterCompareStatus) bool {
	switch status {
	case "", DigitalCompareOK, DigitalCompareDeviation, DigitalCompareMissing, DigitalCompareNew,
		DigitalCompareConflict:
		return true
	default:
		return false
	}
}

func digitalCenterWorkItemContains(item DigitalCenterWorkItem, query string) bool {
	fields := []string{
		item.Name, item.Protocol, item.VehicleID, item.CenterObjectID, strconv.Itoa(item.Address),
	}
	for _, field := range fields {
		if strings.Contains(strings.ToLower(field), query) {
			return true
		}
	}
	return false
}

func digitalCenterTimestamp() string {
	return time.Now().UTC().Format(time.RFC3339)
}
