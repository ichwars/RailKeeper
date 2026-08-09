package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"railkeeper/backend/internal/application"
)

func (a *App) listLayouts(w http.ResponseWriter, r *http.Request) {
	layouts, err := a.layoutService.ListLayouts(r.Context())
	if err != nil {
		a.layoutError(w, err, "list layouts")
		return
	}
	respondJSON(w, http.StatusOK, layouts)
}

func (a *App) createLayout(w http.ResponseWriter, r *http.Request) {
	var input application.CreateLayoutInput
	if !decodeLayoutJSON(w, r, &input) {
		return
	}
	layout, err := a.layoutService.CreateLayout(r.Context(), input, actorUserID(r))
	if err != nil {
		a.layoutError(w, err, "create layout")
		return
	}
	respondJSON(w, http.StatusCreated, layout)
}

func (a *App) getLayout(w http.ResponseWriter, r *http.Request) {
	layout, err := a.layoutService.GetLayout(r.Context(), r.PathValue("id"))
	if err != nil {
		a.layoutError(w, err, "get layout")
		return
	}
	respondJSON(w, http.StatusOK, layout)
}

func (a *App) getLayoutTwin(w http.ResponseWriter, r *http.Request) {
	twin, err := a.layoutService.GetTwin(r.Context(), r.PathValue("id"), application.LayoutTwinSelection{
		ConfigurationID: r.URL.Query().Get("configurationId"),
		UnitID:          r.URL.Query().Get("unitId"),
	})
	if err != nil {
		a.layoutError(w, err, "get layout twin")
		return
	}
	respondJSON(w, http.StatusOK, twin)
}

func (a *App) updateLayout(w http.ResponseWriter, r *http.Request) {
	var input application.UpdateLayoutInput
	if !decodeLayoutJSON(w, r, &input) {
		return
	}
	layout, err := a.layoutService.UpdateLayout(r.Context(), r.PathValue("id"), input, actorUserID(r))
	if err != nil {
		a.layoutError(w, err, "update layout")
		return
	}
	respondJSON(w, http.StatusOK, layout)
}

func (a *App) listLayoutUnits(w http.ResponseWriter, r *http.Request) {
	units, err := a.layoutService.ListUnits(r.Context(), r.PathValue("id"))
	if err != nil {
		a.layoutError(w, err, "list layout units")
		return
	}
	respondJSON(w, http.StatusOK, units)
}

func (a *App) createLayoutUnit(w http.ResponseWriter, r *http.Request) {
	var input application.CreateLayoutUnitInput
	if !decodeLayoutJSON(w, r, &input) {
		return
	}
	unit, err := a.layoutService.CreateUnit(r.Context(), r.PathValue("id"), input, actorUserID(r))
	if err != nil {
		a.layoutError(w, err, "create layout unit")
		return
	}
	respondJSON(w, http.StatusCreated, unit)
}

func (a *App) updateLayoutUnit(w http.ResponseWriter, r *http.Request) {
	var input application.UpdateLayoutUnitInput
	if !decodeLayoutJSON(w, r, &input) {
		return
	}
	unit, err := a.layoutService.UpdateUnit(r.Context(), r.PathValue("id"), input, actorUserID(r))
	if err != nil {
		a.layoutError(w, err, "update layout unit")
		return
	}
	respondJSON(w, http.StatusOK, unit)
}

func (a *App) listLayoutUnitPorts(w http.ResponseWriter, r *http.Request) {
	ports, err := a.layoutService.ListUnitPorts(r.Context(), r.PathValue("id"))
	if err != nil {
		a.layoutError(w, err, "list layout unit ports")
		return
	}
	respondJSON(w, http.StatusOK, ports)
}

func (a *App) createLayoutUnitPort(w http.ResponseWriter, r *http.Request) {
	var input application.CreateLayoutUnitPortInput
	if !decodeLayoutJSON(w, r, &input) {
		return
	}
	port, err := a.layoutService.CreateUnitPort(r.Context(), r.PathValue("id"), input, actorUserID(r))
	if err != nil {
		a.layoutError(w, err, "create layout unit port")
		return
	}
	respondJSON(w, http.StatusCreated, port)
}

func (a *App) updateLayoutUnitPort(w http.ResponseWriter, r *http.Request) {
	var input application.UpdateLayoutUnitPortInput
	if !decodeLayoutJSON(w, r, &input) {
		return
	}
	port, err := a.layoutService.UpdateUnitPort(r.Context(), r.PathValue("id"), input, actorUserID(r))
	if err != nil {
		a.layoutError(w, err, "update layout unit port")
		return
	}
	respondJSON(w, http.StatusOK, port)
}

func (a *App) listLayoutConfigurations(w http.ResponseWriter, r *http.Request) {
	configurations, err := a.layoutService.ListConfigurations(r.Context(), r.PathValue("id"))
	if err != nil {
		a.layoutError(w, err, "list layout configurations")
		return
	}
	respondJSON(w, http.StatusOK, configurations)
}

func (a *App) createLayoutConfiguration(w http.ResponseWriter, r *http.Request) {
	var input application.SaveLayoutConfigurationInput
	if !decodeLayoutJSON(w, r, &input) {
		return
	}
	input.ID = ""
	configuration, err := a.layoutService.SaveConfiguration(
		r.Context(), r.PathValue("id"), input, actorUserID(r),
	)
	if err != nil {
		a.layoutError(w, err, "create layout configuration")
		return
	}
	respondJSON(w, http.StatusCreated, configuration)
}

func (a *App) updateLayoutConfiguration(w http.ResponseWriter, r *http.Request) {
	var input application.SaveLayoutConfigurationInput
	if !decodeLayoutJSON(w, r, &input) {
		return
	}
	input.ID = r.PathValue("id")
	configuration, err := a.layoutService.SaveConfiguration(
		r.Context(), "", input, actorUserID(r),
	)
	if err != nil {
		a.layoutError(w, err, "update layout configuration")
		return
	}
	respondJSON(w, http.StatusOK, configuration)
}

func decodeLayoutJSON(w http.ResponseWriter, r *http.Request, target any) bool {
	if err := json.NewDecoder(r.Body).Decode(target); err != nil {
		respondProblem(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return false
	}
	return true
}

func (a *App) layoutError(w http.ResponseWriter, err error, action string) {
	switch {
	case errors.Is(err, application.ErrLayoutValidation):
		respondProblem(w, http.StatusBadRequest, "layout_validation", "Layout data is invalid.")
	case errors.Is(err, application.ErrLayoutNotFound):
		respondProblem(w, http.StatusNotFound, "layout_not_found", "Layout resource not found.")
	case errors.Is(err, application.ErrLayoutVersionConflict):
		respondProblem(w, http.StatusConflict, "layout_version_conflict", "Layout data has changed.")
	case errors.Is(err, application.ErrLayoutPositionNotFound):
		respondProblem(w, http.StatusNotFound, "layout_position_not_found", "Layout position not found.")
	case errors.Is(err, application.ErrLayoutPositionProductNotFound):
		respondProblem(w, http.StatusNotFound, "layout_position_product_not_found", "Accessory product not found.")
	case errors.Is(err, application.ErrLayoutPositionVersionConflict):
		respondProblem(w, http.StatusConflict, "layout_position_version_conflict", "Layout position has changed.")
	case errors.Is(err, application.ErrPlanRevisionImmutable):
		respondProblem(w, http.StatusConflict, "plan_revision_immutable", "Published plan revisions are immutable.")
	case errors.Is(err, application.ErrPlanRevisionConflict):
		respondProblem(w, http.StatusConflict, "plan_revision_conflict", "Plan revision has changed.")
	default:
		a.logger.Error("layout operation failed", "action", action, "error", err)
		respondProblem(w, http.StatusInternalServerError, "layout_operation_failed", "Layout operation failed.")
	}
}
