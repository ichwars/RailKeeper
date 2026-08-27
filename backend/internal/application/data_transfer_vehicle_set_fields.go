package application

var vehicleSetTransferFields = []VehicleTransferField{
	vehicleTransferField("vehicleSetInventoryNumber", "Set-Inventarnummer", "Set inventory number", VehicleTransferString),
	vehicleTransferField("vehicleSetName", "Set-Bezeichnung", "Set name", VehicleTransferString),
	vehicleTransferField("vehicleSetManufacturer", "Set-Hersteller", "Set manufacturer", VehicleTransferString),
	vehicleTransferField("vehicleSetArticleNumber", "Set-Artikelnummer", "Set article number", VehicleTransferString),
	vehicleTransferField("vehicleSetArticleSourceUrl", "Set-Artikelquelle", "Set article source URL", VehicleTransferString),
	vehicleTransferField("vehicleSetGauge", "Set-Spurweite", "Set gauge", VehicleTransferString),
	vehicleTransferField("vehicleSetEpoch", "Set-Epoche", "Set epoch", VehicleTransferString),
	vehicleTransferField("vehicleSetRailwayCompany", "Set-Bahngesellschaft", "Set railway company", VehicleTransferString),
	vehicleTransferField("vehicleSetCategory", "Set-Kategorie", "Set category", VehicleTransferString),
	vehicleTransferField("vehicleSetGattung", "Set-Gattung", "Set class", VehicleTransferString),
	vehicleTransferField("vehicleSetDescription", "Set-Beschreibung", "Set description", VehicleTransferString),
	vehicleTransferField("vehicleSetEAN", "Set-EAN", "Set EAN", VehicleTransferString),
	vehicleTransferField("vehicleSetProductionPeriod", "Set-Produktionszeit", "Set production period", VehicleTransferString),
	vehicleTransferField("vehicleSetListPrice", "Set-Listenpreis", "Set list price", VehicleTransferString),
	vehicleTransferField("vehicleSetAcquisitionType", "Set-Erwerbsart", "Set acquisition type", VehicleTransferString),
	vehicleTransferField("vehicleSetAcquiredFrom", "Set-Erworben von/bei", "Set acquired from", VehicleTransferString),
	vehicleTransferField("vehicleSetPurchasePrice", "Set-Kaufpreis", "Set purchase price", VehicleTransferString),
	vehicleTransferField("vehicleSetPurchaseDate", "Set-Kaufdatum", "Set purchase date", VehicleTransferString),
	vehicleTransferField("vehicleSetStorageLocation", "Set-Lagerort", "Set storage location", VehicleTransferString),
	vehicleTransferField("vehicleSetStorageDetails", "Set-Lagerdetails", "Set storage details", VehicleTransferString),
	vehicleTransferField("vehicleSetCondition", "Set-Zustand", "Set condition", VehicleTransferString),
	vehicleTransferField("vehicleSetConditionDetails", "Set-Zustandsdetails", "Set condition details", VehicleTransferString),
	vehicleTransferField("vehicleSetPackaging", "Set-Verpackung", "Set packaging", VehicleTransferString),
	vehicleTransferField("vehicleSetPosition", "Set-Position", "Set position", VehicleTransferInteger),
	vehicleTransferField("vehicleSetMemberCount", "Set-Mitgliederzahl", "Set member count", VehicleTransferInteger),
	vehicleTransferField("vehicleSetMemberLabel", "Set-Mitgliedsbezeichnung", "Set member label", VehicleTransferString),
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
