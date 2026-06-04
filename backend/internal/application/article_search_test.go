package application

import (
	"context"
	"strings"
	"testing"
)

type fakeArticleAdapter struct {
	results []ArticleSearchResult
}

func (f fakeArticleAdapter) Search(context.Context, ArticleSearchInput, string) ([]ArticleSearchResult, error) {
	return f.results, nil
}

func TestArticleSearchSortsAndMarksConflicts(t *testing.T) {
	service := &ArticleSearchService{
		adapters: []ArticleSearchAdapter{
			fakeArticleAdapter{results: []ArticleSearchResult{
				{
					Source: "fake",
					Title:  "Weak",
					URL:    "https://example.test/weak",
					Score:  10,
					Fields: map[string]ArticleSearchField{"name": {Label: "Bezeichnung", Value: "Andere Lok"}},
				},
				{
					Source: "fake",
					Title:  "Strong",
					URL:    "https://example.test/strong",
					Score:  30,
					Fields: map[string]ArticleSearchField{"articleNumber": {Label: "Artikel-Nr.", Value: "47284"}},
				},
			}},
		},
		timeout: 0,
	}

	result, err := service.Search(context.Background(), ArticleSearchInput{
		Manufacturer:  "Piko",
		ArticleNumber: "11111",
		Name:          "V180",
		Gauge:         "TT",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Results) != 2 || result.Results[0].Title != "Strong" {
		t.Fatalf("unexpected result order: %#v", result.Results)
	}
	if len(result.Results[0].Conflicts) != 1 || result.Results[0].Conflicts[0] != "articleNumber" {
		t.Fatalf("expected article number conflict, got %#v", result.Results[0].Conflicts)
	}
}

func TestArticleSearchResponseIncludesSearchTrace(t *testing.T) {
	service := &ArticleSearchService{
		adapters: []ArticleSearchAdapter{
			fakeArticleAdapter{results: []ArticleSearchResult{}},
		},
		timeout: 0,
	}

	result, err := service.Search(context.Background(), ArticleSearchInput{
		Manufacturer:     "Acme",
		ArticleNumber:    "12345",
		Name:             "Testlok",
		Gauge:            "H0",
		SearchSources:    []string{"manufacturer", "dealers"},
		PreferredDomains: []string{"https://www.acme.example/produkte"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(result.Sources, ",") != "manufacturer,dealers" {
		t.Fatalf("unexpected sources %#v", result.Sources)
	}
	if len(result.ManufacturerDomains) != 1 || result.ManufacturerDomains[0] != "acme.example" {
		t.Fatalf("unexpected manufacturer domains %#v", result.ManufacturerDomains)
	}
	if len(result.Queries) == 0 || result.Queries[0].Source != "Herstellerseiten" {
		t.Fatalf("expected query trace to start with manufacturer search, got %#v", result.Queries)
	}
}

func TestArticleSearchDetailLoadErrorMessage(t *testing.T) {
	if detailLoadError(nil, "") != "empty response" {
		t.Fatalf("expected empty response marker")
	}
}

func TestArticleSearchEnrichmentPrioritizesCatalogResults(t *testing.T) {
	results := []ArticleSearchResult{}
	for index := 0; index < 8; index++ {
		results = append(results, ArticleSearchResult{
			URL:    "https://example.test/result-" + string(rune('a'+index)),
			Fields: map[string]ArticleSearchField{},
		})
	}
	results = append(results, ArticleSearchResult{
		URL:    "https://www.modellbahn-fokus.de/product/TT/Piko/47302",
		Fields: map[string]ArticleSearchField{},
	})

	indices := articleResultEnrichmentIndices(ArticleSearchInput{Manufacturer: "Piko", ArticleNumber: "47302"}, results)
	if len(indices) == 0 || indices[0] != 8 {
		t.Fatalf("catalog result should be enriched first even when raw rank is lower, got %#v", indices)
	}
}

func TestArticleSearchQueryUsesFocusedModelPattern(t *testing.T) {
	query := articleSearchQuery(ArticleSearchInput{
		Manufacturer:  "Piko Spielwaren",
		ArticleNumber: "47284",
		Name:          "V180 4-achsig",
		Gauge:         "TT",
		Fields: map[string]string{
			"ean":         "4015615472841",
			"epoch":       "IV",
			"description": "soll nicht in die Suchanfrage",
		},
	})

	expected := "V180 4-achsig 47284 4015615472841 Piko Spielwaren TT"
	if query != expected {
		t.Fatalf("expected %q, got %q", expected, query)
	}
}

func TestArticleSearchQueryAllowsEANOnlyPattern(t *testing.T) {
	input := ArticleSearchInput{
		Fields: map[string]string{
			"ean": "4012501136399",
		},
	}
	query := articleSearchQuery(input)

	if query != "4012501136399" {
		t.Fatalf("expected EAN-only query, got %q", query)
	}
	if !isEANOnlyArticleSearch(input, query) {
		t.Fatal("expected EAN-only search to be detected")
	}
}

func TestArticleSearchQueriesPreferFocusedManufacturerAndRawSearch(t *testing.T) {
	input := ArticleSearchInput{
		Manufacturer:  "Tillig",
		ArticleNumber: "13639",
		Name:          "Y-Wagen Nirosta",
		Gauge:         "TT",
		SearchSources: []string{"manufacturer", "web", "wiki"},
	}
	queries := articleSearchQueries(input, articleSearchQuery(input))

	expectedStart := []articleSearchQuerySpec{
		{Query: "13639 Tillig TT site:tillig.com", Source: "Herstellerseiten"},
		{Query: "Y-Wagen Nirosta 13639 Tillig TT site:tillig.com", Source: "Herstellerseiten"},
	}
	for index, expected := range expectedStart {
		if len(queries) <= index || queries[index] != expected {
			t.Fatalf("expected query %d to be %#v, got %#v", index, expected, queries)
		}
	}
	if !containsArticleSearchQuery(queries, articleSearchQuerySpec{Query: "13639 Tillig TT", Source: "DuckDuckGo"}) {
		t.Fatalf("expected focused web query, got %#v", queries)
	}
	if !containsArticleSearchQuery(queries, articleSearchQuerySpec{Query: "Y-Wagen Nirosta 13639 Tillig TT site:modellbau-wiki.de", Source: "Modellbau-Wiki"}) {
		t.Fatalf("expected wiki query, got %#v", queries)
	}
}

func TestArticleSearchMatchesManufacturerAliases(t *testing.T) {
	entries := []MasterDataEntry{
		{
			Key:   "piko-spielwaren",
			Label: "Piko",
			Metadata: map[string]any{
				"aliases": []any{"PIKO", "Piko Spielwaren"},
			},
		},
	}

	entry, ok := matchManufacturerEntry("Piko Spielwaren", entries)
	if !ok || entry.Key != "piko-spielwaren" {
		t.Fatalf("expected alias match for Piko Spielwaren, got ok=%v entry=%#v", ok, entry)
	}
}

func TestArticleSearchRecognizesPriorityURLsInGenericResults(t *testing.T) {
	results := []ArticleSearchResult{{URL: "https://www.modellbahn-fokus.de/product/TT/Piko/47302"}}
	if !hasPriorityArticleURL(ArticleSearchInput{Manufacturer: "Piko"}, results) {
		t.Fatal("Modellbahn-Fokus URLs from generic search should be enriched immediately")
	}
}

func TestArticleSearchRecognizesPrioritySources(t *testing.T) {
	if !isPriorityArticleSource("Modellbahn-Fokus") {
		t.Fatal("Modellbahn-Fokus should be enriched immediately")
	}
	if !isPriorityArticleSource("Herstellerseiten") {
		t.Fatal("manufacturer pages should be enriched immediately")
	}
	if isPriorityArticleSource("DuckDuckGo") {
		t.Fatal("generic web search should not be a priority enrichment source")
	}
}

func TestArticleSearchQueriesIncludeModellbahnFokusCatalog(t *testing.T) {
	input := ArticleSearchInput{
		Manufacturer:  "Piko",
		ArticleNumber: "47284",
		Name:          "V180",
		Gauge:         "TT",
		SearchSources: []string{"catalogs", "web"},
	}
	queries := articleSearchQueries(input, articleSearchQuery(input))

	if !containsArticleSearchQuery(queries, articleSearchQuerySpec{Query: "47284 Piko TT site:modellbahn-fokus.de", Source: "Modellbahn-Fokus"}) {
		t.Fatalf("expected focused Modellbahn-Fokus query, got %#v", queries)
	}
	if !containsArticleSearchQuery(queries, articleSearchQuerySpec{Query: "V180 47284 Piko TT site:modellbahn-fokus.de", Source: "Modellbahn-Fokus"}) {
		t.Fatalf("expected raw Modellbahn-Fokus query, got %#v", queries)
	}
}

func TestArticleSearchQueriesPreferConfiguredManufacturerDomains(t *testing.T) {
	input := ArticleSearchInput{
		Manufacturer:     "Acme",
		ArticleNumber:    "12345",
		Name:             "Testlok",
		Gauge:            "H0",
		SearchSources:    []string{"manufacturer", "web", "wiki"},
		PreferredDomains: []string{"modellbau-wiki.de", "https://www.acme.example/produkte", "shop.acme.example"},
	}
	queries := articleSearchQueries(input, articleSearchQuery(input))

	expectedStart := []articleSearchQuerySpec{
		{Query: "12345 Acme H0 site:acme.example", Source: "Herstellerseiten"},
		{Query: "Testlok 12345 Acme H0 site:acme.example", Source: "Herstellerseiten"},
		{Query: "12345 Acme H0 site:shop.acme.example", Source: "Herstellerseiten"},
		{Query: "Testlok 12345 Acme H0 site:shop.acme.example", Source: "Herstellerseiten"},
	}
	for index, expected := range expectedStart {
		if len(queries) <= index || queries[index] != expected {
			t.Fatalf("expected query %d to be %#v, got %#v", index, expected, queries)
		}
	}
}

func TestArticleSearchDomainsIgnoreImportedReferenceSites(t *testing.T) {
	domains := uniqueDomains([]string{
		"https://www.eisenbahnfreunde-sonneberg.de/",
		"https://forum.spurnull-magazin.de/thread/33000-wer-ist-allmo/",
		"https://www.piko-shop.de/",
	})

	if len(domains) != 1 || domains[0] != "piko-shop.de" {
		t.Fatalf("expected only official manufacturer domain, got %#v", domains)
	}
}

func TestArticleSearchDefaultSourcesDoNotIncludeWiki(t *testing.T) {
	sources := cleanArticleSearchSources(nil)
	for _, source := range sources {
		if source == "wiki" {
			t.Fatalf("wiki should be optional, got default sources %#v", sources)
		}
	}
}

func containsArticleSearchQuery(queries []articleSearchQuerySpec, expected articleSearchQuerySpec) bool {
	for _, query := range queries {
		if query == expected {
			return true
		}
	}
	return false
}

func TestDuckDuckGoSearchURLUsesGermanRegion(t *testing.T) {
	requestURL := duckDuckGoSearchURL("Piko TT")

	if !strings.Contains(requestURL, "kl=de-de") || !strings.Contains(requestURL, "kad=de_DE") {
		t.Fatalf("expected German DuckDuckGo region/language, got %s", requestURL)
	}
}

func TestArticleSearchBoostsManufacturerDomains(t *testing.T) {
	input := ArticleSearchInput{Manufacturer: "Piko", ArticleNumber: "47284", Name: "V180", Gauge: "TT"}
	fields := map[string]ArticleSearchField{"articleNumber": {Label: "Artikel-Nr.", Value: "47284", Confidence: 90}}

	manufacturerScore := scoreArticleResult(input, "Piko V180", "https://www.piko.de/DE/index.php/de/piko-shop.html", "47284 TT", fields)
	marketplaceScore := scoreArticleResult(input, "Piko V180", "https://www.idealo.de/preisvergleich/piko-v180.html", "47284 TT", fields)

	if manufacturerScore <= marketplaceScore {
		t.Fatalf("manufacturer domain should rank higher, got manufacturer=%d marketplace=%d", manufacturerScore, marketplaceScore)
	}
}

func TestArticleSearchRanksCatalogAboveDealerForArticleMatches(t *testing.T) {
	input := ArticleSearchInput{Manufacturer: "Piko", ArticleNumber: "47284", Name: "V180", Gauge: "TT"}
	fields := map[string]ArticleSearchField{"articleNumber": {Label: "Artikel-Nr.", Value: "47284", Confidence: 90}}

	catalogScore := scoreArticleResult(input, "Piko V180", "https://www.modellbahn-fokus.de/product/H0/47284", "47284 TT", fields)
	dealerScore := scoreArticleResult(input, "Piko V180", "https://www.modellbahnshop-lippe.com/piko-47284", "47284 TT", fields)

	if catalogScore <= dealerScore {
		t.Fatalf("catalog should rank above dealer for exact article matches, got catalog=%d dealer=%d", catalogScore, dealerScore)
	}
}

func TestArticleSearchRanksDealerAboveWikiAndMarketplace(t *testing.T) {
	input := ArticleSearchInput{Manufacturer: "Piko", ArticleNumber: "47284", Name: "V180", Gauge: "TT"}
	fields := map[string]ArticleSearchField{"articleNumber": {Label: "Artikel-Nr.", Value: "47284", Confidence: 90}}

	dealerScore := scoreArticleResult(input, "Piko V180", "https://www.modellbahnshop-lippe.com/piko-47284", "47284 TT", fields)
	wikiScore := scoreArticleResult(input, "Piko V180", "https://www.modellbau-wiki.de/wiki/Piko", "47284 TT", fields)
	marketplaceScore := scoreArticleResult(input, "Piko V180", "https://www.ebay.de/itm/piko-47284", "47284 TT", fields)

	if dealerScore <= wikiScore || dealerScore <= marketplaceScore {
		t.Fatalf("dealer should rank above wiki and marketplace, got dealer=%d wiki=%d marketplace=%d", dealerScore, wikiScore, marketplaceScore)
	}
}

func TestArticleSearchBoostsConfiguredManufacturerDomains(t *testing.T) {
	input := ArticleSearchInput{
		Manufacturer:     "Acme",
		ArticleNumber:    "12345",
		Name:             "Testlok",
		Gauge:            "H0",
		PreferredDomains: []string{"acme.example"},
	}
	fields := map[string]ArticleSearchField{"articleNumber": {Label: "Artikel-Nr.", Value: "12345", Confidence: 90}}

	manufacturerScore := scoreArticleResult(input, "Acme Testlok", "https://www.acme.example/modelle/12345", "H0", fields)
	dealerScore := scoreArticleResult(input, "Acme Testlok", "https://shop.example.test/acme-12345", "H0", fields)

	if manufacturerScore <= dealerScore {
		t.Fatalf("configured manufacturer domain should rank higher, got manufacturer=%d dealer=%d", manufacturerScore, dealerScore)
	}
}

func TestArticleSearchPenalizesMissingArticleNumber(t *testing.T) {
	input := ArticleSearchInput{Manufacturer: "Tillig", ArticleNumber: "13639", Name: "Y-Wagen Nirosta", Gauge: "TT"}

	exactScore := scoreArticleResult(input, "Tillig 13639 Y-Wagen Nirosta TT", "https://shop.example.test/tillig-13639.html", "TT Modellwagen", map[string]ArticleSearchField{})
	weakScore := scoreArticleResult(input, "Tillig Y-Wagen Nirosta TT", "https://shop.example.test/tillig-y-wagen.html", "TT Modellwagen", map[string]ArticleSearchField{})

	if exactScore <= weakScore {
		t.Fatalf("article-number match should rank higher, got exact=%d weak=%d", exactScore, weakScore)
	}
}

func TestArticleSearchRequiresArticleNumberForPreferredManufacturerBoost(t *testing.T) {
	input := ArticleSearchInput{Manufacturer: "Piko", ArticleNumber: "47284", Name: "V180", Gauge: "TT"}

	exactDealerScore := scoreArticleResult(input,
		"Piko TT 47284 Diesellok V 180 DR",
		"https://www.conrad.de/de/p/piko-tt-47284-diesellok-v-180.html",
		"Spur TT",
		map[string]ArticleSearchField{"articleNumber": {Label: "Artikel-Nr.", Value: "47284", Confidence: 90}},
	)
	wrongManufacturerScore := scoreArticleResult(input,
		"PIKO Spielwaren GmbH - Spur TT E-Lok BR 243 DR IV #47490ff",
		"https://www.piko.de/DE/index.php/de/piko-news/modellvorstellungen/2250-spur-tt-e-lok-br-243-dr-iv-47490ff.html",
		"Spur TT BR 243 DR",
		map[string]ArticleSearchField{},
	)

	if exactDealerScore <= wrongManufacturerScore {
		t.Fatalf("exact article number should outrank unrelated manufacturer page, got exact=%d wrong=%d", exactDealerScore, wrongManufacturerScore)
	}
}

func TestArticleSearchBoostsExactEANMatches(t *testing.T) {
	input := ArticleSearchInput{Fields: map[string]string{"ean": "4012501136399"}}
	fields := map[string]ArticleSearchField{"ean": {Label: "EAN-Nr.", Value: "4012501136399", Confidence: 60}}

	exactScore := scoreArticleResult(input, "Tillig Y Wagen Nirosta", "https://example.test/13639", "EAN 4012501136399", fields)
	weakScore := scoreArticleResult(input, "Tillig Y Wagen Nirosta", "https://example.test/13639", "passender Modellbahnwagen", map[string]ArticleSearchField{})

	if exactScore <= weakScore {
		t.Fatalf("exact EAN match should rank higher, got exact=%d weak=%d", exactScore, weakScore)
	}
}

func TestBuildArticleFieldsExtractsModellbahnFokusDetails(t *testing.T) {
	input := ArticleSearchInput{Manufacturer: "Piko", ArticleNumber: "47302", Name: "V15", Gauge: "TT"}
	fields := buildArticleFields(input,
		"Piko 47302 Baureihe V 15 V15 2231 Diesellok TT Modellbahn Katalog",
		"https://www.modellbahn-fokus.de/product/TT/Piko/47302",
		`Beschreibung:
		Diesellokomotive Baureihe V 15 der Deutschen Reichsbahn (DR), Epoche III.
		Ausf?hrung in blauer Farbgebung. Betriebsnr.: V15 2231.
		Modell mit PluX16 Schnittstelle f?r Decoder nach NEM 658. Motor mit Schwungmasse.
		Spitzenbeleuchtung wei?/rot, mit der Fahrtrichtung wechselnd.
		Daten & Details:
		Hersteller: Piko
		Art.-Nr. 47302
		EAN: 4015615473022
		Spur TT 1:120
		Bahn-Gesellschaft: DR
		Epoche: III
		Stromsystem DC
		Schnittstelle: Elektrische Schnittstelle f?r Triebfahrzeuge PluX16
		Motor 5-pol. Motor
		Schwungmasse Ja
		Hersteller-Preis: 114,99 ?`,
	)

	if fields["ean"].Value != "4015615473022" {
		t.Fatalf("expected EAN, got %#v", fields["ean"])
	}
	if fields["epoch"].Value != "III" {
		t.Fatalf("expected epoch III, got %#v", fields["epoch"])
	}
	if fields["railwayCompany"].Value != "DR" {
		t.Fatalf("expected DR, got %#v", fields["railwayCompany"])
	}
	if fields["powerPickup"].Value != "DC" {
		t.Fatalf("expected DC, got %#v", fields["powerPickup"])
	}
	if fields["adapter"].Value == "" || !strings.Contains(fields["adapter"].Value, "PluX16") {
		t.Fatalf("expected PluX16 adapter, got %#v", fields["adapter"])
	}
	if fields["driveDescription"].Value != "5-pol" && fields["driveDescription"].Value != "5-pol. Motor" {
		t.Fatalf("expected motor description, got %#v", fields["driveDescription"])
	}
	if fields["headlightsDescription"].Value == "" {
		t.Fatal("expected headlight description")
	}
	if fields["listPrice"].Value != "114.99" {
		t.Fatalf("expected normalized price, got %#v", fields["listPrice"])
	}
	if !strings.Contains(fields["description"].Value, "Diesellokomotive Baureihe V 15") {
		t.Fatalf("expected catalog description, got %#v", fields["description"])
	}
}

func TestBuildArticleFieldsKeepsProductDataClean(t *testing.T) {
	input := ArticleSearchInput{Manufacturer: "Piko", ArticleNumber: "47284", Name: "V 180", Gauge: "TT"}
	fields := buildArticleFields(input,
		"TT Diesellok V 180 DR III, 4-achsig - PIKO Spielwaren GmbH Webshop",
		"https://www.piko-shop.de/de/artikel/tt-diesellok-v180-47284.html",
		`TT Diesellok V 180 kaufen {"mandatory":true,"google_analytics":true}
		 Neuheit 2021: Druckvariante der B 118 als V180 der DR in Epoche III.
		 Mass [mm]: 162. Anzahl Haftreifen: 2. Digitale Schnittstelle: NEM 658 PluX16.
		 Lichtwechsel: Fahrtrichtungsabhaengiger Lichtwechsel weiss / rot.
		 Soundgenerator: Sound laut Artikeldaten.`,
	)

	if fields["name"].Value == "TT Diesellok V 180 DR III, 4-achsig - PIKO Spielwaren GmbH Webshop" {
		t.Fatal("source suffix should not be part of the model name")
	}
	if fields["description"].Value == "" || strings.Contains(fields["description"].Value, "google_analytics") {
		t.Fatalf("unexpected description %q", fields["description"].Value)
	}
	if !strings.HasPrefix(fields["description"].Value, "Neuheit 2021") {
		t.Fatalf("expected manufacturer novelty text, got %q", fields["description"].Value)
	}
	if fields["lengthMm"].Value != "162" {
		t.Fatalf("expected length 162, got %#v", fields["lengthMm"])
	}
	if fields["tractionTireCount"].Value != "2" {
		t.Fatalf("expected two traction tires, got %#v", fields["tractionTireCount"])
	}
	if _, ok := fields["digital"]; ok {
		t.Fatal("digital interface must not be interpreted as digital decoder")
	}
	if fields["adapter"].Value != "NEM 658 PluX16" {
		t.Fatalf("expected adapter information, got %#v", fields["adapter"])
	}
	if fields["headlightsDescription"].Value == "" {
		t.Fatal("expected light description")
	}
	if _, ok := fields["lightingDescription"]; ok {
		t.Fatal("directional headlight description must not be copied to general lighting")
	}
	if fields["soundGeneratorDescription"].Value == "" {
		t.Fatal("expected sound description")
	}
}

func TestBuildArticleFieldsKeepsTechnicalContextSeparate(t *testing.T) {
	input := ArticleSearchInput{Manufacturer: "Piko", ArticleNumber: "47284", Name: "V 180", Gauge: "TT"}
	fields := buildArticleFields(input,
		"TT Diesellok V 180 DR III - PIKO Webshop",
		"https://www.piko-shop.de/de/artikel/tt-diesellok-v180-47284.html",
		`Neuheit 2021: Druckvariante der B 118 als V180 der DR in Epoche III.
		 Maß [mm]: 162. Anzahl Haftreifen: 2.
		 Digitale Schnittstelle: NEM 658 PluX16.
		 Lichtwechsel: Fahrtrichtungsabhängiger Lichtwechsel weiß / rot.
		 Beleuchtung Beschreibung: Fahrtrichtungsabhängiger Lichtwechsel weiß / rot.
		 Sound: PIKO Sound-Modul nachrüstbar #46552 DE | EN Menü Sprunggröße wählen.`,
	)

	if fields["description"].Value != "Neuheit 2021: Druckvariante der B 118 als V180 der DR in Epoche III" {
		t.Fatalf("unexpected description %q", fields["description"].Value)
	}
	if fields["lengthMm"].Value != "162" {
		t.Fatalf("expected length 162, got %#v", fields["lengthMm"])
	}
	if fields["tractionTireCount"].Value != "2" {
		t.Fatalf("expected two traction tires, got %#v", fields["tractionTireCount"])
	}
	if _, ok := fields["digital"]; ok {
		t.Fatal("digital interface must not be interpreted as digital decoder")
	}
	if fields["headlightsDescription"].Value == "" {
		t.Fatal("expected directional headlight description")
	}
	if _, ok := fields["lightingDescription"]; ok {
		t.Fatal("directional headlight description must not be copied to general lighting")
	}
	if fields["soundGeneratorDescription"].Value != "PIKO Sound-Modul nachrüstbar #46552" {
		t.Fatalf("unexpected sound description %q", fields["soundGeneratorDescription"].Value)
	}
}

func TestBuildArticleFieldsRejectsImplausibleLength(t *testing.T) {
	input := ArticleSearchInput{Manufacturer: "Piko", ArticleNumber: "47284", Name: "V 180", Gauge: "TT"}
	fields := buildArticleFields(input,
		"TT Diesellok V 180 DR III",
		"https://shop.example.test/piko-47284.html",
		`Länge (mm): 2026. Maß [mm]: 162.`,
	)

	if fields["lengthMm"].Value != "162" {
		t.Fatalf("expected plausible model length 162, got %#v", fields["lengthMm"])
	}
}

func TestBuildArticleFieldsRejectsWrongContextValues(t *testing.T) {
	input := ArticleSearchInput{Manufacturer: "Piko", ArticleNumber: "47284", Name: "V 180", Gauge: "TT"}
	fields := buildArticleFields(input,
		"TT Diesellok V 180 DR III",
		"https://shop.example.test/piko-47284.html",
		`Länge (mm) 2026 Die Absicht ist, Anzeigen zu zeigen.
		 Beleuchtung Beschreibung: Fahrtrichtungsabhängiger Lichtwechsel weiß / rot.
		 Digitale Schnittstelle: NEM 658 PluX16. Ohne Sound.`,
	)

	if _, ok := fields["lengthMm"]; ok {
		t.Fatalf("year-like value must not be used as length: %#v", fields["lengthMm"])
	}
	if _, ok := fields["description"]; ok {
		t.Fatalf("advertising text must not be used as description: %#v", fields["description"])
	}
	if _, ok := fields["lightingDescription"]; ok {
		t.Fatalf("directional headlight text must not be used as general lighting: %#v", fields["lightingDescription"])
	}
	if _, ok := fields["soundGeneratorEnabled"]; ok {
		t.Fatal("without sound must not enable sound generator")
	}
}

func TestBuildArticleFieldsExtractsTechnicalSentencesWithoutColon(t *testing.T) {
	input := ArticleSearchInput{Manufacturer: "Piko", ArticleNumber: "47284", Name: "V 180", Gauge: "TT"}
	fields := buildArticleFields(input,
		"TT Diesellok V 180 DR III",
		"https://shop.example.test/piko-47284.html",
		`Digitale Schnittstelle NEM 658 PluX16.
		 Fahrtrichtungsabhängiger Lichtwechsel weiß / rot.
		 PIKO Sound-Modul nachrüstbar #46552.`,
	)

	if fields["adapter"].Value != "NEM 658 PluX16" {
		t.Fatalf("expected combined adapter, got %#v", fields["adapter"])
	}
	if fields["headlightsDescription"].Value != "Fahrtrichtungsabhängiger Lichtwechsel weiß / rot" {
		t.Fatalf("unexpected headlight description %q", fields["headlightsDescription"].Value)
	}
	if fields["soundGeneratorDescription"].Value != "PIKO Sound-Modul nachrüstbar #46552" {
		t.Fatalf("unexpected sound description %q", fields["soundGeneratorDescription"].Value)
	}
}

func TestArticleImagesIgnorePlaceholders(t *testing.T) {
	images := articleImagesFromHTML(`
		<img src="/assets/placeholder.png">
		<img src="/assets/no-image.webp">
		<img src="/images/flaggen/i_ital.jpg">
		<img src="/images/Versandkostenfrei_2023_Start_.png">
		<img src="/media/47284-v180-product.jpg">
	`, "https://shop.example.test/product/47284", "Piko V180")

	if len(images) != 1 {
		t.Fatalf("expected one product image, got %#v", images)
	}
	if !strings.Contains(images[0].URL, "47284-v180-product.jpg") {
		t.Fatalf("unexpected image selected: %#v", images)
	}
}

func TestArticleImagesReadLazyImagesAndSrcset(t *testing.T) {
	body := strings.Repeat(`<img src="/assets/logo.png">`, 12) + `
		<img src="/assets/lazy.png" data-src="/media/47284-v180-product.jpg">
		<img src="/assets/lazy.png" data-srcset="/media/47284-v180-detail-small.webp 320w, /media/47284-v180-detail-large.webp 1024w">
	`
	images := articleImagesFromHTML(body, "https://shop.example.test/product/47284", "Piko V180")

	if len(images) != 2 {
		t.Fatalf("expected lazy image and srcset image, got %#v", images)
	}
	if !strings.Contains(images[0].URL, "47284-v180-product.jpg") {
		t.Fatalf("unexpected lazy image selected: %#v", images)
	}
	if !strings.Contains(images[1].URL, "47284-v180-detail-large.webp") {
		t.Fatalf("unexpected srcset image selected: %#v", images)
	}
}

func TestArticleSparePartDescriptionRemovesLabelsAndCartText(t *testing.T) {
	description := cleanArticleSparePartDescription("Nummer: Replacement loudspeaker, oval * Add to shopping cart")

	if description != "Replacement loudspeaker, oval" {
		t.Fatalf("unexpected cleaned description %q", description)
	}
}

func TestArticleSparePartFromRowRejectsDocumentRows(t *testing.T) {
	if _, ok := articleSparePartFromRow("47284 Bedienungsanl./Ersatzteilliste", "https://www.piko-shop.de/download/47284.pdf"); ok {
		t.Fatal("manual/spare-parts-list rows must not be treated as spare parts")
	}
}

func TestArticleSparePartsOnlyScrapeTrustedPageHTML(t *testing.T) {
	input := ArticleSearchInput{Manufacturer: "Piko", ArticleNumber: "47284", Name: "V180", Gauge: "TT"}

	if !shouldExtractPageSpareParts(input, "https://www.piko-shop.de/de/artikel/tt-diesellok-v180-47284.html") {
		t.Fatal("manufacturer pages may expose spare-part HTML")
	}
	if !shouldExtractPageSpareParts(input, "https://www.modellbahn-fokus.de/product/TT/Piko/47284") {
		t.Fatal("catalog pages may expose spare-part HTML")
	}
	if shouldExtractPageSpareParts(input, "https://www.modellbahnshop-lippe.com/piko-47284") {
		t.Fatal("dealer pages must not be scraped for spare-part HTML")
	}
	if shouldExtractPageSpareParts(input, "https://www.ebay.de/itm/piko-47284") {
		t.Fatal("marketplaces must not be scraped for spare-part HTML")
	}
}

func TestSparePartDocumentsPreferManufacturerPDFs(t *testing.T) {
	input := ArticleSearchInput{Manufacturer: "Piko", ArticleNumber: "47284"}
	documents := []ArticleSearchDocument{
		{Title: "Shop Ersatzteile", URL: "https://www.modellbahnshop-lippe.com/download/47284-ersatzteile.pdf", Kind: "spare-parts"},
		{Title: "Manual", URL: "https://www.piko-shop.de/download/47284-manual.pdf", Kind: "manual"},
		{Title: "Ersatzteilliste", URL: "https://www.piko-shop.de/download/47284-ersatzteile.pdf", Kind: "spare-parts"},
		{Title: "Explosionszeichnung", URL: "https://example.test/47284.png", Kind: "document"},
	}

	prioritized := prioritizedSparePartDocuments(input, documents)

	if len(prioritized) != 3 {
		t.Fatalf("expected only spare-part-like documents, got %#v", prioritized)
	}
	if prioritized[0].URL != "https://www.piko-shop.de/download/47284-ersatzteile.pdf" {
		t.Fatalf("manufacturer spare-parts PDF should be first, got %#v", prioritized)
	}
	if prioritized[len(prioritized)-1].URL != "https://www.modellbahnshop-lippe.com/download/47284-ersatzteile.pdf" {
		t.Fatalf("dealer document should stay behind manufacturer documents, got %#v", prioritized)
	}
}

func TestArticleSparePartsFromDocumentTextRequiresMatchingArticleNumber(t *testing.T) {
	text := `
		PIKO Bedienungsanleitung 47284
		47284 Bedienungsanl./Ersatzteilliste
		56026 Haftreifen 10 x 6,4 mm (10 Stueck)
		56333 Nummer: Replacement loudspeaker, oval * Add to shopping cart
	`
	parts := articleSparePartsFromDocumentText(text, "47284", "https://www.piko-shop.de/download/47284.pdf")

	if len(parts) != 2 {
		t.Fatalf("expected two real spare parts, got %#v", parts)
	}
	if parts[0].ArticleNumber != "56026" || parts[0].Description != "Haftreifen 10 x 6,4 mm (10 Stueck)" {
		t.Fatalf("unexpected first part %#v", parts[0])
	}
	if parts[1].ArticleNumber != "56333" || strings.Contains(strings.ToLower(parts[1].Description), "nummer") || strings.Contains(strings.ToLower(parts[1].Description), "shopping cart") {
		t.Fatalf("unexpected cleaned second part %#v", parts[1])
	}

	if parts := articleSparePartsFromDocumentText(text, "99999", "https://www.piko-shop.de/download/47284.pdf"); len(parts) != 0 {
		t.Fatalf("wrong locomotive article number must block document extraction, got %#v", parts)
	}
}

func TestArticleSparePartsFromDocumentTextReadsPikoSpareList(t *testing.T) {
	text := `
		47284 Gleichstrom DC
		ERSATZTEILE DIESELLOKOMOTIVE BR 118_TT
		Bezeichnung / Description ET-Nr. / spare part N° PG*
		Gehäuse, dekoriert (mit Fenster) / Body, decorated (with Windows) 47284-08 13
		Frontfenster / Front windows 47284-10 9
		Klammer / Clamp 47280-37 6
		Kurzkupplung / Short Coupling 46042
		Decodertyp XP 5.1 / Decoder type XP 5.1 56502
	`
	parts := articleSparePartsFromDocumentText(text, "47284", "Ersatzteilliste-47284.pdf")

	numbers := map[string]string{}
	for _, part := range parts {
		numbers[part.ArticleNumber] = part.Description
		if strings.Contains(part.Description, "PG") || strings.HasSuffix(part.Description, " 13") {
			t.Fatalf("description still contains table residue: %#v", part)
		}
	}
	for _, number := range []string{"47284-08", "47284-10", "47280-37", "46042", "56502"} {
		if numbers[number] == "" {
			t.Fatalf("expected spare part %s in %#v", number, parts)
		}
	}
}

func TestPikoSparePartsFromHTMLReadsPriceLinkAndAvailability(t *testing.T) {
	body := `<div class="artikel_ersatzteil__list"><div class="artikel_ersatzteil__list_item">
<div class="artikel_ersatzteil__image"><a href="https://www.piko-shop.de/de/artikel/gehaeuse-bedr.m.fenster-34828.html"><img alt="Geh&auml;use bedr.m.Fenster"></a></div>
<div class="artikel_ersatzteil__description"><h3>Geh&auml;use bedr.m.Fenster</h3> Artikelnummer: ET47284-08<br /></div>
<div class="artikel_ersatzteil__prices"><div class="artikel_ersatzteil__price"> 32,30 € </div>
<span class="element_artikel_delivery__availability_status lieferstatus lieferstatus1 availability1"> weniger als 10 verf&uuml;gbar (Auslieferung erfolgt innerhalb von 3 Werktagen) </span></div>
</div></div>`

	parts := pikoSparePartsFromHTML(body, "https://www.piko-shop.de/de/artikel/ersatzteil/xref_suchtext-47284.html")

	if len(parts) != 1 {
		t.Fatalf("expected one piko spare part, got %#v", parts)
	}
	if parts[0].ArticleNumber != "ET47284-08" || parts[0].Description != "Gehäuse bedr.m.Fenster" || parts[0].Price != "32.30" {
		t.Fatalf("unexpected parsed spare part %#v", parts[0])
	}
	if parts[0].URL != "https://www.piko-shop.de/de/artikel/gehaeuse-bedr.m.fenster-34828.html" {
		t.Fatalf("unexpected link %q", parts[0].URL)
	}
	if !strings.Contains(parts[0].Availability, "weniger als 10 verfügbar") {
		t.Fatalf("unexpected availability %q", parts[0].Availability)
	}
}

func TestRocoSparePartsFromHTMLReadsPriceLinkAndAvailability(t *testing.T) {
	body := `<div class="row table-row-et">
<div class="col-xs-6 col-sm-6 col-md-2 col-lg-2 tdList art-nr" style="text-align: left;">110904</div>
<div class="col-xs-6 col-sm-6 col-md-4 col-lg-4 tdList art-bz" style="text-align: left;">Steuerung kpl. 43202 / 43327</div>
<div class="col-xs-6 col-sm-6 col-md-2 col-lg-2 tdList art-pr" style="text-align: right;">106,00 &euro; </div>
<div class="col-xs-6 col-sm-6 col-md-2 col-lg-2 tdList art-vf" style="text-align: right;">
<img class="produkt-head-verfuegbarkeit" src="/static/verfuegbarkeit/lieferbar-gering.svg" alt="Auslaufartikel" title="Auslaufartikel! (geringe Menge lieferbar)">
</div>
</div>`

	parts := rocoSparePartsFromHTML(body, "https://www.roco.cc/rde/ersatzteile?et=110904")

	if len(parts) != 1 {
		t.Fatalf("expected one roco spare part, got %#v", parts)
	}
	if parts[0].ArticleNumber != "110904" || parts[0].Description != "Steuerung kpl. 43202 / 43327" || parts[0].Price != "106.00" {
		t.Fatalf("unexpected parsed spare part %#v", parts[0])
	}
	if parts[0].URL != "https://www.roco.cc/rde/ersatzteile?et=110904" {
		t.Fatalf("unexpected link %q", parts[0].URL)
	}
	if !strings.Contains(parts[0].Availability, "geringe Menge lieferbar") {
		t.Fatalf("unexpected availability %q", parts[0].Availability)
	}
}

func TestRocoSparePartsFromHTMLReadsLiteralEuroPrice(t *testing.T) {
	body := `<div class="row table-row-et">
<div class="col-xs-6 col-sm-6 col-md-2 col-lg-2 tdList art-nr" style="text-align: left;">120735</div>
<div class="col-xs-6 col-sm-6 col-md-4 col-lg-4 tdList art-bz" style="text-align: left;">Messingscheibe</div>
<div class="col-xs-6 col-sm-6 col-md-2 col-lg-2 tdList art-pr" style="text-align: right;">12,25 € </div>
<img class="produkt-head-verfuegbarkeit" title="Lieferbar">
</div>`

	parts := rocoSparePartsFromHTML(body, "https://www.roco.cc/rde/ersatzteile?et=36270")

	if len(parts) != 1 {
		t.Fatalf("expected one roco spare part, got %#v", parts)
	}
	if parts[0].Price != "12.25" {
		t.Fatalf("expected literal euro price, got %#v", parts[0])
	}
}

func TestArticleSparePartsFromDocumentTextReadsManufacturerFixtures(t *testing.T) {
	tests := []struct {
		name          string
		articleNumber string
		text          string
		expected      map[string]string
	}{
		{
			name:          "Roco",
			articleNumber: "71399",
			text: `
				Roco Ersatzteilliste 71399 Dampflokomotive
				Pos. Art.-Nr. Bezeichnung Preisgruppe
				1 100644 Kupplung komplett 6
				2 129524 Puffer gewölbt 4
				3 85111 Haftreifen 8,4-10,3 mm 3
			`,
			expected: map[string]string{
				"100644": "Kupplung komplett",
				"129524": "Puffer gewölbt",
				"85111":  "Haftreifen 8,4-10,3 mm",
			},
		},
		{
			name:          "Maerklin",
			articleNumber: "39030",
			text: `
				Märklin Ersatzteile 39030
				Bestell-Nr. Bezeichnung
				E123456 Schleifer
				E786790 Haftreifen
				144133 Schraube M1,6 x 5
			`,
			expected: map[string]string{
				"E123456": "Schleifer",
				"E786790": "Haftreifen",
				"144133":  "Schraube M1,6 x 5",
			},
		},
		{
			name:          "Tillig",
			articleNumber: "04930",
			text: `
				TILLIG Ersatzteilliste TT-Modell 04930
				Art.-Nr. Benennung
				300980 Radsatz mit Zahnrad
				322000 Kupplungsaufnahme
				08828 Stromabnehmer rot
			`,
			expected: map[string]string{
				"300980": "Radsatz mit Zahnrad",
				"322000": "Kupplungsaufnahme",
				"08828":  "Stromabnehmer rot",
			},
		},
		{
			name:          "ESU",
			articleNumber: "31033",
			text: `
				ESU spare parts list 31033
				item number description
				51968 LokSound 5 decoder
				54671 Speaker 11 x 15 mm
				50708 Wheel contact spring
			`,
			expected: map[string]string{
				"51968": "LokSound 5 decoder",
				"54671": "Speaker 11 x 15 mm",
				"50708": "Wheel contact spring",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			parts := articleSparePartsFromDocumentText(test.text, test.articleNumber, test.name+".pdf")
			byNumber := map[string]string{}
			for _, part := range parts {
				byNumber[part.ArticleNumber] = part.Description
			}
			for number, description := range test.expected {
				if byNumber[number] != description {
					t.Fatalf("expected %s -> %q, got %#v", number, description, parts)
				}
			}
		})
	}
}

func TestArticleSparePartsFromScannedPDFUsesOCRFallback(t *testing.T) {
	previousExtractor := pdfOCRTextExtractor
	defer func() { pdfOCRTextExtractor = previousExtractor }()

	called := false
	pdfOCRTextExtractor = func(data []byte) string {
		called = true
		return `
			47284 Gleichstrom DC
			ERSATZTEILE DIESELLOKOMOTIVE BR 118_TT
			Gehäuse, dekoriert 47284-08 13
			Frontfenster 47284-10 9
		`
	}

	scannedPDF := []byte("%PDF-1.4\n1 0 obj\n<< /Type /XObject /Subtype /Image /Width 1200 /Height 800 >>\nendobj\n%%EOF")
	parts := ArticleSparePartsFromDocumentData(scannedPDF, "47284", "scan.pdf")

	if !called {
		t.Fatal("expected OCR fallback to be called for PDF without text layer")
	}
	if len(parts) != 2 {
		t.Fatalf("expected OCR spare parts, got %#v", parts)
	}
	if parts[0].ArticleNumber != "47284-08" || parts[1].ArticleNumber != "47284-10" {
		t.Fatalf("unexpected OCR spare parts %#v", parts)
	}
}

func TestArticleSparePartsFromTextPDFSkipsOCRFallback(t *testing.T) {
	previousExtractor := pdfOCRTextExtractor
	defer func() { pdfOCRTextExtractor = previousExtractor }()

	pdfOCRTextExtractor = func(data []byte) string {
		t.Fatal("text PDFs should not call OCR fallback")
		return ""
	}

	textPDF := []byte(`%PDF-1.4
1 0 obj
<< /Length 280 >>
stream
BT
(47284 Gleichstrom DC) Tj
0 -14 Td
(ERSATZTEILE DIESELLOKOMOTIVE BR 118_TT) Tj
0 -14 Td
(Bezeichnung / Description ET-Nr. / spare part N° PG*) Tj
0 -14 Td
(Gehäuse, dekoriert 47284-08 13) Tj
0 -14 Td
(Frontfenster 47284-10 9) Tj
ET
endstream
endobj
%%EOF`)
	parts := ArticleSparePartsFromDocumentData(textPDF, "47284", "text.pdf")

	if len(parts) != 2 {
		t.Fatalf("expected text-layer spare parts, got %#v", parts)
	}
}
