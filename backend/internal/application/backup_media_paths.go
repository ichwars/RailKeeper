package application

import (
	"fmt"
	"strings"
)

// Historical Windows paths retain their stored text; only the read path is normalized.
func portableBackupMediaPath(value string) (string, error) {
	value = strings.ReplaceAll(value, `\`, "/")
	if err := validateBackupFilePath(value); err != nil {
		return "", err
	}
	return value, nil
}

func validateBackupMediaPaths(table string, rows []map[string]any) []string {
	var paths map[string]string
	switch table {
	case "vehicle_images":
		paths = map[string]string{"storage_path": "blob_id", "thumbnail_path": "thumbnail_blob_id"}
	case "vehicle_attachments", "vehicle_cv_files":
		paths = map[string]string{"storage_path": "blob_id"}
	default:
		return nil
	}
	for index, row := range rows {
		for field, blobField := range paths {
			if blob, ok := row[blobField].(string); ok && blob != "" {
				continue
			}
			if row[field] == nil || row[field] == "" {
				continue
			}
			value, ok := row[field].(string)
			if _, err := portableBackupMediaPath(value); !ok || err != nil {
				return []string{fmt.Sprintf("%s Zeile %d: ungültiger Medienpfad %s. / Invalid upload-media path.", table, index+1, field)}
			}
		}
	}
	return nil
}
