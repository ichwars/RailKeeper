package api

import (
	"net/http"

	"railkeeper/backend/internal/application"
)

type planRevisionTransitionInput struct {
	ExpectedVersion int `json:"expectedVersion"`
}

func (a *App) listPlanVariants(w http.ResponseWriter, r *http.Request) {
	variants, err := a.layoutService.ListVariants(r.Context(), r.PathValue("id"))
	if err != nil {
		a.layoutError(w, err, "list plan variants")
		return
	}
	respondJSON(w, http.StatusOK, variants)
}

func (a *App) createPlanVariant(w http.ResponseWriter, r *http.Request) {
	var input application.CreatePlanVariantInput
	if !decodeLayoutJSON(w, r, &input) {
		return
	}
	variant, err := a.layoutService.CreateVariant(r.Context(), r.PathValue("id"), input, actorUserID(r))
	if err != nil {
		a.layoutError(w, err, "create plan variant")
		return
	}
	respondJSON(w, http.StatusCreated, variant)
}

func (a *App) createPlanRevision(w http.ResponseWriter, r *http.Request) {
	var input application.CreatePlanRevisionInput
	if !decodeLayoutJSON(w, r, &input) {
		return
	}
	revision, err := a.layoutService.CreateDraft(r.Context(), r.PathValue("id"), input, actorUserID(r))
	if err != nil {
		a.layoutError(w, err, "create plan revision")
		return
	}
	respondJSON(w, http.StatusCreated, revision)
}

func (a *App) submitPlanRevision(w http.ResponseWriter, r *http.Request) {
	var input planRevisionTransitionInput
	if !decodeLayoutJSON(w, r, &input) {
		return
	}
	revision, err := a.layoutService.SubmitRevision(
		r.Context(), r.PathValue("id"), input.ExpectedVersion, actorUserID(r),
	)
	if err != nil {
		a.layoutError(w, err, "submit plan revision")
		return
	}
	respondJSON(w, http.StatusOK, revision)
}

func (a *App) publishPlanRevision(w http.ResponseWriter, r *http.Request) {
	var input planRevisionTransitionInput
	if !decodeLayoutJSON(w, r, &input) {
		return
	}
	revision, err := a.layoutService.PublishRevision(
		r.Context(), r.PathValue("id"), input.ExpectedVersion, actorUserID(r),
	)
	if err != nil {
		a.layoutError(w, err, "publish plan revision")
		return
	}
	respondJSON(w, http.StatusOK, revision)
}
