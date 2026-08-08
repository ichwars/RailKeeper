package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"unicode"
	"unicode/utf8"

	"railkeeper/backend/internal/application"
	"railkeeper/backend/internal/domain"
)

func (a *App) listAccessoryProducts(w http.ResponseWriter, r *http.Request) {
	values, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		a.accessoryError(w, application.ErrAccessoryValidation, "list accessory products")
		return
	}
	query, valid := parseAccessoryArticleListQuery(values)
	if !valid {
		a.accessoryError(w, application.ErrAccessoryValidation, "list accessory products")
		return
	}
	products, err := a.accessoryService.ListArticles(r.Context(), query)
	if err != nil {
		a.accessoryError(w, err, "list accessory products")
		return
	}
	respondJSON(w, http.StatusOK, products)
}

func parseAccessoryArticleListQuery(values url.Values) (application.AccessoryArticleListQuery, bool) {
	query, validQuery := parseAccessoryArticleScalar(values, "query", 200)
	manufacturer, validManufacturer := parseAccessoryArticleScalar(values, "manufacturer", 200)
	locationID, validLocationID := parseAccessoryArticleScalar(values, "locationId", 128)
	if !validQuery || !validManufacturer || !validLocationID {
		return application.AccessoryArticleListQuery{}, false
	}

	articleTypes := make([]domain.AccessoryArticleType, len(values["articleType"]))
	for index, value := range values["articleType"] {
		articleTypes[index] = domain.AccessoryArticleType(value)
	}
	statuses := make([]application.AccessoryArticleStatus, len(values["status"]))
	for index, value := range values["status"] {
		statuses[index] = application.AccessoryArticleStatus(value)
	}
	return application.AccessoryArticleListQuery{
		Query:        query,
		ArticleTypes: articleTypes,
		Gauges:       values["gauge"],
		Statuses:     statuses,
		Manufacturer: manufacturer,
		LocationID:   locationID,
		Sort:         values.Get("sort"),
		Direction:    values.Get("direction"),
	}, true
}

func parseAccessoryArticleScalar(values url.Values, name string, maxRunes int) (string, bool) {
	rawValues, present := values[name]
	if !present {
		return "", true
	}
	if len(rawValues) != 1 {
		return "", false
	}
	rawValue := rawValues[0]
	if !utf8.ValidString(rawValue) || strings.TrimSpace(rawValue) == "" ||
		utf8.RuneCountInString(rawValue) > maxRunes || strings.IndexFunc(rawValue, unicode.IsControl) >= 0 {
		return "", false
	}
	return strings.TrimSpace(rawValue), true
}

func (a *App) checkAccessoryProductDuplicates(w http.ResponseWriter, r *http.Request) {
	var input application.AccessoryDuplicateCheckInput
	if !decodeAccessoryJSON(w, r, &input) {
		return
	}
	result, err := a.accessoryService.CheckDuplicateProducts(r.Context(), input)
	if err != nil {
		a.accessoryError(w, err, "check accessory product duplicates")
		return
	}
	respondJSON(w, http.StatusOK, result)
}

func (a *App) createAccessoryProduct(w http.ResponseWriter, r *http.Request) {
	var input application.CreateAccessoryProductInput
	if !decodeAccessoryJSON(w, r, &input) {
		return
	}
	product, err := a.accessoryService.CreateProduct(r.Context(), input, actorUserID(r))
	if err != nil {
		a.accessoryError(w, err, "create accessory product")
		return
	}
	respondJSON(w, http.StatusCreated, product)
}

func (a *App) getAccessoryProduct(w http.ResponseWriter, r *http.Request) {
	product, err := a.accessoryService.GetProduct(r.Context(), r.PathValue("id"))
	if err != nil {
		a.accessoryError(w, err, "get accessory product")
		return
	}
	respondJSON(w, http.StatusOK, product)
}

func (a *App) updateAccessoryProduct(w http.ResponseWriter, r *http.Request) {
	var input application.UpdateAccessoryProductInput
	if !decodeAccessoryJSON(w, r, &input) {
		return
	}
	product, err := a.accessoryService.UpdateProduct(r.Context(), r.PathValue("id"), input, actorUserID(r))
	if err != nil {
		a.accessoryError(w, err, "update accessory product")
		return
	}
	respondJSON(w, http.StatusOK, product)
}

func (a *App) archiveAccessoryProduct(w http.ResponseWriter, r *http.Request) {
	a.setAccessoryProductArchived(w, r, true)
}

func (a *App) restoreAccessoryProduct(w http.ResponseWriter, r *http.Request) {
	a.setAccessoryProductArchived(w, r, false)
}

func (a *App) setAccessoryProductArchived(w http.ResponseWriter, r *http.Request, archived bool) {
	product, err := a.accessoryService.SetProductArchived(
		r.Context(), r.PathValue("id"), archived, actorUserID(r),
	)
	if err != nil {
		a.accessoryError(w, err, "set accessory product archived")
		return
	}
	respondJSON(w, http.StatusOK, product)
}

func (a *App) getAccessoryStock(w http.ResponseWriter, r *http.Request) {
	stock, err := a.accessoryService.GetStock(r.Context(), r.PathValue("id"))
	if err != nil {
		a.accessoryError(w, err, "get accessory stock")
		return
	}
	respondJSON(w, http.StatusOK, stock)
}

func (a *App) adjustAccessoryStock(w http.ResponseWriter, r *http.Request) {
	var input application.StockAdjustmentInput
	if !decodeAccessoryJSON(w, r, &input) {
		return
	}
	stock, err := a.accessoryService.AdjustStock(r.Context(), r.PathValue("id"), input, actorUserID(r))
	if err != nil {
		a.accessoryError(w, err, "adjust accessory stock")
		return
	}
	respondJSON(w, http.StatusOK, stock)
}

func (a *App) listAccessoryStockMovements(w http.ResponseWriter, r *http.Request) {
	movements, err := a.accessoryService.ListStockMovements(r.Context(), r.PathValue("id"))
	if err != nil {
		a.accessoryError(w, err, "list accessory stock movements")
		return
	}
	respondJSON(w, http.StatusOK, movements)
}

func (a *App) transferAccessoryStock(w http.ResponseWriter, r *http.Request) {
	var input application.TransferAccessoryStockInput
	if !decodeAccessoryJSON(w, r, &input) {
		return
	}
	stock, err := a.accessoryService.TransferStock(
		r.Context(), r.PathValue("id"), input, actorUserID(r),
	)
	if err != nil {
		a.accessoryError(w, err, "transfer accessory stock")
		return
	}
	respondJSON(w, http.StatusOK, stock)
}

func (a *App) listAccessoryPurchases(w http.ResponseWriter, r *http.Request) {
	purchases, err := a.accessoryService.ListPurchases(r.Context(), r.PathValue("id"))
	if err != nil {
		a.accessoryError(w, err, "list accessory purchases")
		return
	}
	respondJSON(w, http.StatusOK, purchases)
}

func (a *App) createAccessoryPurchase(w http.ResponseWriter, r *http.Request) {
	var input application.CreateAccessoryPurchaseInput
	if !decodeAccessoryJSON(w, r, &input) {
		return
	}
	purchase, err := a.accessoryService.CreatePurchase(
		r.Context(), r.PathValue("id"), input, actorUserID(r),
	)
	if err != nil {
		a.accessoryError(w, err, "create accessory purchase")
		return
	}
	respondJSON(w, http.StatusCreated, purchase)
}

func (a *App) individualizeAccessoryProduct(w http.ResponseWriter, r *http.Request) {
	var input application.IndividualizeAccessoryInput
	if !decodeAccessoryJSON(w, r, &input) {
		return
	}
	asset, err := a.accessoryService.Individualize(
		r.Context(), r.PathValue("id"), input, actorUserID(r),
	)
	if err != nil {
		a.accessoryError(w, err, "individualize accessory product")
		return
	}
	respondJSON(w, http.StatusCreated, asset)
}

func (a *App) listAccessoryAssets(w http.ResponseWriter, r *http.Request) {
	assets, err := a.accessoryService.ListAssets(r.Context(), r.PathValue("id"))
	if err != nil {
		a.accessoryError(w, err, "list accessory assets")
		return
	}
	respondJSON(w, http.StatusOK, assets)
}

func (a *App) createAccessoryAsset(w http.ResponseWriter, r *http.Request) {
	var input application.CreateAccessoryAssetInput
	if !decodeAccessoryJSON(w, r, &input) {
		return
	}
	asset, err := a.accessoryService.CreateAsset(r.Context(), r.PathValue("id"), input, actorUserID(r))
	if err != nil {
		a.accessoryError(w, err, "create accessory asset")
		return
	}
	respondJSON(w, http.StatusCreated, asset)
}

func (a *App) updateAccessoryAsset(w http.ResponseWriter, r *http.Request) {
	var input application.UpdateAccessoryAssetInput
	if !decodeAccessoryJSON(w, r, &input) {
		return
	}
	asset, err := a.accessoryService.UpdateAsset(r.Context(), r.PathValue("id"), input, actorUserID(r))
	if err != nil {
		a.accessoryError(w, err, "update accessory asset")
		return
	}
	respondJSON(w, http.StatusOK, asset)
}

func (a *App) listStorageLocations(w http.ResponseWriter, r *http.Request) {
	locations, err := a.accessoryService.ListLocations(r.Context())
	if err != nil {
		a.accessoryError(w, err, "list storage locations")
		return
	}
	respondJSON(w, http.StatusOK, locations)
}

func (a *App) createStorageLocation(w http.ResponseWriter, r *http.Request) {
	var input application.CreateStorageLocationInput
	if !decodeAccessoryJSON(w, r, &input) {
		return
	}
	location, err := a.accessoryService.CreateLocation(r.Context(), input, actorUserID(r))
	if err != nil {
		a.accessoryError(w, err, "create storage location")
		return
	}
	respondJSON(w, http.StatusCreated, location)
}

func (a *App) updateStorageLocation(w http.ResponseWriter, r *http.Request) {
	var input application.UpdateStorageLocationInput
	if !decodeAccessoryJSON(w, r, &input) {
		return
	}
	location, err := a.accessoryService.UpdateLocation(r.Context(), r.PathValue("id"), input, actorUserID(r))
	if err != nil {
		a.accessoryError(w, err, "update storage location")
		return
	}
	respondJSON(w, http.StatusOK, location)
}

func decodeAccessoryJSON(w http.ResponseWriter, r *http.Request, target any) bool {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		respondProblem(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return false
	}
	return true
}

func (a *App) accessoryError(w http.ResponseWriter, err error, action string) {
	switch {
	case errors.Is(err, application.ErrAccessoryValidation):
		message := "Accessory data is invalid."
		if err != application.ErrAccessoryValidation {
			message = err.Error()
		}
		respondProblem(w, http.StatusBadRequest, "accessory_validation", message)
	case errors.Is(err, application.ErrAccessoryNotFound):
		respondProblem(w, http.StatusNotFound, "accessory_not_found", "Accessory resource not found.")
	case errors.Is(err, application.ErrAccessoryConflict):
		respondProblem(w, http.StatusConflict, "accessory_conflict", "Accessory data conflicts with existing data.")
	case errors.Is(err, application.ErrAccessoryInsufficientStock):
		respondProblem(w, http.StatusConflict, "accessory_insufficient_stock", "Insufficient accessory stock.")
	case errors.Is(err, application.ErrAccessoryTrackingMode):
		respondProblem(w, http.StatusConflict, "accessory_tracking_mode", "Operation is invalid for this tracking mode.")
	default:
		a.logger.Error("accessory operation failed", "action", action, "error", err)
		respondProblem(w, http.StatusInternalServerError, "accessory_operation_failed", "Accessory operation failed.")
	}
}
