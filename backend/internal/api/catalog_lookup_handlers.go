package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"railkeeper/backend/internal/application"
)

func (a *App) searchArticleData(w http.ResponseWriter, r *http.Request) {
	var input application.ArticleSearchInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondProblem(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}

	result, err := a.articleSearch.Search(r.Context(), input)
	if err != nil {
		if errors.Is(err, application.ErrArticleSearchValidation) {
			respondProblem(w, http.StatusBadRequest, "article_search_validation", "At least one search field is required.")
			return
		}
		a.logger.Error("article search failed", "error", err)
		respondProblem(w, http.StatusGatewayTimeout, "article_search_failed", "Artikeldaten-Websuche konnte nicht abgeschlossen werden.")
		return
	}

	respondJSON(w, http.StatusOK, result)
}

func (a *App) listInventoryNumberSchemes(w http.ResponseWriter, r *http.Request) {
	schemes, err := a.inventoryNumbers.List(r.Context())
	if err != nil {
		a.logger.Error("inventory number scheme list failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "inventory_number_scheme_list_failed", "Could not list inventory number schemes.")
		return
	}

	respondJSON(w, http.StatusOK, schemes)
}

func (a *App) createInventoryNumberScheme(w http.ResponseWriter, r *http.Request) {
	var input application.InventoryNumberSchemeCreateInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondProblem(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}

	scheme, err := a.inventoryNumbers.Create(r.Context(), input)
	if err != nil {
		switch {
		case errors.Is(err, application.ErrInventoryNumberValidation):
			respondProblem(w, http.StatusBadRequest, "inventory_number_validation", "Category, prefix, next number and padding are required.")
		case errors.Is(err, application.ErrInventoryNumberConflict):
			respondProblem(w, http.StatusConflict, "inventory_number_scheme_exists", "Inventory number scheme already exists.")
		default:
			a.logger.Error("inventory number scheme create failed", "error", err)
			respondProblem(w, http.StatusInternalServerError, "inventory_number_scheme_create_failed", "Could not create inventory number scheme.")
		}
		return
	}

	respondJSON(w, http.StatusCreated, scheme)
}

func (a *App) updateInventoryNumberScheme(w http.ResponseWriter, r *http.Request) {
	var input application.InventoryNumberSchemeInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondProblem(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}

	scheme, err := a.inventoryNumbers.Update(r.Context(), r.PathValue("category"), input)
	if err != nil {
		switch {
		case errors.Is(err, application.ErrInventoryNumberValidation):
			respondProblem(w, http.StatusBadRequest, "inventory_number_validation", "Prefix, next number and padding are required.")
		case errors.Is(err, application.ErrInventoryNumberNotFound):
			respondProblem(w, http.StatusNotFound, "inventory_number_scheme_not_found", "Inventory number scheme not found.")
		default:
			a.logger.Error("inventory number scheme update failed", "error", err)
			respondProblem(w, http.StatusInternalServerError, "inventory_number_scheme_update_failed", "Could not update inventory number scheme.")
		}
		return
	}

	respondJSON(w, http.StatusOK, scheme)
}
