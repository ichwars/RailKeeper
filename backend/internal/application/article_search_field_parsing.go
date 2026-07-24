package application

import (
	"regexp"
	"strings"
)

func labeledValue(text string, labels []string) string {
	text = repairMojibake(text)
	lines := regexp.MustCompile(`[

]+`).Split(text, -1)
	for _, line := range lines {
		line = normalizeWhitespace(line)
		lowerLine := strings.ToLower(line)
		for _, label := range labels {
			lowerLabel := strings.ToLower(label)
			if lowerLine == lowerLabel {
				continue
			}
			if strings.HasPrefix(lowerLine, lowerLabel+":") || strings.HasPrefix(lowerLine, lowerLabel+" ") {
				value := strings.TrimSpace(line[len(label):])
				value = strings.Trim(value, " -:;,.	")
				if value != "" {
					return value
				}
			}
		}
	}
	if value := compactLabeledValue(text, labels); value != "" {
		return value
	}
	for _, label := range labels {
		pattern := regexp.MustCompile(`(?im)^\s*` + regexp.QuoteMeta(label) + `\s*:?\s*([^\n\r.;|]{1,90})`)
		for _, match := range pattern.FindAllStringSubmatch(text, -1) {
			if len(match) < 2 {
				continue
			}
			value := normalizeWhitespace(match[1])
			value = strings.Trim(value, " -:;,.	")
			if value != "" {
				return value
			}
		}
	}
	return ""
}

func compactLabeledValue(text string, labels []string) string {
	searchText := text
	lower := strings.ToLower(text)
	if detailsStart := strings.Index(lower, "daten & details:"); detailsStart >= 0 {
		searchText = text[detailsStart:]
		lower = strings.ToLower(searchText)
	}
	for _, label := range labels {
		lowerLabel := strings.ToLower(label)
		index := strings.Index(lower, lowerLabel)
		for index >= 0 {
			start := index + len(label)
			if start < len(searchText) {
				for start < len(searchText) && (searchText[start] == ' ' || searchText[start] == ':' || searchText[start] == '\t' || searchText[start] == '\u00a0') {
					start++
				}
				end := len(searchText)
				for _, nextLabel := range catalogDetailLabels {
					if strings.EqualFold(nextLabel, label) {
						continue
					}
					nextLower := strings.ToLower(nextLabel)
					if next := strings.Index(lower[start:], nextLower); next > 0 && start+next < end {
						end = start + next
					}
				}
				value := normalizeWhitespace(searchText[start:end])
				value = strings.Trim(value, " -:;,.	")
				if value != "" && len(value) <= 90 {
					return value
				}
			}
			nextIndex := strings.Index(lower[index+len(lowerLabel):], lowerLabel)
			if nextIndex < 0 {
				break
			}
			index += len(lowerLabel) + nextIndex
		}
	}
	return ""
}

func catalogDescription(value, resultURL string) string {
	if !isCatalogURL(resultURL) {
		return ""
	}
	value = repairMojibake(value)
	lower := strings.ToLower(value)
	start := strings.Index(lower, "beschreibung:")
	for start > 0 && strings.Contains(lower[maxInt(0, start-14):start], "beleuchtung") {
		next := strings.Index(lower[start+len("beschreibung:"):], "beschreibung:")
		if next < 0 {
			return ""
		}
		start += len("beschreibung:") + next
	}
	if start < 0 {
		return ""
	}
	description := value[start+len("beschreibung:"):]
	if end := strings.Index(strings.ToLower(description), "daten & details:"); end >= 0 {
		description = description[:end]
	}
	description = normalizeWhitespace(description)
	description = strings.Trim(description, " -:;,.	")
	if len(description) > 520 {
		description = strings.TrimSpace(description[:520])
	}
	if len(description) < 20 {
		return ""
	}
	return description
}

func maxInt(left, right int) int {
	if left > right {
		return left
	}
	return right
}

func extractPrice(value string) string {
	for _, label := range []string{"Hersteller-Preis", "Herstellerpreis", "Preis", "UVP"} {
		labeled := labeledValue(value, []string{label})
		if labeled == "" {
			continue
		}
		if match := regexp.MustCompile(`\d{1,4}(?:[,.]\d{2})?`).FindString(labeled); match != "" {
			return normalizePrice(match)
		}
	}
	if value := firstRegexValue(pricePattern, value); value != "" {
		return normalizePrice(value)
	}
	return ""
}

func normalizePrice(value string) string {
	value = strings.ReplaceAll(strings.TrimSpace(value), ".", "")
	value = strings.ReplaceAll(value, ",", ".")
	return value
}

func extractLengthMM(value string) string {
	for _, match := range lengthPattern.FindAllStringSubmatch(value, -1) {
		if len(match) < 2 {
			continue
		}
		candidate := strings.ReplaceAll(strings.TrimSpace(match[1]), ",", ".")
		whole := strings.TrimSpace(match[0])
		if !looksLikeModelLength(candidate, whole) {
			continue
		}
		return candidate
	}
	return ""
}

func looksLikeModelLength(candidate, context string) bool {
	normalized := strings.ReplaceAll(candidate, ",", ".")
	parts := strings.Split(normalized, ".")
	number := parts[0]
	if len(number) == 4 && strings.HasPrefix(number, "20") {
		return false
	}
	var integer int
	for _, char := range number {
		if char < '0' || char > '9' {
			return false
		}
		integer = integer*10 + int(char-'0')
	}
	if integer < 20 || integer > 600 {
		return false
	}
	lower := strings.ToLower(context)
	return strings.Contains(lower, "mm") ||
		strings.Contains(lower, "laenge") ||
		strings.Contains(lower, "länge") ||
		strings.Contains(lower, "laenge") ||
		strings.Contains(lower, "length") ||
		strings.Contains(lower, "mass") ||
		strings.Contains(lower, "maß") ||
		strings.Contains(lower, "luep")
}

func extractHeadlightDescription(value string) string {
	description := firstRegexValue(headlightDescriptionPattern, value)
	if description == "" {
		description = sentenceForKeywords(value, []string{"lichtwechsel", "fahrlicht", "spitzenlicht", "spitzenbeleuchtung", "schlusslicht"})
	}
	if description == "" {
		return ""
	}
	return cleanTechnicalDescription(description)
}

func extractLightingDescription(value string) string {
	description := firstRegexValue(lightingDescriptionPattern, value)
	if description == "" {
		return ""
	}
	lower := strings.ToLower(description)
	if strings.Contains(lower, "fahrtrichtung") || strings.Contains(lower, "lichtwechsel") {
		return ""
	}
	return cleanTechnicalDescription(description)
}

func extractSoundDescription(value string) string {
	lower := strings.ToLower(value)
	if strings.Contains(lower, "ohne sound") || strings.Contains(lower, "kein sound") {
		return ""
	}
	description := firstRegexValue(soundDescriptionPattern, value)
	if cleaned := cleanTechnicalDescription(description); cleaned != "" {
		return cleaned
	}
	if description == "" || cleanTechnicalDescription(description) == "" {
		description = sentenceForKeywords(value, []string{"sound-modul", "soundmodul", "sounddecoder", "soundgenerator", "geräuschmodul", "geraeuschmodul"})
	}
	if description == "" {
		return ""
	}
	return cleanTechnicalDescription(description)
}

func extractAdapterInfo(value string) string {
	matches := adapterPattern.FindAllString(value, -1)
	if len(matches) == 0 {
		return ""
	}
	parts := []string{}
	for _, match := range matches {
		part := normalizeWhitespace(match)
		if part != "" {
			parts = append(parts, part)
		}
	}
	return strings.Join(uniqueNonEmpty(parts), " ")
}

func sentenceForKeywords(value string, keywords []string) string {
	for _, candidate := range regexp.MustCompile(`[.;\n\r]+`).Split(value, -1) {
		candidate = normalizeWhitespace(candidate)
		if candidate == "" {
			continue
		}
		lower := strings.ToLower(candidate)
		for _, keyword := range keywords {
			if strings.Contains(lower, keyword) {
				return candidate
			}
		}
	}
	return ""
}

func cleanTechnicalDescription(value string) string {
	value = normalizeWhitespace(repairMojibake(value))
	value = trimTechnicalNoise(value)
	value = strings.Trim(value, " -:;,.")
	if !looksLikeTechnicalDescription(value) {
		return ""
	}
	return value
}

func trimTechnicalNoise(value string) string {
	lower := strings.ToLower(value)
	end := len(value)
	for _, marker := range []string{
		" downloads", " bedienungsanleitung", " altersempfehlung", " de | en",
		" menü", " menue", " menu", " sprunggröße", " sprunggroesse",
		" wählen sie", " waehlen sie",
	} {
		if index := strings.Index(lower, marker); index > 0 && index < end {
			end = index
		}
	}
	return strings.TrimSpace(value[:end])
}

func looksLikeTechnicalDescription(value string) bool {
	value = strings.TrimSpace(value)
	if len(value) < 3 || len(value) > 220 {
		return false
	}
	lower := strings.ToLower(value)
	badTokens := []string{"google_analytics", "cookie", "mandatory", "preferences", "statistics", "marketing", "function", "const ", "new map", "document.", "window.", "{", "};", "class ", "anzeigen zu zeigen", "personalisierte anzeigen", "absicht ist", "menü", "menue", "menu", "sprunggröße", "sprunggroesse", "wählen sie", "waehlen sie", "downloads", "bedienungsanleitung", "altersempfehlung"}
	if strings.HasPrefix(lower, "//") || strings.Contains(lower, "://") {
		return false
	}
	for _, token := range badTokens {
		if strings.Contains(lower, token) {
			return false
		}
	}
	return true
}

func hasExplicitHeadlight(value string) bool {
	return strings.Contains(value, "lichtwechsel") ||
		strings.Contains(value, "spitzenlicht") ||
		strings.Contains(value, "schlusslicht") ||
		strings.Contains(value, "fahrlicht")
}

func hasExplicitInteriorLighting(value string) bool {
	return strings.Contains(value, "innenbeleuchtung") ||
		strings.Contains(value, "fuehrerstandsbeleuchtung") ||
		strings.Contains(value, "führerstandsbeleuchtung") ||
		strings.Contains(value, "kabinenbeleuchtung")
}

func hasExplicitSoundGenerator(value string) bool {
	if strings.Contains(value, "ohne sound") || strings.Contains(value, "kein sound") {
		return false
	}
	return strings.Contains(value, "soundgenerator") ||
		strings.Contains(value, "sounddecoder") ||
		strings.Contains(value, "sound-modul") ||
		strings.Contains(value, "soundmodul") ||
		strings.Contains(value, "sound laut artikeldaten") ||
		strings.Contains(value, "geraeuschmodul") ||
		strings.Contains(value, "geräuschmodul")
}
