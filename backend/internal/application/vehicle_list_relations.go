package application

import (
	"context"
	"encoding/json"

	"railkeeper/backend/internal/decoder/workflow"
	"railkeeper/backend/internal/infrastructure/vehicledata"
)

func (s *VehicleService) attachVehicleRelations(ctx context.Context, vehicles []Vehicle) error {
	if len(vehicles) == 0 {
		return nil
	}
	ids := make([]string, len(vehicles))
	for i, vehicle := range vehicles {
		ids[i] = vehicle.ID
	}
	encodedIDs, err := json.Marshal(ids)
	if err != nil {
		return err
	}
	scope := string(encodedIDs)
	images, err := vehicledata.ReadImagesForVehicles(ctx, s.db, scope)
	if err != nil {
		return err
	}
	attachments, err := vehicledata.ReadAttachmentsForVehicles(ctx, s.db, scope)
	if err != nil {
		return err
	}
	maintenance, err := s.workshopRows.ListMaintenanceForVehicles(ctx, scope)
	if err != nil {
		return err
	}
	spareParts, err := s.workshopRows.ListSparePartsForVehicles(ctx, scope)
	if err != nil {
		return err
	}
	functions, err := vehicledata.ReadFunctionsForVehicles(ctx, s.db, scope)
	if err != nil {
		return err
	}
	cvRecords, err := vehicledata.ReadCVValuesForVehicles(ctx, s.db, scope)
	if err != nil {
		return err
	}
	cvHistories, err := vehicledata.ReadCVHistoriesForVehicles(ctx, s.db, scope)
	if err != nil {
		return err
	}
	cvFiles, err := vehicledata.ReadCVFilesForVehicles(ctx, s.db, scope)
	if err != nil {
		return err
	}
	mappings, err := vehicledata.ReadMappingsForVehicles(ctx, s.db, scope)
	if err != nil {
		return err
	}

	imagesByVehicle := groupVehicleRows(images, func(row VehicleImage) string { return row.VehicleID })
	attachmentsByVehicle := groupVehicleRows(attachments, func(row VehicleAttachment) string { return row.VehicleID })
	maintenanceByVehicle := groupVehicleRows(maintenance, func(row VehicleMaintenance) string { return row.VehicleID })
	sparePartsByVehicle := groupVehicleRows(spareParts, func(row VehicleSparePart) string { return row.VehicleID })
	functionsByVehicle := groupVehicleRows(functions, func(row VehicleFunction) string { return row.VehicleID })
	cvValuesByVehicle := groupVehicleRows(workflow.CVValuesWithHistory(cvRecords, cvHistories),
		func(row VehicleCVValue) string { return row.VehicleID })
	cvFilesByVehicle := groupVehicleRows(cvFiles, func(row VehicleCVFile) string { return row.VehicleID })
	mappingsByVehicle := groupVehicleRows(mappings, func(row VehicleExternalMap) string { return row.VehicleID })
	for i := range vehicles {
		vehicle := &vehicles[i]
		vehicle.Images = imagesByVehicle.forVehicle(vehicle.ID)
		vehicle.ImagesRevision = vehicleImagesRevision(vehicle.Images)
		vehicle.Attachments = attachmentsByVehicle.forVehicle(vehicle.ID)
		vehicle.Maintenance = maintenanceByVehicle.forVehicle(vehicle.ID)
		vehicle.SpareParts = sparePartsByVehicle.forVehicle(vehicle.ID)
		vehicle.Functions = functionsByVehicle.forVehicle(vehicle.ID)
		vehicle.CVValues = cvValuesByVehicle.forVehicle(vehicle.ID)
		vehicle.CVFiles = cvFilesByVehicle.forVehicle(vehicle.ID)
		vehicle.ExternalMappings = mappingsByVehicle.forVehicle(vehicle.ID)
	}
	return nil
}

type vehicleRows[T any] map[string][]T

func groupVehicleRows[T any](rows []T, owner func(T) string) vehicleRows[T] {
	groups := vehicleRows[T]{}
	for _, row := range rows {
		id := owner(row)
		groups[id] = append(groups[id], row)
	}
	return groups
}

func (rows vehicleRows[T]) forVehicle(id string) []T {
	if items, ok := rows[id]; ok {
		return items
	}
	return []T{}
}
