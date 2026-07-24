package application

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
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
