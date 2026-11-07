package api

import (
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"railkeeper/backend/internal/application"
)

// These records feed generated transport types. Compare their actual JSON fields, including
// optional fields and allOf parents, so contract cleanup cannot silently drop existing data.
func TestOpenAPITransportFieldsMatchRuntimeJSON(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "..", "openapi", "railkeeper.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	for name, record := range map[string]any{
		"Vehicle":                        application.Vehicle{},
		"VehicleCatalogAssignment":       application.VehicleCatalogAssignment{},
		"VehicleCatalogBackfillReport":   application.VehicleCatalogBackfillReport{},
		"VehicleCatalogContext":          application.VehicleCatalogContext{},
		"VehicleCatalogRevision":         application.VehicleCatalogRevision{},
		"VehicleCatalogRevisionInput":    application.VehicleCatalogRevisionInput{},
		"VehicleCatalogAdoptionInput":    application.VehicleCatalogAdoptionInput{},
		"VehicleImage":                   application.VehicleImage{},
		"VehicleImageCreation":           application.VehicleImageCreation{},
		"VehicleImageInput":              application.VehicleImageInput{},
		"VehicleImageImportInput":        importVehicleImageInput{},
		"VehicleAttachmentImportInput":   importVehicleAttachmentInput{},
		"VehicleAttachment":              application.VehicleAttachment{},
		"VehicleAttachmentUpdateRequest": application.VehicleAttachmentUpdateInput{},
		"VehicleMaintenance":             application.VehicleMaintenance{},
		"VehicleMaintenanceInput":        application.VehicleMaintenanceInput{},
		"VehicleSparePart":               application.VehicleSparePart{},
		"VehicleSparePartInput":          application.VehicleSparePartInput{},
		"VehicleExternalMapping":         application.VehicleExternalMap{},
		"VehicleExternalMappingInput":    application.VehicleExternalMapInput{},
		"VehicleFunction":                application.VehicleFunction{},
		"VehicleFunctionInput":           application.VehicleFunctionInput{},
		"VehicleFunctionChange":          application.VehicleFunctionChange{},
		"VehicleCVValue":                 application.VehicleCVValue{},
		"VehicleCVValueInput":            application.VehicleCVValueInput{},
		"VehicleCVChange":                application.VehicleCVChange{},
		"VehicleCVChangesInput":          application.VehicleCVChangesInput{},
		"VehicleCVValueHistory":          application.VehicleCVValueHistory{},
		"VehicleCVFile":                  application.VehicleCVFile{},
		"CreateVehicleRequest":           application.CreateVehicleInput{},
		"CreateMappedVehicleRequest":     createVehicleRequest{},
		"CreateVehicleSetRequest":        application.CreateVehicleSetInput{},
		"DataTransferProfile":            application.DataTransferProfile{},
		"DataTransferJob":                application.DataTransferJob{},
		"DataTransferIssue":              application.DataTransferIssue{},
		"DataTransferArtifact":           application.DataTransferArtifact{},
		"DataTransferVehicleSet":         application.TransferVehicleSet{},
		"DataTransferCSVColumnMapping":   application.DataTransferCSVColumnMapping{},
		"DataTransferCSVMappingInput":    application.DataTransferCSVMappingInput{},
		"VehicleTransferField":           application.VehicleTransferField{},
	} {
		t.Run(name, func(t *testing.T) {
			properties, required := transportSchemaFields(t, string(data), name)
			fields := transportJSONFields(reflect.TypeOf(record))
			for field, optional := range fields {
				if !properties[field] {
					t.Errorf("runtime JSON field %s is missing from %s", field, name)
				}
				// Input JSON tags describe decoding, not validation requirements.
				input := strings.HasSuffix(name, "Input") || strings.HasSuffix(name, "Request") || name == "VehicleFunctionChange"
				if !input && !optional && !required[field] {
					t.Errorf("always emitted field %s is optional in %s", field, name)
				}
			}
			for field := range properties {
				if _, ok := fields[field]; !ok {
					t.Errorf("%s documents JSON field %s which the runtime does not emit", name, field)
				}
			}
		})
	}
}

func transportJSONFields(record reflect.Type) map[string]bool {
	fields := map[string]bool{}
	for index := range record.NumField() {
		field := record.Field(index)
		if field.Anonymous {
			for name, optional := range transportJSONFields(field.Type) {
				fields[name] = optional
			}
			continue
		}
		tag := strings.Split(field.Tag.Get("json"), ",")
		if tag[0] != "" && tag[0] != "-" {
			fields[tag[0]] = len(tag) > 1 && tag[1] == "omitempty"
		}
	}
	return fields
}

func transportSchemaFields(t *testing.T, contract, name string) (map[string]bool, map[string]bool) {
	t.Helper()
	block := openAPIIndentedBlock(t, contract, name, 4)
	properties, required := map[string]bool{}, map[string]bool{}
	// Only allOf parent references extend this record. Property/item references do not.
	parentPattern := regexp.MustCompile(`(?m)^ {8}- \$ref: ["']?#/components/schemas/([A-Za-z0-9_]+)`)
	for _, parent := range parentPattern.FindAllStringSubmatch(block, -1) {
		parentProperties, parentRequired := transportSchemaFields(t, contract, parent[1])
		for field := range parentProperties {
			properties[field] = true
		}
		for field := range parentRequired {
			required[field] = true
		}
	}
	lines := strings.Split(block, "\n")
	for index, line := range lines {
		indent := len(line) - len(strings.TrimLeft(line, " "))
		if strings.TrimSpace(line) != "properties:" || (indent != 6 && indent != 10) {
			continue
		}
		for _, property := range lines[index+1:] {
			if strings.TrimSpace(property) == "" {
				continue
			}
			propertyIndent := len(property) - len(strings.TrimLeft(property, " "))
			if propertyIndent <= indent {
				break
			}
			if propertyIndent == indent+2 {
				field, _, found := strings.Cut(strings.TrimSpace(property), ":")
				if found {
					properties[field] = true
				}
			}
		}
	}
	requiredPattern := regexp.MustCompile(`(?m)^ {6}(?: {4})?required: \[([^\]]*)\]`)
	for _, group := range requiredPattern.FindAllStringSubmatch(block, -1) {
		for _, field := range strings.Split(group[1], ",") {
			required[strings.TrimSpace(field)] = true
		}
	}
	return properties, required
}
