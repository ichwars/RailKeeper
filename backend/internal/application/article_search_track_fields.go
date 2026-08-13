package application

import (
	"regexp"
	"strconv"
	"strings"
)

var trackNumberPattern = regexp.MustCompile(`[-+]?\d+(?:[,.]\d+)?`)

var trackDetailLabels = []string{
	"Gleissystem", "Gleissortiment", "Track system",
	"Länge", "Laenge", "Length", "Radius", "Winkel", "Weichenwinkel", "Angle",
	"Richtung", "Direction", "Herzstückwinkel", "Herzstueckwinkel", "Frog angle",
	"Schwellenart", "Schwellen", "Sleeper type", "Profilhöhe", "Profilhoehe", "Rail height",
	"Bettung", "Gleisbettung", "Roadbed", "Anzahl Anschlüsse", "Anzahl Anschluesse", "Connections",
	"Digitaltauglich", "Digital geeignet", "Digital ready",
}

func buildTrackArticleFields(input ArticleSearchInput, text string) map[string]ArticleSearchField {
	if !strings.EqualFold(strings.TrimSpace(input.Fields["articleType"]), "track") {
		return nil
	}
	fields := map[string]ArticleSearchField{}
	addTextTrackField(fields, "trackSystem", "Gleissystem", text,
		[]string{"Gleissystem", "Gleissortiment", "Track system"})
	addDecimalTrackField(fields, "lengthMm", "Länge (mm)", text,
		[]string{"Länge", "Laenge", "Length"})
	addDecimalTrackField(fields, "radiusMm", "Radius (mm)", text, []string{"Radius"})
	addDecimalTrackField(fields, "angleDegrees", "Winkel (°)", text,
		[]string{"Winkel", "Weichenwinkel", "Angle"})
	addDirectionTrackField(fields, text)
	addDecimalTrackField(fields, "frogAngleDegrees", "Herzstückwinkel (°)", text,
		[]string{"Herzstückwinkel", "Herzstueckwinkel", "Frog angle"})
	addTextTrackField(fields, "sleeperType", "Schwellenart", text,
		[]string{"Schwellenart", "Schwellen", "Sleeper type"})
	addDecimalTrackField(fields, "railHeightMm", "Profilhöhe (mm)", text,
		[]string{"Profilhöhe", "Profilhoehe", "Rail height"})
	addBooleanTrackField(fields, "roadbed", "Bettung", text,
		[]string{"Bettung", "Gleisbettung", "Roadbed"})
	addIntegerTrackField(fields, "connectionCount", "Anzahl Anschlüsse", text,
		[]string{"Anzahl Anschlüsse", "Anzahl Anschluesse", "Connections"})
	addBooleanTrackField(fields, "digitalReady", "Digitaltauglich", text,
		[]string{"Digitaltauglich", "Digital geeignet", "Digital ready"})
	return fields
}

func addTextTrackField(
	fields map[string]ArticleSearchField,
	key, label, text string,
	labels []string,
) {
	value := normalizeWhitespace(trackLabeledValue(text, labels))
	value = strings.Trim(value, " -:;,.\t")
	if value == "" || len(value) > 90 {
		return
	}
	fields[key] = ArticleSearchField{Label: label, Value: value, Confidence: 78}
}

func addDecimalTrackField(
	fields map[string]ArticleSearchField,
	key, label, text string,
	labels []string,
) {
	raw := trackLabeledValue(text, labels)
	match := trackNumberPattern.FindString(raw)
	if match == "" {
		return
	}
	value, err := strconv.ParseFloat(strings.ReplaceAll(match, ",", "."), 64)
	if err != nil || value < 0 {
		return
	}
	fields[key] = ArticleSearchField{
		Label: label, Value: strconv.FormatFloat(value, 'f', -1, 64), Confidence: 82,
	}
}

func addIntegerTrackField(
	fields map[string]ArticleSearchField,
	key, label, text string,
	labels []string,
) {
	raw := trackLabeledValue(text, labels)
	match := regexp.MustCompile(`\d+`).FindString(raw)
	if match == "" {
		return
	}
	value, err := strconv.Atoi(match)
	if err != nil || value < 0 {
		return
	}
	fields[key] = ArticleSearchField{Label: label, Value: strconv.Itoa(value), Confidence: 82}
}

func addDirectionTrackField(fields map[string]ArticleSearchField, text string) {
	value := strings.ToLower(trackLabeledValue(text, []string{"Richtung", "Direction"}))
	normalized := ""
	switch {
	case strings.HasPrefix(value, "links"), strings.HasPrefix(value, "left"):
		normalized = "left"
	case strings.HasPrefix(value, "rechts"), strings.HasPrefix(value, "right"):
		normalized = "right"
	case strings.HasPrefix(value, "symmetrisch"), strings.HasPrefix(value, "symmetric"):
		normalized = "symmetric"
	}
	if normalized != "" {
		fields["direction"] = ArticleSearchField{Label: "Richtung", Value: normalized, Confidence: 82}
	}
}

func addBooleanTrackField(
	fields map[string]ArticleSearchField,
	key, label, text string,
	labels []string,
) {
	value := strings.ToLower(trackLabeledValue(text, labels))
	normalized := ""
	switch {
	case strings.HasPrefix(value, "ja"), strings.HasPrefix(value, "yes"), strings.HasPrefix(value, "true"):
		normalized = "true"
	case strings.HasPrefix(value, "nein"), strings.HasPrefix(value, "no"), strings.HasPrefix(value, "false"):
		normalized = "false"
	}
	if normalized != "" {
		fields[key] = ArticleSearchField{Label: label, Value: normalized, Confidence: 82}
	}
}

func trackLabeledValue(text string, labels []string) string {
	text = repairMojibake(text)
	lower := strings.ToLower(text)
	for _, label := range labels {
		lowerLabel := strings.ToLower(label)
		for offset := 0; offset < len(lower); {
			relative := strings.Index(lower[offset:], lowerLabel)
			if relative < 0 {
				break
			}
			index := offset + relative
			if index > 0 && isTrackLabelCharacter(lower[index-1]) {
				offset = index + len(lowerLabel)
				continue
			}
			start := index + len(label)
			for start < len(text) && strings.ContainsRune(" \t\u00a0:-", rune(text[start])) {
				start++
			}
			end := len(text)
			for _, nextLabel := range trackDetailLabels {
				nextLower := strings.ToLower(nextLabel)
				searchOffset := start
				for searchOffset < len(lower) {
					nextRelative := strings.Index(lower[searchOffset:], nextLower)
					if nextRelative < 0 {
						break
					}
					next := searchOffset + nextRelative
					if next > start && !isTrackLabelCharacter(lower[next-1]) {
						if next < end {
							end = next
						}
						break
					}
					searchOffset = next + len(nextLower)
				}
			}
			value := normalizeWhitespace(text[start:end])
			value = strings.Trim(value, " -:;,.\t")
			if value != "" && len(value) <= 90 {
				return value
			}
			offset = index + len(lowerLabel)
		}
	}
	return ""
}

func isTrackLabelCharacter(value byte) bool {
	return value >= 'a' && value <= 'z'
}
