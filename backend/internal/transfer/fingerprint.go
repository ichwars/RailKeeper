package transfer

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	"railkeeper/backend/internal/domain"
)

const DataTransferTargetFingerprintVersion = 2

func DataTransferTargetFingerprint(value any) string {
	return DataTransferTargetFingerprintForVersion(value, DataTransferTargetFingerprintVersion)
}

func DataTransferTargetFingerprintForVersion(value any, version int) string {
	fingerprintValue := value
	switch accessory := value.(type) {
	case TransferAccessory:
		// Empty contents predate component lists and must retain existing preview hashes.
		fingerprintValue = struct {
			TransferAccessory
			SetContents      []domain.AccessorySetItem         `json:"setContents,omitempty"`
			FingerprintState TransferAccessoryFingerprintState `json:"fingerprintState"`
		}{accessory, accessory.SetContents, accessory.FingerprintState}
	case *TransferAccessory:
		if accessory != nil {
			return DataTransferTargetFingerprintForVersion(*accessory, version)
		}
	}
	if version < DataTransferTargetFingerprintVersion {
		switch vehicle := value.(type) {
		case TransferVehicle:
			fingerprintValue = LegacyTransferVehicle(vehicle)
		case *TransferVehicle:
			if vehicle != nil {
				fingerprintValue = LegacyTransferVehicle(*vehicle)
			}
		}
	}
	payload, err := json.Marshal(fingerprintValue)
	if err != nil {
		return ""
	}
	digest := sha256.Sum256(payload)
	return hex.EncodeToString(digest[:])
}
