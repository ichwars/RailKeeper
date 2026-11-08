package masterdata

func CloneMasterDataMap(input map[string][]MasterDataEntry) map[string][]MasterDataEntry {
	out := make(map[string][]MasterDataEntry, len(input))
	for key, entries := range input {
		cloned := append([]MasterDataEntry(nil), entries...)
		for index := range cloned {
			if cloned[index].Metadata != nil {
				cloned[index].Metadata = cloneMetadataMap(cloned[index].Metadata)
			}
			if cloned[index].Capabilities != nil {
				capabilities := *cloned[index].Capabilities
				cloned[index].Capabilities = &capabilities
			}
		}
		out[key] = cloned
	}
	return out
}

func MasterDataCaches(snapshot map[string][]MasterDataEntry) map[bool]map[string][]MasterDataEntry {
	active := make(map[string][]MasterDataEntry, len(snapshot))
	for typeName, entries := range snapshot {
		for _, entry := range entries {
			if entry.Active {
				active[typeName] = append(active[typeName], entry)
			}
		}
	}
	return map[bool]map[string][]MasterDataEntry{
		false: CloneMasterDataMap(snapshot),
		true:  CloneMasterDataMap(active),
	}
}

// Stored metadata has JSON shapes. Copy mutable containers while preserving exact
// scalar types and the distinction between omitted and explicitly empty arrays.
func cloneMetadata(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		return cloneMetadataMap(typed)
	case []any:
		if typed == nil {
			return typed
		}
		out := make([]any, len(typed))
		for index, item := range typed {
			out[index] = cloneMetadata(item)
		}
		return out
	case []string:
		if typed == nil {
			return typed
		}
		out := make([]string, len(typed))
		copy(out, typed)
		return out
	default:
		return value
	}
}

func cloneMetadataMap(input map[string]any) map[string]any {
	if input == nil {
		return nil
	}
	out := make(map[string]any, len(input))
	for key, item := range input {
		out[key] = cloneMetadata(item)
	}
	return out
}
