package domain

import "errors"

var ErrVehicleValidation = errors.New("vehicle validation failed")
var ErrVehicleNotFound = errors.New("vehicle not found")

type VehicleMaintenance struct {
	ID              string `json:"id"`
	VehicleID       string `json:"vehicleId"`
	Kind            string `json:"kind"`
	Status          string `json:"status"`
	ConditionRating string `json:"conditionRating,omitempty"`
	DueDate         string `json:"dueDate,omitempty"`
	CompletedAt     string `json:"completedAt,omitempty"`
	Cost            string `json:"cost,omitempty"`
	Notes           string `json:"notes,omitempty"`
	CreatedAt       string `json:"createdAt"`
	UpdatedAt       string `json:"updatedAt"`
}

type VehicleMaintenanceInput struct {
	Kind            string `json:"kind"`
	Status          string `json:"status"`
	ConditionRating string `json:"conditionRating"`
	DueDate         string `json:"dueDate"`
	CompletedAt     string `json:"completedAt"`
	Cost            string `json:"cost"`
	Notes           string `json:"notes"`
}

type VehicleSparePart struct {
	ID            string `json:"id"`
	VehicleID     string `json:"vehicleId"`
	ArticleNumber string `json:"articleNumber"`
	Description   string `json:"description"`
	Price         string `json:"price,omitempty"`
	URL           string `json:"url,omitempty"`
	CreatedAt     string `json:"createdAt"`
	UpdatedAt     string `json:"updatedAt"`
}

type VehicleSparePartInput struct {
	ArticleNumber string `json:"articleNumber"`
	Description   string `json:"description"`
	Price         string `json:"price"`
	URL           string `json:"url"`
}
