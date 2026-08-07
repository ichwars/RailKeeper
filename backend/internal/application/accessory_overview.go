package application

import (
	"context"
	"strings"

	"railkeeper/backend/internal/domain"
)

type AccessoryArticleStatus string

const (
	AccessoryArticleAvailable      AccessoryArticleStatus = "available"
	AccessoryArticleReserved       AccessoryArticleStatus = "reserved"
	AccessoryArticleInstalled      AccessoryArticleStatus = "installed"
	AccessoryArticleMaintenanceDue AccessoryArticleStatus = "maintenance_due"
	AccessoryArticleDefective      AccessoryArticleStatus = "defective"
	AccessoryArticleArchived       AccessoryArticleStatus = "archived"
)

type AccessoryArticleListQuery struct {
	Query        string                        `json:"query,omitempty"`
	ArticleTypes []domain.AccessoryArticleType `json:"articleType,omitempty"`
	Gauges       []string                      `json:"gauge,omitempty"`
	Statuses     []AccessoryArticleStatus      `json:"status,omitempty"`
	Manufacturer string                        `json:"manufacturer,omitempty"`
	LocationID   string                        `json:"locationId,omitempty"`
	Sort         string                        `json:"sort,omitempty"`
	Direction    string                        `json:"direction,omitempty"`
}

type AccessoryArticleListItem struct {
	ID                string                            `json:"id"`
	Manufacturer      string                            `json:"manufacturer"`
	ArticleNumber     string                            `json:"articleNumber"`
	Name              string                            `json:"name"`
	ArticleType       domain.AccessoryArticleType       `json:"articleType"`
	Subtype           string                            `json:"subtype"`
	Gauges            []string                          `json:"gauges"`
	InventoryStrategy domain.AccessoryInventoryStrategy `json:"inventoryStrategy"`
	Archived          bool                              `json:"archived"`
	Owned             int                               `json:"owned"`
	Available         int                               `json:"available"`
	Reserved          int                               `json:"reserved"`
	Installed         int                               `json:"installed"`
	LocationNames     []string                          `json:"locationNames"`
	HasUsageHistory   bool                              `json:"hasUsageHistory"`
	CareHintCount     int                               `json:"careHintCount"`
	UpdatedAt         string                            `json:"updatedAt"`
	Attributes        []domain.AccessoryAttributeValue  `json:"attributes"`
}

type AccessoryOverviewMetrics struct {
	ArticleCount     int `json:"articleCount"`
	ArticleTypeCount int `json:"articleTypeCount"`
	Available        int `json:"available"`
	LocationCount    int `json:"locationCount"`
	Reserved         int `json:"reserved"`
	Installed        int `json:"installed"`
	CareHintCount    int `json:"careHintCount"`
}

type AccessoryArticleFilterOptions struct {
	Manufacturers    []string                      `json:"manufacturers"`
	ArticleTypes     []domain.AccessoryArticleType `json:"articleTypes"`
	Gauges           []string                      `json:"gauges"`
	StorageLocations []StorageLocationOption       `json:"storageLocations"`
}

type StorageLocationOption struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type AccessoryArticleListResult struct {
	Items         []AccessoryArticleListItem    `json:"items"`
	Metrics       AccessoryOverviewMetrics      `json:"metrics"`
	FilterOptions AccessoryArticleFilterOptions `json:"filterOptions"`
}

func (s *AccessoryService) ListArticles(
	ctx context.Context,
	query AccessoryArticleListQuery,
) (*AccessoryArticleListResult, error) {
	query.Query = strings.TrimSpace(query.Query)
	query.Manufacturer = strings.TrimSpace(query.Manufacturer)
	query.LocationID = strings.TrimSpace(query.LocationID)
	query.Gauges = cleanStringArray(query.Gauges)
	if query.Sort == "" {
		query.Sort = "article"
	}
	if query.Direction == "" {
		query.Direction = "asc"
	}
	if !validAccessoryArticleQuery(query) {
		return nil, ErrAccessoryValidation
	}
	return s.repository.ListArticles(ctx, query)
}

func (s *AccessoryService) CheckDuplicateProducts(
	ctx context.Context,
	input AccessoryDuplicateCheckInput,
) (*AccessoryDuplicateCheckResult, error) {
	input.Manufacturer = strings.TrimSpace(input.Manufacturer)
	input.ArticleNumber = strings.TrimSpace(input.ArticleNumber)
	input.ExcludeID = strings.TrimSpace(input.ExcludeID)
	if input.Manufacturer == "" || input.ArticleNumber == "" {
		return nil, ErrAccessoryValidation
	}
	candidates, err := s.repository.FindDuplicateCandidates(
		ctx, input.Manufacturer, input.ArticleNumber, input.ExcludeID,
	)
	if err != nil {
		return nil, err
	}
	return &AccessoryDuplicateCheckResult{Candidates: candidates}, nil
}

func validAccessoryArticleQuery(query AccessoryArticleListQuery) bool {
	for _, articleType := range query.ArticleTypes {
		if !articleType.Valid() {
			return false
		}
	}
	for _, status := range query.Statuses {
		switch status {
		case AccessoryArticleAvailable, AccessoryArticleReserved, AccessoryArticleInstalled,
			AccessoryArticleMaintenanceDue, AccessoryArticleDefective, AccessoryArticleArchived:
		default:
			return false
		}
	}
	switch query.Sort {
	case "article", "type", "gauge", "stock", "storage", "updatedAt":
	default:
		return false
	}
	return query.Direction == "asc" || query.Direction == "desc"
}
