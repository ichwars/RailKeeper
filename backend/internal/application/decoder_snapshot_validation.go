package application

import (
	"railkeeper/backend/internal/decoder"
)

func strictDecoderSnapshotJSON(raw string, target any) error {
	return decoder.StrictSnapshotJSON(raw, target)
}

func checkDecoderSnapshotPayload(raw, vehicle, profile string, sources map[string]bool) error {
	return decoder.CheckSnapshotPayload(raw, vehicle, profile, sources)
}
