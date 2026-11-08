package masterdata

import (
	"fmt"

	"railkeeper/backend/internal/decoderlibrary"
)

func ValidateDecoderProfile(typeName, key string, metadata map[string]any) error {
	if typeName == decoderlibrary.ReferenceMasterDataType {
		reference, err := decoderlibrary.ParseReference(metadata)
		if err != nil || reference.ID != key {
			return fmt.Errorf("%w: invalid decoder CV reference", ErrMasterDataValidation)
		}
		return nil
	}
	if typeName == decoderlibrary.InventoryMasterDataType {
		if _, err := decoderlibrary.ParseInventory(metadata); err != nil || key != decoderlibrary.InventoryKey {
			return fmt.Errorf("%w: invalid decoder inventory", ErrMasterDataValidation)
		}
		return nil
	}
	if typeName != decoderlibrary.MasterDataType {
		return nil
	}
	profile, err := decoderlibrary.Parse(metadata)
	if err != nil || profile.ID != key {
		return fmt.Errorf("%w: invalid decoder profile %s", ErrMasterDataValidation, key)
	}
	return nil
}
