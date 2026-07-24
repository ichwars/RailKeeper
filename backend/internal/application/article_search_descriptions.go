package application

import (
	"html"
	"net/url"
	"regexp"
	"strings"
	"unicode/utf8"
)

func visibleArticleText(value string) string {
	value = regexp.MustCompile(`(?is)</(?:tr|li|p|div|h[1-6]|dd|dt)>`).ReplaceAllString(value, ". ")
	value = scriptStylePattern.ReplaceAllString(value, " ")
	return cleanHTML(value)
}

func cleanArticleName(title, resultURL string) string {
	value := cleanHTML(title)
	if isCatalogURL(resultURL) {
		value = cleanCatalogArticleName(value)
	}
	sourceParts := []string{
		" - " + sourceDisplayName(resultURL),
		" | " + sourceDisplayName(resultURL),
		" - PIKO Spielwaren GmbH Webshop",
		" - PIKO Webshop",
		" - Amazon.de",
		" - eBay",
		" - idealo",
	}
	for _, part := range sourceParts {
		if part != " - " && part != " | " && strings.HasSuffix(strings.ToLower(value), strings.ToLower(part)) {
			value = strings.TrimSpace(value[:len(value)-len(part)])
		}
	}
	return strings.Trim(value, " -|")
}

func cleanCatalogArticleName(value string) string {
	value = regexp.MustCompile(`(?i)^\s*\S+\s+\d{3,8}\s+`).ReplaceAllString(value, "")
	value = regexp.MustCompile(`(?i)\s+(Diesellok|E-Lok|Elektrolok|Dampflok|Triebwagen|Dieseltriebwagen|Wagen|G?terwagen|Personenwagen)\s+[A-Z0-9]+\s+Modellbahn\s+Katalog\s*$`).ReplaceAllString(value, "")
	return strings.Trim(value, " -:;,.\t")
}

func sourceDisplayName(resultURL string) string {
	parsed, err := url.Parse(resultURL)
	if err != nil || parsed.Host == "" {
		return "Quelle"
	}
	host := strings.TrimPrefix(strings.ToLower(parsed.Host), "www.")
	parts := strings.Split(host, ".")
	if len(parts) == 0 || parts[0] == "" {
		return host
	}
	return parts[0]
}

func bestArticleDescription(input ArticleSearchInput, name, text, resultURL string) string {
	text = normalizeWhitespace(text)
	if len(text) < 20 {
		return ""
	}
	if preferred := preferredArticleDescription(text); preferred != "" {
		return preferred
	}
	candidates := splitDescriptionCandidates(text)
	best := ""
	bestScore := -1
	for _, candidate := range candidates {
		candidate = normalizeWhitespace(candidate)
		if !looksLikeHumanDescription(candidate) {
			continue
		}
		score := 0
		lower := strings.ToLower(candidate)
		for _, token := range uniqueNonEmpty([]string{input.ArticleNumber, input.Name, input.Gauge, input.Manufacturer, "neuheit", "druckvariante", "epoche", "dr", "db"}) {
			if strings.Contains(lower, strings.ToLower(token)) {
				score += 8
			}
		}
		if strings.Contains(strings.ToLower(resultURL), "piko") || strings.Contains(strings.ToLower(resultURL), "roco") || strings.Contains(strings.ToLower(resultURL), "tillig") {
			score += 4
		}
		if len(candidate) > 60 && len(candidate) < 280 {
			score += 3
		}
		if score > bestScore {
			bestScore = score
			best = candidate
		}
	}
	if best == "" {
		return ""
	}
	if len(best) > 320 {
		best = best[:320]
	}
	return strings.TrimSpace(best)
}

func preferredArticleDescription(text string) string {
	text = normalizeWhitespace(repairMojibake(text))
	lower := strings.ToLower(text)
	start := -1
	for _, marker := range []string{"neuheit ", "druckvariante "} {
		if index := strings.Index(lower, marker); index >= 0 && (start < 0 || index < start) {
			start = index
		}
	}
	if start < 0 {
		return ""
	}
	candidate := text[start:]
	candidateLower := strings.ToLower(candidate)
	end := len(candidate)
	for _, marker := range []string{
		" maß ", " mass ", " länge ", " laenge ", " digitale schnittstelle",
		" lichtwechsel", " fahrlicht", " soundgenerator", " sounddecoder",
		" downloads", " bedienungsanleitung", " altersempfehlung", " ean ",
	} {
		if index := strings.Index(candidateLower, marker); index > 30 && index < end {
			end = index
		}
	}
	if period := strings.Index(candidate, "."); period > 40 && period+1 < end {
		end = period + 1
	}
	candidate = strings.TrimSpace(candidate[:end])
	candidate = strings.Trim(candidate, " -:;,.")
	if !looksLikeHumanDescription(candidate) {
		return ""
	}
	return candidate
}

func splitDescriptionCandidates(text string) []string {
	parts := regexp.MustCompile(`[.!?]\s+|\s{2,}`).Split(text, -1)
	out := []string{}
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	if len(out) == 0 && text != "" {
		out = append(out, text)
	}
	return out
}

func looksLikeHumanDescription(value string) bool {
	value = strings.TrimSpace(value)
	if len(value) < 20 || len(value) > 600 {
		return false
	}
	lower := strings.ToLower(value)
	badTokens := []string{"google_analytics", "cookie", "mandatory", "preferences", "statistics", "marketing", "function", "const ", "new map", "document.", "window.", "{", "};", "class ", "anzeigen zu zeigen", "personalisierte anzeigen", "absicht ist", "menü", "menue", "menu", "sprunggröße", "sprunggroesse", "wählen sie", "waehlen sie", "downloads", "bedienungsanleitung", "altersempfehlung"}
	for _, token := range badTokens {
		if strings.Contains(lower, token) {
			return false
		}
	}
	technicalStarts := []string{"digitale schnittstelle", "schnittstelle", "laenge", "mass", "gewicht", "haftreifen", "ean", "artikelnummer", "artikel-nr", "beleuchtung", "fahrlicht", "lichtwechsel", "soundgenerator", "sound", "altersempfehlung", "downloads", "bedienungsanleitung"}
	for _, token := range technicalStarts {
		if strings.HasPrefix(lower, token) {
			return false
		}
	}
	return true
}

func firstRegexValue(pattern *regexp.Regexp, value string) string {
	matches := pattern.FindStringSubmatch(value)
	if len(matches) < 2 {
		return ""
	}
	return strings.TrimSpace(matches[1])
}

func cleanHTML(value string) string {
	value = tagPattern.ReplaceAllString(value, " ")
	value = html.UnescapeString(value)
	value = repairMojibake(value)
	value = strings.Join(strings.Fields(value), " ")
	return strings.TrimSpace(value)
}

func repairMojibake(value string) string {
	if !strings.ContainsAny(value, "ÃÂâ") {
		return value
	}
	bytes := make([]byte, 0, len(value))
	for _, char := range value {
		if char > 255 {
			return value
		}
		bytes = append(bytes, byte(char))
	}
	if !utf8.Valid(bytes) {
		return value
	}
	return string(bytes)
}

func decodeDuckDuckGoURL(value string) string {
	value = html.UnescapeString(value)
	parsed, err := url.Parse(value)
	if err == nil {
		if raw := parsed.Query().Get("uddg"); raw != "" {
			if decoded, err := url.QueryUnescape(raw); err == nil {
				return decoded
			}
			return raw
		}
		if parsed.Scheme != "" {
			return parsed.String()
		}
	}
	return strings.TrimSpace(value)
}
