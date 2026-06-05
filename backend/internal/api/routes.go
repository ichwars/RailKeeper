package api

import "net/http"

type routeSpec struct {
	Method string
	Path   string
}

func apiRouteSpecs() []routeSpec {
	return []routeSpec{
		{"GET", "/health"},
		{"GET", "/api/v1/version"},
		{"GET", "/api/v1/system/storage"},
		{"POST", "/api/v1/system/storage/optimize"},
		{"GET", "/api/v1/system/printers"},
		{"GET", "/api/v1/system/audit-log"},
		{"GET", "/api/v1/system/smtp"},
		{"PUT", "/api/v1/system/smtp"},
		{"POST", "/api/v1/system/smtp/test"},
		{"GET", "/api/v1/system/digital-settings"},
		{"PUT", "/api/v1/system/digital-settings"},
		{"GET", "/api/v1/setup/status"},
		{"POST", "/api/v1/setup/admin"},
		{"POST", "/api/v1/auth/login"},
		{"POST", "/api/v1/auth/password-reset"},
		{"POST", "/api/v1/auth/password-reset/confirm"},
		{"POST", "/api/v1/auth/logout"},
		{"GET", "/api/v1/auth/session"},
		{"GET", "/api/v1/profile/settings"},
		{"PUT", "/api/v1/profile/settings"},
		{"PUT", "/api/v1/auth/password"},
		{"GET", "/api/v1/auth/two-factor"},
		{"POST", "/api/v1/auth/two-factor/setup"},
		{"POST", "/api/v1/auth/two-factor/enable"},
		{"POST", "/api/v1/auth/two-factor/disable"},
		{"GET", "/api/v1/roles"},
		{"GET", "/api/v1/users"},
		{"POST", "/api/v1/users"},
		{"PUT", "/api/v1/users/{id}"},
		{"DELETE", "/api/v1/users/{id}"},
		{"GET", "/api/v1/sessions"},
		{"PUT", "/api/v1/sessions/{id}/revoke"},
		{"POST", "/api/v1/ecos/test"},
		{"POST", "/api/v1/ecos/locomotives/count"},
		{"POST", "/api/v1/ecos/locomotives/raw"},
		{"POST", "/api/v1/digital-centers/ecos/locomotives/sync"},
		{"POST", "/api/v1/digital-centers/z21/test"},
		{"POST", "/api/v1/digital-centers/z21/probe"},
		{"POST", "/api/v1/digital-centers/intellibox3/test"},
		{"POST", "/api/v1/digital-centers/intellibox3/probe"},
		{"POST", "/api/v1/digital-centers/cs3/test"},
		{"GET", "/api/v1/digital-centers/ecos/live/status"},
		{"POST", "/api/v1/digital-centers/ecos/live/start"},
		{"POST", "/api/v1/digital-centers/ecos/live/stop"},
		{"GET", "/api/v1/vehicles"},
		{"POST", "/api/v1/vehicles"},
		{"GET", "/api/v1/vehicles/{id}"},
		{"PUT", "/api/v1/vehicles/{id}"},
		{"DELETE", "/api/v1/vehicles/{id}"},
		{"POST", "/api/v1/vehicles/{id}/external-mappings"},
		{"POST", "/api/v1/vehicles/{id}/images"},
		{"POST", "/api/v1/vehicles/{id}/images/import-url"},
		{"DELETE", "/api/v1/vehicles/{id}/images/{imageID}"},
		{"GET", "/api/v1/vehicles/{id}/images/{imageID}/file"},
		{"GET", "/api/v1/vehicles/{id}/images/{imageID}/thumbnail"},
		{"POST", "/api/v1/vehicles/{id}/attachments"},
		{"PUT", "/api/v1/vehicles/{id}/attachments/{attachmentID}"},
		{"DELETE", "/api/v1/vehicles/{id}/attachments/{attachmentID}"},
		{"GET", "/api/v1/vehicles/{id}/attachments/{attachmentID}/download"},
		{"POST", "/api/v1/vehicles/{id}/attachments/import-url"},
		{"GET", "/api/v1/vehicles/{id}/maintenance"},
		{"POST", "/api/v1/vehicles/{id}/maintenance"},
		{"PUT", "/api/v1/vehicles/{id}/maintenance/{maintenanceID}"},
		{"DELETE", "/api/v1/vehicles/{id}/maintenance/{maintenanceID}"},
		{"GET", "/api/v1/vehicles/{id}/spare-parts"},
		{"GET", "/api/v1/vehicles/{id}/spare-parts/suggestions"},
		{"POST", "/api/v1/vehicles/{id}/spare-parts"},
		{"PUT", "/api/v1/vehicles/{id}/spare-parts/{sparePartID}"},
		{"DELETE", "/api/v1/vehicles/{id}/spare-parts/{sparePartID}"},
		{"GET", "/api/v1/vehicles/{id}/functions"},
		{"PUT", "/api/v1/vehicles/{id}/functions/{functionKey}"},
		{"DELETE", "/api/v1/vehicles/{id}/functions/{functionKey}"},
		{"GET", "/api/v1/vehicles/{id}/cv-values"},
		{"POST", "/api/v1/vehicles/{id}/cv-values"},
		{"PUT", "/api/v1/vehicles/{id}/cv-values/{cvValueID}"},
		{"DELETE", "/api/v1/vehicles/{id}/cv-values/{cvValueID}"},
		{"POST", "/api/v1/cv-files/preview"},
		{"POST", "/api/v1/vehicles/{id}/cv-files"},
		{"DELETE", "/api/v1/vehicles/{id}/cv-files/{cvFileID}"},
		{"GET", "/api/v1/vehicles/{id}/cv-files/{cvFileID}/download"},
		{"POST", "/api/v1/article-search"},
		{"GET", "/api/v1/inventory-number-schemes"},
		{"POST", "/api/v1/inventory-number-schemes"},
		{"PUT", "/api/v1/inventory-number-schemes/{category}"},
		{"GET", "/api/v1/master-data-all"},
		{"GET", "/api/v1/master-data/export"},
		{"POST", "/api/v1/master-data/import"},
		{"GET", "/api/v1/master-data/{type}"},
		{"POST", "/api/v1/master-data/{type}"},
		{"PUT", "/api/v1/master-data/{type}/{key}"},
		{"DELETE", "/api/v1/master-data/{type}/{key}"},
		{"GET", "/api/v1/master-data-relations"},
		{"GET", "/api/v1/backup/export"},
		{"POST", "/api/v1/backup/validate"},
		{"POST", "/api/v1/backup/restore"},
		{"GET", "/api/v1/exhibition-lists"},
		{"POST", "/api/v1/exhibition-lists"},
		{"GET", "/api/v1/exhibition-lists/{id}"},
		{"PUT", "/api/v1/exhibition-lists/{id}"},
		{"DELETE", "/api/v1/exhibition-lists/{id}"},
		{"PUT", "/api/v1/exhibition-lists/{id}/lock"},
		{"GET", "/api/v1/exhibition-lists/{id}/entries"},
		{"POST", "/api/v1/exhibition-lists/{id}/entries"},
		{"PUT", "/api/v1/exhibition-lists/{id}/entries/{entryID}"},
		{"DELETE", "/api/v1/exhibition-lists/{id}/entries/{entryID}"},
	}
}

func (a *App) registerRoutes(mux *http.ServeMux) {
	a.registerHealthRoutes(mux)
	a.registerSystemRoutes(mux)
	a.registerAuthRoutes(mux)
	a.registerVehicleRoutes(mux)
	a.registerImportExportRoutes(mux)
	a.registerExhibitionRoutes(mux)
}

func (a *App) registerHealthRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
}

func (a *App) registerSystemRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/version", a.versionInfo)
	mux.HandleFunc("GET /api/v1/system/storage", a.require("Admin", a.systemStorage))
	mux.HandleFunc("POST /api/v1/system/storage/optimize", a.require("Admin", a.optimizeSystemStorage))
	mux.HandleFunc("GET /api/v1/system/printers", a.require("Admin", a.systemPrinters))
	mux.HandleFunc("GET /api/v1/system/audit-log", a.require("Admin", a.systemAuditLog))
	mux.HandleFunc("GET /api/v1/system/smtp", a.require("Admin", a.getSMTPSettings))
	mux.HandleFunc("PUT /api/v1/system/smtp", a.require("Admin", a.updateSMTPSettings))
	mux.HandleFunc("POST /api/v1/system/smtp/test", a.require("Admin", a.testSMTPSettings))
	mux.HandleFunc("GET /api/v1/system/digital-settings", a.require("Admin", a.getDigitalSettings))
	mux.HandleFunc("PUT /api/v1/system/digital-settings", a.require("Admin", a.updateDigitalSettings))
}

func (a *App) registerAuthRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/setup/status", a.setupStatus)
	mux.HandleFunc("POST /api/v1/setup/admin", a.createAdmin)
	mux.HandleFunc("POST /api/v1/auth/login", a.login)
	mux.HandleFunc("POST /api/v1/auth/password-reset", a.requestPasswordReset)
	mux.HandleFunc("POST /api/v1/auth/password-reset/confirm", a.confirmPasswordReset)
	mux.HandleFunc("POST /api/v1/auth/logout", a.logout)
	mux.HandleFunc("GET /api/v1/auth/session", a.session)
	mux.HandleFunc("GET /api/v1/profile/settings", a.require("Viewer", a.getProfileSettings))
	mux.HandleFunc("PUT /api/v1/profile/settings", a.require("Viewer", a.updateProfileSettings))
	mux.HandleFunc("PUT /api/v1/auth/password", a.require("Viewer", a.changePassword))
	mux.HandleFunc("GET /api/v1/auth/two-factor", a.require("Viewer", a.twoFactorStatus))
	mux.HandleFunc("POST /api/v1/auth/two-factor/setup", a.require("Viewer", a.setupTwoFactor))
	mux.HandleFunc("POST /api/v1/auth/two-factor/enable", a.require("Viewer", a.enableTwoFactor))
	mux.HandleFunc("POST /api/v1/auth/two-factor/disable", a.require("Viewer", a.disableTwoFactor))
	mux.HandleFunc("GET /api/v1/roles", a.require("Admin", a.listRoles))
	mux.HandleFunc("GET /api/v1/users", a.require("Admin", a.listUsers))
	mux.HandleFunc("POST /api/v1/users", a.require("Admin", a.createUser))
	mux.HandleFunc("PUT /api/v1/users/{id}", a.require("Admin", a.updateUser))
	mux.HandleFunc("DELETE /api/v1/users/{id}", a.require("Admin", a.deleteUser))
	mux.HandleFunc("GET /api/v1/sessions", a.require("Admin", a.listSessions))
	mux.HandleFunc("PUT /api/v1/sessions/{id}/revoke", a.require("Admin", a.revokeSession))
}

func (a *App) registerVehicleRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/ecos/test", a.require("Admin", a.testECoSConnection))
	mux.HandleFunc("POST /api/v1/ecos/locomotives/count", a.require("Admin", a.countECoSLocomotives))
	mux.HandleFunc("POST /api/v1/ecos/locomotives/raw", a.require("Admin", a.probeECoSLocomotiveRaw))
	mux.HandleFunc("POST /api/v1/digital-centers/ecos/locomotives/sync", a.require("Admin", a.syncECoSLocomotive))
	mux.HandleFunc("POST /api/v1/digital-centers/z21/test", a.require("Admin", a.testZ21Connection))
	mux.HandleFunc("POST /api/v1/digital-centers/z21/probe", a.require("Admin", a.probeZ21Connection))
	mux.HandleFunc("POST /api/v1/digital-centers/intellibox3/test", a.require("Admin", a.testIntellibox3Connection))
	mux.HandleFunc("POST /api/v1/digital-centers/intellibox3/probe", a.require("Admin", a.probeIntellibox3Connection))
	mux.HandleFunc("POST /api/v1/digital-centers/cs3/test", a.require("Admin", a.testCS3Connection))
	mux.HandleFunc("GET /api/v1/digital-centers/ecos/live/status", a.require("Admin", a.eCoSLiveStatus))
	mux.HandleFunc("POST /api/v1/digital-centers/ecos/live/start", a.require("Admin", a.startECoSLive))
	mux.HandleFunc("POST /api/v1/digital-centers/ecos/live/stop", a.require("Admin", a.stopECoSLive))
	mux.HandleFunc("GET /api/v1/vehicles", a.require("Viewer", a.listVehicles))
	mux.HandleFunc("POST /api/v1/vehicles", a.require("Editor", a.createVehicle))
	mux.HandleFunc("GET /api/v1/vehicles/{id}", a.require("Viewer", a.getVehicle))
	mux.HandleFunc("PUT /api/v1/vehicles/{id}", a.require("Editor", a.updateVehicle))
	mux.HandleFunc("DELETE /api/v1/vehicles/{id}", a.require("Editor", a.deleteVehicle))
	mux.HandleFunc("POST /api/v1/vehicles/{id}/external-mappings", a.require("Editor", a.upsertVehicleExternalMapping))
	mux.HandleFunc("POST /api/v1/vehicles/{id}/images", a.require("Editor", a.uploadVehicleImage))
	mux.HandleFunc("POST /api/v1/vehicles/{id}/images/import-url", a.require("Editor", a.importVehicleImageFromURL))
	mux.HandleFunc("DELETE /api/v1/vehicles/{id}/images/{imageID}", a.require("Editor", a.deleteVehicleImage))
	mux.HandleFunc("GET /api/v1/vehicles/{id}/images/{imageID}/file", a.require("Viewer", a.downloadVehicleImage))
	mux.HandleFunc("GET /api/v1/vehicles/{id}/images/{imageID}/thumbnail", a.require("Viewer", a.downloadVehicleImageThumbnail))
	mux.HandleFunc("POST /api/v1/vehicles/{id}/attachments", a.require("Editor", a.uploadVehicleAttachment))
	mux.HandleFunc("PUT /api/v1/vehicles/{id}/attachments/{attachmentID}", a.require("Editor", a.updateVehicleAttachment))
	mux.HandleFunc("DELETE /api/v1/vehicles/{id}/attachments/{attachmentID}", a.require("Editor", a.deleteVehicleAttachment))
	mux.HandleFunc("GET /api/v1/vehicles/{id}/attachments/{attachmentID}/download", a.require("Viewer", a.downloadVehicleAttachment))
	mux.HandleFunc("POST /api/v1/vehicles/{id}/attachments/import-url", a.require("Editor", a.importVehicleAttachmentFromURL))
	mux.HandleFunc("GET /api/v1/vehicles/{id}/maintenance", a.require("Viewer", a.listVehicleMaintenance))
	mux.HandleFunc("POST /api/v1/vehicles/{id}/maintenance", a.require("Editor", a.createVehicleMaintenance))
	mux.HandleFunc("PUT /api/v1/vehicles/{id}/maintenance/{maintenanceID}", a.require("Editor", a.updateVehicleMaintenance))
	mux.HandleFunc("DELETE /api/v1/vehicles/{id}/maintenance/{maintenanceID}", a.require("Editor", a.deleteVehicleMaintenance))
	mux.HandleFunc("GET /api/v1/vehicles/{id}/spare-parts", a.require("Viewer", a.listVehicleSpareParts))
	mux.HandleFunc("GET /api/v1/vehicles/{id}/spare-parts/suggestions", a.require("Viewer", a.suggestVehicleSpareParts))
	mux.HandleFunc("POST /api/v1/vehicles/{id}/spare-parts", a.require("Editor", a.createVehicleSparePart))
	mux.HandleFunc("PUT /api/v1/vehicles/{id}/spare-parts/{sparePartID}", a.require("Editor", a.updateVehicleSparePart))
	mux.HandleFunc("DELETE /api/v1/vehicles/{id}/spare-parts/{sparePartID}", a.require("Editor", a.deleteVehicleSparePart))
	mux.HandleFunc("GET /api/v1/vehicles/{id}/functions", a.require("Viewer", a.listVehicleFunctions))
	mux.HandleFunc("PUT /api/v1/vehicles/{id}/functions/{functionKey}", a.require("Editor", a.upsertVehicleFunction))
	mux.HandleFunc("DELETE /api/v1/vehicles/{id}/functions/{functionKey}", a.require("Editor", a.deleteVehicleFunction))
	mux.HandleFunc("GET /api/v1/vehicles/{id}/cv-values", a.require("Viewer", a.listVehicleCVValues))
	mux.HandleFunc("POST /api/v1/vehicles/{id}/cv-values", a.require("Editor", a.createVehicleCVValue))
	mux.HandleFunc("PUT /api/v1/vehicles/{id}/cv-values/{cvValueID}", a.require("Editor", a.updateVehicleCVValue))
	mux.HandleFunc("DELETE /api/v1/vehicles/{id}/cv-values/{cvValueID}", a.require("Editor", a.deleteVehicleCVValue))
	mux.HandleFunc("POST /api/v1/cv-files/preview", a.require("Editor", a.previewVehicleCVFile))
	mux.HandleFunc("POST /api/v1/vehicles/{id}/cv-files", a.require("Editor", a.uploadVehicleCVFile))
	mux.HandleFunc("DELETE /api/v1/vehicles/{id}/cv-files/{cvFileID}", a.require("Editor", a.deleteVehicleCVFile))
	mux.HandleFunc("GET /api/v1/vehicles/{id}/cv-files/{cvFileID}/download", a.require("Viewer", a.downloadVehicleCVFile))
}

func (a *App) registerImportExportRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/article-search", a.require("Viewer", a.searchArticleData))
	mux.HandleFunc("GET /api/v1/inventory-number-schemes", a.require("Viewer", a.listInventoryNumberSchemes))
	mux.HandleFunc("POST /api/v1/inventory-number-schemes", a.require("Editor", a.createInventoryNumberScheme))
	mux.HandleFunc("PUT /api/v1/inventory-number-schemes/{category}", a.require("Editor", a.updateInventoryNumberScheme))
	mux.HandleFunc("GET /api/v1/master-data-all", a.require("Viewer", a.listAllMasterData))
	mux.HandleFunc("GET /api/v1/master-data/export", a.require("Admin", a.exportMasterData))
	mux.HandleFunc("POST /api/v1/master-data/import", a.require("Admin", a.importMasterData))
	mux.HandleFunc("GET /api/v1/master-data/{type}", a.requireMasterDataRead(a.listMasterData))
	mux.HandleFunc("POST /api/v1/master-data/{type}", a.require("Editor", a.createMasterData))
	mux.HandleFunc("PUT /api/v1/master-data/{type}/{key}", a.require("Editor", a.updateMasterData))
	mux.HandleFunc("DELETE /api/v1/master-data/{type}/{key}", a.require("Editor", a.deleteMasterData))
	mux.HandleFunc("GET /api/v1/master-data-relations", a.require("Viewer", a.listMasterDataRelations))
	mux.HandleFunc("GET /api/v1/backup/export", a.require("Admin", a.exportBackup))
	mux.HandleFunc("POST /api/v1/backup/validate", a.require("Admin", a.validateBackup))
	mux.HandleFunc("POST /api/v1/backup/restore", a.require("Admin", a.restoreBackup))
}

func (a *App) registerExhibitionRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/exhibition-lists", a.require("Messe", a.listExhibitionLists))
	mux.HandleFunc("POST /api/v1/exhibition-lists", a.require("Admin", a.createExhibitionList))
	mux.HandleFunc("GET /api/v1/exhibition-lists/{id}", a.require("Messe", a.getExhibitionList))
	mux.HandleFunc("PUT /api/v1/exhibition-lists/{id}", a.require("Admin", a.updateExhibitionList))
	mux.HandleFunc("DELETE /api/v1/exhibition-lists/{id}", a.require("Admin", a.deleteExhibitionList))
	mux.HandleFunc("PUT /api/v1/exhibition-lists/{id}/lock", a.require("Admin", a.setExhibitionListLocked))
	mux.HandleFunc("GET /api/v1/exhibition-lists/{id}/entries", a.require("Messe", a.listExhibitionEntries))
	mux.HandleFunc("POST /api/v1/exhibition-lists/{id}/entries", a.require("Messe", a.createExhibitionEntry))
	mux.HandleFunc("PUT /api/v1/exhibition-lists/{id}/entries/{entryID}", a.require("Messe", a.updateExhibitionEntry))
	mux.HandleFunc("DELETE /api/v1/exhibition-lists/{id}/entries/{entryID}", a.require("Admin", a.deleteExhibitionEntry))
}
