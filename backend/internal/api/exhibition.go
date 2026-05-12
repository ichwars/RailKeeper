package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"railkeeper2/backend/internal/application"
)

type exhibitionLockInput struct {
	Locked bool `json:"locked"`
}

func (a *App) listExhibitionLists(w http.ResponseWriter, r *http.Request) {
	lists, err := a.exhibitionService.List(r.Context())
	if err != nil {
		a.logger.Error("exhibition list failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "exhibition_list_failed", "Could not list exhibition lists.")
		return
	}
	respondJSON(w, http.StatusOK, lists)
}

func (a *App) createExhibitionList(w http.ResponseWriter, r *http.Request) {
	var input application.ExhibitionListInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondProblem(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}
	list, err := a.exhibitionService.Create(r.Context(), input)
	if err != nil {
		handleExhibitionError(a, w, err, "exhibition_create_failed", "Could not create exhibition list.")
		return
	}
	respondJSON(w, http.StatusCreated, list)
}

func (a *App) getExhibitionList(w http.ResponseWriter, r *http.Request) {
	list, err := a.exhibitionService.Get(r.Context(), r.PathValue("id"))
	if err != nil {
		handleExhibitionError(a, w, err, "exhibition_get_failed", "Could not read exhibition list.")
		return
	}
	respondJSON(w, http.StatusOK, list)
}

func (a *App) updateExhibitionList(w http.ResponseWriter, r *http.Request) {
	var input application.ExhibitionListInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondProblem(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}
	list, err := a.exhibitionService.Update(r.Context(), r.PathValue("id"), input)
	if err != nil {
		handleExhibitionError(a, w, err, "exhibition_update_failed", "Could not update exhibition list.")
		return
	}
	respondJSON(w, http.StatusOK, list)
}

func (a *App) deleteExhibitionList(w http.ResponseWriter, r *http.Request) {
	if err := a.exhibitionService.Delete(r.Context(), r.PathValue("id")); err != nil {
		handleExhibitionError(a, w, err, "exhibition_delete_failed", "Could not delete exhibition list.")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *App) setExhibitionListLocked(w http.ResponseWriter, r *http.Request) {
	var input exhibitionLockInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondProblem(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}
	list, err := a.exhibitionService.SetLocked(r.Context(), r.PathValue("id"), input.Locked)
	if err != nil {
		handleExhibitionError(a, w, err, "exhibition_lock_failed", "Could not update exhibition lock state.")
		return
	}
	respondJSON(w, http.StatusOK, list)
}

func (a *App) listExhibitionEntries(w http.ResponseWriter, r *http.Request) {
	entries, err := a.exhibitionService.ListEntries(r.Context(), r.PathValue("id"))
	if err != nil {
		handleExhibitionError(a, w, err, "exhibition_entries_failed", "Could not list exhibition entries.")
		return
	}
	respondJSON(w, http.StatusOK, entries)
}

func (a *App) createExhibitionEntry(w http.ResponseWriter, r *http.Request) {
	var input application.ExhibitionEntryInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondProblem(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}
	entry, err := a.exhibitionService.CreateEntry(r.Context(), r.PathValue("id"), input)
	if err != nil {
		handleExhibitionError(a, w, err, "exhibition_entry_create_failed", "Could not create exhibition entry.")
		return
	}
	respondJSON(w, http.StatusCreated, entry)
}

func (a *App) updateExhibitionEntry(w http.ResponseWriter, r *http.Request) {
	var input application.ExhibitionEntryInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondProblem(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}
	entry, err := a.exhibitionService.UpdateEntry(r.Context(), r.PathValue("id"), r.PathValue("entryID"), input)
	if err != nil {
		handleExhibitionError(a, w, err, "exhibition_entry_update_failed", "Could not update exhibition entry.")
		return
	}
	respondJSON(w, http.StatusOK, entry)
}

func (a *App) deleteExhibitionEntry(w http.ResponseWriter, r *http.Request) {
	if err := a.exhibitionService.DeleteEntry(r.Context(), r.PathValue("id"), r.PathValue("entryID")); err != nil {
		handleExhibitionError(a, w, err, "exhibition_entry_delete_failed", "Could not delete exhibition entry.")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func handleExhibitionError(a *App, w http.ResponseWriter, err error, code, message string) {
	switch {
	case errors.Is(err, application.ErrExhibitionValidation):
		respondProblem(w, http.StatusBadRequest, "exhibition_validation", "Bezeichnung, Datum, Besitzer und Lok-Bezeichnung sind erforderlich.")
	case errors.Is(err, application.ErrExhibitionNotFound):
		respondProblem(w, http.StatusNotFound, "exhibition_not_found", "Exhibition list or entry not found.")
	case errors.Is(err, application.ErrExhibitionLocked):
		respondProblem(w, http.StatusConflict, "exhibition_locked", "Diese Messeliste ist gesperrt.")
	default:
		a.logger.Error(message, "error", err)
		respondProblem(w, http.StatusInternalServerError, code, message)
	}
}
