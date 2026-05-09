package application

import (
	"context"
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

func TestArticleSearchQueryUsesFocusedModelPattern(t *testing.T) {
	query := articleSearchQuery(ArticleSearchInput{
		Manufacturer:  "Piko Spielwaren",
		ArticleNumber: "47284",
		Name:          "V180 4-achsig",
		Gauge:         "TT",
		Fields: map[string]string{
			"epoch":       "IV",
			"description": "soll nicht in die Suchanfrage",
		},
	})

	expected := "V180 4-achsig 47284 Piko Spielwaren TT"
	if query != expected {
		t.Fatalf("expected %q, got %q", expected, query)
	}
}

func TestArticleSearchBoostsManufacturerDomains(t *testing.T) {
	input := ArticleSearchInput{Manufacturer: "Piko", ArticleNumber: "47284", Name: "V180", Gauge: "TT"}
	fields := map[string]ArticleSearchField{"articleNumber": {Label: "Artikel-Nr.", Value: "47284", Confidence: 90}}

	manufacturerScore := scoreArticleResult(input, "Piko V180", "https://www.piko.de/DE/index.php/de/piko-shop.html", "47284 TT", fields)
	shopScore := scoreArticleResult(input, "Piko V180", "https://example-shop.test/piko-v180", "47284 TT", fields)

	if manufacturerScore <= shopScore {
		t.Fatalf("manufacturer domain should rank higher, got manufacturer=%d shop=%d", manufacturerScore, shopScore)
	}
}
