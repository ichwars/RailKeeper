// Package decoderlibrary contains declarative, local decoder documentation.
package decoderlibrary

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"
)

const MasterDataType = "decoder_profile"

//go:embed catalog.json
var bundledJSON []byte

type Text struct {
	DE string `json:"de"`
	EN string `json:"en"`
}

type Source struct {
	Title       string `json:"title"`
	URL         string `json:"url"`
	Version     string `json:"version"`
	Scope       string `json:"scope"`
	SHA256      string `json:"sha256,omitempty"`
	RetrievedAt string `json:"retrievedAt,omitempty"`
}

type Bit struct {
	Bit  int  `json:"bit"`
	Name Text `json:"name"`
	Off  Text `json:"off"`
	On   Text `json:"on"`
}

// Reference identifies the reviewed table or section within a profile source.
type Reference struct {
	SourceIndex int    `json:"sourceIndex"`
	Location    string `json:"location"`
}

type Definition struct {
	CV             int        `json:"cv"`
	CVEnd          int        `json:"cvEnd,omitempty"`
	CVStride       int        `json:"cvStride,omitempty"`
	FirmwareMin    string     `json:"firmwareMin,omitempty"`
	IndexCV31      *int       `json:"indexCV31,omitempty"`
	IndexCV32      *int       `json:"indexCV32,omitempty"`
	IndexCV32End   *int       `json:"indexCV32End,omitempty"`
	Name           Text       `json:"name"`
	Help           Text       `json:"help"`
	Topic          string     `json:"topic"`
	Minimum        int        `json:"minimum"`
	Maximum        int        `json:"maximum"`
	Bits           []Bit      `json:"bits,omitempty"`
	Reference      *Reference `json:"reference,omitempty"`
	FactoryDefault *int       `json:"factoryDefault,omitempty"`
	Access         string     `json:"access,omitempty"`
	Unit           *Text      `json:"unit,omitempty"`
	AllowedValues  []int      `json:"allowedValues,omitempty"`
	Reserved       bool       `json:"reserved,omitempty"`
}

type BitRule struct {
	CV   int `json:"cv"`
	Mask int `json:"mask"`
}

type BrakeRule struct {
	ForwardCV int `json:"forwardCV"`
	ReverseCV int `json:"reverseCV"`
	EnableCV  int `json:"enableCV,omitempty"`
}

type Rules struct {
	ModelDefinitionsOnly bool       `json:"modelDefinitionsOnly,omitempty"`
	PIKO56500            bool       `json:"piko56500,omitempty"`
	Roco10882            bool       `json:"roco10882,omitempty"`
	Uhlenbrock76420      bool       `json:"uhlenbrock76420,omitempty"`
	ZIMOMN               bool       `json:"zimoMN,omitempty"`
	LenzStandardV2       bool       `json:"lenzStandardV2,omitempty"`
	KuehnN45             bool       `json:"kuehnN45,omitempty"`
	MotorCurveCV         int        `json:"motorCurveCV,omitempty"`
	ABC                  *BitRule   `json:"abc,omitempty"`
	DCAnalog             *BitRule   `json:"dcAnalog,omitempty"`
	ConstantBrake        *BrakeRule `json:"constantBrake,omitempty"`
}

type Mapping struct {
	Rows          int `json:"rows"`
	RowsPerPage   int `json:"rowsPerPage"`
	Stride        int `json:"stride"`
	StartCV       int `json:"startCV"`
	IndexCV31     int `json:"indexCV31"`
	ConditionPage int `json:"conditionPage"`
	ActionPage    int `json:"actionPage"`
}

type Profile struct {
	ID             string       `json:"id"`
	Name           string       `json:"name"`
	NameEN         string       `json:"nameEn"`
	Revision       string       `json:"revision"`
	Protocol       string       `json:"protocol"`
	Kind           string       `json:"kind"`
	ManufacturerID *int         `json:"manufacturerId,omitempty"`
	Family         string       `json:"family,omitempty"`
	Model          string       `json:"model,omitempty"`
	ParentID       string       `json:"parentId,omitempty"`
	FirmwareMin    string       `json:"firmwareMin,omitempty"`
	Aliases        []string     `json:"aliases"`
	FamilyAliases  []string     `json:"familyAliases"`
	Sources        []Source     `json:"sources"`
	Definitions    []Definition `json:"definitions"`
	Rules          *Rules       `json:"rules,omitempty"`
	Mapping        *Mapping     `json:"mapping,omitempty"`
	Active         bool         `json:"active"`
}

type Catalog struct {
	SchemaVersion     int                 `json:"schemaVersion"`
	Profiles          []Profile           `json:"profiles"`
	Inventory         *Inventory          `json:"inventory,omitempty"`
	ReferenceCoverage []ReferenceCoverage `json:"referenceCoverage,omitempty"`
}

var identifier = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,79}$`)
var version = regexp.MustCompile(`^\d{1,4}(\.\d{1,4}){1,3}$`)
var documentDigest = regexp.MustCompile(`^[0-9a-f]{64}$`)

func validSourceFingerprint(source Source) bool {
	if source.SHA256 != "" && !documentDigest.MatchString(source.SHA256) {
		return false
	}
	if source.RetrievedAt != "" {
		_, err := time.Parse("2006-01-02", source.RetrievedAt)
		return err == nil
	}
	return true
}

func Bundled() Catalog {
	var catalog Catalog
	if err := json.Unmarshal(bundledJSON, &catalog); err != nil {
		panic(err)
	}
	for i := range catalog.Profiles {
		catalog.Profiles[i].Active = true
		if err := catalog.Profiles[i].Validate(); err != nil {
			panic(err)
		}
	}
	return catalog
}

func Parse(metadata map[string]any) (Profile, error) {
	var profile Profile
	body, err := json.Marshal(metadata)
	if err != nil || len(body) > 128*1024 {
		return profile, errors.New("invalid decoder profile metadata")
	}
	if err := json.Unmarshal(body, &profile); err != nil {
		return profile, err
	}
	if profile.Aliases == nil {
		profile.Aliases = []string{}
	}
	if profile.FamilyAliases == nil {
		profile.FamilyAliases = []string{}
	}
	return profile, profile.Validate()
}

func validText(value Text) bool {
	return strings.TrimSpace(value.DE) != "" && strings.TrimSpace(value.EN) != "" && len(value.DE) <= 1500 && len(value.EN) <= 1500
}

func (p Profile) Validate() error {
	if !identifier.MatchString(p.ID) || p.Name == "" || p.NameEN == "" || len(p.Name) > 200 || len(p.NameEN) > 200 ||
		p.Revision == "" || len(p.Revision) > 100 || p.Protocol != "DCC" || (p.Kind != "standard" && p.Kind != "model") ||
		(p.ParentID != "" && (!identifier.MatchString(p.ParentID) || p.ParentID == p.ID)) ||
		(p.FirmwareMin != "" && !version.MatchString(p.FirmwareMin)) || len(p.Definitions) > 256 ||
		len(p.Sources) == 0 || len(p.Sources) > 10 || len(p.Aliases)+len(p.FamilyAliases) > 64 {
		return errors.New("invalid decoder profile identity or limits")
	}
	if p.Kind == "model" && (p.ManufacturerID == nil || *p.ManufacturerID < 0 || *p.ManufacturerID > 255 || p.Model == "" || p.ParentID != "dcc-basics") {
		return errors.New("decoder model requires manufacturer and model")
	}
	if p.Kind == "standard" && (p.ID != "dcc-basics" || p.Rules != nil || p.Mapping != nil) {
		return errors.New("unsupported DCC foundation")
	}
	if len(p.Family) > 200 || len(p.Model) > 200 || len(p.Definitions) == 0 {
		return errors.New("invalid decoder profile description")
	}
	for _, alias := range append(append([]string{}, p.Aliases...), p.FamilyAliases...) {
		if strings.TrimSpace(alias) == "" || len(alias) > 200 {
			return errors.New("invalid decoder profile alias")
		}
	}
	for _, source := range p.Sources {
		u, err := url.Parse(source.URL)
		if err != nil || u.Scheme != "https" || u.Hostname() == "" || u.User != nil || source.Title == "" || len(source.URL) > 2000 || !validSourceFingerprint(source) {
			return errors.New("decoder source requires a valid HTTPS reference")
		}
	}
	seen := map[string]bool{}
	for _, d := range p.Definitions {
		if err := d.validate(); err != nil {
			return err
		}
		if d.Reference != nil && (d.Reference.SourceIndex < 0 || d.Reference.SourceIndex >= len(p.Sources) ||
			strings.TrimSpace(d.Reference.Location) == "" || len(d.Reference.Location) > 200) {
			return errors.New("invalid CV documentation reference")
		}
		end := d.CVEnd
		if end == 0 {
			end = d.CV
		}
		firstPage, lastPage := -1, -1
		if d.IndexCV32 != nil {
			firstPage, lastPage = *d.IndexCV32, *d.IndexCV32
		}
		if d.IndexCV32End != nil {
			lastPage = *d.IndexCV32End
		}
		stride := max(1, d.CVStride)
		if len(seen)+((end-d.CV)/stride+1)*(lastPage-firstPage+1) > 65536 {
			return errors.New("decoder definitions exceed the supported address count")
		}
		for cv := d.CV; cv <= end; cv += stride {
			for page := firstPage; page <= lastPage; page++ {
				index := -1
				if d.IndexCV31 != nil {
					index = *d.IndexCV31
				}
				key := fmt.Sprintf("%d/%d/%d", cv, index, page)
				if seen[key] {
					return errors.New("overlapping CV definitions")
				}
				seen[key] = true
			}
		}
	}
	if p.Rules != nil {
		if p.Rules.PIKO56500 && (p.ManufacturerID == nil || *p.ManufacturerID != 162 ||
			p.Kind != "model" || p.Family != "SmartDecoder XP 5.1" || p.Model != "SmartDecoder XP 5.1 PluX22" ||
			p.Rules.Roco10882 || p.Rules.Uhlenbrock76420 || p.Rules.KuehnN45 || p.Rules.LenzStandardV2 || p.Rules.ZIMOMN ||
			p.Rules.MotorCurveCV != 0 || p.Rules.ConstantBrake != nil || p.Mapping != nil ||
			p.Rules.ABC == nil || p.Rules.ABC.CV != 27 || p.Rules.ABC.Mask != 3 ||
			p.Rules.DCAnalog == nil || p.Rules.DCAnalog.CV != 12 || p.Rules.DCAnalog.Mask != 1) {
			return errors.New("unsupported PIKO 56500 evaluation scope")
		}
		if p.Rules.Roco10882 && (p.ManufacturerID == nil || *p.ManufacturerID != 145 ||
			p.Kind != "model" || p.Family != "Roco 10882" || p.Model != "10882 · PluX16-Decoder (NEM 658)" ||
			p.Rules.Uhlenbrock76420 || p.Rules.KuehnN45 || p.Rules.LenzStandardV2 || p.Rules.ZIMOMN ||
			p.Rules.MotorCurveCV != 0 || p.Rules.ABC != nil || p.Rules.DCAnalog != nil ||
			p.Rules.ConstantBrake != nil || p.Mapping != nil) {
			return errors.New("unsupported Roco 10882 evaluation scope")
		}
		if p.Rules.Uhlenbrock76420 && (p.ManufacturerID == nil || *p.ManufacturerID != 85 ||
			p.Kind != "model" || p.Family != "IntelliDrive" || p.Model != "IntelliDrive 76420" ||
			p.Rules.KuehnN45 || p.Rules.LenzStandardV2 || p.Rules.ZIMOMN || p.Rules.MotorCurveCV != 0 ||
			p.Rules.ABC != nil || p.Rules.DCAnalog != nil || p.Rules.ConstantBrake != nil || p.Mapping != nil) {
			return errors.New("unsupported Uhlenbrock 76420 evaluation scope")
		}
		if p.Rules.KuehnN45 && (p.ManufacturerID == nil || *p.ManufacturerID != 157 ||
			p.Family != "N45" || p.Model != "N45" || p.Rules.LenzStandardV2 || p.Rules.ZIMOMN ||
			p.Rules.MotorCurveCV != 0 || p.Rules.ConstantBrake != nil || p.Rules.DCAnalog != nil || p.Mapping != nil ||
			p.Rules.ABC == nil || p.Rules.ABC.CV != 27 || p.Rules.ABC.Mask != 3) {
			return errors.New("unsupported Kühn N45 evaluation scope")
		}
		if p.Rules.LenzStandardV2 && (p.ManufacturerID == nil || *p.ManufacturerID != 99 ||
			p.Family != "STANDARD+ V2" || p.Model != "STANDARD+ V2" || p.Rules.ZIMOMN ||
			p.Rules.MotorCurveCV != 0 || p.Rules.ConstantBrake != nil || p.Rules.DCAnalog != nil || p.Mapping != nil ||
			p.Rules.ABC == nil || p.Rules.ABC.CV != 51 || p.Rules.ABC.Mask != 2) {
			return errors.New("unsupported Lenz STANDARD+ V2 evaluation scope")
		}
		if p.Rules.ZIMOMN && (p.ManufacturerID == nil || *p.ManufacturerID != 145 ||
			p.Family != "MN" || p.Model != "MN180N18" || p.FirmwareMin == "" ||
			!versionAtLeast(p.FirmwareMin, "5.29") || p.Rules.MotorCurveCV != 0 ||
			p.Rules.ConstantBrake != nil || p.Mapping != nil) {
			return errors.New("unsupported ZIMO MN evaluation scope")
		}
		if cv := p.Rules.MotorCurveCV; cv != 0 {
			found := false
			for _, d := range p.Definitions {
				if d.CV == cv && d.CVEnd == 0 && d.IndexCV31 == nil && d.Topic == "motor" && !d.Reserved {
					found = true
				}
			}
			if cv < 1 || cv > 256 || !found {
				return errors.New("motor characteristic requires a documented base motor CV")
			}
		}
		for _, rule := range []*BitRule{p.Rules.ABC, p.Rules.DCAnalog} {
			if rule != nil && (rule.CV < 1 || rule.CV > 1024 || rule.Mask < 1 || rule.Mask > 255) {
				return errors.New("invalid decoder bit rule")
			}
		}
		if r := p.Rules.ConstantBrake; r != nil && (r.ForwardCV < 1 || r.ForwardCV > 1024 ||
			r.ReverseCV < 1 || r.ReverseCV > 1024 || r.EnableCV < 0 || r.EnableCV > 1024) {
			return errors.New("invalid brake distance rule")
		}
	}
	// Only the reviewed LokPilot-5 algorithm is supported; metadata cannot inject executable rules.
	if m := p.Mapping; m != nil && (p.ManufacturerID == nil || *p.ManufacturerID != 151 || p.Family != "LokPilot 5" ||
		m.Rows != 72 || m.RowsPerPage != 16 || m.Stride != 16 || m.StartCV != 257 || m.IndexCV31 != 16 || m.ConditionPage != 3 || m.ActionPage != 8) {
		return errors.New("unsupported decoder mapping layout")
	}
	return nil
}

func (d Definition) validate() error {
	end := d.CVEnd
	if end == 0 {
		end = d.CV
	}
	if d.CV < 1 || end < d.CV || end > 1024 || d.Minimum < 0 || d.Maximum > 255 || d.Minimum > d.Maximum ||
		!validText(d.Name) || !validText(d.Help) || len(d.Bits) > 8 {
		return fmt.Errorf("invalid CV definition %d", d.CV)
	}
	if (d.FactoryDefault != nil && (*d.FactoryDefault < d.Minimum || *d.FactoryDefault > d.Maximum)) ||
		(d.Access != "" && d.Access != "read-only" && d.Access != "read-write" && d.Access != "command") ||
		(d.Unit != nil && !validText(*d.Unit)) {
		return fmt.Errorf("invalid CV documented value metadata %d", d.CV)
	}
	if d.Reserved && (d.Reference == nil || len(d.Bits) > 0 || d.FactoryDefault != nil ||
		len(d.AllowedValues) > 0 || d.Unit != nil || d.Access != "") {
		return errors.New("reserved CV requires a source and cannot declare functional semantics")
	}
	if len(d.AllowedValues) > 256 {
		return errors.New("too many CV allowed values")
	}
	values := map[int]bool{}
	for _, value := range d.AllowedValues {
		if value < d.Minimum || value > d.Maximum || values[value] {
			return errors.New("invalid CV allowed value")
		}
		values[value] = true
	}
	if len(values) > 0 && d.FactoryDefault != nil && !values[*d.FactoryDefault] {
		return errors.New("CV factory default not in allowed values")
	}
	if d.CVStride < 0 || d.CVStride > 256 || (d.CVStride > 0 &&
		(d.CVEnd == 0 || (end-d.CV)%d.CVStride != 0)) ||
		(d.FirmwareMin != "" && !version.MatchString(d.FirmwareMin)) {
		return fmt.Errorf("invalid CV stride or firmware scope %d", d.CV)
	}
	if !strings.Contains("|address|motor|braking|outputs|communication|", "|"+d.Topic+"|") {
		return errors.New("invalid CV topic")
	}
	if (d.IndexCV31 == nil) != (d.IndexCV32 == nil) || (d.IndexCV32End != nil && d.IndexCV32 == nil) {
		return errors.New("incomplete CV index pair")
	}
	for _, index := range []*int{d.IndexCV31, d.IndexCV32, d.IndexCV32End} {
		if index != nil && (*index < 0 || *index > 255 || d.CV < 257 || end > 512) {
			return errors.New("invalid indexed CV definition")
		}
	}
	if d.IndexCV32End != nil && *d.IndexCV32End < *d.IndexCV32 {
		return errors.New("invalid CV page range")
	}
	bits := map[int]bool{}
	for _, bit := range d.Bits {
		if bit.Bit < 0 || bit.Bit > 7 || bits[bit.Bit] || !validText(bit.Name) || !validText(bit.Off) || !validText(bit.On) {
			return errors.New("invalid CV bit definition")
		}
		bits[bit.Bit] = true
	}
	return nil
}
