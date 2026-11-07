package application

import (
	"context"
	"errors"
	"fmt"
)

func (s *ECoSService) SyncLocomotive(ctx context.Context, input ECoSLocomotiveSyncInput) (*ECoSLocomotiveSyncResult, error) {
	target, err := normalizeECoSInput(ECoSConnectionInput{Host: input.Host, Port: input.Port})
	if err != nil {
		return nil, err
	}
	if input.ObjectID <= 0 {
		return nil, errors.New("ECoS-Objekt-ID ist erforderlich")
	}

	current, err := s.fetchLocomotiveDetails(ctx, target.Host, target.Port, input.ObjectID)
	if err != nil {
		return nil, err
	}
	desired := cleanECoSLocomotiveSyncDesired(input.Desired)
	changes, commands, err := buildECoSLocomotiveSyncCommands(input.ObjectID, current, desired)
	if err != nil {
		return nil, err
	}

	result := &ECoSLocomotiveSyncResult{
		Host:     target.Host,
		Port:     target.Port,
		ObjectID: input.ObjectID,
		DryRun:   input.DryRun || !input.Confirm,
		Current: ECoSLocomotiveSyncSnapshot{
			Name:     current.Name,
			Address:  current.Address,
			Protocol: current.Protocol,
		},
		Desired: desired,
		Changes: changes,
		Message: "ECoS-Sync-Vorschau erstellt.",
	}
	if len(commands) > 0 {
		result.Commands = append([]string(nil), commands...)
	}
	if len(changes) == 0 {
		result.Message = "ECoS-Lok ist bereits synchron."
		return result, nil
	}
	if result.DryRun {
		return result, nil
	}

	writeCommands := make([]struct {
		command string
		fields  []string
	}, 0, len(commands))
	for index, command := range commands {
		writeCommands = append(writeCommands, struct {
			command string
			fields  []string
		}{command: command, fields: eCoSLocomotiveSyncFields(changes[index : index+1])})
	}
	probes, err := s.exchangeControlledCommands(ctx, target.Host, target.Port, input.ObjectID, writeCommands)
	if err != nil {
		return nil, err
	}
	if len(probes) == 0 {
		return nil, errors.New("ECoS hat keine Schreibantwort geliefert")
	}
	confirmedWrites := 0
	for _, probe := range probes {
		result.RawLines = append(result.RawLines, probe.RawLines...)
		if probe.OK {
			confirmedWrites++
			continue
		}
		if probe.Error != "" || confirmedWrites > 0 {
			return nil, fmt.Errorf("%w: ECoS-Schreibantwort fehlt", ErrECoSWriteStateUnknown)
		}
		return nil, fmt.Errorf("ECoS-Schreibbefehl nicht bestätigt: %s", firstNonEmpty(probe.Status, "unbekannter Status"))
	}
	if len(probes) != len(commands) {
		return nil, fmt.Errorf("%w: ECoS-Schreibantwort fehlt", ErrECoSWriteStateUnknown)
	}
	result.Applied = confirmedWrites > 0
	result.Message = "ECoS-Lok wurde geschrieben."
	return result, nil
}
