package application

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"railkeeper/backend/internal/safefetch"
)

type DuckDuckGoArticleSearchAdapter struct {
	client *http.Client
}

func NewDuckDuckGoArticleSearchAdapter(client *http.Client) *DuckDuckGoArticleSearchAdapter {
	if client == nil {
		client = safefetch.NewHTTPClient(context.Background(), safefetch.Options{Timeout: 10 * time.Second})
	}
	return &DuckDuckGoArticleSearchAdapter{client: client}
}

func (a *DuckDuckGoArticleSearchAdapter) Search(ctx context.Context, input ArticleSearchInput, query string) ([]ArticleSearchResult, error) {
	if isEANOnlyArticleSearch(input, query) {
		results, err := a.searchDuckDuckGo(ctx, input, query, "DuckDuckGo")
		if err == nil && len(results) > 0 {
			results = dedupeArticleResults(results)
			a.enrichResultsFromPages(ctx, input, results)
			return results, nil
		}

		fallbackResults, fallbackErr := a.searchDuckDuckGo(ctx, input, query+" Modelleisenbahn", "DuckDuckGo")
		if fallbackErr != nil {
			if err != nil {
				return nil, err
			}
			return nil, fallbackErr
		}
		results = dedupeArticleResults(fallbackResults)
		a.enrichResultsFromPages(ctx, input, results)
		return results, nil
	}

	results := []ArticleSearchResult{}
	var firstErr error
	for _, searchQuery := range articleSearchQueries(input, query) {
		searchResults, err := a.searchDuckDuckGo(ctx, input, searchQuery.Query, searchQuery.Source)
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		if isPriorityArticleSource(searchQuery.Source) || hasPriorityArticleURL(input, searchResults) {
			searchResults = dedupeArticleResults(searchResults)
			a.enrichResultsFromPages(ctx, input, searchResults)
		}
		results = append(results, searchResults...)
	}
	if firstErr != nil && len(results) == 0 {
		return nil, firstErr
	}
	results = dedupeArticleResults(results)
	a.enrichResultsFromPages(ctx, input, results)
	return results, nil
}

func isPriorityArticleSource(source string) bool {
	return source == "Herstellerseiten" || source == "Modellbahn-Fokus"
}

func hasPriorityArticleURL(input ArticleSearchInput, results []ArticleSearchResult) bool {
	for _, result := range results {
		if isManufacturerPreferredURL(input, result.URL) || isCatalogURL(result.URL) {
			return true
		}
	}
	return false
}

func articleSearchQueries(input ArticleSearchInput, query string) []articleSearchQuerySpec {
	focused := focusedArticleSearchQuery(input)
	sources := cleanArticleSearchSources(input.SearchSources)
	queries := []articleSearchQuerySpec{}
	hasSource := func(source string) bool {
		for _, selected := range sources {
			if selected == source {
				return true
			}
		}
		return false
	}
	appendQuery := func(searchQuery, source string) {
		if strings.TrimSpace(searchQuery) == "" {
			return
		}
		queries = append(queries, articleSearchQuerySpec{Query: searchQuery, Source: source})
	}

	if hasSource("manufacturer") {
		for _, domain := range preferredManufacturerDomains(input) {
			if focused != "" {
				appendQuery(focused+" site:"+domain, "Herstellerseiten")
			}
			appendQuery(query+" site:"+domain, "Herstellerseiten")
			if len(queries) >= 4 {
				break
			}
		}
	}
	if hasSource("catalogs") {
		for _, domain := range catalogArticleDomains {
			if focused != "" {
				appendQuery(focused+" site:"+domain, "Modellbahn-Fokus")
			}
			appendQuery(query+" site:"+domain, "Modellbahn-Fokus")
		}
	}
	if hasSource("dealers") {
		for _, domain := range dealerArticleDomains {
			appendQuery(query+" site:"+domain, "H?ndlerseiten")
			if len(queries) >= 8 {
				break
			}
		}
	}
	if hasSource("wiki") {
		appendQuery(query+" site:modellbau-wiki.de", "Modellbau-Wiki")
	}
	if hasSource("web") {
		appendQuery(focused, "DuckDuckGo")
		appendQuery(query, "DuckDuckGo")
		appendQuery(query+" Modelleisenbahn", "DuckDuckGo")
	}
	return uniqueArticleSearchQueries(queries, 11)
}

func articleSearchQueryInfo(input ArticleSearchInput, query string) []ArticleSearchQueryInfo {
	if isEANOnlyArticleSearch(input, query) {
		return []ArticleSearchQueryInfo{
			{Source: "DuckDuckGo", Query: query},
			{Source: "DuckDuckGo", Query: query + " Modelleisenbahn"},
		}
	}
	queries := articleSearchQueries(input, query)
	out := make([]ArticleSearchQueryInfo, 0, len(queries))
	for _, item := range queries {
		out = append(out, ArticleSearchQueryInfo{Source: item.Source, Query: item.Query})
	}
	return out
}

func uniqueArticleSearchQueries(queries []articleSearchQuerySpec, limit int) []articleSearchQuerySpec {
	seen := map[string]bool{}
	out := []articleSearchQuerySpec{}
	for _, query := range queries {
		key := strings.ToLower(strings.TrimSpace(query.Query))
		if key == "" || seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, query)
		if limit > 0 && len(out) >= limit {
			break
		}
	}
	return out
}

func (s *ArticleSearchService) searchPikoSpareParts(ctx context.Context, input ArticleSearchInput) *ArticleSearchResult {
	if s == nil || !isPikoManufacturer(input.Manufacturer) || input.Fields["sparePartLookup"] != "piko" {
		return nil
	}
	searchText := pikoSparePartSearchText(input.ArticleNumber, input.Fields)
	if searchText == "" {
		return nil
	}
	client := safefetch.NewHTTPClient(ctx, safefetch.Options{Timeout: 10 * time.Second})
	searchURL := "https://www.piko-shop.de/de/artikel/ersatzteil/xref_suchtext-" + url.PathEscape(searchText) + ".html"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, searchURL, nil)
	if err != nil {
		return nil
	}
	req.Header.Set("User-Agent", "RailKeeper/0.1 piko-spare-part-search")
	req.Header.Set("Accept", "text/html,application/xhtml+xml")
	req.Header.Set("Accept-Language", "de-DE,de;q=0.9,en;q=0.5")
	resp, err := client.Do(req)
	if err != nil {
		return nil
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1024*1024))
	if err != nil || len(body) == 0 {
		return nil
	}
	parts := pikoSparePartsFromHTML(string(body), resp.Request.URL.String())
	if len(parts) == 0 {
		return nil
	}
	return &ArticleSearchResult{
		Source:  "PIKO",
		Title:   "PIKO Ersatzteile " + searchText,
		URL:     resp.Request.URL.String(),
		Snippet: "Direkte PIKO Ersatzteilsuche",
		Score:   1000,
		Fields: map[string]ArticleSearchField{
			"manufacturer":  {Label: "Hersteller", Value: "Piko", Confidence: 95},
			"articleNumber": {Label: "Artikel-Nr.", Value: searchText, Confidence: 90},
		},
		SpareParts: parts,
	}
}

func (s *ArticleSearchService) searchRocoSpareParts(ctx context.Context, input ArticleSearchInput) *ArticleSearchResult {
	if s == nil || !isRocoManufacturer(input.Manufacturer) || input.Fields["sparePartLookup"] != "roco" {
		return nil
	}
	searchText := rocoSparePartSearchText(input.ArticleNumber, input.Fields)
	if searchText == "" {
		return nil
	}
	client := safefetch.NewHTTPClient(ctx, safefetch.Options{Timeout: 10 * time.Second})
	searchURL := "https://www.roco.cc/rde/ersatzteile?et=" + url.QueryEscape(searchText)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, searchURL, nil)
	if err != nil {
		return nil
	}
	req.Header.Set("User-Agent", "RailKeeper/0.1 roco-spare-part-search")
	req.Header.Set("Accept", "text/html,application/xhtml+xml")
	req.Header.Set("Accept-Language", "de-DE,de;q=0.9,en;q=0.5")
	resp, err := client.Do(req)
	if err != nil {
		return nil
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1024*1024))
	if err != nil || len(body) == 0 {
		return nil
	}
	parts := rocoSparePartsFromHTML(string(body), resp.Request.URL.String())
	if len(parts) == 0 {
		return nil
	}
	return &ArticleSearchResult{
		Source:  "ROCO",
		Title:   "ROCO Ersatzteile " + searchText,
		URL:     resp.Request.URL.String(),
		Snippet: "Direkte ROCO Ersatzteilsuche",
		Score:   1000,
		Fields: map[string]ArticleSearchField{
			"manufacturer":  {Label: "Hersteller", Value: "Roco", Confidence: 95},
			"articleNumber": {Label: "Artikel-Nr.", Value: searchText, Confidence: 90},
		},
		SpareParts: parts,
	}
}

func isPikoManufacturer(manufacturer string) bool {
	manufacturer = strings.ToLower(strings.TrimSpace(manufacturer))
	return manufacturer == "piko" || strings.Contains(manufacturer, "piko")
}

func isRocoManufacturer(manufacturer string) bool {
	manufacturer = strings.ToLower(strings.TrimSpace(manufacturer))
	return manufacturer == "roco" || strings.Contains(manufacturer, "roco")
}

func pikoSparePartSearchText(articleNumber string, fields map[string]string) string {
	for _, value := range []string{
		fields["vehicleArticleNumber"],
		fields["modelArticleNumber"],
		fields["locomotiveArticleNumber"],
		fields["lokArticleNumber"],
	} {
		if digits := firstFiveDigits(value); digits != "" {
			return digits
		}
	}
	if digits := firstFiveDigits(articleNumber); digits != "" {
		return digits
	}
	return ""
}

func rocoSparePartSearchText(articleNumber string, fields map[string]string) string {
	for _, value := range []string{articleNumber, fields["articleNumber"], fields["sparePartArticleNumber"]} {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		return strings.ReplaceAll(value, "-", "")
	}
	return ""
}
