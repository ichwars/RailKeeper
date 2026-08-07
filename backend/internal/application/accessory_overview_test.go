package application

import (
	"context"
	"errors"
	"testing"

	"railkeeper/backend/internal/domain"
)

type accessoryCatalogSpy struct {
	AccessoryRepository
	listCalls      int
	listQuery      AccessoryArticleListQuery
	duplicateCalls int
	manufacturer   string
	articleNumber  string
	excludeID      string
}

func (spy *accessoryCatalogSpy) ListArticles(
	_ context.Context,
	query AccessoryArticleListQuery,
) (*AccessoryArticleListResult, error) {
	spy.listCalls++
	spy.listQuery = query
	return &AccessoryArticleListResult{}, nil
}

func (spy *accessoryCatalogSpy) FindDuplicateCandidates(
	_ context.Context,
	manufacturer, articleNumber, excludeID string,
) ([]AccessoryDuplicateCandidate, error) {
	spy.duplicateCalls++
	spy.manufacturer = manufacturer
	spy.articleNumber = articleNumber
	spy.excludeID = excludeID
	return []AccessoryDuplicateCandidate{{ID: "variant-1", Name: "Variant"}}, nil
}

func TestAccessoryOverviewValidatesQueryBeforeRepository(t *testing.T) {
	tests := []AccessoryArticleListQuery{
		{ArticleTypes: []domain.AccessoryArticleType{"vehicle"}},
		{Statuses: []AccessoryArticleStatus{"lost"}},
		{Sort: "name"},
		{Direction: "sideways"},
	}
	for _, query := range tests {
		spy := &accessoryCatalogSpy{}
		service := NewAccessoryService(spy)
		if _, err := service.ListArticles(t.Context(), query); !errors.Is(err, ErrAccessoryValidation) {
			t.Fatalf("ListArticles(%#v) error = %v, want validation", query, err)
		}
		if spy.listCalls != 0 {
			t.Fatalf("invalid query reached repository: %#v", query)
		}
	}
}

func TestAccessoryOverviewNormalizesQuery(t *testing.T) {
	spy := &accessoryCatalogSpy{}
	service := NewAccessoryService(spy)
	_, err := service.ListArticles(t.Context(), AccessoryArticleListQuery{
		Query: "  83125 ", Manufacturer: " Tillig ", Gauges: []string{" TT ", "", "TT"},
		LocationID: " shelf-1 ", Sort: "", Direction: "",
	})
	if err != nil {
		t.Fatal(err)
	}
	query := spy.listQuery
	if query.Query != "83125" || query.Manufacturer != "Tillig" || query.LocationID != "shelf-1" ||
		len(query.Gauges) != 1 || query.Gauges[0] != "TT" || query.Sort != "article" || query.Direction != "asc" {
		t.Fatalf("unexpected normalized query: %#v", query)
	}
}

func TestAccessoryServiceChecksDuplicatesWithoutConflict(t *testing.T) {
	spy := &accessoryCatalogSpy{}
	service := NewAccessoryService(spy)
	result, err := service.CheckDuplicateProducts(t.Context(), AccessoryDuplicateCheckInput{
		Manufacturer: "  Tillig ", ArticleNumber: " 83125 ", ExcludeID: " current-1 ",
	})
	if err != nil {
		t.Fatal(err)
	}
	if spy.manufacturer != "Tillig" || spy.articleNumber != "83125" || spy.excludeID != "current-1" ||
		len(result.Candidates) != 1 {
		t.Fatalf("unexpected duplicate check: %#v, %#v", spy, result)
	}
	if _, err := service.CheckDuplicateProducts(t.Context(), AccessoryDuplicateCheckInput{
		Manufacturer: "Tillig",
	}); !errors.Is(err, ErrAccessoryValidation) || spy.duplicateCalls != 1 {
		t.Fatalf("invalid duplicate check reached repository: calls=%d err=%v", spy.duplicateCalls, err)
	}
}
