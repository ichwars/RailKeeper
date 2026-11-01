package decoderlibrary

type ResolutionContext struct {
	ModelLabel  string
	ECoSProfile string
}

type ECoSProfileMatch struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

func selectECoSProfile(profiles []Profile, label, modelLabel, source string) (*Profile, *ECoSProfileMatch, bool) {
	if normalized(label) != "ecos" || normalized(source) == "" {
		return nil, nil, false
	}
	evidence := &ECoSProfileMatch{Name: source, Status: "unknown"}
	var candidates []*Profile
	for i := range profiles {
		p := &profiles[i]
		if p.Active && p.Kind == "model" &&
			(normalized(p.Name) == normalized(source) || normalized(p.NameEN) == normalized(source) || includes(p.Aliases, source)) {
			candidates = append(candidates, p)
		}
	}
	if len(candidates) == 0 {
		return nil, evidence, false
	}
	profile := candidates[0]
	if profile.ManufacturerID == nil {
		return nil, evidence, false
	}
	manualCompatible := normalized(modelLabel) == normalized(label) ||
		normalized(profile.Name) == normalized(modelLabel) || normalized(profile.NameEN) == normalized(modelLabel) ||
		includes(profile.Aliases, modelLabel) || includes(profile.FamilyAliases, modelLabel)
	if len(candidates) != 1 || !manualCompatible {
		evidence.Status = "conflict"
		return nil, evidence, true
	}
	evidence.Status = "matched"
	return profile, evidence, false
}

func ResolveWithContext(profiles []Profile, label string, values []Reading, inventory *Inventory, context ResolutionContext) Match {
	if context.ModelLabel == "" {
		context.ModelLabel = label
	}
	return resolve(profiles, label, values, inventory, context)
}
