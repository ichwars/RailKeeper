package decoderlibrary

import (
	"regexp"
	"strconv"
	"strings"
)

type Reading struct {
	CVNumber       int    `json:"cvNumber"`
	Value          int    `json:"value"`
	IndexCV31      *int   `json:"indexCV31,omitempty"`
	IndexCV32      *int   `json:"indexCV32,omitempty"`
	Protocol       string `json:"protocol"`
	DecoderProfile string `json:"decoderProfile"`
	Description    string `json:"description"`
}

type Match struct {
	Status             string            `json:"status"`
	ProfileID          string            `json:"profileId,omitempty"`
	ProfileName        string            `json:"profileName,omitempty"`
	Revision           string            `json:"revision,omitempty"`
	ManufacturerID     *int              `json:"manufacturerId,omitempty"`
	Firmware           string            `json:"firmware,omitempty"`
	FirmwareCompatible *bool             `json:"firmwareCompatible,omitempty"`
	KnownValues        int               `json:"knownValues"`
	TotalValues        int               `json:"totalValues"`
	MatchedBy          string            `json:"matchedBy,omitempty"`
	Identification     *IdentityEvidence `json:"identification,omitempty"`
}

var firmwareDescription = regexp.MustCompile(`^LokProgrammer Decoder: .+ \(([\d.]+)\)$`)
var dccProtocol = regexp.MustCompile(`(?i)^DCC(?: ?(?:14|28|128))?$`)

func normalized(value string) string {
	return strings.ToLower(strings.Join(strings.Fields(value), " "))
}

func includes(values []string, value string) bool {
	for _, candidate := range values {
		if normalized(candidate) == normalized(value) {
			return true
		}
	}
	return false
}

func Resolve(profiles []Profile, label string, values []Reading, inventories ...*Inventory) Match {
	match := Match{Status: "standard"}
	manufacturerConflict := false
	var rows []Reading
	for _, value := range values {
		if strings.TrimSpace(value.DecoderProfile) != strings.TrimSpace(label) {
			continue
		}
		rows = append(rows, value)
		if protocol := strings.TrimSpace(value.Protocol); protocol != "" && !dccProtocol.MatchString(protocol) {
			match.Status = "unsupported"
		}
		if value.CVNumber == 8 && value.IndexCV31 == nil && value.IndexCV32 == nil {
			id := value.Value
			if match.ManufacturerID != nil && *match.ManufacturerID != id {
				manufacturerConflict = true
			}
			match.ManufacturerID = &id
		}
		if value.CVNumber == 7 && value.IndexCV31 == nil && value.IndexCV32 == nil {
			if parts := firmwareDescription.FindStringSubmatch(value.Description); len(parts) == 2 {
				match.Firmware = parts[1]
			}
		}
	}
	match.TotalValues = len(rows)
	if match.Status == "unsupported" {
		return match
	}
	var inventory *Inventory
	if len(inventories) > 0 {
		inventory = inventories[0]
	}
	identity, identityConflict := identify(inventory, label, rows, match.ManufacturerID)
	var selected *Profile
	var exact, family []*Profile
	for i := range profiles {
		p := &profiles[i]
		if !p.Active {
			continue
		}
		if p.Kind == "standard" && p.ID == "dcc-basics" {
			selected = p
		}
		if p.Kind != "model" {
			continue
		}
		if includes(p.Aliases, label) {
			exact = append(exact, p)
		}
		if includes(p.FamilyAliases, label) {
			family = append(family, p)
		}
	}
	candidates := exact
	if len(candidates) == 0 {
		candidates = family
	}
	if manufacturerConflict || identityConflict || len(candidates) > 1 {
		match.Status = "conflict"
	} else if len(candidates) == 1 {
		p := candidates[0]
		switch {
		case match.ManufacturerID == nil:
			match.Status = "suggested"
		case p.ManufacturerID == nil || *p.ManufacturerID != *match.ManufacturerID:
			match.Status = "conflict"
		default:
			if identity != nil && !identity.supportsProfile(p.ID) {
				match.Status = "conflict"
				break
			}
			selected = p
			match.MatchedBy = "name"
			match.Status = "family"
			if len(exact) > 0 {
				match.Status = "matched"
			}
		}
	}
	if identity != nil && match.Status != "conflict" {
		match.Identification = &identity.evidence
		if match.MatchedBy == "" {
			match.MatchedBy = "cv"
			match.Status = "identified"
			if identity.evidence.Scope == "family" {
				match.Status = "family"
			}
			if identity.evidence.Scope == "model" {
				for i := range profiles {
					p := &profiles[i]
					if p.Active && p.Kind == "model" && p.ID == identity.profileID && p.ManufacturerID != nil && *p.ManufacturerID == *match.ManufacturerID {
						selected = p
						match.Status = "matched"
					}
				}
			}
		}
	}
	if selected == nil {
		match.Status = "unavailable"
		return match
	}
	if selected.Kind == "standard" && match.Status == "standard" && match.ManufacturerID != nil {
		for _, p := range profiles {
			if p.Active && p.Kind == "model" && p.ManufacturerID != nil && *p.ManufacturerID == *match.ManufacturerID {
				match.Status = "manufacturer"
				break
			}
		}
	}
	match.ProfileID, match.ProfileName, match.Revision = selected.ID, selected.Name, selected.Revision
	if match.Status == "family" {
		match.ProfileName = selected.Family
	}
	if match.MatchedBy == "cv" && match.Identification != nil {
		match.ProfileName = match.Identification.Name
	}
	if selected.Kind == "model" && match.Firmware != "" && selected.FirmwareMin != "" {
		compatible := versionAtLeast(match.Firmware, selected.FirmwareMin)
		match.FirmwareCompatible = &compatible
	}
	for _, value := range rows {
		if definition := FindDefinition(profiles, match, value); definition != nil && !definition.Reserved {
			match.KnownValues++
		}
	}
	return match
}

func versionAtLeast(actual, minimum string) bool {
	a, b := strings.Split(actual, "."), strings.Split(minimum, ".")
	for i := 0; i < max(len(a), len(b)); i++ {
		x, y := 0, 0
		if i < len(a) {
			x, _ = strconv.Atoi(a[i])
		}
		if i < len(b) {
			y, _ = strconv.Atoi(b[i])
		}
		if x != y {
			return x > y
		}
	}
	return true
}

func FindDefinition(profiles []Profile, match Match, value Reading) *Definition {
	if match.Status == "unsupported" || match.Status == "unavailable" {
		return nil
	}
	customCurve := false
	modelDefinitionsOnly := false
	for _, p := range profiles {
		if p.Active && p.ID == match.ProfileID && (match.Status == "matched" || match.Status == "family") &&
			(match.FirmwareCompatible == nil || *match.FirmwareCompatible) && p.Rules != nil {
			customCurve = p.Rules.MotorCurveCV != 0
			modelDefinitionsOnly = p.Rules.ModelDefinitionsOnly || p.Rules.Uhlenbrock76420 || p.Rules.Roco10882 || p.Rules.PIKO56500
		}
	}
	// Model definitions override their DCC foundation. Firmware outside the source scope stays generic.
	for _, model := range []bool{true, false} {
		for _, p := range profiles {
			if !p.Active || (p.Kind == "model") != model || (!model && p.ID != "dcc-basics") {
				continue
			}
			if !model && modelDefinitionsOnly {
				continue
			}
			if !model && customCurve && (value.CVNumber == 6 || (value.CVNumber >= 67 && value.CVNumber <= 94)) {
				continue
			}
			if model && (p.ID != match.ProfileID || (match.Status != "matched" && match.Status != "family") ||
				(match.FirmwareCompatible != nil && !*match.FirmwareCompatible)) {
				continue
			}
			for _, definition := range p.Definitions {
				d := definition
				end := d.CVEnd
				if end == 0 {
					end = d.CV
				}
				if value.CVNumber < d.CV || value.CVNumber > end {
					continue
				}
				if (value.CVNumber-d.CV)%max(1, d.CVStride) != 0 ||
					(d.FirmwareMin != "" && match.Firmware != "" && !versionAtLeast(match.Firmware, d.FirmwareMin)) {
					continue
				}
				if d.IndexCV31 == nil {
					if value.IndexCV31 == nil && value.IndexCV32 == nil {
						return &d
					}
					continue
				}
				if value.IndexCV31 == nil || value.IndexCV32 == nil || *value.IndexCV31 != *d.IndexCV31 {
					continue
				}
				last := *d.IndexCV32
				if d.IndexCV32End != nil {
					last = *d.IndexCV32End
				}
				if *value.IndexCV32 >= *d.IndexCV32 && *value.IndexCV32 <= last {
					return &d
				}
			}
		}
	}
	return nil
}
