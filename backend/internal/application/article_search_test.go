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
