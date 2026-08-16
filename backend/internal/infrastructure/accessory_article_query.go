package infrastructure

import (
	"context"
	"fmt"
	"strings"

	"railkeeper/backend/internal/application"
	"railkeeper/backend/internal/domain"
)

const accessoryArticleAggregationCTE = `
WITH stock_stats AS (
  SELECT product_id, SUM(quantity) AS stored
  FROM accessory_stock GROUP BY product_id
), asset_stats AS (
  SELECT product_id,
         COUNT(*) AS owned,
         SUM(CASE WHEN storage_location_id IS NOT NULL
                       AND lifecycle_state IN ('stored', 'reserved') THEN 1 ELSE 0 END) AS stored,
         MAX(CASE WHEN condition_state='maintenance_due' THEN 1 ELSE 0 END) AS maintenance_due,
         MAX(CASE WHEN condition_state='defective' THEN 1 ELSE 0 END) AS defective
  FROM accessory_assets GROUP BY product_id
), reservation_stats AS (
  SELECT product_id,
         SUM(CASE WHEN status='active' THEN quantity ELSE 0 END) AS reserved,
         1 AS has_history
  FROM accessory_reservations GROUP BY product_id
), installation_stats AS (
  SELECT product_id,
         SUM(CASE WHEN removed_at IS NULL THEN quantity ELSE 0 END) AS installed,
         SUM(CASE WHEN asset_id IS NULL AND removed_at IS NULL THEN quantity ELSE 0 END) AS quantity_installed,
         MAX(CASE WHEN removed_at IS NULL AND condition_state='maintenance_due' THEN 1 ELSE 0 END) AS maintenance_due,
         MAX(CASE WHEN removed_at IS NULL AND condition_state='defective' THEN 1 ELSE 0 END) AS defective,
         1 AS has_history
  FROM accessory_installations GROUP BY product_id
), product_locations AS (
  SELECT product_id, location_id FROM accessory_stock WHERE quantity > 0
  UNION
  SELECT product_id, storage_location_id FROM accessory_assets WHERE storage_location_id IS NOT NULL
), location_stats AS (
  SELECT locations.product_id,
         json_group_array(locations.name) AS location_names,
         MIN(locations.name COLLATE NOCASE) AS storage_sort
  FROM (
    SELECT DISTINCT links.product_id, storage.name
    FROM product_locations links
    JOIN storage_locations storage ON storage.id=links.location_id
    ORDER BY storage.name COLLATE NOCASE
  ) locations
  GROUP BY locations.product_id
), primary_images AS (
  SELECT product_id, MIN(id) AS document_id
  FROM accessory_documents
  WHERE category='image' AND is_primary=1
  GROUP BY product_id
), article_rows AS (
  SELECT product.id, product.inventory_number, product.manufacturer, product.article_number, product.name,
         product.list_price,
         product.article_type, product.subtype, product.gauges_json, product.inventory_strategy,
         product.archived, product.updated_at, COALESCE(primary_images.document_id, '') AS primary_document_id,
         CASE WHEN primary_images.document_id IS NULL THEN 0 ELSE 1 END AS has_primary_image,
         CASE WHEN product.inventory_strategy='individual'
              THEN COALESCE(asset_stats.stored, 0)
              WHEN product.inventory_strategy='quantity_later_individual'
              THEN COALESCE(stock_stats.stored, 0) + COALESCE(asset_stats.stored, 0)
              ELSE COALESCE(stock_stats.stored, 0) END AS stored,
         COALESCE(reservation_stats.reserved, 0) AS reserved,
         COALESCE(installation_stats.installed, 0) AS installed,
         CASE WHEN product.inventory_strategy='individual'
              THEN COALESCE(asset_stats.owned, 0)
              WHEN product.inventory_strategy='quantity_later_individual'
              THEN COALESCE(stock_stats.stored, 0) + COALESCE(asset_stats.owned, 0)
                   + COALESCE(installation_stats.quantity_installed, 0)
              ELSE COALESCE(stock_stats.stored, 0) + COALESCE(installation_stats.installed, 0) END AS owned,
         MAX(CASE WHEN product.inventory_strategy='individual' THEN COALESCE(asset_stats.stored, 0)
                  WHEN product.inventory_strategy='quantity_later_individual'
                  THEN COALESCE(stock_stats.stored, 0) + COALESCE(asset_stats.stored, 0)
                  ELSE COALESCE(stock_stats.stored, 0) END
             - COALESCE(reservation_stats.reserved, 0), 0) AS available,
         COALESCE(location_stats.location_names, '[]') AS location_names,
         COALESCE(location_stats.storage_sort, '') AS storage_sort,
         CASE WHEN reservation_stats.has_history=1 OR installation_stats.has_history=1 THEN 1 ELSE 0 END
           AS has_usage_history,
         COALESCE(asset_stats.maintenance_due, 0) OR COALESCE(installation_stats.maintenance_due, 0)
           AS maintenance_due,
         COALESCE(asset_stats.defective, 0) OR COALESCE(installation_stats.defective, 0) AS defective,
         (CASE WHEN TRIM(product.manufacturer)='' THEN 1 ELSE 0 END
          + CASE WHEN TRIM(product.article_number)='' THEN 1 ELSE 0 END
          + CASE WHEN TRIM(product.article_type)='' THEN 1 ELSE 0 END
		  + CASE WHEN product.article_type IN (
		                 'track', 'signal', 'decoder', 'electrical_control',
		                 'building_equipment', 'lighting'
		               )
		              AND json_array_length(product.gauges_json)=0 THEN 1 ELSE 0 END
          + CASE WHEN TRIM(product.stock_unit)='' THEN 1 ELSE 0 END) AS care_hint_count
  FROM accessory_products product
  LEFT JOIN stock_stats ON stock_stats.product_id=product.id
  LEFT JOIN asset_stats ON asset_stats.product_id=product.id
  LEFT JOIN reservation_stats ON reservation_stats.product_id=product.id
  LEFT JOIN installation_stats ON installation_stats.product_id=product.id
  LEFT JOIN location_stats ON location_stats.product_id=product.id
  LEFT JOIN primary_images ON primary_images.product_id=product.id
)`

var accessoryArticleSortSQL = map[string]string{
	"article":         "manufacturer COLLATE NOCASE, article_number COLLATE NOCASE, name COLLATE NOCASE",
	"image":           "has_primary_image",
	"inventoryNumber": "inventory_number COLLATE NOCASE",
	"manufacturer":    "manufacturer COLLATE NOCASE",
	"articleNumber":   "article_number COLLATE NOCASE",
	"name":            "name COLLATE NOCASE",
	"type": `CASE article_type WHEN 'track' THEN 1 WHEN 'signal' THEN 2 WHEN 'decoder' THEN 3
                    WHEN 'electrical_control' THEN 4 WHEN 'building_equipment' THEN 5
                    WHEN 'landscape_consumable' THEN 6 WHEN 'lighting' THEN 7 ELSE 8 END, subtype COLLATE NOCASE`,
	"gauge":     "gauges_json COLLATE NOCASE",
	"stock":     "available",
	"storage":   "storage_sort COLLATE NOCASE",
	"updatedAt": "updated_at",
}

func (r *AccessoryRepository) ListArticles(
	ctx context.Context,
	query application.AccessoryArticleListQuery,
) (*application.AccessoryArticleListResult, error) {
	where, args := accessoryArticleWhere(query)
	direction := "ASC"
	if query.Direction == "desc" {
		direction = "DESC"
	}
	sortSQL, valid := accessoryArticleSortSQL[query.Sort]
	if !valid {
		return nil, application.ErrAccessoryValidation
	}
	orderParts := strings.Split(sortSQL, ",")
	for index := range orderParts {
		orderParts[index] = strings.TrimSpace(orderParts[index]) + " " + direction
	}
	orderSQL := strings.Join(orderParts, ", ") + ", id " + direction

	rows, err := r.db.QueryContext(ctx, accessoryArticleAggregationCTE+`
SELECT id, inventory_number, primary_document_id, manufacturer, article_number, name, list_price,
       article_type, subtype, gauges_json, inventory_strategy,
       archived, owned, available, reserved, installed, location_names, has_usage_history,
       care_hint_count, updated_at
FROM article_rows`+where+` ORDER BY `+orderSQL, args...)
	if err != nil {
		return nil, fmt.Errorf("list accessory articles: %w", err)
	}
	defer func() { _ = rows.Close() }()
	result := &application.AccessoryArticleListResult{Items: []application.AccessoryArticleListItem{}}
	productIDs := make([]string, 0)
	for rows.Next() {
		item := application.AccessoryArticleListItem{}
		var gauges, locations, primaryDocumentID string
		var archived, usageHistory int
		if err := rows.Scan(&item.ID, &item.InventoryNumber, &primaryDocumentID, &item.Manufacturer,
			&item.ArticleNumber, &item.Name, &item.ListPrice, &item.ArticleType,
			&item.Subtype, &gauges, &item.InventoryStrategy, &archived, &item.Owned, &item.Available,
			&item.Reserved, &item.Installed, &locations, &usageHistory, &item.CareHintCount,
			&item.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan accessory article: %w", err)
		}
		if err := decodeAccessoryStringArray(gauges, &item.Gauges); err != nil {
			return nil, err
		}
		if err := decodeAccessoryStringArray(locations, &item.LocationNames); err != nil {
			return nil, err
		}
		item.Archived = archived != 0
		item.HasUsageHistory = usageHistory != 0
		if primaryDocumentID != "" {
			item.PrimaryImageURL = fmt.Sprintf(
				"/api/v1/accessory-products/%s/documents/%s/download",
				item.ID,
				primaryDocumentID,
			)
		}
		result.Items = append(result.Items, item)
		productIDs = append(productIDs, item.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate accessory articles: %w", err)
	}
	attributes, err := loadAccessoryAttributes(ctx, r.db, productIDs)
	if err != nil {
		return nil, err
	}
	for index := range result.Items {
		result.Items[index].Attributes = attributes[result.Items[index].ID]
	}
	if err := r.loadAccessoryOverviewMetrics(ctx, &result.Metrics); err != nil {
		return nil, err
	}
	if err := r.loadAccessoryFilterOptions(ctx, &result.FilterOptions); err != nil {
		return nil, err
	}
	return result, nil
}

func accessoryArticleWhere(query application.AccessoryArticleListQuery) (string, []any) {
	clauses := []string{}
	args := []any{}
	archivedRequested := false
	for _, status := range query.Statuses {
		if status == application.AccessoryArticleArchived {
			archivedRequested = true
		}
	}
	if !archivedRequested {
		clauses = append(clauses, "archived=0")
	}
	if query.Query != "" {
		clauses = append(clauses, `(inventory_number LIKE '%' || ? || '%' COLLATE NOCASE
      OR manufacturer LIKE '%' || ? || '%' COLLATE NOCASE
      OR article_number LIKE '%' || ? || '%' COLLATE NOCASE
      OR name LIKE '%' || ? || '%' COLLATE NOCASE
      OR EXISTS(SELECT 1 FROM accessory_products product WHERE product.id=article_rows.id
                AND product.ean LIKE '%' || ? || '%' COLLATE NOCASE))`)
		args = append(args, query.Query, query.Query, query.Query, query.Query, query.Query)
	}
	if query.Manufacturer != "" {
		clauses = append(clauses, "manufacturer=? COLLATE NOCASE")
		args = append(args, query.Manufacturer)
	}
	if len(query.ArticleTypes) > 0 {
		clauses = append(clauses, "article_type IN ("+sqlPlaceholders(len(query.ArticleTypes))+")")
		for _, value := range query.ArticleTypes {
			args = append(args, value)
		}
	}
	if len(query.Gauges) > 0 {
		gaugeClauses := make([]string, len(query.Gauges))
		for index, value := range query.Gauges {
			gaugeClauses[index] = "EXISTS(SELECT 1 FROM json_each(gauges_json) WHERE value=? COLLATE NOCASE)"
			args = append(args, value)
		}
		clauses = append(clauses, "("+strings.Join(gaugeClauses, " OR ")+")")
	}
	if query.LocationID != "" {
		clauses = append(clauses, "EXISTS(SELECT 1 FROM product_locations WHERE product_id=article_rows.id AND location_id=?)")
		args = append(args, query.LocationID)
	}
	if len(query.Statuses) > 0 {
		statuses := make([]string, 0, len(query.Statuses))
		for _, status := range query.Statuses {
			switch status {
			case application.AccessoryArticleAvailable:
				statuses = append(statuses, "available>0")
			case application.AccessoryArticleReserved:
				statuses = append(statuses, "reserved>0")
			case application.AccessoryArticleInstalled:
				statuses = append(statuses, "installed>0")
			case application.AccessoryArticleMaintenanceDue:
				statuses = append(statuses, "maintenance_due=1")
			case application.AccessoryArticleDefective:
				statuses = append(statuses, "defective=1")
			case application.AccessoryArticleArchived:
				statuses = append(statuses, "archived=1")
			}
		}
		clauses = append(clauses, "("+strings.Join(statuses, " OR ")+")")
	}
	if len(clauses) == 0 {
		return "", args
	}
	return " WHERE " + strings.Join(clauses, " AND "), args
}

func sqlPlaceholders(count int) string {
	return strings.TrimSuffix(strings.Repeat("?,", count), ",")
}

func (r *AccessoryRepository) loadAccessoryOverviewMetrics(
	ctx context.Context,
	metrics *application.AccessoryOverviewMetrics,
) error {
	err := r.db.QueryRowContext(ctx, accessoryArticleAggregationCTE+`
SELECT COUNT(*), COUNT(DISTINCT article_type), COALESCE(SUM(available), 0),
       (SELECT COUNT(DISTINCT location_id) FROM product_locations
        JOIN accessory_products product ON product.id=product_locations.product_id WHERE product.archived=0),
       COALESCE(SUM(reserved), 0), COALESCE(SUM(installed), 0), COALESCE(SUM(care_hint_count), 0)
FROM article_rows WHERE archived=0`).Scan(&metrics.ArticleCount, &metrics.ArticleTypeCount, &metrics.Available,
		&metrics.LocationCount, &metrics.Reserved, &metrics.Installed, &metrics.CareHintCount)
	if err != nil {
		return fmt.Errorf("load accessory overview metrics: %w", err)
	}
	return nil
}

func (r *AccessoryRepository) loadAccessoryFilterOptions(
	ctx context.Context,
	options *application.AccessoryArticleFilterOptions,
) error {
	options.Manufacturers = []string{}
	options.ArticleTypes = []domain.AccessoryArticleType{}
	options.Gauges = []string{}
	options.StorageLocations = []application.StorageLocationOption{}
	rows, err := r.db.QueryContext(ctx, `SELECT DISTINCT manufacturer FROM accessory_products
WHERE archived=0 AND TRIM(manufacturer)<>'' ORDER BY manufacturer COLLATE NOCASE`)
	if err != nil {
		return fmt.Errorf("list accessory manufacturer filters: %w", err)
	}
	if err := appendAccessoryStringFilterOptions(
		rows, "accessory manufacturer filters", &options.Manufacturers,
	); err != nil {
		_ = rows.Close()
		return err
	}
	_ = rows.Close()
	rows, err = r.db.QueryContext(ctx, `SELECT DISTINCT article_type FROM accessory_products
WHERE archived=0 AND TRIM(article_type)<>'' ORDER BY article_type COLLATE NOCASE`)
	if err != nil {
		return fmt.Errorf("list accessory type filters: %w", err)
	}
	if err := appendAccessoryStringFilterOptions(
		rows, "accessory article type filters", &options.ArticleTypes,
	); err != nil {
		_ = rows.Close()
		return err
	}
	_ = rows.Close()
	rows, err = r.db.QueryContext(ctx, `SELECT DISTINCT gauge.value
FROM accessory_products product, json_each(product.gauges_json) gauge
WHERE product.archived=0 AND gauge.value IS NOT NULL
ORDER BY gauge.value COLLATE NOCASE`)
	if err != nil {
		return fmt.Errorf("list accessory gauge filters: %w", err)
	}
	if err := appendAccessoryStringFilterOptions(rows, "accessory gauge filters", &options.Gauges); err != nil {
		_ = rows.Close()
		return err
	}
	_ = rows.Close()
	rows, err = r.db.QueryContext(ctx, `SELECT id, name FROM storage_locations
WHERE archived=0 ORDER BY name COLLATE NOCASE, id`)
	if err != nil {
		return fmt.Errorf("list accessory location filters: %w", err)
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var option application.StorageLocationOption
		if err := rows.Scan(&option.ID, &option.Name); err != nil {
			return err
		}
		options.StorageLocations = append(options.StorageLocations, option)
	}
	return rows.Err()
}

type accessoryFilterRows interface {
	Next() bool
	Scan(...any) error
	Err() error
}

func appendAccessoryStringFilterOptions[T ~string](
	rows accessoryFilterRows,
	label string,
	options *[]T,
) error {
	for rows.Next() {
		var value T
		if err := rows.Scan(&value); err != nil {
			return fmt.Errorf("scan %s: %w", label, err)
		}
		*options = append(*options, value)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate %s: %w", label, err)
	}
	return nil
}

func (r *AccessoryRepository) FindDuplicateCandidates(
	ctx context.Context,
	manufacturer, articleNumber, excludeID string,
) ([]application.AccessoryDuplicateCandidate, error) {
	rows, err := r.db.QueryContext(ctx, `
SELECT id, manufacturer, article_number, name, article_type, subtype
FROM accessory_products
WHERE LOWER(TRIM(manufacturer))=LOWER(TRIM(?)) AND LOWER(TRIM(article_number))=LOWER(TRIM(?))
  AND id<>?
ORDER BY name COLLATE NOCASE, id`, manufacturer, articleNumber, excludeID)
	if err != nil {
		return nil, fmt.Errorf("find accessory duplicate candidates: %w", err)
	}
	defer func() { _ = rows.Close() }()
	candidates := []application.AccessoryDuplicateCandidate{}
	for rows.Next() {
		var candidate application.AccessoryDuplicateCandidate
		if err := rows.Scan(&candidate.ID, &candidate.Manufacturer, &candidate.ArticleNumber,
			&candidate.Name, &candidate.ArticleType, &candidate.Subtype); err != nil {
			return nil, fmt.Errorf("scan accessory duplicate candidate: %w", err)
		}
		candidates = append(candidates, candidate)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate accessory duplicate candidates: %w", err)
	}
	return candidates, nil
}
