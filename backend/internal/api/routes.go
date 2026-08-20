package api

import (
	"errors"
	"net/http"

	"railkeeper/backend/internal/application"
)

type routeAccess string

const (
	routeAccessPublic          routeAccess = "public"
	routeAccessAdmin           routeAccess = "Admin"
	routeAccessEditor          routeAccess = "Editor"
	routeAccessViewer          routeAccess = "Viewer"
	routeAccessMesse           routeAccess = "Messe"
	routeAccessPlanner         routeAccess = "Planner"
	routeAccessEditorOrPlanner routeAccess = "EditorOrPlanner"
)

type routeHandler func(*App, http.ResponseWriter, *http.Request)
type routeAuthorizer func(*App, http.HandlerFunc) http.HandlerFunc

type routeSpec struct {
	Method    string
	Path      string
	Access    routeAccess
	Handler   routeHandler
	Authorize routeAuthorizer
}

func apiRouteSpecs() []routeSpec {
	return []routeSpec{
		{http.MethodGet, "/health", routeAccessPublic, (*App).health, nil},
		{http.MethodGet, "/api/v1/version", routeAccessPublic, (*App).versionInfo, nil},
		{http.MethodGet, "/api/v1/system/storage", routeAccessAdmin, (*App).systemStorage, nil},
		{http.MethodPost, "/api/v1/system/storage/optimize", routeAccessAdmin, (*App).optimizeSystemStorage, nil},
		{http.MethodGet, "/api/v1/system/storage/info", routeAccessAdmin, (*App).storageLocationInfo, nil},
		{http.MethodPost, "/api/v1/system/storage/open-folder", routeAccessAdmin, (*App).openStorageFolder, nil},
		{http.MethodPost, "/api/v1/system/storage/migration-receipt/acknowledge", routeAccessAdmin, (*App).acknowledgeStorageMigration, nil},
		{http.MethodGet, "/api/v1/system/printers", routeAccessAdmin, (*App).systemPrinters, nil},
		{http.MethodGet, "/api/v1/system/audit-log", routeAccessAdmin, (*App).systemAuditLog, nil},
		{http.MethodGet, "/api/v1/system/smtp", routeAccessAdmin, (*App).getSMTPSettings, nil},
		{http.MethodPut, "/api/v1/system/smtp", routeAccessAdmin, (*App).updateSMTPSettings, nil},
		{http.MethodPost, "/api/v1/system/smtp/test", routeAccessAdmin, (*App).testSMTPSettings, nil},
		{http.MethodGet, "/api/v1/system/digital-settings", routeAccessAdmin, (*App).getDigitalSettings, nil},
		{http.MethodPut, "/api/v1/system/digital-settings", routeAccessAdmin, (*App).updateDigitalSettings, nil},
		{http.MethodGet, "/api/v1/setup/status", routeAccessPublic, (*App).setupStatus, nil},
		{http.MethodPost, "/api/v1/setup/admin", routeAccessPublic, (*App).createAdmin, nil},
		{http.MethodPost, "/api/v1/auth/login", routeAccessPublic, (*App).login, nil},
		{http.MethodPost, "/api/v1/auth/password-reset", routeAccessPublic, (*App).requestPasswordReset, nil},
		{http.MethodPost, "/api/v1/auth/password-reset/confirm", routeAccessPublic, (*App).confirmPasswordReset, nil},
		{http.MethodPost, "/api/v1/auth/logout", routeAccessPublic, (*App).logout, nil},
		{http.MethodGet, "/api/v1/auth/session", routeAccessPublic, (*App).session, nil},
		{http.MethodGet, "/api/v1/profile/settings", routeAccessViewer, (*App).getProfileSettings, nil},
		{http.MethodPut, "/api/v1/profile/settings", routeAccessViewer, (*App).updateProfileSettings, nil},
		{http.MethodPut, "/api/v1/auth/password", routeAccessViewer, (*App).changePassword, nil},
		{http.MethodGet, "/api/v1/auth/two-factor", routeAccessViewer, (*App).twoFactorStatus, nil},
		{http.MethodPost, "/api/v1/auth/two-factor/setup", routeAccessViewer, (*App).setupTwoFactor, nil},
		{http.MethodPost, "/api/v1/auth/two-factor/enable", routeAccessViewer, (*App).enableTwoFactor, nil},
		{http.MethodPost, "/api/v1/auth/two-factor/disable", routeAccessViewer, (*App).disableTwoFactor, nil},
		{http.MethodGet, "/api/v1/roles", routeAccessAdmin, (*App).listRoles, nil},
		{http.MethodGet, "/api/v1/users", routeAccessAdmin, (*App).listUsers, nil},
		{http.MethodPost, "/api/v1/users", routeAccessAdmin, (*App).createUser, nil},
		{http.MethodPut, "/api/v1/users/{id}", routeAccessAdmin, (*App).updateUser, nil},
		{http.MethodDelete, "/api/v1/users/{id}", routeAccessAdmin, (*App).deleteUser, nil},
		{http.MethodGet, "/api/v1/sessions", routeAccessAdmin, (*App).listSessions, nil},
		{http.MethodPut, "/api/v1/sessions/{id}/revoke", routeAccessAdmin, (*App).revokeSession, nil},
		{http.MethodPost, "/api/v1/ecos/test", routeAccessAdmin, (*App).testECoSConnection, nil},
		{http.MethodPost, "/api/v1/ecos/locomotives/count", routeAccessAdmin, (*App).countECoSLocomotives, nil},
		{http.MethodPost, "/api/v1/ecos/locomotives/raw", routeAccessAdmin, (*App).probeECoSLocomotiveRaw, nil},
		{http.MethodPost, "/api/v1/digital-centers/ecos/locomotives/sync", routeAccessAdmin, (*App).syncECoSLocomotive, nil},
		{http.MethodPost, "/api/v1/digital-centers/z21/test", routeAccessAdmin, (*App).testZ21Connection, nil},
		{http.MethodPost, "/api/v1/digital-centers/z21/probe", routeAccessAdmin, (*App).probeZ21Connection, nil},
		{http.MethodPost, "/api/v1/digital-centers/intellibox3/test", routeAccessAdmin, (*App).testIntellibox3Connection, nil},
		{http.MethodPost, "/api/v1/digital-centers/intellibox3/probe", routeAccessAdmin, (*App).probeIntellibox3Connection, nil},
		{http.MethodPost, "/api/v1/digital-centers/cs3/test", routeAccessAdmin, (*App).testCS3Connection, nil},
		{http.MethodGet, "/api/v1/digital-centers/ecos/live/status", routeAccessAdmin, (*App).eCoSLiveStatus, nil},
		{http.MethodPost, "/api/v1/digital-centers/ecos/live/start", routeAccessAdmin, (*App).startECoSLive, nil},
		{http.MethodPost, "/api/v1/digital-centers/ecos/live/stop", routeAccessAdmin, (*App).stopECoSLive, nil},
		{http.MethodGet, "/api/v1/vehicles", routeAccessViewer, (*App).listVehicles, nil},
		{http.MethodGet, "/api/v1/overview/valuation", routeAccessViewer, (*App).overviewValuation, nil},
		{http.MethodPost, "/api/v1/vehicles", routeAccessEditor, (*App).createVehicle, nil},
		{http.MethodPost, "/api/v1/vehicle-sets", routeAccessEditor, (*App).createVehicleSet, nil},
		{http.MethodGet, "/api/v1/vehicle-sets/{id}", routeAccessViewer, (*App).getVehicleSet, nil},
		{http.MethodPatch, "/api/v1/vehicle-sets/{id}", routeAccessEditor, (*App).updateVehicleSet, nil},
		{http.MethodGet, "/api/v1/vehicles/{id}", routeAccessViewer, (*App).getVehicle, nil},
		{http.MethodPut, "/api/v1/vehicles/{id}", routeAccessEditor, (*App).updateVehicle, nil},
		{http.MethodDelete, "/api/v1/vehicles/{id}", routeAccessEditor, (*App).deleteVehicle, nil},
		{http.MethodPost, "/api/v1/vehicles/{id}/external-mappings", routeAccessEditor, (*App).upsertVehicleExternalMapping, nil},
		{http.MethodPost, "/api/v1/vehicles/{id}/images", routeAccessEditor, (*App).uploadVehicleImage, nil},
		{http.MethodPost, "/api/v1/vehicles/{id}/images/import-url", routeAccessEditor, (*App).importVehicleImageFromURL, nil},
		{http.MethodDelete, "/api/v1/vehicles/{id}/images/{imageID}", routeAccessEditor, (*App).deleteVehicleImage, nil},
		{http.MethodGet, "/api/v1/vehicles/{id}/images/{imageID}/file", routeAccessViewer, (*App).downloadVehicleImage, nil},
		{http.MethodGet, "/api/v1/vehicles/{id}/images/{imageID}/thumbnail", routeAccessViewer, (*App).downloadVehicleImageThumbnail, nil},
		{http.MethodPost, "/api/v1/vehicles/{id}/attachments", routeAccessEditor, (*App).uploadVehicleAttachment, nil},
		{http.MethodPut, "/api/v1/vehicles/{id}/attachments/{attachmentID}", routeAccessEditor, (*App).updateVehicleAttachment, nil},
		{http.MethodDelete, "/api/v1/vehicles/{id}/attachments/{attachmentID}", routeAccessEditor, (*App).deleteVehicleAttachment, nil},
		{http.MethodGet, "/api/v1/vehicles/{id}/attachments/{attachmentID}/download", routeAccessViewer, (*App).downloadVehicleAttachment, nil},
		{http.MethodPost, "/api/v1/vehicles/{id}/attachments/import-url", routeAccessEditor, (*App).importVehicleAttachmentFromURL, nil},
		{http.MethodGet, "/api/v1/vehicles/{id}/maintenance", routeAccessViewer, (*App).listVehicleMaintenance, nil},
		{http.MethodPost, "/api/v1/vehicles/{id}/maintenance", routeAccessEditor, (*App).createVehicleMaintenance, nil},
		{http.MethodPut, "/api/v1/vehicles/{id}/maintenance/{maintenanceID}", routeAccessEditor, (*App).updateVehicleMaintenance, nil},
		{http.MethodDelete, "/api/v1/vehicles/{id}/maintenance/{maintenanceID}", routeAccessEditor, (*App).deleteVehicleMaintenance, nil},
		{http.MethodGet, "/api/v1/vehicles/{id}/spare-parts", routeAccessViewer, (*App).listVehicleSpareParts, nil},
		{http.MethodGet, "/api/v1/vehicles/{id}/spare-parts/suggestions", routeAccessViewer, (*App).suggestVehicleSpareParts, nil},
		{http.MethodPost, "/api/v1/vehicles/{id}/spare-parts", routeAccessEditor, (*App).createVehicleSparePart, nil},
		{http.MethodPut, "/api/v1/vehicles/{id}/spare-parts/{sparePartID}", routeAccessEditor, (*App).updateVehicleSparePart, nil},
		{http.MethodDelete, "/api/v1/vehicles/{id}/spare-parts/{sparePartID}", routeAccessEditor, (*App).deleteVehicleSparePart, nil},
		{http.MethodGet, "/api/v1/vehicles/{id}/functions", routeAccessViewer, (*App).listVehicleFunctions, nil},
		{http.MethodPut, "/api/v1/vehicles/{id}/functions/{functionKey}", routeAccessEditor, (*App).upsertVehicleFunction, nil},
		{http.MethodDelete, "/api/v1/vehicles/{id}/functions/{functionKey}", routeAccessEditor, (*App).deleteVehicleFunction, nil},
		{http.MethodGet, "/api/v1/vehicles/{id}/cv-values", routeAccessViewer, (*App).listVehicleCVValues, nil},
		{http.MethodPost, "/api/v1/vehicles/{id}/cv-values", routeAccessEditor, (*App).createVehicleCVValue, nil},
		{http.MethodPut, "/api/v1/vehicles/{id}/cv-values/{cvValueID}", routeAccessEditor, (*App).updateVehicleCVValue, nil},
		{http.MethodDelete, "/api/v1/vehicles/{id}/cv-values/{cvValueID}", routeAccessEditor, (*App).deleteVehicleCVValue, nil},
		{http.MethodPost, "/api/v1/cv-files/preview", routeAccessEditor, (*App).previewVehicleCVFile, nil},
		{http.MethodPost, "/api/v1/vehicles/{id}/cv-files", routeAccessEditor, (*App).uploadVehicleCVFile, nil},
		{http.MethodDelete, "/api/v1/vehicles/{id}/cv-files/{cvFileID}", routeAccessEditor, (*App).deleteVehicleCVFile, nil},
		{http.MethodGet, "/api/v1/vehicles/{id}/cv-files/{cvFileID}/download", routeAccessViewer, (*App).downloadVehicleCVFile, nil},
		{http.MethodPost, "/api/v1/article-search", routeAccessViewer, (*App).searchArticleData, nil},
		{http.MethodGet, "/api/v1/data-transfer/profiles", routeAccessViewer, (*App).listDataTransferProfiles, authorizeDataTransferRead},
		{http.MethodPost, "/api/v1/data-transfer/profiles", routeAccessEditor, (*App).createDataTransferProfile, nil},
		{http.MethodPut, "/api/v1/data-transfer/profiles/{id}", routeAccessEditor, (*App).updateDataTransferProfile, nil},
		{http.MethodDelete, "/api/v1/data-transfer/profiles/{id}", routeAccessAdmin, (*App).disableDataTransferProfile, nil},
		{http.MethodPost, "/api/v1/data-transfer/jobs/export", routeAccessViewer, (*App).createDataTransferExportJob, authorizeDataTransferRead},
		{http.MethodPost, "/api/v1/data-transfer/jobs/{id}/execute", routeAccessViewer, (*App).executeDataTransferExport, authorizeDataTransferRead},
		{http.MethodPost, "/api/v1/data-transfer/jobs/import", routeAccessEditor, (*App).createDataTransferImportJob, authorizeDataTransferWrite},
		{http.MethodPost, "/api/v1/data-transfer/jobs/{id}/upload", routeAccessEditor, (*App).uploadDataTransferImport, authorizeDataTransferWrite},
		{http.MethodPut, "/api/v1/data-transfer/jobs/{id}/issues/{issueID}", routeAccessEditor, (*App).resolveDataTransferIssue, authorizeDataTransferWrite},
		{http.MethodPost, "/api/v1/data-transfer/jobs/{id}/cancel", routeAccessEditor, (*App).cancelDataTransferJob, authorizeDataTransferWrite},
		{http.MethodGet, "/api/v1/data-transfer/artifacts/{id}/download", routeAccessViewer, (*App).downloadDataTransferArtifact, authorizeDataTransferRead},
		{http.MethodDelete, "/api/v1/data-transfer/artifacts/{id}", routeAccessAdmin, (*App).deleteDataTransferArtifact, nil},
		{http.MethodPost, "/api/v1/data-transfer/artifacts/open-folder", routeAccessAdmin, (*App).openDataTransferArtifactFolder, nil},
		{http.MethodGet, "/api/v1/accessory-products", routeAccessViewer, (*App).listAccessoryProducts, nil},
		{http.MethodPost, "/api/v1/accessory-products", routeAccessEditor, (*App).createAccessoryProduct, nil},
		{http.MethodPost, "/api/v1/accessory-products/duplicate-check", routeAccessEditor, (*App).checkAccessoryProductDuplicates, nil},
		{http.MethodGet, "/api/v1/accessory-products/{id}", routeAccessViewer, (*App).getAccessoryProduct, nil},
		{http.MethodPut, "/api/v1/accessory-products/{id}", routeAccessEditor, (*App).updateAccessoryProduct, nil},
		{http.MethodDelete, "/api/v1/accessory-products/{id}", routeAccessAdmin, (*App).deleteAccessoryProduct, nil},
		{http.MethodPost, "/api/v1/accessory-products/{id}/archive", routeAccessEditor, (*App).archiveAccessoryProduct, nil},
		{http.MethodPost, "/api/v1/accessory-products/{id}/restore", routeAccessEditor, (*App).restoreAccessoryProduct, nil},
		{http.MethodGet, "/api/v1/accessory-products/{id}/stock", routeAccessViewer, (*App).getAccessoryStock, nil},
		{http.MethodPost, "/api/v1/accessory-products/{id}/stock-adjustments", routeAccessEditor, (*App).adjustAccessoryStock, nil},
		{http.MethodGet, "/api/v1/accessory-products/{id}/stock-movements", routeAccessViewer, (*App).listAccessoryStockMovements, nil},
		{http.MethodPost, "/api/v1/accessory-products/{id}/stock-transfers", routeAccessEditor, (*App).transferAccessoryStock, nil},
		{http.MethodGet, "/api/v1/accessory-products/{id}/purchases", routeAccessViewer, (*App).listAccessoryPurchases, nil},
		{http.MethodPost, "/api/v1/accessory-products/{id}/purchases", routeAccessEditor, (*App).createAccessoryPurchase, nil},
		{http.MethodPost, "/api/v1/accessory-products/{id}/individualizations", routeAccessEditor, (*App).individualizeAccessoryProduct, nil},
		{http.MethodGet, "/api/v1/accessory-products/{id}/assets", routeAccessViewer, (*App).listAccessoryAssets, nil},
		{http.MethodPost, "/api/v1/accessory-products/{id}/assets", routeAccessEditor, (*App).createAccessoryAsset, nil},
		{http.MethodPut, "/api/v1/accessory-assets/{id}", routeAccessEditor, (*App).updateAccessoryAsset, nil},
		{http.MethodGet, "/api/v1/accessory-products/{id}/documents", routeAccessViewer, (*App).listAccessoryDocuments, nil},
		{http.MethodPost, "/api/v1/accessory-products/{id}/documents", routeAccessEditor, (*App).uploadAccessoryDocument, nil},
		{http.MethodPost, "/api/v1/accessory-products/{id}/documents/import-url", routeAccessEditor, (*App).importAccessoryDocumentFromURL, nil},
		{http.MethodGet, "/api/v1/accessory-products/{id}/documents/{documentID}", routeAccessViewer, (*App).getAccessoryDocument, nil},
		{http.MethodPut, "/api/v1/accessory-products/{id}/documents/{documentID}", routeAccessEditor, (*App).updateAccessoryDocument, nil},
		{http.MethodDelete, "/api/v1/accessory-products/{id}/documents/{documentID}", routeAccessEditor, (*App).deleteAccessoryDocument, nil},
		{http.MethodGet, "/api/v1/accessory-products/{id}/documents/{documentID}/download", routeAccessViewer, (*App).downloadAccessoryDocument, nil},
		{http.MethodGet, "/api/v1/accessory-products/{id}/allocation-summary", routeAccessViewer, (*App).getAccessoryAllocationSummary, nil},
		{http.MethodGet, "/api/v1/accessory-products/{id}/usage-history", routeAccessViewer, (*App).getAccessoryUsageHistory, nil},
		{http.MethodGet, "/api/v1/accessory-reservations", routeAccessViewer, (*App).listAccessoryReservations, nil},
		{http.MethodPost, "/api/v1/accessory-reservations", routeAccessEditorOrPlanner, (*App).createAccessoryReservation, nil},
		{http.MethodPost, "/api/v1/accessory-reservations/{id}/cancel", routeAccessEditorOrPlanner, (*App).cancelAccessoryReservation, nil},
		{http.MethodGet, "/api/v1/accessory-installations", routeAccessViewer, (*App).listAccessoryInstallations, nil},
		{http.MethodPost, "/api/v1/accessory-installations", routeAccessEditor, (*App).createAccessoryInstallation, nil},
		{http.MethodPost, "/api/v1/accessory-installations/{id}/remove", routeAccessEditor, (*App).removeAccessoryInstallation, nil},
		{http.MethodPut, "/api/v1/accessory-installations/{id}/condition", routeAccessEditor, (*App).updateAccessoryInstallationCondition, nil},
		{http.MethodGet, "/api/v1/storage-locations", routeAccessViewer, (*App).listStorageLocations, nil},
		{http.MethodPost, "/api/v1/storage-locations", routeAccessEditor, (*App).createStorageLocation, nil},
		{http.MethodPut, "/api/v1/storage-locations/{id}", routeAccessEditor, (*App).updateStorageLocation, nil},
		{http.MethodGet, "/api/v1/layouts", routeAccessViewer, (*App).listLayouts, nil},
		{http.MethodPost, "/api/v1/layouts", routeAccessPlanner, (*App).createLayout, nil},
		{http.MethodGet, "/api/v1/layouts/{id}", routeAccessViewer, (*App).getLayout, nil},
		{http.MethodGet, "/api/v1/layouts/{id}/twin", routeAccessViewer, (*App).getLayoutTwin, nil},
		{http.MethodPut, "/api/v1/layouts/{id}", routeAccessPlanner, (*App).updateLayout, nil},
		{http.MethodGet, "/api/v1/layouts/{id}/units", routeAccessViewer, (*App).listLayoutUnits, nil},
		{http.MethodPost, "/api/v1/layouts/{id}/units", routeAccessPlanner, (*App).createLayoutUnit, nil},
		{http.MethodPut, "/api/v1/layout-units/{id}", routeAccessPlanner, (*App).updateLayoutUnit, nil},
		{http.MethodGet, "/api/v1/layout-units/{id}/ports", routeAccessViewer, (*App).listLayoutUnitPorts, nil},
		{http.MethodPost, "/api/v1/layout-units/{id}/ports", routeAccessPlanner, (*App).createLayoutUnitPort, nil},
		{http.MethodPut, "/api/v1/layout-unit-ports/{id}", routeAccessPlanner, (*App).updateLayoutUnitPort, nil},
		{http.MethodPut, "/api/v1/layout-units/{id}/outline", routeAccessPlanner, (*App).updateLayoutUnitOutline, nil},
		{http.MethodGet, "/api/v1/layout-units/{id}/technical-positions", routeAccessViewer, (*App).listLayoutTechnicalPositions, nil},
		{http.MethodPost, "/api/v1/layout-units/{id}/technical-positions", routeAccessPlanner, (*App).createLayoutTechnicalPosition, nil},
		{http.MethodPut, "/api/v1/layout-technical-positions/{id}", routeAccessPlanner, (*App).updateLayoutTechnicalPosition, nil},
		{http.MethodGet, "/api/v1/layouts/{id}/configurations", routeAccessViewer, (*App).listLayoutConfigurations, nil},
		{http.MethodPost, "/api/v1/layouts/{id}/configurations", routeAccessPlanner, (*App).createLayoutConfiguration, nil},
		{http.MethodPut, "/api/v1/layout-configurations/{id}", routeAccessPlanner, (*App).updateLayoutConfiguration, nil},
		{http.MethodGet, "/api/v1/layout-configurations/{id}/port-analysis", routeAccessViewer, (*App).analyzeLayoutConfigurationPorts, nil},
		{http.MethodPost, "/api/v1/layout-configurations/{id}/unit-snap-preview", routeAccessPlanner, (*App).previewLayoutConfigurationUnitSnap, nil},
		{http.MethodGet, "/api/v1/layout-units/{id}/plan-variants", routeAccessViewer, (*App).listPlanVariants, nil},
		{http.MethodPost, "/api/v1/layout-units/{id}/plan-variants", routeAccessPlanner, (*App).createPlanVariant, nil},
		{http.MethodPost, "/api/v1/plan-variants/{id}/revisions", routeAccessPlanner, (*App).createPlanRevision, nil},
		{http.MethodPost, "/api/v1/plan-revisions/{id}/submit", routeAccessPlanner, (*App).submitPlanRevision, nil},
		{http.MethodPost, "/api/v1/plan-revisions/{id}/publish", routeAccessPlanner, (*App).publishPlanRevision, nil},
		{http.MethodGet, "/api/v1/track-geometries", routeAccessViewer, (*App).listTrackGeometries, nil},
		{http.MethodGet, "/api/v1/track-libraries", routeAccessViewer, (*App).listTrackLibraries, nil},
		{http.MethodGet, "/api/v1/track-libraries/{id}/export", routeAccessViewer, (*App).exportTrackLibrary, nil},
		{http.MethodPost, "/api/v1/track-libraries/import/preview", routeAccessAdmin, (*App).previewTrackLibraryImport, nil},
		{http.MethodPost, "/api/v1/track-libraries/import", routeAccessAdmin, (*App).importTrackLibrary, nil},
		{http.MethodPut, "/api/v1/track-libraries/{id}/status", routeAccessAdmin, (*App).updateTrackLibraryStatus, nil},
		{http.MethodGet, "/api/v1/plan-revisions/{id}/track-plan", routeAccessViewer, (*App).getTrackPlan, nil},
		{http.MethodGet, "/api/v1/plan-revisions/{id}/track-analysis", routeAccessViewer, (*App).getTrackPlanAnalysis, nil},
		{http.MethodGet, "/api/v1/plan-revisions/{id}/track-change-preview", routeAccessViewer, (*App).getTrackPlanChangePreview, nil},
		{http.MethodPost, "/api/v1/plan-revisions/{id}/track-reservations", routeAccessPlanner, (*App).reserveTrackPlanMaterials, nil},
		{http.MethodPost, "/api/v1/plan-revisions/{id}/track-objects", routeAccessPlanner, (*App).createPlanTrackObject, nil},
		{http.MethodPost, "/api/v1/plan-revisions/{id}/free-objects", routeAccessPlanner, (*App).createFreePlanObject, nil},
		{http.MethodPut, "/api/v1/plan-track-objects/{id}", routeAccessPlanner, (*App).updatePlanTrackObject, nil},
		{http.MethodPut, "/api/v1/plan-free-objects/{id}", routeAccessPlanner, (*App).updateFreePlanObject, nil},
		{http.MethodPost, "/api/v1/plan-track-objects/{id}/flex-preview", routeAccessPlanner, (*App).previewFlexTrackPath, nil},
		{http.MethodPost, "/api/v1/plan-track-objects/{id}/transition-preview", routeAccessPlanner, (*App).previewTransitionCurve, nil},
		{http.MethodDelete, "/api/v1/plan-track-objects/{id}", routeAccessPlanner, (*App).deletePlanTrackObject, nil},
		{http.MethodDelete, "/api/v1/plan-free-objects/{id}", routeAccessPlanner, (*App).deleteFreePlanObject, nil},
		{http.MethodGet, "/api/v1/inventory-number-schemes", routeAccessViewer, (*App).listInventoryNumberSchemes, nil},
		{http.MethodPost, "/api/v1/inventory-number-schemes", routeAccessEditor, (*App).createInventoryNumberScheme, nil},
		{http.MethodPut, "/api/v1/inventory-number-schemes/{category}", routeAccessEditor, (*App).updateInventoryNumberScheme, nil},
		{http.MethodGet, "/api/v1/master-data-all", routeAccessViewer, (*App).listAllMasterData, nil},
		{http.MethodGet, "/api/v1/master-data/export", routeAccessAdmin, (*App).exportMasterData, nil},
		{http.MethodPost, "/api/v1/master-data/import", routeAccessAdmin, (*App).importMasterData, nil},
		{http.MethodGet, "/api/v1/master-data/{type}", routeAccessViewer, (*App).listMasterData, authorizeMasterDataRead},
		{http.MethodPost, "/api/v1/master-data/{type}", routeAccessEditor, (*App).createMasterData, nil},
		{http.MethodPut, "/api/v1/master-data/{type}/{key}", routeAccessEditor, (*App).updateMasterData, nil},
		{http.MethodPatch, "/api/v1/master-data/{type}/{key}/active", routeAccessEditor, (*App).setMasterDataActive, nil},
		{http.MethodDelete, "/api/v1/master-data/{type}/{key}", routeAccessEditor, (*App).deleteMasterData, nil},
		{http.MethodGet, "/api/v1/master-data-relations", routeAccessViewer, (*App).listMasterDataRelations, nil},
		{http.MethodGet, "/api/v1/backup/export", routeAccessAdmin, (*App).exportBackup, nil},
		{http.MethodPost, "/api/v1/backup/validate", routeAccessAdmin, (*App).validateBackup, nil},
		{http.MethodPost, "/api/v1/backup/restore", routeAccessAdmin, (*App).restoreBackup, nil},
		{http.MethodGet, "/api/v1/exhibition-lists", routeAccessMesse, (*App).listExhibitionLists, nil},
		{http.MethodPost, "/api/v1/exhibition-lists", routeAccessAdmin, (*App).createExhibitionList, nil},
		{http.MethodGet, "/api/v1/exhibition-lists/{id}", routeAccessMesse, (*App).getExhibitionList, nil},
		{http.MethodPut, "/api/v1/exhibition-lists/{id}", routeAccessAdmin, (*App).updateExhibitionList, nil},
		{http.MethodDelete, "/api/v1/exhibition-lists/{id}", routeAccessAdmin, (*App).deleteExhibitionList, nil},
		{http.MethodPut, "/api/v1/exhibition-lists/{id}/lock", routeAccessAdmin, (*App).setExhibitionListLocked, nil},
		{http.MethodGet, "/api/v1/exhibition-lists/{id}/entries", routeAccessMesse, (*App).listExhibitionEntries, nil},
		{http.MethodPost, "/api/v1/exhibition-lists/{id}/entries", routeAccessMesse, (*App).createExhibitionEntry, nil},
		{http.MethodPut, "/api/v1/exhibition-lists/{id}/entries/{entryID}", routeAccessMesse, (*App).updateExhibitionEntry, nil},
		{http.MethodDelete, "/api/v1/exhibition-lists/{id}/entries/{entryID}", routeAccessAdmin, (*App).deleteExhibitionEntry, nil},
	}
}

func (a *App) registerRoutes(mux *http.ServeMux) {
	for _, route := range apiRouteSpecs() {
		route := route
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			route.Handler(a, w, r)
		})
		if route.Authorize != nil {
			handler = route.Authorize(a, handler)
		} else if route.Access == routeAccessEditorOrPlanner {
			handler = a.requireAny([]string{"Editor", "Planner"}, handler)
		} else if route.Access != routeAccessPublic {
			handler = a.require(string(route.Access), handler)
		}
		mux.HandleFunc(route.Method+" "+route.Path, handler)
	}
}

func authorizeMasterDataRead(a *App, next http.HandlerFunc) http.HandlerFunc {
	return a.requireMasterDataRead(next)
}

func authorizeDataTransferRead(a *App, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := a.authService.RequireAnyRole(r.Context(), cookieValue(r, "rk_session"), "Viewer", "Messe")
		if err != nil {
			authorizationError(a, w, err)
			return
		}
		session, err := a.authService.CurrentSession(r.Context(), cookieValue(r, "rk_session"))
		if err != nil {
			authorizationError(a, w, err)
			return
		}
		next.ServeHTTP(w, withDataTransferScope(withActorUserID(r, userID), dataTransferMesseScope(session.Roles)))
	}
}

func authorizeDataTransferWrite(a *App, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := a.authService.RequireAnyRole(r.Context(), cookieValue(r, "rk_session"), "Editor", "Messe")
		if err != nil {
			authorizationError(a, w, err)
			return
		}
		session, err := a.authService.CurrentSession(r.Context(), cookieValue(r, "rk_session"))
		if err != nil {
			authorizationError(a, w, err)
			return
		}
		next.ServeHTTP(w, withDataTransferScope(withActorUserID(r, userID), dataTransferMesseScope(session.Roles)))
	}
}

func authorizationError(a *App, w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, application.ErrUnauthorized):
		respondProblem(w, http.StatusUnauthorized, "unauthorized", "Not logged in.")
	case errors.Is(err, application.ErrForbidden):
		respondProblem(w, http.StatusForbidden, "forbidden", "Insufficient role.")
	default:
		a.logger.Error("role check failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "role_check_failed", "Could not verify permissions.")
	}
}

func (a *App) health(w http.ResponseWriter, _ *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
