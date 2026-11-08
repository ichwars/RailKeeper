package catalog

import (
	"context"
	"sort"
	"time"
)

func BuildReport(rows []RawAssignment) (*VehicleCatalogBackfillReport, error) {
	report := &VehicleCatalogBackfillReport{Assignments: []VehicleCatalogAssignment{}}
	for _, row := range rows {
		assignment := row.Assignment
		legacyJSON, fieldsJSON, overridesJSON := row.LegacyJSON, row.FieldsJSON, row.OverridesJSON
		legacy, err := DecodeFields(legacyJSON, true)
		if err != nil {
			return nil, err
		}
		assignment.Name = legacy["name"]
		assignment.OverrideFields, assignment.DifferentFields = []string{}, []string{}
		if assignment.RevisionID != "" {
			fields, err := DecodeFields(fieldsJSON, true)
			if err != nil {
				return nil, err
			}
			overrides, err := DecodeFields(overridesJSON, false)
			if err != nil {
				return nil, err
			}
			for field, value := range overrides {
				assignment.OverrideFields = append(assignment.OverrideFields, field)
				fields[field] = value
			}
			for field, value := range legacy {
				if fields[field] != value {
					assignment.DifferentFields = append(assignment.DifferentFields, field)
				}
			}
		}
		sort.Strings(assignment.OverrideFields)
		sort.Strings(assignment.DifferentFields)
		if assignment.TargetType == "vehicle" {
			report.Vehicles++
			if assignment.RevisionID != "" {
				report.LinkedVehicles++
			}
		} else {
			report.Sets++
			if assignment.RevisionID != "" {
				report.LinkedSets++
			}
		}
		if len(assignment.DifferentFields) > 0 {
			report.DifferentRepresentations++
		}
		report.Assignments = append(report.Assignments, assignment)
	}
	return report, nil
}
func (s *Service) BackfillLegacyCatalog(ctx context.Context, actor string) (*VehicleCatalogBackfillReport, error) {
	tx, err := s.repository.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	before, err := tx.LinkCount(ctx)
	if err != nil {
		return nil, err
	}
	if err := tx.Backfill(ctx); err != nil {
		return nil, err
	}
	rows, err := tx.Report(ctx)
	if err != nil {
		return nil, err
	}
	report, err := BuildReport(rows)
	if err != nil {
		return nil, err
	}
	if added := report.LinkedVehicles + report.LinkedSets - before; added > 0 {
		if err := tx.AuditBackfill(ctx, added, actor, time.Now().UTC().Format(time.RFC3339)); err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return report, nil
}
