package backup

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"math"
	"path/filepath"
	"sort"
	"strings"
)

func BackupNumberValue(value any) (float64, bool) {
	switch typed := value.(type) {
	case json.Number:
		number, err := typed.Float64()
		return number, err == nil && !math.IsNaN(number) && !math.IsInf(number, 0)
	case float64:
		return typed, !math.IsNaN(typed) && !math.IsInf(typed, 0)
	case float32:
		return float64(typed), !math.IsNaN(float64(typed)) && !math.IsInf(float64(typed), 0)
	case int:
		return float64(typed), true
	case int8:
		return float64(typed), true
	case int16:
		return float64(typed), true
	case int32:
		return float64(typed), true
	case int64:
		return float64(typed), true
	case uint:
		return float64(typed), true
	case uint8:
		return float64(typed), true
	case uint16:
		return float64(typed), true
	case uint32:
		return float64(typed), true
	case uint64:
		return float64(typed), true
	default:
		return 0, false
	}
}

func BackupBooleanValue(value any) (bool, bool) {
	if typed, ok := value.(bool); ok {
		return typed, true
	}
	number, valid := BackupNumberValue(value)
	if !valid || (number != 0 && number != 1) {
		return false, false
	}
	return number == 1, true
}

func BackupNonEmptyString(value any) (string, bool) {
	text, ok := value.(string)
	text = strings.TrimSpace(text)
	return text, ok && text != ""
}

func ValidBackupNumber(value any) bool {
	_, valid := BackupNumberValue(value)
	return valid
}

func ValidBackupBoolean(value any) bool {
	_, valid := BackupBooleanValue(value)
	return valid
}

func SortedKeys(values map[string]struct{}) []string {
	out := make([]string, 0, len(values))
	for value := range values {
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}

func FinishBackupValidation(result *BackupValidationResult) *BackupValidationResult {
	result.Compatible = len(result.Errors) == 0
	return result
}

func ValidateBackupFilePath(value string) error {
	value = filepath.ToSlash(strings.TrimSpace(value))
	if value == "" || filepath.IsAbs(value) || strings.Contains(value, "\x00") {
		return ErrBackupPath
	}
	cleaned := PathClean(value)
	if cleaned != value || strings.HasPrefix(cleaned, "../") || cleaned == ".." {
		return ErrBackupPath
	}
	if !strings.HasPrefix(cleaned, "uploads/") {
		return ErrBackupPath
	}
	return nil
}

func ValidateBackupFiles(files []BackupFile) error {
	for _, file := range files {
		if err := ValidateBackupFilePath(file.Path); err != nil {
			return err
		}
		data, err := base64.StdEncoding.DecodeString(file.ContentBase64)
		if err != nil {
			return ErrBackupInvalid
		}
		if file.SizeBytes > 0 && int64(len(data)) != file.SizeBytes {
			return ErrBackupInvalid
		}
		sum := sha256.Sum256(data)
		if file.SHA256 != "" && !strings.EqualFold(file.SHA256, hex.EncodeToString(sum[:])) {
			return ErrBackupInvalid
		}
	}
	return nil
}

func PathClean(value string) string {
	cleaned := filepath.ToSlash(filepath.Clean(filepath.FromSlash(value)))
	if cleaned == "." {
		return ""
	}
	return cleaned
}
