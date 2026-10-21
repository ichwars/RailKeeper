package application

import (
	"context"
	"fmt"

	"railkeeper/backend/internal/decoderlibrary"
)

func (s *MasterDataService) DecoderLibrary(ctx context.Context, activeOnly bool) (decoderlibrary.Catalog, error) {
	entries, err := s.List(ctx, decoderlibrary.MasterDataType, activeOnly)
	catalog := decoderlibrary.Catalog{SchemaVersion: 1, Profiles: []decoderlibrary.Profile{}}
	if err != nil {
		return catalog, err
	}
	for _, entry := range entries {
		profile, err := decoderlibrary.Parse(entry.Metadata)
		if err != nil || profile.ID != entry.Key {
			return catalog, fmt.Errorf("invalid decoder profile %s", entry.Key)
		}
		profile.Active = entry.Active
		catalog.Profiles = append(catalog.Profiles, profile)
	}
	inventories, err := s.List(ctx, decoderlibrary.InventoryMasterDataType, true)
	if err != nil {
		return catalog, err
	}
	for _, entry := range inventories {
		if entry.Key != decoderlibrary.InventoryKey {
			continue
		}
		inventory, parseErr := decoderlibrary.ParseInventory(entry.Metadata)
		if parseErr != nil {
			return catalog, fmt.Errorf("invalid decoder inventory: %w", parseErr)
		}
		catalog.Inventory = &inventory
	}
	references, err := s.DecoderReferences(ctx, "", activeOnly)
	if err != nil {
		return catalog, err
	}
	for _, reference := range references {
		catalog.ReferenceCoverage = append(catalog.ReferenceCoverage, reference.Coverage())
	}
	return catalog, nil
}

func (s *MasterDataService) DecoderReferences(ctx context.Context, modelID string, activeOnly bool) ([]decoderlibrary.CVReference, error) {
	entries, err := s.List(ctx, decoderlibrary.ReferenceMasterDataType, activeOnly)
	result := []decoderlibrary.CVReference{}
	if err != nil {
		return result, err
	}
	for _, entry := range entries {
		reference, err := decoderlibrary.ParseReference(entry.Metadata)
		if err != nil || reference.ID != entry.Key {
			return nil, fmt.Errorf("invalid decoder CV reference %s", entry.Key)
		}
		reference.Active = entry.Active
		if modelID == "" {
			result = append(result, reference)
			continue
		}
		for _, id := range reference.ModelIDs {
			if id == modelID {
				result = append(result, reference)
				break
			}
		}
	}
	return result, nil
}

func validateDecoderProfile(typeName, key string, metadata map[string]any) error {
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
