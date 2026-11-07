package transfer

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

const DataTransferTargetFingerprintVersion = 2

func DataTransferTargetFingerprint(value any) string {
	return DataTransferTargetFingerprintForVersion(value, DataTransferTargetFingerprintVersion)
}

func DataTransferTargetFingerprintForVersion(value any, version int) string {
	fingerprintValue := value
	switch accessory := value.(type) {
	case TransferAccessory:
		fingerprintValue = struct {
			TransferAccessory
			FingerprintState TransferAccessoryFingerprintState `json:"fingerprintState"`
		}{TransferAccessory: accessory, FingerprintState: accessory.FingerprintState}
	case *TransferAccessory:
		if accessory != nil {
			fingerprintValue = struct {
				TransferAccessory
				FingerprintState TransferAccessoryFingerprintState `json:"fingerprintState"`
			}{TransferAccessory: *accessory, FingerprintState: accessory.FingerprintState}
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
