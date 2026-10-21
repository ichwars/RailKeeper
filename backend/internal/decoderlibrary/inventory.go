package decoderlibrary

import (
	_ "embed"
	"encoding/json"
	"errors"
	"net/url"
	"slices"
	"strings"
	"time"
)

const InventoryMasterDataType = "decoder_catalog"
const InventoryKey = "library"

//go:embed inventory.json
var inventoryJSON []byte

type Manufacturer struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	CV8      int    `json:"cv8"`
	Website  string `json:"website"`
	Source   Source `json:"source"`
	Coverage *Text  `json:"coverage,omitempty"`
}

type Family struct {
	ID             string   `json:"id"`
	ManufacturerID string   `json:"manufacturerId"`
	Name           string   `json:"name"`
	Sources        []Source `json:"sources"`
	Notes          Text     `json:"notes"`
	Identification Text     `json:"identification"`
}

type Model struct {
	ID             string   `json:"id"`
	FamilyID       string   `json:"familyId"`
	Name           string   `json:"name"`
	Kind           string   `json:"kind"`
	Interfaces     []string `json:"interfaces"`
	ArticleNumbers []string `json:"articleNumbers"`
	ProfileID      string   `json:"profileId,omitempty"`
	// DocumentationProfileID links branded products to a manual without enabling identity matching.
	DocumentationProfileID string        `json:"documentationProfileId,omitempty"`
	Notes                  *Text         `json:"notes,omitempty"`
	Sources                []Source      `json:"sources,omitempty"`
	Identification         *IdentityRule `json:"identification,omitempty"`
	Variants               []Variant     `json:"variants,omitempty"`
	BrandOnly              bool          `json:"brandOnly,omitempty"`
	DocumentStatus         string        `json:"documentStatus,omitempty"`
}

// Variant pairs an orderable article with its documented connection, never by list position.
type Variant struct {
	ArticleNumber string `json:"articleNumber"`
	Interface     string `json:"interface"`
	Source        Source `json:"source"`
}

type IdentityRule struct {
	CV     int    `json:"cv"`
	Values []int  `json:"values"`
	Scope  string `json:"scope"`
	Name   string `json:"name"`
}

// Inventory lists documented products. Only a linked, active Profile supplies CV interpretation.
type Inventory struct {
	Revision      string         `json:"revision"`
	ReviewedAt    string         `json:"reviewedAt"`
	Manufacturers []Manufacturer `json:"manufacturers"`
	Families      []Family       `json:"families"`
	Models        []Model        `json:"models"`
}

func BundledInventory() Inventory {
	var inventory Inventory
	if err := json.Unmarshal(inventoryJSON, &inventory); err != nil {
		panic(err)
	}
	if err := inventory.Validate(); err != nil {
		panic(err)
	}
	return inventory
}

func ParseInventory(metadata map[string]any) (Inventory, error) {
	var inventory Inventory
	body, err := json.Marshal(metadata)
	if err != nil || len(body) > 512*1024 {
		return inventory, errors.New("invalid decoder inventory metadata")
	}
	if err := json.Unmarshal(body, &inventory); err != nil {
		return inventory, err
	}
	for index := range inventory.Models {
		if inventory.Models[index].Interfaces == nil {
			inventory.Models[index].Interfaces = []string{}
		}
		if inventory.Models[index].ArticleNumbers == nil {
			inventory.Models[index].ArticleNumbers = []string{}
		}
	}
	return inventory, inventory.Validate()
}

func validHTTPS(value string) bool {
	u, err := url.Parse(value)
	return err == nil && u.Scheme == "https" && u.Hostname() != "" && u.User == nil && len(value) <= 2000
}

func validSource(source Source) bool {
	return validHTTPS(source.URL) && validLabel(source.Title) && validLabel(source.Version) &&
		strings.TrimSpace(source.Scope) != "" && len(source.Scope) <= 1500 && validSourceFingerprint(source)
}

func validLabel(value string) bool { return strings.TrimSpace(value) != "" && len(value) <= 200 }

func (i Inventory) Validate() error {
	if _, err := time.Parse("2006-01-02", i.ReviewedAt); err != nil || !validLabel(i.Revision) ||
		len(i.Manufacturers) < 1 || len(i.Manufacturers) > 128 || len(i.Families) < 1 || len(i.Families) > 512 ||
		len(i.Models) < 1 || len(i.Models) > 2048 {
		return errors.New("invalid decoder inventory revision or limits")
	}
	makers, families, models := map[string]bool{}, map[string]bool{}, map[string]bool{}
	for _, maker := range i.Manufacturers {
		if !identifier.MatchString(maker.ID) || makers[maker.ID] || !validLabel(maker.Name) ||
			maker.CV8 < 0 || maker.CV8 > 255 || !validHTTPS(maker.Website) || !validSource(maker.Source) ||
			(maker.Coverage != nil && !validText(*maker.Coverage)) {
			return errors.New("invalid decoder manufacturer")
		}
		makers[maker.ID] = true
	}
	for _, family := range i.Families {
		if !identifier.MatchString(family.ID) || families[family.ID] || !makers[family.ManufacturerID] ||
			!validLabel(family.Name) || !validText(family.Notes) || !validText(family.Identification) ||
			len(family.Sources) == 0 || len(family.Sources) > 10 {
			return errors.New("invalid decoder family")
		}
		for _, source := range family.Sources {
			if !validSource(source) {
				return errors.New("invalid decoder family source")
			}
		}
		families[family.ID] = true
	}
	for _, model := range i.Models {
		if !identifier.MatchString(model.ID) || models[model.ID] || !families[model.FamilyID] ||
			!validLabel(model.Name) || (model.Kind != "motor" && model.Kind != "sound" && model.Kind != "function") ||
			(model.ProfileID != "" && !identifier.MatchString(model.ProfileID)) ||
			(model.DocumentationProfileID != "" && (!identifier.MatchString(model.DocumentationProfileID) || !model.BrandOnly || model.ProfileID != "")) ||
			(model.Notes != nil && !validText(*model.Notes)) || len(model.Interfaces)+len(model.ArticleNumbers) > 32 || len(model.Sources) > 10 ||
			len(model.Variants) > 64 || (model.BrandOnly && (model.Identification != nil || model.ProfileID != "")) {
			return errors.New("invalid decoder model")
		}
		if model.DocumentStatus != "" && model.DocumentStatus != "historical" && model.DocumentStatus != "announced" {
			return errors.New("invalid decoder documentation status")
		}
		for _, label := range append(append([]string{}, model.Interfaces...), model.ArticleNumbers...) {
			if !validLabel(label) {
				return errors.New("invalid decoder model variant")
			}
		}
		for _, source := range model.Sources {
			if !validSource(source) {
				return errors.New("invalid decoder model source")
			}
		}
		articles := map[string]bool{}
		for _, variant := range model.Variants {
			if !validLabel(variant.ArticleNumber) || !validLabel(variant.Interface) || !validSource(variant.Source) ||
				articles[variant.ArticleNumber] || !slices.Contains(model.ArticleNumbers, variant.ArticleNumber) ||
				!slices.Contains(model.Interfaces, variant.Interface) {
				return errors.New("invalid documented decoder variant")
			}
			articles[variant.ArticleNumber] = true
		}
		if rule := model.Identification; rule != nil {
			if rule.CV < 1 || rule.CV > 1024 || rule.CV == 8 || rule.CV == 31 || rule.CV == 32 ||
				len(rule.Values) < 1 || len(rule.Values) > 16 || !validLabel(rule.Name) || (rule.Scope != "model" && rule.Scope != "family") {
				return errors.New("invalid decoder identity rule")
			}
			seen := map[int]bool{}
			for _, value := range rule.Values {
				if value < 0 || value > 255 || seen[value] {
					return errors.New("invalid decoder identity values")
				}
				seen[value] = true
			}
		}
		models[model.ID] = true
	}
	return nil
}
