package decoderlibrary

import "slices"

type IdentityEvidence struct {
	Scope    string `json:"scope"`
	Name     string `json:"name"`
	CVNumber int    `json:"cvNumber"`
	Value    int    `json:"value"`
	ModelID  string `json:"modelId,omitempty"`
}

type identityMatch struct {
	evidence  IdentityEvidence
	profileID string
	profiles  []string
}

func (m *identityMatch) supportsProfile(id string) bool { return slices.Contains(m.profiles, id) }

func identify(inventory *Inventory, label string, rows []Reading, manufacturer *int) (*identityMatch, bool) {
	if inventory == nil || manufacturer == nil {
		return nil, false
	}
	base := map[int]int{}
	conflicts := map[int]bool{}
	for _, row := range rows {
		if row.IndexCV31 != nil || row.IndexCV32 != nil {
			continue
		}
		if previous, ok := base[row.CVNumber]; ok && previous != row.Value {
			conflicts[row.CVNumber] = true
		}
		base[row.CVNumber] = row.Value
	}
	var candidates []identityMatch
	for _, model := range inventory.Models {
		var maker *Manufacturer
		for _, family := range inventory.Families {
			if family.ID != model.FamilyID {
				continue
			}
			for i := range inventory.Manufacturers {
				if inventory.Manufacturers[i].ID == family.ManufacturerID {
					maker = &inventory.Manufacturers[i]
				}
			}
		}
		if maker == nil {
			continue
		}
		named := normalized(label) == normalized(model.Name) || normalized(label) == normalized(maker.Name+" "+model.Name)
		if named && !model.BrandOnly && maker.CV8 != *manufacturer {
			return nil, true
		}
		rule := model.Identification
		if rule == nil || maker.CV8 != *manufacturer {
			continue
		}
		value, present := base[rule.CV]
		if conflicts[rule.CV] {
			return nil, true
		}
		if !present {
			continue
		}
		if !slices.Contains(rule.Values, value) {
			if named {
				return nil, true
			}
			continue
		}
		evidence := IdentityEvidence{Scope: rule.Scope, Name: rule.Name, CVNumber: rule.CV, Value: value}
		if rule.Scope == "model" {
			evidence.ModelID = model.ID
		}
		candidates = append(candidates, identityMatch{evidence: evidence, profileID: model.ProfileID})
	}
	if len(candidates) == 0 {
		return nil, false
	}
	first := candidates[0]
	for _, candidate := range candidates {
		first.profiles = append(first.profiles, candidate.profileID)
	}
	for _, candidate := range candidates[1:] {
		if first.evidence.Scope != "family" || candidate.evidence.Scope != "family" ||
			first.evidence.Name != candidate.evidence.Name || first.evidence.CVNumber != candidate.evidence.CVNumber ||
			first.evidence.Value != candidate.evidence.Value {
			return nil, true
		}
	}
	return &first, false
}
