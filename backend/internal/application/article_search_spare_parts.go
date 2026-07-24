package application

import (
	"bytes"
	"html"
	"regexp"
	"strings"
)

func ArticleSparePartsFromDocumentData(data []byte, articleNumber, source string) []ArticleSearchSparePart {
	if len(data) == 0 {
		return nil
	}
	text := ""
	if bytes.HasPrefix(bytes.TrimSpace(data), []byte("%PDF")) {
		text = extractPDFTextWithOCRFallback(data, articleNumber)
	} else {
		raw := string(data)
		if strings.Contains(strings.ToLower(raw), "<html") || strings.Contains(strings.ToLower(raw), "<body") {
			text = visibleArticleLines(raw)
		} else {
			text = normalizeWhitespacePreservingLines(raw)
		}
	}
	return articleSparePartsFromDocumentText(text, articleNumber, source)
}

func articleSparePartDocumentRows(text string) []string {
	text = repairMojibake(text)
	rows := []string{}
	previousDescription := ""
	for _, line := range regexp.MustCompile(`[\n\r]+`).Split(text, -1) {
		line = normalizeWhitespace(line)
		if line == "" {
			continue
		}
		match := sparePartArticlePattern.FindStringIndex(line)
		if match != nil {
			if match[0] == 0 && previousDescription != "" {
				rows = append(rows, previousDescription+" "+line)
			} else {
				rows = append(rows, line)
			}
			previousDescription = ""
			continue
		}
		rows = append(rows, line)
		if looksLikeSparePartDescriptionFragment(line) {
			previousDescription = line
		} else {
			previousDescription = ""
		}
	}
	return rows
}

func looksLikeSparePartDescriptionFragment(line string) bool {
	line = strings.TrimSpace(line)
	if len([]rune(line)) < 4 || !regexp.MustCompile(`[A-Za-zÄÖÜäöüß]`).MatchString(line) {
		return false
	}
	lower := strings.ToLower(line)
	return !containsAny(lower, []string{
		"ersatzteile", "spare parts", "pièces de rechange", "náhradní díly", "bezeichnung / description",
		"et-nr", "spare part n", "preisgruppe", "price category", "bitte immer", "please order",
		"bestell-nr", "bestell nr", "bestellnummer", "art.-nr", "art nr", "artikel-nr", "artikel nr",
		"benennung", "item number", "description", "pos.", "position",
	})
}

func articleSparePartFromRow(row, pageURL string) (ArticleSearchSparePart, bool) {
	text := cleanHTML(row)
	if len(text) < 6 {
		return ArticleSearchSparePart{}, false
	}
	lower := strings.ToLower(text)
	price := extractPrice(text)
	if price == "" && !containsAny(lower, []string{"ersatzteil", "spare", "kuppl", "motor", "radsatz", "puffer", "decoder", "leiterplatte", "geh\u00e4use", "gehaeuse", "schraube", "stromabnehmer", "getriebe", "reifen", "haftreifen", "lautsprecher", "loudspeaker", "coupler", "speaker"}) {
		return ArticleSearchSparePart{}, false
	}
	match := sparePartArticlePattern.FindStringSubmatch(text)
	if len(match) < 2 {
		return ArticleSearchSparePart{}, false
	}
	number := strings.TrimSpace(match[1])
	description := strings.TrimSpace(strings.Replace(text, match[0], " ", 1))
	if price != "" {
		description = strings.TrimSpace(strings.Replace(description, price, " ", 1))
	}
	description = cleanArticleSparePartDescription(description)
	if !looksLikeRealArticleSparePart(description, price) {
		return ArticleSearchSparePart{}, false
	}
	if len(description) > 180 {
		description = strings.TrimSpace(description[:180])
	}
	link := ""
	if linkMatch := linkHrefAttrPattern.FindStringSubmatch(row); len(linkMatch) >= 2 {
		link = resolveURL(pageURL, html.UnescapeString(linkMatch[1]))
	}
	if description == "" && link == "" && price == "" {
		return ArticleSearchSparePart{}, false
	}
	return ArticleSearchSparePart{ArticleNumber: number, Description: description, Price: price, URL: link, Source: sourceDisplayName(pageURL)}, true
}

func articleSparePartFromConfirmedDocumentRow(row, pageURL string) (ArticleSearchSparePart, bool) {
	text := cleanHTML(row)
	if len(text) < 6 {
		return ArticleSearchSparePart{}, false
	}
	match := sparePartArticlePattern.FindStringSubmatch(text)
	if len(match) < 2 {
		return ArticleSearchSparePart{}, false
	}
	number := strings.TrimSpace(match[1])
	if strings.Count(number, "-") > 1 || strings.Contains(number, "-90") {
		return ArticleSearchSparePart{}, false
	}
	description := strings.TrimSpace(strings.Replace(text, match[0], " ", 1))
	price := extractPrice(text)
	if price != "" {
		description = strings.TrimSpace(strings.Replace(description, price, " ", 1))
	}
	description = cleanArticleSparePartDescription(description)
	description = cleanConfirmedSparePartDescription(description)
	if !looksLikeConfirmedDocumentSparePart(description) {
		return ArticleSearchSparePart{}, false
	}
	if len(description) > 180 {
		description = strings.TrimSpace(description[:180])
	}
	return ArticleSearchSparePart{ArticleNumber: number, Description: description, Price: price, URL: pageURL, Source: sourceDisplayName(pageURL)}, true
}

func cleanArticleSparePartDescription(description string) string {
	description = strings.Trim(description, " -:;,.\t")
	replacements := []*regexp.Regexp{
		regexp.MustCompile(`(?i)^(GER|DE|ENG|EN)\s*[:/-]\s*`),
		regexp.MustCompile(`(?i)\b(ersatzteil|spare part|artikel|artikelnummer|nummer|number|no\.?|item number|item no\.?|art\.?\s*nr\.?|nr\.?)\s*[:#-]*\s*`),
		regexp.MustCompile(`(?i)\b(preis|price)\s*[:#-]?\s*\d+(?:[,.]\d{1,2})?\s*(\x{20ac}|EUR)?`),
		regexp.MustCompile(`(?i)\d+(?:[,.]\d{1,2})?\s*(\x{20ac}|EUR)`),
		regexp.MustCompile(`(?i)\*?\s*\b(in den warenkorb|zum warenkorb hinzuf(?:\x{fc}|u|ue)gen|in den einkaufswagen|add to cart|add to shopping cart|add to basket|add to bag|ajouter au panier|anadir al carrito|aggiungi al carrello|in winkelwagen|toevoegen aan winkelwagen)\b`),
	}
	for _, pattern := range replacements {
		description = pattern.ReplaceAllString(description, "")
	}
	description = strings.Join(strings.Fields(description), " ")
	return strings.Trim(description, " -:;,.|")
}

func looksLikeRealArticleSparePart(description, price string) bool {
	lower := strings.ToLower(description)
	if len([]rune(strings.TrimSpace(description))) < 3 {
		return false
	}
	if containsAny(lower, []string{"bedienungsanl", "bedienungsanleitung", "ersatzteilliste", "ersatzteilblatt", "spare parts list", "manual", "download", "katalog", "catalog", "et-blatt", "explosionszeichnung", "serviceblatt"}) {
		return false
	}
	if price != "" {
		return true
	}
	return containsAny(lower, []string{"kuppl", "lautsprecher", "decoder", "reifen", "haftreifen", "radsatz", "motor", "puffer", "schraube", "stromabnehmer", "getriebe", "geh\xC3\xA4use", "gehaeuse", "leiterplatte", "feder", "achse", "traction tire", "loudspeaker", "coupler", "speaker"})
}

func cleanConfirmedSparePartDescription(description string) string {
	description = regexp.MustCompile(`(?i)\b(?:PG\*?|Preisgruppe|price category)\b\s*[:#-]?\s*\d{1,3}\s*$`).ReplaceAllString(description, "")
	description = stripTrailingSparePartPriceGroup(description)
	description = regexp.MustCompile(`(?i)^\s*\d{1,3}\s+`).ReplaceAllString(description, "")
	description = regexp.MustCompile(`(?i)\b(?:ET-Nr\.?|spare part N.?|Bezeichnung / Description|Bestell-Nr\.?|Bestellnummer|Art\.-Nr\.?|Artikel-Nr\.?|item number|description|Benennung)\b`).ReplaceAllString(description, "")
	description = strings.Join(strings.Fields(description), " ")
	return strings.Trim(description, " -:;,.|")
}

func stripTrailingSparePartPriceGroup(description string) string {
	description = strings.TrimSpace(description)
	if regexp.MustCompile(`(?i)(?:\bx|×)\s+\d{1,3}$`).MatchString(description) {
		return description
	}
	if regexp.MustCompile(`(?i)\b(?:m\d+(?:[,.]\d+)?|gewinde)\s*x\s*\d{1,3}$`).MatchString(description) {
		return description
	}
	return regexp.MustCompile(`\s+\d{1,3}\s*$`).ReplaceAllString(description, "")
}

func looksLikeConfirmedDocumentSparePart(description string) bool {
	description = strings.TrimSpace(description)
	if len([]rune(description)) < 3 || len([]rune(description)) > 180 {
		return false
	}
	lower := strings.ToLower(description)
	badTokens := []string{
		"bedienungsanl", "bedienungsanleitung", "instructions for use", "manuel d", "návod", "ersatzteilliste", "spare parts list",
		"piko spielwaren", "lutherstraße", "lutherstrasse", "germany", "www.", "http", "tel.", "telefon", "sicherheitshinweise",
		"please note", "hinweis", "aviso", "uwaga", "nota:", "bei ersatzteilanforderung", "please order", "vollständige",
		"preisgruppe", "price category", "nicht enthalten", "not included", "non compris", "neobsahuje",
		"demontage", "disassembly", "einbau", "installing", "installation", "ölen sie", "oelen sie", "if used frequently",
	}
	if containsAny(lower, badTokens) {
		return false
	}
	return regexp.MustCompile(`[A-Za-zÄÖÜäöüß]`).MatchString(description)
}

func looksLikeSparePartsDocumentText(text, articleNumber string) bool {
	if !documentTextMatchesArticleNumber(text, articleNumber) {
		return false
	}
	lower := strings.ToLower(text)
	return containsAny(lower, []string{
		"ersatzteile", "ersatzteilliste", "spare parts", "pièces de rechange", "náhradní díly", "et-nr", "spare part n", "bezeichnung / description",
	})
}

func sparePartsDocumentSection(text string) string {
	lower := strings.ToLower(text)
	start := -1
	for _, marker := range []string{"ersatzteile", "spare parts", "pièces de rechange", "náhradní díly", "bezeichnung / description"} {
		if index := strings.Index(lower, marker); index >= 0 && (start < 0 || index < start) {
			start = index
		}
	}
	if start < 0 {
		return text
	}
	section := text[start:]
	sectionLower := strings.ToLower(section)
	for _, marker := range []string{"\nsoundeinbau", "\ninstalling sound", "\ndecodereinbau", "\ninstalling decoder", "\nhaftreifenwechsel", "\nchange the traction tires"} {
		if index := strings.Index(sectionLower, marker); index > 0 {
			section = section[:index]
			sectionLower = sectionLower[:index]
		}
	}
	return section
}

func looksLikeSparePartDocument(document ArticleSearchDocument) bool {
	lower := strings.ToLower(document.Kind + " " + document.Title + " " + document.URL)
	return containsAny(lower, []string{"spare-parts", "ersatzteil", "ersatzteilliste", "spare", "et-blatt", "explosionszeichnung", "serviceblatt", "bedienungsanl", "manual"}) &&
		strings.Contains(lower, ".pdf")
}

func documentTextMatchesArticleNumber(text, articleNumber string) bool {
	needle := normalizedArticleNumber(articleNumber)
	if needle == "" {
		return false
	}
	return strings.Contains(normalizedArticleNumber(text), needle)
}

func normalizedArticleNumber(value string) string {
	builder := strings.Builder{}
	for _, char := range value {
		if char >= '0' && char <= '9' {
			builder.WriteRune(char)
		}
	}
	return builder.String()
}

func mergeArticleSpareParts(base, extra []ArticleSearchSparePart, limit int) []ArticleSearchSparePart {
	if limit <= 0 {
		limit = 80
	}
	out := []ArticleSearchSparePart{}
	seen := map[string]bool{}
	add := func(part ArticleSearchSparePart) bool {
		key := strings.ToLower(part.ArticleNumber + "|" + part.Description + "|" + part.URL)
		if key == "||" || seen[key] {
			return len(out) >= limit
		}
		seen[key] = true
		out = append(out, part)
		return len(out) >= limit
	}
	for _, part := range base {
		if add(part) {
			return out
		}
	}
	for _, part := range extra {
		if add(part) {
			return out
		}
	}
	return out
}
