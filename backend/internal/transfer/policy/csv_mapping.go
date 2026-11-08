package policy

import (
	"fmt"
	"strings"
)

func DefaultDataTransferCSVMapping(
	header []string,
	profileDefaults map[string]string,
) []DataTransferCSVColumnMapping {
	aliases := vehicleTransferCSVAliases()
	mapping := make([]DataTransferCSVColumnMapping, len(header))
	usedTargets := map[string]bool{}
	for index, sourceHeader := range header {
		normalized := normalizeTransferCSVHeader(sourceHeader)
		target, hasProfileDefault := profileDefaults[normalized]
		if hasProfileDefault && target != dataTransferCSVIgnoreTarget {
			if _, valid := VehicleCSVTransferFieldByKey(target); !valid {
				hasProfileDefault = false
				target = ""
			}
		}
		origin := CSVMappingProfile
		if !hasProfileDefault {
			target = aliases[normalized]
			origin = CSVMappingAlias
		}
		if hasProfileDefault && target == dataTransferCSVIgnoreTarget {
			target = ""
			origin = CSVMappingIgnored
		}
		if target == "" && origin != CSVMappingIgnored {
			origin = CSVMappingUnmapped
		} else if target != "" && usedTargets[target] {
			target = ""
			origin = CSVMappingUnmapped
		} else if target != "" {
			usedTargets[target] = true
		}
		mapping[index] = DataTransferCSVColumnMapping{
			Index: index, SourceHeader: sourceHeader, NormalizedHeader: normalized,
			TargetField: target, Origin: origin,
		}
	}
	return mapping
}

func dataTransferCSVMappingDefaults(options map[string]any) map[string]string {
	defaults := map[string]string{}
	value, ok := options[dataTransferCSVMappingOption]
	if !ok {
		return defaults
	}
	values, ok := value.(map[string]any)
	if !ok {
		if typed, typedOK := value.(map[string]string); typedOK {
			for source, target := range typed {
				defaults[normalizeTransferCSVHeader(source)] = target
			}
		}
		return defaults
	}
	for source, value := range values {
		target, ok := value.(string)
		if ok {
			defaults[normalizeTransferCSVHeader(source)] = strings.TrimSpace(target)
		}
	}
	return defaults
}

func DataTransferOptionsWithCSVMapping(
	options map[string]any,
	mapping []DataTransferCSVColumnMapping,
) map[string]any {
	updated := CloneTransferOptions(options)
	defaults := make(map[string]any, len(mapping))
	for _, column := range mapping {
		switch {
		case column.TargetField != "":
			defaults[column.NormalizedHeader] = column.TargetField
		case column.Origin == CSVMappingIgnored:
			defaults[column.NormalizedHeader] = dataTransferCSVIgnoreTarget
		}
	}
	updated[dataTransferCSVMappingOption] = defaults
	return updated
}

func vehicleTransferCSVAliases() map[string]string {
	fields := VehicleCSVTransferFields()
	aliases := make(map[string]string, len(fields)*3)
	for _, field := range fields {
		for _, alias := range field.Aliases {
			aliases[normalizeTransferCSVHeader(alias)] = field.Key
		}
	}
	return aliases
}

func ValidateDataTransferCSVMapping(header []string, mapping []DataTransferCSVColumnMapping) error {
	if len(header) == 0 || len(mapping) != len(header) {
		return fmt.Errorf("%w: mapping must contain every CSV column", ErrDataTransferValidation)
	}
	seenIndexes := map[int]bool{}
	seenTargets := map[string]bool{}
	validOrigins := map[DataTransferCSVMappingOrigin]bool{
		CSVMappingAlias: true, CSVMappingProfile: true, CSVMappingManual: true,
		CSVMappingIgnored: true, CSVMappingUnmapped: true,
	}
	for _, column := range mapping {
		if column.Index < 0 || column.Index >= len(header) || seenIndexes[column.Index] {
			return fmt.Errorf("%w: invalid or repeated CSV mapping index", ErrDataTransferValidation)
		}
		seenIndexes[column.Index] = true
		if column.SourceHeader != header[column.Index] ||
			column.NormalizedHeader != normalizeTransferCSVHeader(column.SourceHeader) {
			return fmt.Errorf("%w: CSV mapping header changed", ErrDataTransferValidation)
		}
		if !validOrigins[column.Origin] {
			return fmt.Errorf("%w: invalid CSV mapping origin", ErrDataTransferValidation)
		}
		if column.Origin == CSVMappingIgnored || column.Origin == CSVMappingUnmapped {
			if column.TargetField != "" {
				return fmt.Errorf("%w: ignored or unmapped CSV column must not have a target", ErrDataTransferValidation)
			}
			continue
		}
		if column.TargetField == "" {
			return fmt.Errorf("%w: mapped CSV column must have a target", ErrDataTransferValidation)
		}
		if _, ok := VehicleCSVTransferFieldByKey(column.TargetField); !ok {
			return fmt.Errorf("%w: unsupported CSV target %q", ErrDataTransferValidation, column.TargetField)
		}
		if seenTargets[column.TargetField] {
			return fmt.Errorf("%w: repeated CSV target %q", ErrDataTransferValidation, column.TargetField)
		}
		seenTargets[column.TargetField] = true
	}
	return nil
}

func ValidateRequestedDataTransferCSVMapping(mapping []DataTransferCSVColumnMapping) error {
	for _, column := range mapping {
		if column.Origin == CSVMappingUnmapped || (column.TargetField == "" && column.Origin != CSVMappingIgnored) {
			return fmt.Errorf(
				"%w: CSV column %q must be mapped or explicitly ignored",
				ErrDataTransferValidation,
				column.SourceHeader,
			)
		}
	}
	return nil
}
