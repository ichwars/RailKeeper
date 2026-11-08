package masterdata

import (
	"regexp"
	"strings"
)

func CleanMasterDataInput(input MasterDataInput) MasterDataInput {
	input.Key = strings.TrimSpace(input.Key)
	input.Label = strings.TrimSpace(input.Label)
	input.SourceURL = strings.TrimSpace(input.SourceURL)
	if input.Metadata == nil {
		input.Metadata = map[string]any{}
	}
	return input
}

var slugPattern = regexp.MustCompile(`[^a-z0-9]+`)

func SlugKey(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.NewReplacer("\u00e4", "ae", "\u00f6", "oe", "\u00fc", "ue", "\u00df", "ss").Replace(value)
	value = slugPattern.ReplaceAllString(value, "-")
	return strings.Trim(value, "-")
}
