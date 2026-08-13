package application

import (
	"net/url"
	"regexp"
	"strings"
)

func parseDuckDuckGoResults(body string, input ArticleSearchInput, source string) []ArticleSearchResult {
	blocks := resultBlockPattern.FindAllString(body, 12)
	results := []ArticleSearchResult{}
	for rank, block := range blocks {
		linkMatch := resultLinkPattern.FindStringSubmatch(block)
		if len(linkMatch) < 3 {
			continue
		}
		resultURL := decodeDuckDuckGoURL(linkMatch[1])
		title := cleanHTML(linkMatch[2])
		snippet := ""
		if snippetMatch := snippetPattern.FindStringSubmatch(block); len(snippetMatch) > 0 {
			snippet = cleanHTML(strings.Join(snippetMatch[1:], " "))
		}
		if title == "" || resultURL == "" {
			continue
		}
		fields := buildArticleFields(input, title, resultURL, snippet)
		score := scoreArticleResult(input, title, resultURL, snippet, fields)
		score += duckDuckGoRankBonus(rank)
		results = append(results, ArticleSearchResult{
			Source:  source,
			Title:   title,
			URL:     resultURL,
			Snippet: snippet,
			Score:   score,
			Fields:  fields,
		})
	}
	return results
}

func duckDuckGoRankBonus(rank int) int {
	bonus := 48 - rank*6
	if bonus < 0 {
		return 0
	}
	return bonus
}

func buildArticleFields(input ArticleSearchInput, title, resultURL, snippet string) map[string]ArticleSearchField {
	cleanName := cleanArticleName(title, resultURL)
	fields := map[string]ArticleSearchField{
		"name": {
			Label:      "Bezeichnung",
			Value:      cleanName,
			Confidence: 60,
		},
		"articleSourceUrl": {
			Label:      "Quelle",
			Value:      resultURL,
			Confidence: 100,
		},
	}
	combined := repairMojibake(title + " " + snippet + " " + resultURL)
	combinedLower := strings.ToLower(combined)
	if containsManufacturerTerm(input, combinedLower) {
		fields["manufacturer"] = ArticleSearchField{Label: "Hersteller", Value: input.Manufacturer, Confidence: 80}
	}
	if input.ArticleNumber != "" && strings.Contains(strings.ToLower(combined), strings.ToLower(input.ArticleNumber)) {
		fields["articleNumber"] = ArticleSearchField{Label: "Artikel-Nr.", Value: input.ArticleNumber, Confidence: 90}
	}
	if input.ArticleNumber == "" {
		if value := labeledValue(combined, []string{"Art.-Nr.", "Artikel-Nr.", "Artikelnummer"}); value != "" {
			fields["articleNumber"] = ArticleSearchField{Label: "Artikel-Nr.", Value: value, Confidence: 78}
		}
	}
	if input.Gauge != "" && strings.Contains(strings.ToLower(combined), strings.ToLower(input.Gauge)) {
		fields["gauge"] = ArticleSearchField{Label: "Spurweite", Value: input.Gauge, Confidence: 80}
	}
	if description := bestArticleDescription(input, cleanName, snippet, resultURL); description != "" {
		fields["description"] = ArticleSearchField{Label: "Beschreibung", Value: description, Confidence: 65}
	}
	if description := catalogDescription(snippet, resultURL); description != "" {
		if existing, ok := fields["description"]; !ok || len(description) > len(existing.Value) {
			fields["description"] = ArticleSearchField{Label: "Beschreibung", Value: description, Confidence: 72}
		}
	}
	if value := firstRegexValue(eanPattern, combined); value != "" && value != input.ArticleNumber {
		fields["ean"] = ArticleSearchField{Label: "EAN-Nr.", Value: value, Confidence: 60}
	}
	if value := firstRegexValue(epochPattern, combined); value != "" {
		fields["epoch"] = ArticleSearchField{Label: "Epoche", Value: strings.ToUpper(value), Confidence: 60}
	}
	if value := firstRegexValue(railwayPattern, combined); value != "" {
		fields["railwayCompany"] = ArticleSearchField{Label: "Bahngesellschaft", Value: strings.ToUpper(value), Confidence: 55}
	}
	if value := extractPrice(combined); value != "" {
		fields["listPrice"] = ArticleSearchField{Label: "Listenpreis", Value: value, Confidence: 55}
	}
	if value := labeledValue(combined, []string{"Bahn-Gesellschaft", "Bahngesellschaft"}); value != "" {
		fields["railwayCompany"] = ArticleSearchField{Label: "Bahngesellschaft", Value: strings.ToUpper(value), Confidence: 70}
	}
	if value := labeledValue(combined, []string{"Stromsystem"}); value != "" {
		fields["powerPickup"] = ArticleSearchField{Label: "Stromsystem", Value: normalizeWhitespace(value), Confidence: 62}
	}
	if value := extractLengthMM(combined); value != "" {
		fields["lengthMm"] = ArticleSearchField{Label: "Länge (mm)", Value: value, Confidence: 62}
	}
	if value := firstRegexValue(weightPattern, combined); value != "" {
		fields["weightG"] = ArticleSearchField{Label: "Gewicht (g)", Value: strings.ReplaceAll(value, ",", "."), Confidence: 55}
	}
	if value := firstRegexValue(tractionTirePattern, combined); value != "" {
		fields["tractionTireCount"] = ArticleSearchField{Label: "Anzahl Haftreifen", Value: value, Confidence: 58}
	}
	if value := extractAdapterInfo(combined); value != "" {
		fields["adapter"] = ArticleSearchField{Label: "Schnittstelle / Adapter", Value: normalizeWhitespace(value), Confidence: 60}
	}
	if value := labeledValue(combined, []string{"Motor"}); value != "" {
		fields["driveDescription"] = ArticleSearchField{Label: "Antrieb Beschreibung", Value: normalizeWhitespace(value), Confidence: 55}
	}
	if value := firstRegexValue(powerPattern, combined); value != "" {
		fields["powerPickup"] = ArticleSearchField{Label: "Stromsystem", Value: normalizeWhitespace(value), Confidence: 50}
	}
	if digitalPositivePattern.MatchString(combined) {
		fields["digital"] = ArticleSearchField{Label: "Digital", Value: "Ja", Confidence: 48}
	}
	if soundDescription := extractSoundDescription(combined); soundDescription != "" {
		fields["soundGeneratorEnabled"] = ArticleSearchField{Label: "Soundgenerator", Value: "Ja", Confidence: 48}
		fields["soundGeneratorDescription"] = ArticleSearchField{Label: "Soundgenerator Beschreibung", Value: normalizeWhitespace(soundDescription), Confidence: 55}
	} else if hasExplicitSoundGenerator(combinedLower) {
		fields["soundGeneratorEnabled"] = ArticleSearchField{Label: "Soundgenerator", Value: "Ja", Confidence: 38}
	}
	if lightDescription := extractHeadlightDescription(combined); lightDescription != "" {
		fields["headlightsEnabled"] = ArticleSearchField{Label: "Fahrlicht", Value: "Ja", Confidence: 42}
		fields["headlightsDescription"] = ArticleSearchField{Label: "Fahrlicht Beschreibung", Value: normalizeWhitespace(lightDescription), Confidence: 55}
	} else if hasExplicitHeadlight(combinedLower) {
		fields["headlightsEnabled"] = ArticleSearchField{Label: "Fahrlicht", Value: "Ja", Confidence: 36}
	}
	if lightingDescription := extractLightingDescription(combined); lightingDescription != "" {
		fields["lightingEnabled"] = ArticleSearchField{Label: "Beleuchtung", Value: "Ja", Confidence: 36}
		fields["lightingDescription"] = ArticleSearchField{Label: "Beleuchtung Beschreibung", Value: normalizeWhitespace(lightingDescription), Confidence: 52}
	} else if hasExplicitInteriorLighting(combinedLower) {
		fields["lightingEnabled"] = ArticleSearchField{Label: "Beleuchtung", Value: "Ja", Confidence: 34}
	}
	for key, field := range buildTrackArticleFields(input, combined) {
		if existing, ok := fields[key]; !ok || field.Confidence > existing.Confidence {
			fields[key] = field
		}
	}
	return fields
}

func scoreArticleResult(input ArticleSearchInput, title, resultURL, snippet string, fields map[string]ArticleSearchField) int {
	haystack := strings.ToLower(title + " " + resultURL + " " + snippet)
	score := len(fields) * 10
	manufacturer := strings.ToLower(strings.TrimSpace(input.Manufacturer))
	articleNumber := strings.ToLower(strings.TrimSpace(input.ArticleNumber))
	gauge := strings.ToLower(strings.TrimSpace(input.Gauge))
	name := strings.ToLower(strings.TrimSpace(input.Name))

	if manufacturer != "" && containsManufacturerTerm(input, haystack) {
		score += 35
	}
	hasArticleNumber := articleNumber != "" && strings.Contains(haystack, articleNumber)
	if hasArticleNumber {
		score += 105
	} else if articleNumber != "" {
		score -= 165
	}
	if gauge != "" && containsGaugeToken(haystack, gauge) {
		score += 35
	}
	if name != "" && strings.Contains(haystack, name) {
		score += 30
	}
	score += articleNameTokenScore(name, haystack)

	if isManufacturerPreferredURL(input, resultURL) {
		if articleNumber == "" || hasArticleNumber {
			score += 140
		} else {
			score += 15
		}
	} else if isCatalogURL(resultURL) && (articleNumber == "" || hasArticleNumber) {
		score += 70
	} else if isDealerURL(resultURL) && (articleNumber == "" || hasArticleNumber) {
		score += 35
	} else if strings.Contains(haystack, manufacturerDomainToken(input.Manufacturer)) {
		score += 20
	}
	resultDomain := domainFromURL(resultURL)
	if isMarketplaceURL(resultURL) {
		score -= 30
	}
	if isWikiDomain(resultDomain) {
		score -= 35
	}
	if isBlockedManufacturerDomain(resultDomain) {
		score -= 45
	}
	ean := strings.ToLower(strings.TrimSpace(input.Fields["ean"]))
	if ean != "" && strings.Contains(haystack, ean) {
		score += 160
		if field, ok := fields["ean"]; ok && strings.EqualFold(strings.TrimSpace(field.Value), ean) {
			score += 120
		}
	}
	for _, value := range input.Fields {
		value = strings.ToLower(strings.TrimSpace(value))
		if value != "" && strings.Contains(haystack, value) {
			score += 8
		}
	}
	return score
}

func containsManufacturerTerm(input ArticleSearchInput, haystack string) bool {
	haystack = strings.ToLower(haystack)
	for _, term := range manufacturerSearchTerms(input) {
		if strings.Contains(haystack, strings.ToLower(term)) {
			return true
		}
	}
	return false
}

func manufacturerSearchTerms(input ArticleSearchInput) []string {
	terms := []string{input.Manufacturer}
	terms = append(terms, input.ManufacturerAliases...)
	return uniqueNonEmpty(terms)
}

func articleNameTokenScore(name, haystack string) int {
	if name == "" {
		return 0
	}
	score := 0
	for _, token := range uniqueSearchTokens(name) {
		if strings.Contains(haystack, token) {
			score += 10
		}
	}
	if score > 40 {
		return 40
	}
	return score
}

func uniqueSearchTokens(value string) []string {
	tokens := []string{}
	for _, token := range strings.FieldsFunc(strings.ToLower(value), func(r rune) bool {
		return (r < 'a' || r > 'z') && (r < '0' || r > '9') && r != 'ä' && r != 'ö' && r != 'ü' && r != 'ß'
	}) {
		if len(token) >= 3 {
			tokens = append(tokens, token)
		}
	}
	return uniqueNonEmpty(tokens)
}

func containsGaugeToken(haystack, gauge string) bool {
	if gauge == "" {
		return false
	}
	return regexp.MustCompile(`(?i)(^|[^a-z0-9])` + regexp.QuoteMeta(gauge) + `([^a-z0-9]|$)`).MatchString(haystack)
}

func resolveURL(baseURL, raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" || strings.HasPrefix(raw, "data:") {
		return ""
	}
	parsed, err := url.Parse(raw)
	if err == nil && parsed.Scheme != "" {
		return parsed.String()
	}
	base, err := url.Parse(baseURL)
	if err != nil {
		return ""
	}
	relative, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	return base.ResolveReference(relative).String()
}

func isManufacturerPreferredURL(input ArticleSearchInput, resultURL string) bool {
	resultURL = strings.ToLower(resultURL)
	for _, domain := range preferredManufacturerDomains(input) {
		if strings.Contains(resultURL, domain) {
			return true
		}
	}
	return false
}

func isCatalogURL(resultURL string) bool {
	domain := domainFromURL(resultURL)
	for _, catalog := range catalogArticleDomains {
		if domain == catalog || strings.HasSuffix(domain, "."+catalog) {
			return true
		}
	}
	return false
}

func isDealerURL(resultURL string) bool {
	domain := domainFromURL(resultURL)
	for _, dealer := range dealerArticleDomains {
		if domain == dealer || strings.HasSuffix(domain, "."+dealer) {
			return true
		}
	}
	return false
}

func isMarketplaceURL(resultURL string) bool {
	parsed, err := url.Parse(resultURL)
	if err != nil {
		return false
	}
	host := strings.TrimPrefix(strings.ToLower(parsed.Host), "www.")
	marketplaces := []string{"amazon.", "ebay.", "idealo.", "kaufland.", "kleinanzeigen."}
	for _, marketplace := range marketplaces {
		if strings.Contains(host, marketplace) {
			return true
		}
	}
	return false
}

func preferredManufacturerDomains(input ArticleSearchInput) []string {
	domains := uniqueDomains(input.PreferredDomains)
	manufacturer := strings.ToLower(strings.TrimSpace(input.Manufacturer))
	for key, staticDomains := range manufacturerDomains {
		if manufacturer == "" || !strings.Contains(manufacturer, key) {
			continue
		}
		domains = append(domains, staticDomains...)
	}
	return uniqueDomains(domains)
}

func uniqueDomains(values []string) []string {
	out := []string{}
	seen := map[string]bool{}
	for _, value := range values {
		domain := domainFromURL(value)
		if domain == "" || isWikiDomain(domain) || isBlockedManufacturerDomain(domain) || seen[domain] {
			continue
		}
		seen[domain] = true
		out = append(out, domain)
	}
	return out
}

func domainFromURL(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	if value == "" {
		return ""
	}
	if !strings.Contains(value, "://") {
		value = "https://" + value
	}
	parsed, err := url.Parse(value)
	if err != nil {
		return ""
	}
	host := strings.TrimPrefix(strings.TrimSpace(parsed.Hostname()), "www.")
	return strings.Trim(host, ".")
}

func isWikiDomain(domain string) bool {
	domain = strings.TrimPrefix(strings.ToLower(strings.TrimSpace(domain)), "www.")
	return domain == "modellbau-wiki.de" || strings.HasSuffix(domain, ".wikipedia.org") || strings.HasSuffix(domain, ".wikimedia.org")
}

func isBlockedManufacturerDomain(domain string) bool {
	domain = strings.TrimPrefix(strings.ToLower(strings.TrimSpace(domain)), "www.")
	blockedDomains := []string{
		"altemodellbahnen.de",
		"berliner-tt-bahner.de",
		"eisenbahnfreunde-sonneberg.de",
		"facebook.com",
		"maetrix.net",
		"modellbahnarchiv.de",
		"modellbahninfo.org",
		"radiomuseum.org",
		"spurnull-magazin.de",
		"web.archive.org",
	}
	for _, blocked := range blockedDomains {
		if domain == blocked || strings.HasSuffix(domain, "."+blocked) {
			return true
		}
	}
	return strings.HasPrefix(domain, "forum.")
}

func manufacturerDomainToken(manufacturer string) string {
	manufacturer = strings.ToLower(strings.TrimSpace(manufacturer))
	for key := range manufacturerDomains {
		if strings.Contains(manufacturer, key) {
			return key
		}
	}
	return manufacturer
}

func normalizeWhitespace(value string) string {
	return strings.Join(strings.Fields(value), " ")
}
