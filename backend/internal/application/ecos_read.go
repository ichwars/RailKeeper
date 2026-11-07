package application

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

func (s *ECoSService) ProbeLocomotiveRaw(ctx context.Context, input ECoSConnectionInput) (*ECoSRawProbe, error) {
	target, err := normalizeECoSInput(input)
	if err != nil {
		return nil, err
	}
	lines, err := s.exchange(ctx, target.Host, target.Port, eCoSLocomotiveListCommand)
	if err != nil {
		return nil, err
	}
	locomotives := parseECoSLocomotives(lines)
	probeFields := eCoSRawProbeFields()
	rawLocomotives := make([]ECoSRawLocomotive, 0, len(locomotives))
	for _, locomotive := range locomotives {
		raw := ECoSRawLocomotive{
			ObjectID:   locomotive.ObjectID,
			Name:       locomotive.Name,
			Address:    locomotive.Address,
			Protocol:   locomotive.Protocol,
			Attributes: map[string][]string{},
		}
		if strings.EqualFold(strings.TrimSpace(locomotive.Protocol), "MULTI") {
			rawLocomotives = append(rawLocomotives, raw)
			continue
		}
		probes, err := s.fetchLocomotiveRawProbes(ctx, target.Host, target.Port, locomotive.ObjectID)
		if err != nil {
			raw.DetailError = err.Error()
			rawLocomotives = append(rawLocomotives, raw)
			continue
		}
		raw.Probes = probes
		supported := map[string]bool{}
		missing := map[string]bool{}
		for _, probe := range probes {
			if !probe.OK || probe.Error != "" {
				continue
			}
			for _, detail := range parseECoSLocomotives(probe.RawLines) {
				if detail.ObjectID == raw.ObjectID {
					mergeECoSRawLocomotive(&raw, detail)
				}
			}
			objectLines := []string{}
			for _, line := range probe.RawLines {
				if strings.HasPrefix(line, strconv.Itoa(raw.ObjectID)+" ") {
					objectLines = append(objectLines, line)
				}
			}
			for key, values := range parseECoSAttributes(objectLines) {
				raw.Attributes[key] = append(raw.Attributes[key], values...)
				supported[strings.ToLower(key)] = true
			}
			for _, field := range probe.Fields {
				field = strings.ToLower(field)
				if !supported[field] {
					missing[field] = true
				}
			}
		}
		raw.SupportedFields = sortedECoSFieldNames(supported)
		raw.MissingFields = sortedMissingECoSFieldNames(missing, supported)
		raw.CVs = parseECoSCVValues(raw.Attributes)
		raw.InterestingFields = interestingECoSFields(raw.Attributes)
		rawLocomotives = append(rawLocomotives, raw)
	}
	return &ECoSRawProbe{
		Host:        target.Host,
		Port:        target.Port,
		ProbeFields: probeFields,
		Locomotives: rawLocomotives,
		RawLines:    lines,
		Message:     fmt.Sprintf("%d ECoS-Lokomotiven roh geprüft.", len(rawLocomotives)),
	}, nil
}

func (s *ECoSService) CountLocomotives(ctx context.Context, input ECoSConnectionInput) (*ECoSLocomotiveSummary, error) {
	target, err := normalizeECoSInput(input)
	if err != nil {
		return nil, err
	}
	lines, err := s.exchange(ctx, target.Host, target.Port, eCoSLocomotiveListCommand)
	if err != nil {
		return nil, err
	}
	count := len(parseECoSLocomotives(lines))
	return &ECoSLocomotiveSummary{
		Host:    target.Host,
		Port:    target.Port,
		Count:   count,
		Message: fmt.Sprintf("%d ECoS-Lokdatensätze gefunden.", count),
	}, nil
}

func (s *ECoSService) fetchLocomotiveDetails(ctx context.Context, host string, port int, objectID int) (*ECoSLocomotive, error) {
	command := eCoSLocomotiveDetailCommand(objectID)
	lines, err := s.exchangeRequestedGet(ctx, host, port, objectID, command)
	if err != nil {
		return nil, err
	}
	locomotives := parseECoSLocomotives(lines)
	if len(locomotives) == 0 {
		return nil, errors.New("keine Detaildaten gelesen")
	}
	for index := range locomotives {
		if locomotives[index].ObjectID == objectID {
			return &locomotives[index], nil
		}
	}
	return nil, errors.New("ECoS-Detailantwort enthält das angeforderte Objekt nicht")
}

func (s *ECoSService) fetchLocomotiveRawProbes(ctx context.Context, host string, port int, objectID int) ([]ECoSRawCommandProbe, error) {
	commands := []struct {
		command string
		fields  []string
	}{
		{
			command: eCoSLocomotiveDetailCommand(objectID),
			fields:  []string{"profile", "protocol", "name", "addr", "funcdesc"},
		},
	}
	for _, field := range eCoSRawProbeFields() {
		commands = append(commands, struct {
			command string
			fields  []string
		}{
			command: fmt.Sprintf("get(%d, %s)", objectID, field),
			fields:  []string{field},
		})
	}
	for _, field := range eCoSTargetedCVProbeFields() {
		commands = append(commands, struct {
			command string
			fields  []string
		}{
			command: fmt.Sprintf("get(%d, %s)", objectID, field),
			fields:  []string{field},
		})
	}
	return s.exchangeRequestedCommands(ctx, host, port, objectID, commands)
}
