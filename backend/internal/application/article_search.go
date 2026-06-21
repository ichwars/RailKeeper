package application

import (
	"bytes"
	"compress/zlib"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"railkeeper/backend/internal/safefetch"
)

var ErrArticleSearchValidation = errors.New("article search validation failed")

var pdfOCRTextExtractor = extractPDFOCRText

type ArticleSearchInput struct {
	Manufacturer        string            `json:"manufacturer"`
	ArticleNumber       string            `json:"articleNumber"`
	Name                string            `json:"name"`
	Gauge               string            `json:"gauge"`
	SearchSources       []string          `json:"searchSources"`
	Fields              map[string]string `json:"fields"`
	PreferredDomains    []string          `json:"preferredDomains,omitempty"`
	ManufacturerAliases []string          `json:"manufacturerAliases,omitempty"`
}

type ArticleSearchField struct {
	Label      string `json:"label"`
	Value      string `json:"value"`
	Confidence int    `json:"confidence"`
}

type ArticleSearchImage struct {
	URL    string `json:"url"`
	Title  string `json:"title"`
	Source string `json:"source"`
}

type ArticleSearchSparePart struct {
	ArticleNumber string `json:"articleNumber"`
	Description   string `json:"description"`
	Price         string `json:"price,omitempty"`
	URL           string `json:"url,omitempty"`
	Source        string `json:"source,omitempty"`
	Availability  string `json:"availability,omitempty"`
}

type ArticleSearchDocument struct {
	Title  string `json:"title"`
	URL    string `json:"url"`
	Source string `json:"source,omitempty"`
	Kind   string `json:"kind,omitempty"`
}

type ArticleSearchResult struct {
	Source     string                        `json:"source"`
	Title      string                        `json:"title"`
	URL        string                        `json:"url"`
	Snippet    string                        `json:"snippet"`
	Score      int                           `json:"score"`
	Fields     map[string]ArticleSearchField `json:"fields"`
	Images     []ArticleSearchImage          `json:"images,omitempty"`
	SpareParts []ArticleSearchSparePart      `json:"spareParts,omitempty"`
	Documents  []ArticleSearchDocument       `json:"documents,omitempty"`
	Trace      ArticleSearchResultTrace      `json:"trace"`
	Conflicts  []string                      `json:"conflicts,omitempty"`
}

type ArticleSearchResultTrace struct {
	DetailLoaded     bool   `json:"detailLoaded"`
	DetailFields     int    `json:"detailFields"`
	DetailImages     int    `json:"detailImages"`
	DetailSpareParts int    `json:"detailSpareParts"`
	DetailDocuments  int    `json:"detailDocuments"`
	FinalURL         string `json:"finalUrl,omitempty"`
	Error            string `json:"error,omitempty"`
}

type ArticleSearchResponse struct {
	Query               string                   `json:"query"`
	Sources             []string                 `json:"sources"`
	ManufacturerDomains []string                 `json:"manufacturerDomains,omitempty"`
	Queries             []ArticleSearchQueryInfo `json:"queries,omitempty"`
	Results             []ArticleSearchResult    `json:"results"`
}

type ArticleSearchQueryInfo struct {
	Source string `json:"source"`
	Query  string `json:"query"`
}

type ArticleSearchAdapter interface {
	Search(ctx context.Context, input ArticleSearchInput, query string) ([]ArticleSearchResult, error)
}

type ArticleSearchService struct {
	adapters   []ArticleSearchAdapter
	timeout    time.Duration
	masterData *MasterDataService
}

type articleSearchQuerySpec struct {
	Query  string
	Source string
}

func NewArticleSearchService(masterData ...*MasterDataService) *ArticleSearchService {
	var masterDataService *MasterDataService
	if len(masterData) > 0 {
		masterDataService = masterData[0]
	}
	return &ArticleSearchService{
		adapters: []ArticleSearchAdapter{
			NewDuckDuckGoArticleSearchAdapter(nil),
		},
		timeout:    10 * time.Second,
		masterData: masterDataService,
	}
}

func (s *ArticleSearchService) Search(ctx context.Context, input ArticleSearchInput) (*ArticleSearchResponse, error) {
	input = cleanArticleSearchInput(input)
	input = s.withManufacturerMetadata(ctx, input)
	query := articleSearchQuery(input)
	if query == "" {
		return nil, ErrArticleSearchValidation
	}

	sources := cleanArticleSearchSources(input.SearchSources)
	manufacturerDomains := preferredManufacturerDomains(input)
	queryPlan := articleSearchQueryInfo(input, query)

	searchCtx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	results := []ArticleSearchResult{}
	if pikoResult := s.searchPikoSpareParts(searchCtx, input); pikoResult != nil {
		results = append(results, *pikoResult)
	}
	if rocoResult := s.searchRocoSpareParts(searchCtx, input); rocoResult != nil {
		results = append(results, *rocoResult)
	}
	for _, adapter := range s.adapters {
		adapterResults, err := adapter.Search(searchCtx, input, query)
		if err != nil && len(results) == 0 {
			return nil, err
		}
		results = append(results, adapterResults...)
	}

	for index := range results {
		results[index].Conflicts = articleSearchConflicts(input, results[index].Fields)
	}
	sort.SliceStable(results, func(left, right int) bool {
		return results[left].Score > results[right].Score
	})
	results = dedupeArticleResults(results)
	if len(results) > 10 {
		results = results[:10]
	}

	return &ArticleSearchResponse{
		Query:               query,
		Sources:             sources,
		ManufacturerDomains: manufacturerDomains,
		Queries:             queryPlan,
		Results:             results,
	}, nil
}

func cleanArticleSearchInput(input ArticleSearchInput) ArticleSearchInput {
	input.Manufacturer = strings.TrimSpace(input.Manufacturer)
	input.ArticleNumber = strings.TrimSpace(input.ArticleNumber)
	input.Name = strings.TrimSpace(input.Name)
	input.Gauge = strings.TrimSpace(input.Gauge)
	input.SearchSources = cleanArticleSearchSources(input.SearchSources)
	cleanFields := map[string]string{}
	for key, value := range input.Fields {
		value = strings.TrimSpace(value)
		if value != "" {
			cleanFields[key] = value
		}
	}
	input.Fields = cleanFields
	return input
}

func cleanArticleSearchSources(sources []string) []string {
	allowed := map[string]bool{
		"web":          true,
		"manufacturer": true,
		"catalogs":     true,
		"dealers":      true,
		"wiki":         true,
	}
	cleaned := []string{}
	for _, source := range sources {
		source = strings.ToLower(strings.TrimSpace(source))
		if allowed[source] {
			cleaned = append(cleaned, source)
		}
	}
	cleaned = uniqueNonEmpty(cleaned)
	if len(cleaned) == 0 {
		return []string{"manufacturer", "catalogs", "dealers", "web"}
	}
	return cleaned
}

func (s *ArticleSearchService) withManufacturerMetadata(ctx context.Context, input ArticleSearchInput) ArticleSearchInput {
	if s == nil || s.masterData == nil || strings.TrimSpace(input.Manufacturer) == "" {
		return input
	}
	entries, err := s.masterData.List(ctx, "manufacturer", true)
	if err != nil {
		return input
	}
	entry, ok := matchManufacturerEntry(input.Manufacturer, entries)
	if !ok {
		return input
	}
	aliases := metadataStringList(entry.Metadata, "aliases")
	input.ManufacturerAliases = uniqueNonEmpty(append(input.ManufacturerAliases, aliases...))
	domains := metadataStringList(entry.Metadata, "searchDomains")
	if website := metadataStringValue(entry.Metadata, "website"); website != "" {
		domains = append(domains, domainFromURL(website))
	}
	input.PreferredDomains = uniqueDomains(append(input.PreferredDomains, domains...))
	return input
}

func matchManufacturerEntry(manufacturer string, entries []MasterDataEntry) (MasterDataEntry, bool) {
	needle := slugKey(manufacturer)
	if needle == "" {
		return MasterDataEntry{}, false
	}
	for _, entry := range entries {
		if slugKey(entry.Label) == needle || slugKey(entry.Key) == needle {
			return entry, true
		}
		for _, alias := range metadataStringList(entry.Metadata, "aliases") {
			if slugKey(alias) == needle {
				return entry, true
			}
		}
	}
	return MasterDataEntry{}, false
}

func metadataStringValue(metadata map[string]any, key string) string {
	if metadata == nil {
		return ""
	}
	value, ok := metadata[key]
	if !ok {
		return ""
	}
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	default:
		return strings.TrimSpace(fmt.Sprint(typed))
	}
}

func metadataStringList(metadata map[string]any, key string) []string {
	if metadata == nil {
		return nil
	}
	value, ok := metadata[key]
	if !ok {
		return nil
	}
	items := []string{}
	switch typed := value.(type) {
	case []any:
		for _, item := range typed {
			items = append(items, strings.TrimSpace(fmt.Sprint(item)))
		}
	case []string:
		items = append(items, typed...)
	case string:
		items = strings.Split(typed, ",")
	}
	return uniqueNonEmpty(items)
}

func articleSearchQuery(input ArticleSearchInput) string {
	parts := []string{}
	for _, value := range []string{input.Name, input.ArticleNumber, input.Fields["ean"], input.Manufacturer, input.Gauge} {
		if value != "" {
			parts = append(parts, value)
		}
	}

	return strings.Join(uniqueNonEmpty(parts), " ")
}

func isEANOnlyArticleSearch(input ArticleSearchInput, query string) bool {
	ean := strings.TrimSpace(input.Fields["ean"])
	if ean == "" || query != ean {
		return false
	}
	return input.Manufacturer == "" && input.ArticleNumber == "" && input.Name == "" && input.Gauge == ""
}

func articleSearchConflicts(input ArticleSearchInput, fields map[string]ArticleSearchField) []string {
	current := map[string]string{
		"manufacturer":  input.Manufacturer,
		"articleNumber": input.ArticleNumber,
		"name":          input.Name,
		"gauge":         input.Gauge,
	}
	for key, value := range input.Fields {
		current[key] = value
	}

	conflicts := []string{}
	for key, field := range fields {
		existing := strings.TrimSpace(current[key])
		if existing == "" || field.Value == "" {
			continue
		}
		if !strings.EqualFold(existing, field.Value) {
			conflicts = append(conflicts, key)
		}
	}
	sort.Strings(conflicts)
	return conflicts
}

func uniqueNonEmpty(values []string) []string {
	seen := map[string]bool{}
	result := []string{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		key := strings.ToLower(value)
		if seen[key] {
			continue
		}
		seen[key] = true
		result = append(result, value)
	}
	return result
}

func dedupeArticleResults(results []ArticleSearchResult) []ArticleSearchResult {
	seen := map[string]bool{}
	out := []ArticleSearchResult{}
	for _, result := range results {
		key := strings.ToLower(strings.TrimSpace(result.URL))
		if key == "" || seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, result)
	}
	return out
}

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

func firstFiveDigits(value string) string {
	digits := normalizedArticleNumber(value)
	if len(digits) < 5 {
		return ""
	}
	return digits[:5]
}

func pikoSparePartsFromHTML(body, pageURL string) []ArticleSearchSparePart {
	seen := map[string]bool{}
	parts := []ArticleSearchSparePart{}
	for index, block := range strings.Split(body, `<div class="artikel_ersatzteil__list_item">`) {
		if index == 0 {
			continue
		}
		block = `<div class="artikel_ersatzteil__list_item">` + block
		link := ""
		if match := linkHrefAttrPattern.FindStringSubmatch(block); len(match) >= 2 {
			link = resolveURL(pageURL, html.UnescapeString(match[1]))
		}
		description := ""
		if match := pikoSparePartTitlePattern.FindStringSubmatch(block); len(match) >= 2 {
			description = cleanHTML(match[1])
		}
		articleNumber := ""
		if match := pikoSparePartNumberPattern.FindStringSubmatch(block); len(match) >= 2 {
			articleNumber = strings.TrimSpace(html.UnescapeString(match[1]))
		}
		price := ""
		if match := pikoSparePartPriceLoosePattern.FindStringSubmatch(block); len(match) >= 2 {
			price = strings.ReplaceAll(strings.TrimSpace(match[1]), ",", ".")
		}
		availability := ""
		if match := pikoSparePartAvailabilityPattern.FindStringSubmatch(block); len(match) >= 2 {
			availability = cleanHTML(match[1])
		}
		if articleNumber == "" || (price == "" && link == "") {
			continue
		}
		key := strings.ToLower(articleNumber + "|" + link)
		if seen[key] {
			continue
		}
		seen[key] = true
		parts = append(parts, ArticleSearchSparePart{
			ArticleNumber: articleNumber,
			Description:   description,
			Price:         price,
			URL:           link,
			Source:        "PIKO",
			Availability:  availability,
		})
		if len(parts) >= 120 {
			break
		}
	}
	return parts
}

func rocoSparePartsFromHTML(body, pageURL string) []ArticleSearchSparePart {
	seen := map[string]bool{}
	parts := []ArticleSearchSparePart{}
	for index, block := range strings.Split(body, `<div class="row table-row-et">`) {
		if index == 0 {
			continue
		}
		block = `<div class="row table-row-et">` + block
		articleNumber := ""
		if match := rocoSparePartNumberPattern.FindStringSubmatch(block); len(match) >= 2 {
			articleNumber = cleanHTML(match[1])
		}
		description := ""
		if match := rocoSparePartDescriptionPattern.FindStringSubmatch(block); len(match) >= 2 {
			description = cleanHTML(match[1])
		}
		price := ""
		if match := rocoSparePartPriceLoosePattern.FindStringSubmatch(block); len(match) >= 2 {
			price = strings.ReplaceAll(strings.TrimSpace(match[1]), ",", ".")
		}
		availability := ""
		if match := rocoSparePartAvailabilityPattern.FindStringSubmatch(block); len(match) >= 2 {
			availability = cleanHTML(match[1])
		}
		if articleNumber == "" || (price == "" && availability == "") {
			continue
		}
		link := rocoSparePartURL(pageURL, articleNumber)
		key := strings.ToLower(articleNumber + "|" + link)
		if seen[key] {
			continue
		}
		seen[key] = true
		parts = append(parts, ArticleSearchSparePart{
			ArticleNumber: articleNumber,
			Description:   description,
			Price:         price,
			URL:           link,
			Source:        "ROCO",
			Availability:  availability,
		})
		if len(parts) >= 80 {
			break
		}
	}
	return parts
}

func rocoSparePartURL(pageURL, articleNumber string) string {
	parsed, err := url.Parse(pageURL)
	if err != nil {
		return pageURL
	}
	parsed.RawQuery = url.Values{"et": []string{articleNumber}}.Encode()
	return parsed.String()
}

func focusedArticleSearchQuery(input ArticleSearchInput) string {
	parts := []string{}
	for _, value := range []string{input.ArticleNumber, input.Manufacturer, input.Gauge} {
		if value != "" {
			parts = append(parts, value)
		}
	}
	return strings.Join(uniqueNonEmpty(parts), " ")
}

func (a *DuckDuckGoArticleSearchAdapter) searchDuckDuckGo(ctx context.Context, input ArticleSearchInput, query string, source string) ([]ArticleSearchResult, error) {
	requestURL := duckDuckGoSearchURL(query)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("build article search request: %w", err)
	}
	req.Header.Set("User-Agent", "RailKeeper/0.1 article-search")
	req.Header.Set("Accept", "text/html,application/xhtml+xml")
	req.Header.Set("Accept-Language", "de-DE,de;q=0.9,en;q=0.5")

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("article search request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("article search returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 2*1024*1024))
	if err != nil {
		return nil, fmt.Errorf("read article search response: %w", err)
	}
	results := parseDuckDuckGoResults(string(body), input, source)
	return results, nil
}

func duckDuckGoSearchURL(query string) string {
	values := url.Values{}
	values.Set("q", query)
	values.Set("kl", "de-de")
	values.Set("kad", "de_DE")
	return "https://duckduckgo.com/html/?" + values.Encode()
}

var (
	resultBlockPattern               = regexp.MustCompile(`(?s)<div class="result results_links.*?</div>\s*</div>`)
	resultLinkPattern                = regexp.MustCompile(`(?s)<a[^>]+class="result__a"[^>]+href="([^"]+)"[^>]*>(.*?)</a>`)
	snippetPattern                   = regexp.MustCompile(`(?s)<a[^>]+class="result__snippet"[^>]*>(.*?)</a>|<div[^>]+class="result__snippet"[^>]*>(.*?)</div>`)
	tagPattern                       = regexp.MustCompile(`(?s)<[^>]+>`)
	scriptStylePattern               = regexp.MustCompile(`(?is)<script[^>]*>.*?</script>|<style[^>]*>.*?</style>|<noscript[^>]*>.*?</noscript>|<svg[^>]*>.*?</svg>`)
	pricePattern                     = regexp.MustCompile(`(?i)(?:hersteller[-\s]*preis|preis|uvp)?[^\d]{0,20}(\d{1,4}(?:[,.]\d{2})?)\s?(?:eur|euro|\x{20AC})`)
	lengthPattern                    = regexp.MustCompile(`(?i)(?:l[äa]nge|laenge|length|ma[ßs]|mass|lüp|luep|luep\.)[^\d]{0,30}(\d{2,4}(?:[,.]\d+)?)\s?(?:mm)?`)
	weightPattern                    = regexp.MustCompile(`(?i)(?:gewicht|weight)[^\d]{0,18}(\d{1,5}(?:[,.]\d+)?)\s?g`)
	tractionTirePattern              = regexp.MustCompile(`(?i)(?:haftreifen|traction\s*tire)[^\d]{0,18}(\d{1,2})`)
	eanPattern                       = regexp.MustCompile(`\b(\d{12,14})\b`)
	epochPattern                     = regexp.MustCompile(`(?i)(?:epoche|epoch|ep\.)[^IVX]{0,16}(I{1,3}|IV|V|VI)\b`)
	railwayPattern                   = regexp.MustCompile(`\b(DB AG|DB|DRG|DR|SBB|\x{00D6}BB|OeBB|BLS|SNCF|NS|FS)\b`)
	adapterPattern                   = regexp.MustCompile(`(?i)\b(NEM\s?651|NEM\s?652|NEM\s?658|PluX\s?16|PluX\s?22|MTC\s?21|Next\s?18|8-?polig|21-?polig|DSS\s?8pol|elektrische\s+schnittstelle)\b`)
	powerPattern                     = regexp.MustCompile(`(?i)\b(DC|AC|2-?Leiter|3-?Leiter|Gleichstrom|Wechselstrom)\b`)
	digitalPositivePattern           = regexp.MustCompile(`(?i)(?:\bdigital\s*[:=]\s*(?:ja|yes|true)\b|\bdigitaldecoder\b|\bsounddecoder\b|\bmit\s+(?:dcc\s+)?decoder\b)`)
	headlightDescriptionPattern      = regexp.MustCompile(`(?i)(?:lichtwechsel|fahrlicht|spitzenlicht|spitzenbeleuchtung|schlusslicht)[^\n:;]{0,35}[:]\s*([^.;\n]{3,220})`)
	lightingDescriptionPattern       = regexp.MustCompile(`(?i)(?:innenbeleuchtung|fuehrerstandsbeleuchtung|fuehrerstand|kabinenbeleuchtung|beleuchtung)[^\n:;]{0,35}[:]\s*([^.;\n]{3,180})`)
	soundDescriptionPattern          = regexp.MustCompile(`(?i)(?:soundgenerator|sounddecoder|\bsound\b|sound\s+laut\s+artikeldaten|geräuschmodul|geraeuschmodul|ger..uschmodul)[^\n:;]{0,35}[:]\s*([^.;\n]{3,180})`)
	imageMetaPattern                 = regexp.MustCompile(`(?is)<meta[^>]+(?:property|name)=["'](?:og:image|twitter:image|thumbnail)["'][^>]+content=["']([^"']+)["']`)
	imageMetaAltPattern              = regexp.MustCompile(`(?is)<meta[^>]+content=["']([^"']+)["'][^>]+(?:property|name)=["'](?:og:image|twitter:image|thumbnail)["']`)
	imageTagPattern                  = regexp.MustCompile(`(?is)<img\b[^>]*>`)
	linkTagPattern                   = regexp.MustCompile(`(?is)<a\b[^>]*>.*?</a>`)
	linkHrefAttrPattern              = regexp.MustCompile(`(?is)\bhref=["']([^"']+)["']`)
	rowLikePattern                   = regexp.MustCompile(`(?is)<(?:tr|li)\b[^>]*>.*?</(?:tr|li)>`)
	sparePartArticlePattern          = regexp.MustCompile(`(?i)\b([A-Z]?\d{4,8}(?:[-/][A-Z0-9]+)?)\b`)
	pikoSparePartTitlePattern        = regexp.MustCompile(`(?is)<h3[^>]*>(.*?)</h3>`)
	pikoSparePartNumberPattern       = regexp.MustCompile(`(?is)Artikelnummer:\s*([^<\s]+)`)
	pikoSparePartPricePattern        = regexp.MustCompile(`(?is)<div class="artikel_ersatzteil__price">\s*(\d{1,4}(?:[,.]\d{2})?)\s*(?:€|&euro;|EUR)?\s*</div>`)
	pikoSparePartAvailabilityPattern = regexp.MustCompile(`(?is)<span[^>]+(?:availability|lieferstatus)[^>]*>\s*(.*?)\s*</span>`)
	rocoSparePartNumberPattern       = regexp.MustCompile(`(?is)<div[^>]+class="[^"]*\bart-nr\b[^"]*"[^>]*>\s*([^<]+?)\s*</div>`)
	rocoSparePartDescriptionPattern  = regexp.MustCompile(`(?is)<div[^>]+class="[^"]*\bart-bz\b[^"]*"[^>]*>\s*(.*?)\s*</div>`)
	rocoSparePartPricePattern        = regexp.MustCompile(`(?is)<div[^>]+class="[^"]*\bart-pr\b[^"]*"[^>]*>\s*(\d{1,4}(?:[,.]\d{2})?)\s*(?:â‚¬|&euro;|EUR)?\s*</div>`)
	rocoSparePartAvailabilityPattern = regexp.MustCompile(`(?is)<img[^>]+class="[^"]*\bprodukt-head-verfuegbarkeit\b[^"]*"[^>]+title=["']([^"']+)["']`)
	pikoSparePartPriceLoosePattern   = regexp.MustCompile(`(?is)<div class="artikel_ersatzteil__price">\s*(\d{1,4}(?:[,.]\d{2})?)\s*[^<]{0,16}</div>`)
	rocoSparePartPriceLoosePattern   = regexp.MustCompile(`(?is)<div[^>]+class="[^"]*\bart-pr\b[^"]*"[^>]*>\s*(\d{1,4}(?:[,.]\d{2})?)\s*[^<]{0,16}</div>`)
	imageURLAttrPattern              = regexp.MustCompile(`(?is)\b(?:src|data-src|data-original|data-lazy-src|data-zoom-image)=["']([^"']+)["']`)
	imageSrcSetAttrPattern           = regexp.MustCompile(`(?is)\b(?:srcset|data-srcset)=["']([^"']+)["']`)
	metaDescriptionRegex             = regexp.MustCompile(`(?is)<meta[^>]+(?:name|property)=["'](?:description|og:description)["'][^>]+content=["']([^"']+)["']`)
)

var manufacturerDomains = map[string][]string{
	"arnold":      {"hornby.com"},
	"brawa":       {"brawa.de"},
	"esu":         {"esu.eu"},
	"fleischmann": {"fleischmann.de"},
	"lgb":         {"lgb.de", "maerklin.de"},
	"maerklin":    {"maerklin.de"},
	"piko":        {"piko.de", "piko-shop.de"},
	"roco":        {"roco.cc"},
	"tillig":      {"tillig.com"},
	"trix":        {"trix.de", "maerklin.de"},
	"viessmann":   {"viessmann-modell.com"},
}

var catalogArticleDomains = []string{
	"modellbahn-fokus.de",
}

var dealerArticleDomains = []string{
	"elriwa.de",
	"modellbahnshop-lippe.com",
	"dm-toys.de",
	"haertle.de",
}

var catalogDetailLabels = []string{
	"Hersteller",
	"Art.-Nr.",
	"Artikel-Nr.",
	"Artikelnummer",
	"EAN",
	"Spur",
	"Bahn-Gesellschaft",
	"Bahngesellschaft",
	"Epoche",
	"Stromsystem",
	"Digital-Decoder",
	"Schnittstelle",
	"Motor",
	"Schwungmasse",
	"Haftreifen",
	"L?nge ?ber Puffer",
	"Laenge ueber Puffer",
	"Mindestradius",
	"Spitzenlicht",
	"Vorbild (Land)",
	"Hersteller-Preis",
	"Herstellerpreis",
	"Preis",
	"UVP",
	"Produktlinie",
	"Erscheinungsdatum",
}

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
		return !(r >= 'a' && r <= 'z') && !(r >= '0' && r <= '9') && r != 'ä' && r != 'ö' && r != 'ü' && r != 'ß'
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

func extractPDFText(data []byte) string {
	if len(data) == 0 {
		return ""
	}
	parts := []string{}
	for _, stream := range pdfStreams(data) {
		if text := pdfContentText(stream); text != "" {
			parts = append(parts, text)
		}
	}
	text := strings.Join(parts, "\n")
	if text == "" {
		text = printablePDFText(data)
	}
	return normalizeWhitespacePreservingLines(repairMojibake(text))
}

func extractPDFTextWithOCRFallback(data []byte, articleNumber string) string {
	text := extractPDFText(data)
	if !needsPDFOCRFallback(text, articleNumber) {
		return text
	}
	if ocrText := pdfOCRTextExtractor(data); ocrText != "" {
		return ocrText
	}
	return text
}

func needsPDFOCRFallback(text, articleNumber string) bool {
	text = normalizeWhitespacePreservingLines(text)
	if text == "" || len([]rune(text)) < 40 {
		return true
	}
	if documentTextMatchesArticleNumber(text, articleNumber) && looksLikeSparePartsDocumentText(text, articleNumber) {
		return false
	}
	lower := strings.ToLower(text)
	return !containsAny(lower, []string{
		"ersatzteile", "ersatzteilliste", "spare parts", "pièces de rechange", "náhradní díly", "et-nr", "spare part n", "bezeichnung / description",
	})
}

func extractPDFOCRText(data []byte) string {
	if pdfOCRDisabled() || len(data) == 0 {
		return ""
	}
	tempDir, err := os.MkdirTemp("", "railkeeper-pdf-ocr-*")
	if err != nil {
		return ""
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	inputPath := filepath.Join(tempDir, "input.pdf")
	if err := os.WriteFile(inputPath, data, 0o600); err != nil {
		return ""
	}
	if text := extractPDFOCRTextWithOCRmyPDF(inputPath, tempDir); text != "" {
		return text
	}
	return extractPDFOCRTextWithTesseract(inputPath, tempDir)
}

func pdfOCRDisabled() bool {
	value := strings.ToLower(strings.TrimSpace(os.Getenv("RAILKEEPER_PDF_OCR")))
	return value == "0" || value == "false" || value == "off" || value == "disabled"
}

func extractPDFOCRTextWithOCRmyPDF(inputPath, tempDir string) string {
	ocrmypdf, err := exec.LookPath("ocrmypdf")
	if err != nil {
		return ""
	}
	outputPath := filepath.Join(tempDir, "ocr.pdf")
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, ocrmypdf, "--skip-text", "--optimize", "0", "--quiet", inputPath, outputPath)
	if err := cmd.Run(); err != nil {
		return ""
	}
	data, err := os.ReadFile(outputPath)
	if err != nil || len(data) == 0 {
		return ""
	}
	return extractPDFText(data)
}

func extractPDFOCRTextWithTesseract(inputPath, tempDir string) string {
	pdftoppm, err := exec.LookPath("pdftoppm")
	if err != nil {
		return ""
	}
	tesseract, err := exec.LookPath("tesseract")
	if err != nil {
		return ""
	}
	pageLimit := pdfOCRPageLimit()
	prefix := filepath.Join(tempDir, "page")
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	render := exec.CommandContext(ctx, pdftoppm, "-r", "200", "-png", "-f", "1", "-l", strconv.Itoa(pageLimit), inputPath, prefix)
	if err := render.Run(); err != nil {
		return ""
	}
	images, err := filepath.Glob(prefix + "-*.png")
	if err != nil || len(images) == 0 {
		return ""
	}
	sort.Strings(images)
	if len(images) > pageLimit {
		images = images[:pageLimit]
	}
	parts := []string{}
	for _, imagePath := range images {
		if text := runTesseractImageOCR(tesseract, imagePath); text != "" {
			parts = append(parts, text)
		}
	}
	return normalizeWhitespacePreservingLines(repairMojibake(strings.Join(parts, "\n")))
}

func pdfOCRPageLimit() int {
	value, err := strconv.Atoi(strings.TrimSpace(os.Getenv("RAILKEEPER_PDF_OCR_MAX_PAGES")))
	if err != nil || value <= 0 {
		return 4
	}
	if value > 12 {
		return 12
	}
	return value
}

func runTesseractImageOCR(tesseract, imagePath string) string {
	for _, language := range []string{"deu+eng", "eng", ""} {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		args := []string{imagePath, "stdout"}
		if language != "" {
			args = append(args, "-l", language)
		}
		args = append(args, "--psm", "6")
		output, err := exec.CommandContext(ctx, tesseract, args...).Output()
		cancel()
		if err != nil {
			continue
		}
		text := normalizeWhitespacePreservingLines(repairMojibake(string(output)))
		if text != "" {
			return text
		}
	}
	return ""
}

func pdfStreams(data []byte) [][]byte {
	streamPattern := regexp.MustCompile(`(?s)(<<.*?>>)\s*stream\r?\n(.*?)\r?\nendstream`)
	matches := streamPattern.FindAllSubmatch(data, -1)
	streams := [][]byte{}
	for _, match := range matches {
		if len(match) < 3 {
			continue
		}
		dictionary := strings.ToLower(string(match[1]))
		stream := bytes.Trim(match[2], "\r\n")
		if strings.Contains(dictionary, "flatedecode") {
			reader, err := zlib.NewReader(bytes.NewReader(stream))
			if err != nil {
				continue
			}
			decoded, err := io.ReadAll(io.LimitReader(reader, 8*1024*1024))
			_ = reader.Close()
			if err == nil {
				streams = append(streams, decoded)
			}
			continue
		}
		streams = append(streams, stream)
	}
	return streams
}

func pdfTextStrings(data []byte) []string {
	raw := string(data)
	out := []string{}
	for _, value := range pdfLiteralStrings(raw) {
		value = normalizeWhitespace(value)
		if value != "" {
			out = append(out, value)
		}
	}
	for _, value := range pdfHexStrings(raw) {
		value = normalizeWhitespace(value)
		if value != "" {
			out = append(out, value)
		}
	}
	return out
}

func pdfContentText(data []byte) string {
	raw := string(data)
	if !strings.Contains(raw, "BT") || (!strings.Contains(raw, "Tj") && !strings.Contains(raw, "TJ")) {
		return ""
	}
	builder := strings.Builder{}
	newLine := func() {
		current := builder.String()
		if current != "" && !strings.HasSuffix(current, "\n") {
			builder.WriteByte('\n')
		}
	}
	space := func() {
		current := builder.String()
		if current != "" && !strings.HasSuffix(current, " ") && !strings.HasSuffix(current, "\n") {
			builder.WriteByte(' ')
		}
	}
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.Contains(line, " Tm") {
			newLine()
		}
		if match := pdfTextMovePattern.FindStringSubmatch(line); len(match) == 3 {
			x := parsePDFNumber(match[1])
			y := parsePDFNumber(match[2])
			if y < -0.2 || x < -15 {
				newLine()
			} else if x > 0.5 {
				space()
			}
		}
		text := pdfShowText(line)
		if text == "" {
			continue
		}
		if builder.Len() > 0 && !strings.HasSuffix(builder.String(), "\n") && needsPDFTextSpace(builder.String(), text) {
			builder.WriteByte(' ')
		}
		builder.WriteString(text)
	}
	return builder.String()
}

var pdfTextMovePattern = regexp.MustCompile(`(-?\d+(?:\.\d+)?)\s+(-?\d+(?:\.\d+)?)\s+Td\b`)

func parsePDFNumber(value string) float64 {
	var result float64
	var sign float64 = 1
	if strings.HasPrefix(value, "-") {
		sign = -1
		value = strings.TrimPrefix(value, "-")
	}
	parts := strings.SplitN(value, ".", 2)
	for _, char := range parts[0] {
		if char >= '0' && char <= '9' {
			result = result*10 + float64(char-'0')
		}
	}
	if len(parts) == 2 {
		scale := 10.0
		for _, char := range parts[1] {
			if char >= '0' && char <= '9' {
				result += float64(char-'0') / scale
				scale *= 10
			}
		}
	}
	return result * sign
}

func needsPDFTextSpace(current, next string) bool {
	last := []rune(strings.TrimRight(current, " \t"))
	first := []rune(strings.TrimLeft(next, " \t"))
	if len(last) == 0 || len(first) == 0 {
		return false
	}
	return isPDFWordRune(last[len(last)-1]) && isPDFWordRune(first[0])
}

func isPDFWordRune(char rune) bool {
	return char >= '0' && char <= '9' || char >= 'A' && char <= 'Z' || char >= 'a' && char <= 'z' || char >= 128
}

func pdfShowText(line string) string {
	if !strings.Contains(line, "Tj") && !strings.Contains(line, "TJ") {
		return ""
	}
	parts := []string{}
	for index := 0; index < len(line); index++ {
		switch line[index] {
		case '(':
			value, next := readPDFLiteral(line, index)
			if next > index {
				parts = append(parts, value)
				index = next - 1
			}
		case '<':
			if index+1 < len(line) && line[index+1] == '<' {
				continue
			}
			value, next := readPDFHex(line, index)
			if next > index {
				parts = append(parts, value)
				index = next - 1
			}
		}
	}
	return strings.Join(parts, "")
}

func readPDFLiteral(raw string, start int) (string, int) {
	depth := 1
	escaped := false
	for index := start + 1; index < len(raw); index++ {
		char := raw[index]
		if escaped {
			escaped = false
			continue
		}
		if char == '\\' {
			escaped = true
			continue
		}
		if char == '(' {
			depth++
			continue
		}
		if char == ')' {
			depth--
			if depth == 0 {
				return decodePDFLiteralString(raw[start+1 : index]), index + 1
			}
		}
	}
	return "", start
}

func readPDFHex(raw string, start int) (string, int) {
	end := strings.IndexByte(raw[start+1:], '>')
	if end < 0 {
		return "", start
	}
	end += start + 1
	cleaned := regexp.MustCompile(`\s+`).ReplaceAllString(raw[start+1:end], "")
	if len(cleaned) < 2 || len(cleaned)%2 == 1 {
		return "", end + 1
	}
	decoded, err := hex.DecodeString(cleaned)
	if err != nil || len(decoded) == 0 {
		return "", end + 1
	}
	return decodePDFHexText(decoded), end + 1
}

func pdfLiteralStrings(raw string) []string {
	out := []string{}
	for index := 0; index < len(raw); index++ {
		if raw[index] != '(' {
			continue
		}
		start := index + 1
		depth := 1
		escaped := false
		for index++; index < len(raw); index++ {
			char := raw[index]
			if escaped {
				escaped = false
				continue
			}
			if char == '\\' {
				escaped = true
				continue
			}
			if char == '(' {
				depth++
				continue
			}
			if char == ')' {
				depth--
				if depth == 0 {
					out = append(out, decodePDFLiteralString(raw[start:index]))
					break
				}
			}
		}
	}
	return out
}

func decodePDFLiteralString(value string) string {
	decoded := []byte{}
	for index := 0; index < len(value); index++ {
		char := value[index]
		if char != '\\' || index+1 >= len(value) {
			decoded = append(decoded, char)
			continue
		}
		index++
		switch value[index] {
		case 'n', 'r', 't':
			decoded = append(decoded, ' ')
		case 'b', 'f':
			continue
		case '(', ')', '\\':
			decoded = append(decoded, value[index])
		default:
			if value[index] >= '0' && value[index] <= '7' {
				end := index + 1
				for end < len(value) && end-index < 3 && value[end] >= '0' && value[end] <= '7' {
					end++
				}
				octal := value[index:end]
				var octalByte byte
				for _, digit := range octal {
					octalByte = octalByte*8 + byte(digit-'0')
				}
				decoded = append(decoded, octalByte)
				index = end - 1
				continue
			}
			decoded = append(decoded, value[index])
		}
	}
	return decodePDFByteString(decoded)
}

func pdfHexStrings(raw string) []string {
	matches := regexp.MustCompile(`<([0-9A-Fa-f\s]{4,})>`).FindAllStringSubmatch(raw, -1)
	out := []string{}
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		cleaned := regexp.MustCompile(`\s+`).ReplaceAllString(match[1], "")
		if len(cleaned)%2 == 1 {
			cleaned += "0"
		}
		decoded, err := hex.DecodeString(cleaned)
		if err != nil || len(decoded) == 0 {
			continue
		}
		out = append(out, decodePDFHexText(decoded))
	}
	return out
}

func decodePDFHexText(data []byte) string {
	if len(data) >= 2 && data[0] == 0xFE && data[1] == 0xFF {
		builder := strings.Builder{}
		for index := 2; index+1 < len(data); index += 2 {
			r := rune(data[index])<<8 | rune(data[index+1])
			if r >= 32 {
				builder.WriteRune(r)
			}
		}
		return builder.String()
	}
	return decodePDFByteString(data)
}

func decodePDFByteString(data []byte) string {
	if utf8.Valid(data) {
		return string(data)
	}
	builder := strings.Builder{}
	for _, char := range data {
		if char == 0 {
			continue
		}
		builder.WriteRune(rune(char))
	}
	return builder.String()
}

func printablePDFText(data []byte) string {
	builder := strings.Builder{}
	for _, char := range string(data) {
		if char == '\n' || char == '\r' || char == '\t' || char >= 32 && char < utf8.RuneSelf {
			builder.WriteRune(char)
		} else {
			builder.WriteRune(' ')
		}
	}
	return builder.String()
}

func normalizeWhitespacePreservingLines(value string) string {
	lines := []string{}
	for _, line := range regexp.MustCompile(`[\n\r]+`).Split(value, -1) {
		line = normalizeWhitespace(line)
		if line != "" {
			lines = append(lines, line)
		}
	}
	return strings.Join(lines, "\n")
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
