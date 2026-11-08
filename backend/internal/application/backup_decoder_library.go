package application

import (
	"encoding/json"
	"fmt"

	"railkeeper/backend/internal/decoderlibrary"
)

func validateBackupDecoderLibrary(rows []map[string]any) []string {
	for i, row := range rows {
		typeName, _ := row["type"].(string)
		if typeName != decoderlibrary.MasterDataType && typeName != decoderlibrary.InventoryMasterDataType && typeName != decoderlibrary.ReferenceMasterDataType {
			continue
		}
		metadataJSON, valid := row["metadata_json"].(string)
		key, keyValid := row["key"].(string)
		var metadata map[string]any
		if !valid || !keyValid || len(metadataJSON) > 512*1024 ||
			json.Unmarshal([]byte(metadataJSON), &metadata) != nil ||
			validateDecoderProfile(typeName, key, metadata) != nil {
			return []string{fmt.Sprintf("Decoderprofil in Stammdatenzeile %d ist ungültig. / Invalid decoder profile in master data row %d.", i+1, i+1)}
		}
	}
	return nil
}
