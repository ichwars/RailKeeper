package application

import (
	"context"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"

	"railkeeper/backend/internal/safefetch"
)

func (a *DuckDuckGoArticleSearchAdapter) enrichResultsFromPages(ctx context.Context, input ArticleSearchInput, results []ArticleSearchResult) {
	for _, index := range articleResultEnrichmentIndices(input, results) {
		pageCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		body, finalURL, err := a.fetchArticlePage(pageCtx, results[index].URL)
		cancel()
		if err != nil || body == "" {
			results[index].Trace.Error = detailLoadError(err, body)
			continue
		}
		fieldsBefore := len(results[index].Fields)
		if finalURL != "" {
			results[index].URL = finalURL
			if sourceField, ok := results[index].Fields["articleSourceUrl"]; ok {
				sourceField.Value = finalURL
				results[index].Fields["articleSourceUrl"] = sourceField
			}
		}
		pageText := visibleArticleText(body)
		if pageDescription := firstRegexValue(metaDescriptionRegex, body); pageDescription != "" {
			pageText = cleanHTML(pageDescription) + " " + pageText
		}
		for key, field := range buildArticleFields(input, results[index].Title, results[index].URL, pageText) {
			if existing, ok := results[index].Fields[key]; !ok || field.Confidence > existing.Confidence {
				results[index].Fields[key] = field
			}
		}
		results[index].Documents = articleDocumentsFromHTML(body, results[index].URL)
		documentSpareParts := a.articleSparePartsFromMatchingDocuments(ctx, input, results[index].Documents)
		pageSpareParts := []ArticleSearchSparePart{}
		if shouldExtractPageSpareParts(input, results[index].URL) {
			pageSpareParts = articleSparePartsFromHTML(body, results[index].URL)
		}
		results[index].Images = articleImagesFromHTML(body, results[index].URL, results[index].Title)
		results[index].SpareParts = mergeArticleSpareParts(documentSpareParts, pageSpareParts, 80)
		results[index].Trace = ArticleSearchResultTrace{
			DetailLoaded:     true,
			DetailFields:     len(results[index].Fields) - fieldsBefore,
			DetailImages:     len(results[index].Images),
			DetailSpareParts: len(results[index].SpareParts),
			DetailDocuments:  len(results[index].Documents),
			FinalURL:         results[index].URL,
		}
		results[index].Score = scoreArticleResult(input, results[index].Title, results[index].URL, results[index].Snippet+" "+pageText, results[index].Fields) + duckDuckGoRankBonus(index)
	}
}

func detailLoadError(err error, body string) string {
	if err != nil {
		return err.Error()
	}
	if body == "" {
		return "empty response"
	}
	return ""
}

func articleResultEnrichmentIndices(input ArticleSearchInput, results []ArticleSearchResult) []int {
	const limit = 10
	indices := []int{}
	seen := map[int]bool{}
	add := func(index int) {
		if index < 0 || index >= len(results) || seen[index] || len(indices) >= limit {
			return
		}
		seen[index] = true
		indices = append(indices, index)
	}
	for index, result := range results {
		if isManufacturerPreferredURL(input, result.URL) || isCatalogURL(result.URL) {
			add(index)
		}
	}
	for index := 0; index < len(results) && len(indices) < limit; index++ {
		add(index)
	}
	return indices
}

func (a *DuckDuckGoArticleSearchAdapter) fetchArticlePage(ctx context.Context, pageURL string) (string, string, error) {
	parsed, err := url.Parse(pageURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "", "", fmt.Errorf("invalid article page url")
	}
	if !safefetch.IsPublicHTTPURL(ctx, pageURL) {
		return "", "", fmt.Errorf("article page url is not public http(s)")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, pageURL, nil)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("User-Agent", "RailKeeper/0.1 article-search")
	req.Header.Set("Accept", "text/html,application/xhtml+xml")
	req.Header.Set("Accept-Language", "de-DE,de;q=0.9,en;q=0.5")
	resp, err := a.client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return "", "", fmt.Errorf("article page returned status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 768*1024))
	if err != nil {
		return "", "", err
	}
	return string(body), resp.Request.URL.String(), nil
}

func articleDocumentsFromHTML(body, pageURL string) []ArticleSearchDocument {
	seen := map[string]bool{}
	documents := []ArticleSearchDocument{}
	for _, tag := range linkTagPattern.FindAllString(body, -1) {
		match := linkHrefAttrPattern.FindStringSubmatch(tag)
		if len(match) < 2 {
			continue
		}
		documentURL := resolveURL(pageURL, html.UnescapeString(match[1]))
		if documentURL == "" || seen[strings.ToLower(documentURL)] || !looksLikeArticleDocument(tag, documentURL) {
			continue
		}
		title := cleanDocumentTitle(cleanHTML(tag), documentURL)
		seen[strings.ToLower(documentURL)] = true
		documents = append(documents, ArticleSearchDocument{
			Title:  title,
			URL:    documentURL,
			Source: sourceDisplayName(pageURL),
			Kind:   classifyArticleDocument(tag + " " + documentURL),
		})
		if len(documents) >= 12 {
			break
		}
	}
	return documents
}

func articleSparePartsFromHTML(body, pageURL string) []ArticleSearchSparePart {
	seen := map[string]bool{}
	parts := []ArticleSearchSparePart{}
	rows := rowLikePattern.FindAllString(body, -1)
	rows = append(rows, strings.Split(visibleArticleLines(body), "\n")...)
	for _, row := range rows {
		part, ok := articleSparePartFromRow(row, pageURL)
		if !ok {
			continue
		}
		key := strings.ToLower(part.ArticleNumber + "|" + part.Description + "|" + part.URL)
		if seen[key] {
			continue
		}
		seen[key] = true
		parts = append(parts, part)
		if len(parts) >= 80 {
			break
		}
	}
	return parts
}

func shouldExtractPageSpareParts(input ArticleSearchInput, pageURL string) bool {
	return isManufacturerPreferredURL(input, pageURL) || isCatalogURL(pageURL)
}

func (a *DuckDuckGoArticleSearchAdapter) articleSparePartsFromMatchingDocuments(ctx context.Context, input ArticleSearchInput, documents []ArticleSearchDocument) []ArticleSearchSparePart {
	articleNumber := strings.TrimSpace(input.ArticleNumber)
	if articleNumber == "" {
		return nil
	}
	parts := []ArticleSearchSparePart{}
	for index, document := range prioritizedSparePartDocuments(input, documents) {
		if index >= 4 {
			break
		}
		if len(parts) >= 80 || !looksLikeSparePartDocument(document) {
			continue
		}
		documentCtx, cancel := context.WithTimeout(ctx, 4*time.Second)
		data, err := a.fetchArticleDocument(documentCtx, document.URL)
		cancel()
		if err != nil || len(data) == 0 {
			continue
		}
		documentParts := ArticleSparePartsFromDocumentData(data, articleNumber, document.URL)
		parts = mergeArticleSpareParts(parts, documentParts, 80)
	}
	return parts
}

func prioritizedSparePartDocuments(input ArticleSearchInput, documents []ArticleSearchDocument) []ArticleSearchDocument {
	out := []ArticleSearchDocument{}
	for _, document := range documents {
		if looksLikeSparePartDocument(document) {
			out = append(out, document)
		}
	}
	sort.SliceStable(out, func(left, right int) bool {
		leftScore := sparePartDocumentPriority(input, out[left])
		rightScore := sparePartDocumentPriority(input, out[right])
		if leftScore != rightScore {
			return leftScore > rightScore
		}
		return strings.ToLower(out[left].Title+" "+out[left].URL) < strings.ToLower(out[right].Title+" "+out[right].URL)
	})
	return out
}

func sparePartDocumentPriority(input ArticleSearchInput, document ArticleSearchDocument) int {
	score := 0
	value := strings.ToLower(document.Kind + " " + document.Title + " " + document.URL)
	if isManufacturerPreferredURL(input, document.URL) {
		score += 100
	}
	if strings.Contains(value, ".pdf") {
		score += 35
	}
	if containsAny(value, []string{"ersatzteil", "spare-parts", "spare parts", "et-blatt", "explosionszeichnung", "serviceblatt"}) {
		score += 30
	}
	if containsAny(value, []string{"bedienungsanl", "bedienungsanleitung", "manual"}) {
		score -= 15
	}
	if isDealerURL(document.URL) {
		score -= 20
	}
	return score
}

func (a *DuckDuckGoArticleSearchAdapter) fetchArticleDocument(ctx context.Context, documentURL string) ([]byte, error) {
	parsed, err := url.Parse(documentURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return nil, fmt.Errorf("invalid article document url")
	}
	if !safefetch.IsPublicHTTPURL(ctx, documentURL) {
		return nil, fmt.Errorf("article document url is not public http(s)")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, documentURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "RailKeeper/0.1 article-search")
	req.Header.Set("Accept", "application/pdf,*/*;q=0.7")
	req.Header.Set("Accept-Language", "de-DE,de;q=0.9,en;q=0.5")
	resp, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("article document returned status %d", resp.StatusCode)
	}
	return io.ReadAll(io.LimitReader(resp.Body, 4*1024*1024))
}

func articleSparePartsFromDocumentText(text, articleNumber, documentURL string) []ArticleSearchSparePart {
	if !looksLikeSparePartsDocumentText(text, articleNumber) {
		return nil
	}
	text = sparePartsDocumentSection(text)
	seen := map[string]bool{}
	parts := []ArticleSearchSparePart{}
	for _, row := range articleSparePartDocumentRows(text) {
		part, ok := articleSparePartFromConfirmedDocumentRow(row, documentURL)
		if !ok || normalizedArticleNumber(part.ArticleNumber) == normalizedArticleNumber(articleNumber) {
			continue
		}
		if part.URL == "" {
			part.URL = documentURL
		}
		key := normalizedArticleNumber(part.ArticleNumber)
		if key == "" {
			key = strings.ToLower(part.ArticleNumber + "|" + part.Description + "|" + part.URL)
		}
		if seen[key] {
			continue
		}
		seen[key] = true
		parts = append(parts, part)
		if len(parts) >= 80 {
			break
		}
	}
	return parts
}

func looksLikeArticleDocument(tag, documentURL string) bool {
	lower := strings.ToLower(tag + " " + documentURL)
	if strings.Contains(lower, ".pdf") {
		return true
	}
	return containsAny(lower, []string{"bedienungsanleitung", "anleitung", "manual", "ersatzteil", "spare", "et-blatt", "explosionszeichnung", "serviceblatt", "beipackzettel", "download"})
}

func classifyArticleDocument(value string) string {
	lower := strings.ToLower(value)
	if containsAny(lower, []string{"ersatzteil", "spare", "et-blatt", "explosionszeichnung", "serviceblatt"}) {
		return "spare-parts"
	}
	if containsAny(lower, []string{"bedienungsanleitung", "anleitung", "manual", "beipackzettel"}) {
		return "manual"
	}
	return "document"
}

func cleanDocumentTitle(title, documentURL string) string {
	title = strings.Trim(title, " -:;,.\t")
	if title == "" || strings.EqualFold(title, "download") {
		if parsed, err := url.Parse(documentURL); err == nil {
			parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
			if len(parts) > 0 {
				title = parts[len(parts)-1]
			}
		}
	}
	if title == "" {
		title = "Dokument"
	}
	return title
}

func visibleArticleLines(value string) string {
	value = regexp.MustCompile(`(?is)</(?:tr|li|p|div|h[1-6]|dd|dt|br)>`).ReplaceAllString(value, "\n")
	value = scriptStylePattern.ReplaceAllString(value, " ")
	value = tagPattern.ReplaceAllString(value, " ")
	value = html.UnescapeString(value)
	value = repairMojibake(value)
	lines := []string{}
	for _, line := range strings.Split(value, "\n") {
		line = strings.Join(strings.Fields(line), " ")
		if line != "" {
			lines = append(lines, line)
		}
	}
	return strings.Join(lines, "\n")
}

func containsAny(value string, tokens []string) bool {
	for _, token := range tokens {
		if strings.Contains(value, token) {
			return true
		}
	}
	return false
}

func articleImagesFromHTML(body, pageURL, title string) []ArticleSearchImage {
	seen := map[string]bool{}
	images := []ArticleSearchImage{}
	addImage := func(raw string) bool {
		imageURL := resolveURL(pageURL, html.UnescapeString(raw))
		if imageURL == "" || seen[strings.ToLower(imageURL)] || !looksLikeArticleImage(imageURL) {
			return false
		}
		seen[strings.ToLower(imageURL)] = true
		images = append(images, ArticleSearchImage{URL: imageURL, Title: title, Source: pageURL})
		return len(images) >= 4
	}

	for _, pattern := range []*regexp.Regexp{imageMetaPattern, imageMetaAltPattern} {
		for _, match := range pattern.FindAllStringSubmatch(body, -1) {
			if len(match) < 2 {
				continue
			}
			if addImage(match[1]) {
				return images
			}
		}
	}
	for _, tag := range imageTagPattern.FindAllString(body, -1) {
		for _, match := range imageURLAttrPattern.FindAllStringSubmatch(tag, -1) {
			if len(match) >= 2 && addImage(match[1]) {
				return images
			}
		}
		for _, match := range imageSrcSetAttrPattern.FindAllStringSubmatch(tag, -1) {
			if len(match) < 2 {
				continue
			}
			for _, candidate := range imageURLsFromSrcset(match[1]) {
				if addImage(candidate) {
					return images
				}
			}
		}
	}
	return images
}

func imageURLsFromSrcset(srcset string) []string {
	best := ""
	for _, candidate := range strings.Split(srcset, ",") {
		parts := strings.Fields(strings.TrimSpace(candidate))
		if len(parts) > 0 {
			best = parts[0]
		}
	}
	if best == "" {
		return nil
	}
	return []string{best}
}

func looksLikeArticleImage(imageURL string) bool {
	lower := strings.ToLower(imageURL)
	badTokens := []string{
		"badge", "banner", "blank", "dummy", "flaggen", "/flag", "icon", "i_ital",
		"lazy", "loading", "logo", "no-image", "noimage", "payment", "placeholder",
		"pixel", "shipping", "spacer", "sprite", "tracking", "transparent", "versandkostenfrei",
	}
	for _, token := range badTokens {
		if strings.Contains(lower, token) {
			return false
		}
	}
	if strings.Contains(lower, "1x1") || strings.Contains(lower, "clear.gif") {
		return false
	}
	return strings.Contains(lower, ".jpg") || strings.Contains(lower, ".jpeg") || strings.Contains(lower, ".png") || strings.Contains(lower, ".webp")
}
