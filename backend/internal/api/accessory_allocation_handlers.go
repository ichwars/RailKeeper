package api

import (
	"net/http"

	"railkeeper/backend/internal/application"
)

func (a *App) listAccessoryReservations(w http.ResponseWriter, r *http.Request) {
	reservations, err := a.accessoryAllocationService.ListReservations(r.Context(), r.URL.Query().Get("productId"))
	if err != nil {
		a.accessoryError(w, err, "list accessory reservations")
		return
	}
	respondJSON(w, http.StatusOK, reservations)
}

func (a *App) createAccessoryReservation(w http.ResponseWriter, r *http.Request) {
	var input application.CreateAccessoryReservationInput
	if !decodeAccessoryJSON(w, r, &input) {
		return
	}
	reservation, err := a.accessoryAllocationService.CreateReservation(r.Context(), input, actorUserID(r))
	if err != nil {
		a.accessoryError(w, err, "create accessory reservation")
		return
	}
	respondJSON(w, http.StatusCreated, reservation)
}

func (a *App) cancelAccessoryReservation(w http.ResponseWriter, r *http.Request) {
	reservation, err := a.accessoryAllocationService.CancelReservation(
		r.Context(), r.PathValue("id"), actorUserID(r),
	)
	if err != nil {
		a.accessoryError(w, err, "cancel accessory reservation")
		return
	}
	respondJSON(w, http.StatusOK, reservation)
}

func (a *App) listAccessoryInstallations(w http.ResponseWriter, r *http.Request) {
	installations, err := a.accessoryAllocationService.ListInstallations(r.Context(), r.URL.Query().Get("productId"))
	if err != nil {
		a.accessoryError(w, err, "list accessory installations")
		return
	}
	respondJSON(w, http.StatusOK, installations)
}

func (a *App) createAccessoryInstallation(w http.ResponseWriter, r *http.Request) {
	var input application.CreateAccessoryInstallationInput
	if !decodeAccessoryJSON(w, r, &input) {
		return
	}
	installation, err := a.accessoryAllocationService.Install(r.Context(), input, actorUserID(r))
	if err != nil {
		a.accessoryError(w, err, "create accessory installation")
		return
	}
	respondJSON(w, http.StatusCreated, installation)
}

func (a *App) removeAccessoryInstallation(w http.ResponseWriter, r *http.Request) {
	var input application.RemoveAccessoryInstallationInput
	if !decodeAccessoryJSON(w, r, &input) {
		return
	}
	installation, err := a.accessoryAllocationService.RemoveInstallation(
		r.Context(), r.PathValue("id"), input, actorUserID(r),
	)
	if err != nil {
		a.accessoryError(w, err, "remove accessory installation")
		return
	}
	respondJSON(w, http.StatusOK, installation)
}

func (a *App) updateAccessoryInstallationCondition(w http.ResponseWriter, r *http.Request) {
	var input application.UpdateAccessoryInstallationConditionInput
	if !decodeAccessoryJSON(w, r, &input) {
		return
	}
	installation, err := a.accessoryAllocationService.UpdateInstallationCondition(
		r.Context(), r.PathValue("id"), input, actorUserID(r),
	)
	if err != nil {
		a.accessoryError(w, err, "update accessory installation condition")
		return
	}
	respondJSON(w, http.StatusOK, installation)
}

func (a *App) getAccessoryAllocationSummary(w http.ResponseWriter, r *http.Request) {
	summary, err := a.accessoryAllocationService.GetAllocationSummary(r.Context(), r.PathValue("id"))
	if err != nil {
		a.accessoryError(w, err, "get accessory allocation summary")
		return
	}
	respondJSON(w, http.StatusOK, summary)
}
