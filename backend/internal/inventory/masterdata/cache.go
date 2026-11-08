package masterdata

func CloneMasterDataMap(input map[string][]MasterDataEntry) map[string][]MasterDataEntry {
	out := make(map[string][]MasterDataEntry, len(input))
	for key, entries := range input {
		cloned := append([]MasterDataEntry(nil), entries...)
		for index := range cloned {
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
		true:  active,
	}
}
