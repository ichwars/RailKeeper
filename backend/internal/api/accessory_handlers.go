package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"railkeeper/backend/internal/application"
)

func (a *App) listAccessoryProducts(w http.ResponseWriter, r *http.Request) {
	products, err := a.accessoryService.ListProducts(r.Context(), r.URL.Query().Get("query"))
	if err != nil {
		a.accessoryError(w, err, "list accessory products")
		return
	}
	respondJSON(w, http.StatusOK, products)
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
		respondProblem(w, http.StatusBadRequest, "accessory_validation", "Accessory data is invalid.")
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
