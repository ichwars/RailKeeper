package policy

var vehicleSetTransferFields = []VehicleTransferField{
	newVehicleTransferField("vehicleSetInventoryNumber", "Set-Inventarnummer", "Set inventory number", VehicleTransferString),
	newVehicleTransferField("vehicleSetName", "Set-Bezeichnung", "Set name", VehicleTransferString),
	newVehicleTransferField("vehicleSetManufacturer", "Set-Hersteller", "Set manufacturer", VehicleTransferString),
	newVehicleTransferField("vehicleSetArticleNumber", "Set-Artikelnummer", "Set article number", VehicleTransferString),
	newVehicleTransferField("vehicleSetArticleSourceUrl", "Set-Artikelquelle", "Set article source URL", VehicleTransferString),
	newVehicleTransferField("vehicleSetGauge", "Set-Spurweite", "Set gauge", VehicleTransferString),
	newVehicleTransferField("vehicleSetEpoch", "Set-Epoche", "Set epoch", VehicleTransferString),
	newVehicleTransferField("vehicleSetRailwayCompany", "Set-Bahngesellschaft", "Set railway company", VehicleTransferString),
	newVehicleTransferField("vehicleSetCategory", "Set-Kategorie", "Set category", VehicleTransferString),
	newVehicleTransferField("vehicleSetGattung", "Set-Gattung", "Set class", VehicleTransferString),
	newVehicleTransferField("vehicleSetDescription", "Set-Beschreibung", "Set description", VehicleTransferString),
	newVehicleTransferField("vehicleSetEAN", "Set-EAN", "Set EAN", VehicleTransferString),
	newVehicleTransferField("vehicleSetProductionPeriod", "Set-Produktionszeit", "Set production period", VehicleTransferString),
	newVehicleTransferField("vehicleSetListPrice", "Set-Listenpreis", "Set list price", VehicleTransferString),
	newVehicleTransferField("vehicleSetAcquisitionType", "Set-Erwerbsart", "Set acquisition type", VehicleTransferString),
	newVehicleTransferField("vehicleSetAcquiredFrom", "Set-Erworben von/bei", "Set acquired from", VehicleTransferString),
	newVehicleTransferField("vehicleSetPurchasePrice", "Set-Kaufpreis", "Set purchase price", VehicleTransferString),
	newVehicleTransferField("vehicleSetPurchaseDate", "Set-Kaufdatum", "Set purchase date", VehicleTransferString),
	newVehicleTransferField("vehicleSetStorageLocation", "Set-Lagerort", "Set storage location", VehicleTransferString),
	newVehicleTransferField("vehicleSetStorageDetails", "Set-Lagerdetails", "Set storage details", VehicleTransferString),
	newVehicleTransferField("vehicleSetCondition", "Set-Zustand", "Set condition", VehicleTransferString),
	newVehicleTransferField("vehicleSetConditionDetails", "Set-Zustandsdetails", "Set condition details", VehicleTransferString),
	newVehicleTransferField("vehicleSetPackaging", "Set-Verpackung", "Set packaging", VehicleTransferString),
	newVehicleTransferField("vehicleSetPosition", "Set-Position", "Set position", VehicleTransferInteger),
	newVehicleTransferField("vehicleSetMemberCount", "Set-Mitgliederzahl", "Set member count", VehicleTransferInteger),
	newVehicleTransferField("vehicleSetMemberLabel", "Set-Mitgliedsbezeichnung", "Set member label", VehicleTransferString),
}

func VehicleSetTransferFields() []VehicleTransferField {
	return cloneVehicleTransferFields(vehicleSetTransferFields)
}

func VehicleCSVTransferFields() []VehicleTransferField {
	fields := make([]VehicleTransferField, 0, len(vehicleTransferFields)+len(vehicleSetTransferFields))
	fields = append(fields, cloneVehicleTransferFields(vehicleTransferFields)...)
	fields = append(fields, cloneVehicleTransferFields(vehicleSetTransferFields)...)
	return fields
}

func VehicleCSVTransferFieldByKey(key string) (VehicleTransferField, bool) {
	for _, field := range VehicleCSVTransferFields() {
		if field.Key == key {
			return field, true
		}
	}
	return VehicleTransferField{}, false
}

func PreserveUnmappedTransferVehicleSetFields(
	incoming TransferVehicleSet,
	existing TransferVehicleSet,
	mapping []DataTransferCSVColumnMapping,
) TransferVehicleSet {
	mappedFields := make(map[string]bool, len(mapping))
	for _, column := range mapping {
		if column.TargetField != "" {
			mappedFields[column.TargetField] = true
		}
	}
	fields := []struct {
		key      string
		incoming *string
		existing string
	}{
		{"vehicleSetInventoryNumber", &incoming.InventoryNumber, existing.InventoryNumber},
		{"vehicleSetName", &incoming.Name, existing.Name},
		{"vehicleSetManufacturer", &incoming.Manufacturer, existing.Manufacturer},
		{"vehicleSetArticleNumber", &incoming.ArticleNumber, existing.ArticleNumber},
		{"vehicleSetArticleSourceUrl", &incoming.ArticleSourceURL, existing.ArticleSourceURL},
		{"vehicleSetGauge", &incoming.Gauge, existing.Gauge},
		{"vehicleSetEpoch", &incoming.Epoch, existing.Epoch},
		{"vehicleSetRailwayCompany", &incoming.RailwayCompany, existing.RailwayCompany},
		{"vehicleSetCategory", &incoming.Category, existing.Category},
		{"vehicleSetGattung", &incoming.Gattung, existing.Gattung},
		{"vehicleSetDescription", &incoming.Description, existing.Description},
		{"vehicleSetEAN", &incoming.EAN, existing.EAN},
		{"vehicleSetProductionPeriod", &incoming.ProductionPeriod, existing.ProductionPeriod},
		{"vehicleSetListPrice", &incoming.ListPrice, existing.ListPrice},
		{"vehicleSetAcquisitionType", &incoming.AcquisitionType, existing.AcquisitionType},
		{"vehicleSetAcquiredFrom", &incoming.AcquiredFrom, existing.AcquiredFrom},
		{"vehicleSetPurchasePrice", &incoming.PurchasePrice, existing.PurchasePrice},
		{"vehicleSetPurchaseDate", &incoming.PurchaseDate, existing.PurchaseDate},
		{"vehicleSetStorageLocation", &incoming.StorageLocation, existing.StorageLocation},
		{"vehicleSetStorageDetails", &incoming.StorageDetails, existing.StorageDetails},
		{"vehicleSetCondition", &incoming.Condition, existing.Condition},
		{"vehicleSetConditionDetails", &incoming.ConditionDetails, existing.ConditionDetails},
		{"vehicleSetPackaging", &incoming.Packaging, existing.Packaging},
	}
	for _, field := range fields {
		if !mappedFields[field.key] {
			*field.incoming = field.existing
		}
	}
	return incoming
}
