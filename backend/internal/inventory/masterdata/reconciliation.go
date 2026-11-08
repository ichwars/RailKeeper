package masterdata

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"railkeeper/backend/internal/credentials"
)

type Identity struct {
	typeName string
	key      string
}
type RelationIdentity struct {
	parent Identity
	child  Identity
}

func BuildDesiredEntries(
	ctx context.Context,
	usage Usage,
	current map[Identity]MasterDataEntry,
	entriesByType map[string][]MasterDataEntry,
	now string,
	articleTypesWereImported bool,
	accessorySubtypesWereImported bool,
) (map[Identity]MasterDataEntry, int, error) {
	desired := map[Identity]MasterDataEntry{}
	importedCount := 0
	for bucketType, entries := range entriesByType {
		for _, entry := range entries {
			normalized, identity, err := normalizeMasterDataImportEntry(bucketType, entry, current, now)
			if err != nil {
				return nil, 0, err
			}
			if _, duplicate := desired[identity]; duplicate {
				return nil, 0, fmt.Errorf("%w: duplicate master data %s/%s",
					ErrMasterDataValidation, identity.typeName, identity.key)
			}
			desired[identity] = normalized
			preservedArticleType := identity.typeName == StandardArticleType && !articleTypesWereImported
			preservedSubtype := identity.typeName == StandardAccessorySubtype && !accessorySubtypesWereImported
			if !preservedArticleType && !preservedSubtype {
				importedCount++
			}
		}
	}
	for identity, entry := range current {
		if _, imported := desired[identity]; imported {
			continue
		}
		if entry.Origin == MasterDataOriginBundled {
			desired[identity] = entry
			continue
		}
		used, err := usage(ctx, entry)
		if err != nil {
			return nil, 0, fmt.Errorf("check omitted master data %s/%s: %w",
				identity.typeName, identity.key, err)
		}
		if used {
			return nil, 0, fmt.Errorf("%w: omitted referenced master data %s/%s",
				ErrMasterDataInUse, identity.typeName, identity.key)
		}
	}
	return desired, importedCount, nil
}
func normalizeMasterDataImportEntry(
	bucketType string,
	entry MasterDataEntry,
	current map[Identity]MasterDataEntry,
	now string,
) (MasterDataEntry, Identity, error) {
	entry.Type = EffectiveMasterDataType(strings.TrimSpace(bucketType), entry)
	entry.Key = strings.TrimSpace(entry.Key)
	entry.Label = strings.TrimSpace(entry.Label)
	if entry.Type == "" || entry.Label == "" {
		return entry, Identity{}, ErrMasterDataValidation
	}
	if entry.Key == "" {
		entry.Key = SlugKey(entry.Label)
	}
	identity := Identity{entry.Type, entry.Key}
	entry.ID = strings.TrimSpace(entry.ID)
	if entry.ID == "" {
		entry.ID = entry.Type + ":" + entry.Key
	}
	entry.SourceURL = strings.TrimSpace(entry.SourceURL)
	if entry.Metadata == nil {
		entry.Metadata = map[string]any{}
	}
	entry.CreatedAt = strings.TrimSpace(entry.CreatedAt)
	if entry.CreatedAt == "" {
		entry.CreatedAt = now
	}
	entry.UpdatedAt = strings.TrimSpace(entry.UpdatedAt)
	if entry.UpdatedAt == "" {
		entry.UpdatedAt = now
	}
	entry.Capabilities = nil
	if currentEntry, exists := current[identity]; exists && currentEntry.Origin == MasterDataOriginBundled {
		entry.Origin = MasterDataOriginBundled
	} else {
		entry.Origin = MasterDataOriginCustom
	}
	return entry, identity, nil
}
func EntriesByType(
	entries map[Identity]MasterDataEntry,
) map[string][]MasterDataEntry {
	byType := map[string][]MasterDataEntry{}
	for _, entry := range entries {
		byType[entry.Type] = append(byType[entry.Type], entry)
	}
	return byType
}
func BuildDesiredRelations(
	imported []MasterDataRelation,
	currentRelations []MasterDataRelation,
	currentEntries map[Identity]MasterDataEntry,
	desiredEntries map[Identity]MasterDataEntry,
) ([]MasterDataRelation, int, error) {
	relations := map[RelationIdentity]MasterDataRelation{}
	for _, relation := range imported {
		normalized, identity, err := normalizeMasterDataRelation(relation)
		if err != nil {
			return nil, 0, err
		}
		if _, duplicate := relations[identity]; duplicate {
			return nil, 0, fmt.Errorf("%w: duplicate master data relation", ErrMasterDataValidation)
		}
		relations[identity] = normalized
	}
	for _, relation := range currentRelations {
		identity := masterDataRelationKey(relation)
		if _, imported := relations[identity]; imported {
			continue
		}
		parent, parentExists := currentEntries[identity.parent]
		child, childExists := currentEntries[identity.child]
		_, parentRetained := desiredEntries[identity.parent]
		_, childRetained := desiredEntries[identity.child]
		if parentExists && childExists && parentRetained && childRetained &&
			parent.Origin == MasterDataOriginBundled && child.Origin == MasterDataOriginBundled {
			relations[identity] = relation
		}
	}
	ordered := make([]MasterDataRelation, 0, len(relations))
	for identity, relation := range relations {
		if _, exists := desiredEntries[identity.parent]; !exists {
			return nil, 0, fmt.Errorf("%w: missing master data relation parent %s/%s",
				ErrMasterDataValidation, identity.parent.typeName, identity.parent.key)
		}
		if _, exists := desiredEntries[identity.child]; !exists {
			return nil, 0, fmt.Errorf("%w: missing master data relation child %s/%s",
				ErrMasterDataValidation, identity.child.typeName, identity.child.key)
		}
		ordered = append(ordered, relation)
	}
	sort.Slice(ordered, func(i, j int) bool {
		left := masterDataRelationKey(ordered[i])
		right := masterDataRelationKey(ordered[j])
		if left.parent.typeName != right.parent.typeName {
			return left.parent.typeName < right.parent.typeName
		}
		if left.parent.key != right.parent.key {
			return left.parent.key < right.parent.key
		}
		if left.child.typeName != right.child.typeName {
			return left.child.typeName < right.child.typeName
		}
		return left.child.key < right.child.key
	})
	return ordered, len(imported), nil
}
func normalizeMasterDataRelation(
	relation MasterDataRelation,
) (MasterDataRelation, RelationIdentity, error) {
	relation.ParentType = strings.TrimSpace(relation.ParentType)
	relation.ParentKey = strings.TrimSpace(relation.ParentKey)
	relation.ChildType = strings.TrimSpace(relation.ChildType)
	relation.ChildKey = strings.TrimSpace(relation.ChildKey)
	if relation.ParentType == "" || relation.ParentKey == "" ||
		relation.ChildType == "" || relation.ChildKey == "" {
		return relation, RelationIdentity{}, ErrMasterDataValidation
	}
	relation.ID = strings.TrimSpace(relation.ID)
	if relation.ID == "" {
		relation.ID = credentials.RandomID()
	}
	return relation, masterDataRelationKey(relation), nil
}
func masterDataRelationKey(relation MasterDataRelation) RelationIdentity {
	return RelationIdentity{
		parent: Identity{relation.ParentType, relation.ParentKey},
		child:  Identity{relation.ChildType, relation.ChildKey},
	}
}
func IdentityOf(entry MasterDataEntry) Identity { return Identity{entry.Type, entry.Key} }
