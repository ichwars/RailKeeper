package application

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

type digitalCenterStoredRosterReader interface {
	ReadStoredLocomotiveRoster(context.Context, ECoSConnectionInput) (*ECoSRawProbe, error)
}

// ReadStoredLocomotiveRoster reads station metadata, saved function keys and the
// HTTP database/backup. It deliberately never invokes the diagnostic CV probes.
func (s *ECoSService) ReadStoredLocomotiveRoster(ctx context.Context, input ECoSConnectionInput) (*ECoSRawProbe, error) {
	target, err := normalizeECoSInput(input)
	if err != nil {
		return nil, err
	}
	sourceHost := target.Host
	web, err := s.storedWeb(ctx, target.Host)
	if err != nil {
		return nil, fmt.Errorf("%w: database transport unavailable", ErrDigitalCenterDeviceOutput)
	}
	web = &storedRosterWeb{storedECoSWeb: web}
	target.Host = web.Address()
	before, err := s.ListLocomotives(ctx, target)
	if err != nil {
		return nil, err
	}
	if len(before) > maxDigitalCenterLocomotives {
		return nil, ErrDigitalCenterDeviceOutput
	}
	backup := s.readStoredBackup(ctx, web)
	probe := &ECoSRawProbe{Host: sourceHost, Port: target.Port,
		ProbeFields: []string{"profile", "protocol", "name", "addr", "funcdesc", "funcexists", "database", "backup"}}
	for _, listed := range before {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		raw := ECoSRawLocomotive{ObjectID: listed.ObjectID, Name: listed.Name, Address: listed.Address, Protocol: listed.Protocol}
		if strings.EqualFold(listed.Protocol, "MULTI") {
			probe.Locomotives = append(probe.Locomotives, raw)
			continue
		}
		detail, readErr := s.ReadLocomotive(ctx, target, listed.ObjectID)
		if readErr == nil && !sameStoredECoSIdentity(listed, detail, false) {
			return nil, ErrDigitalCenterReadNotFresh
		}
		if readErr == nil {
			raw.Profile = detail.Profile
			var snapshot ECoSCVDatabase
			snapshot, readErr = s.readStoredCVSnapshot(ctx, target, web, detail, backup)
			raw.CVSources = snapshot.Sources
			raw.FunctionSnapshot = snapshot.Functions
			if snapshot.Functions != nil {
				snapshot.Functions.Host = sourceHost
			}
			if readErr == nil {
				raw.CVs, raw.CVSources = snapshot.Values, snapshot.Sources
				raw.CVRead = &ECoSCVReadState{Source: "database", NextCV: ECoSStoredCVLimit + 1,
					ReadAt: time.Now().UTC().Format(time.RFC3339), SourceHost: sourceHost}
			}
		}
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if errors.Is(readErr, ErrDigitalCenterReadNotFresh) {
			return nil, readErr
		}
		if readErr != nil {
			raw.DetailError = "Gespeicherte Lokdaten konnten nicht vollständig gelesen werden. / Stored locomotive data could not be fully read."
		}
		probe.Locomotives = append(probe.Locomotives, raw)
	}
	after, err := s.ListLocomotives(ctx, target)
	if err != nil || !sameStoredECoSRoster(before, after) {
		return nil, ErrDigitalCenterReadNotFresh
	}
	return probe, nil
}

func (s *ECoSService) readStoredCVSnapshot(ctx context.Context, target ECoSConnectionInput, web storedECoSWeb,
	before ECoSLocomotive, backup storedECoSBackup,
) (ECoSCVDatabase, error) {
	exported, exportErr := web.Read(ctx, before.ObjectID, before.Name)
	if exportErr != nil {
		exported = nil
	}
	values, sources, sourceErr := mergeStoredCVSources(storedDatabaseCVs(exported), backup.values[before.ObjectID], exportErr == nil, backup.status)
	if sourceErr == nil {
		baseline := storedCVReference(storedDatabaseCVs(exported), backup.values[before.ObjectID])
		values, sourceErr = inspectUnverifiedCVCache(values, baseline, backup.unverified[before.ObjectID], backup.storedStatus, sources)
	}
	if sourceErr != nil {
		values = nil
		warning := "Gespeicherte CV-Werte konnten nicht vollständig geprüft werden. / Stored CV values could not be fully validated."
		sources.Warning = strings.TrimSpace(sources.Warning + " " + warning)
	}
	sources.BackupAdded = countStoredBackupAdditions(values, storedDatabaseCVs(exported))
	functions, functionErr := s.ReadFunctions(ctx, target, before.ObjectID)
	if err := ctx.Err(); err != nil {
		return ECoSCVDatabase{}, err
	}
	after, err := s.ReadLocomotive(ctx, target, before.ObjectID)
	if canceled := ctx.Err(); canceled != nil {
		return ECoSCVDatabase{}, canceled
	}
	if err != nil || !sameStoredECoSIdentity(before, after, true) {
		return ECoSCVDatabase{}, ErrDigitalCenterReadNotFresh
	}
	result := ECoSCVDatabase{Locomotive: before, Values: values, Sources: sources}
	if functionErr == nil {
		result.Functions = &ECoSFunctionSnapshot{ObjectID: before.ObjectID, Host: target.Host,
			ReadAt: time.Now().UTC().Format(time.RFC3339), Functions: functions}
	} else {
		warning := "Gespeicherte Funktionstasten konnten nicht vollständig gelesen werden. / Stored function keys could not be fully read."
		sources.Warning = strings.TrimSpace(sources.Warning + " " + warning)
	}
	if sourceErr != nil || (exportErr != nil && len(backup.values[before.ObjectID]) == 0) {
		// Preserve raw evidence and function keys in a partial roster entry, but
		// never make unverified-only bytes a completed/importable CV snapshot.
		result.Values = nil
		return result, ErrDigitalCenterDeviceOutput
	}
	return result, nil
}

func sameStoredECoSIdentity(left, right ECoSLocomotive, profile bool) bool {
	return left.ObjectID == right.ObjectID && left.Name == right.Name && left.Address == right.Address &&
		left.Protocol == right.Protocol && (!profile || left.Profile == right.Profile)
}

func sameStoredECoSRoster(left, right []ECoSLocomotive) bool {
	if len(left) != len(right) {
		return false
	}
	byID := make(map[int]ECoSLocomotive, len(left))
	for _, item := range left {
		if _, duplicate := byID[item.ObjectID]; duplicate {
			return false
		}
		byID[item.ObjectID] = item
	}
	for _, item := range right {
		expected, found := byID[item.ObjectID]
		if !found || !sameStoredECoSIdentity(expected, item, false) {
			return false
		}
		delete(byID, item.ObjectID)
	}
	return len(byID) == 0
}
