package application

import (
	"railkeeper/backend/internal/decoder"
)

func validCVExportTimestamp(value string) bool { return decoder.ValidCVExportTimestamp(value) }

func validCVNumber(value int) bool { return decoder.ValidCVNumber(value) }

func validCVValue(value int) bool { return decoder.ValidCVValue(value) }
