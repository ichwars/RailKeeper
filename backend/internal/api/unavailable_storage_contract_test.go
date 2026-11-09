package api

import (
	"net/http/httptest"
	"strings"
	"testing"

	"railkeeper/backend/internal/application"
)

// Test the handler boundary separately from authentication: a failed database must
// yield the documented problem response, not accidentally report a successful write.
func TestUnavailableStorageNeverAcknowledgesSuccessfulMutation(t *testing.T) {
	f := newContractFixture(t)
	a := &App{logger: f.config.Logger, dataDir: f.dataDir, authService: f.auth, setupService: f.config.SetupService,
		vehicleService: f.vehicles, settingsService: f.config.SettingsService, inventoryNumbers: f.config.InventoryNumbers,
		masterDataService: f.config.MasterDataService, fileBlobs: f.config.FileBlobService,
		smtpSettingsService: f.config.SMTPSettingsService, databaseMaintenance: f.config.DatabaseMaintenance, backupService: f.config.BackupService, exhibitionService: application.NewExhibitionService(f.db)}
	if err := f.db.Close(); err != nil {
		t.Fatal(err)
	}
	for _, contract := range []struct {
		name, method, body, code string
		handler                  routeHandler
	}{
		{"setup status", "GET", "", "setup_status_failed", (*App).setupStatus},
		{"users", "GET", "", "user_list_failed", (*App).listUsers},
		{"roles", "GET", "", "role_list_failed", (*App).listRoles},
		{"create user", "POST", `{"username":"new","password":"synthetic-new-password","roles":["Viewer"]}`, "user_create_failed", (*App).createUser},
		{"update user", "PUT", `{"username":"new","roles":["Viewer"]}`, "user_update_failed", (*App).updateUser},
		{"delete user", "DELETE", "", "user_delete_failed", (*App).deleteUser},
		{"sessions", "GET", "", "session_list_failed", (*App).listSessions},
		{"revoke session", "PUT", "", "session_revoke_failed", (*App).revokeSession},
		{"audit", "GET", "", "audit_log_failed", (*App).systemAuditLog},
		{"profile load", "GET", "", "profile_settings_failed", (*App).getProfileSettings},
		{"profile save", "PUT", `{"settings":{"theme":"dark"}}`, "profile_settings_failed", (*App).updateProfileSettings},
		{"digital load", "GET", "", "digital_settings_failed", (*App).getDigitalSettings},
		{"digital save", "PUT", `{"provider":"ecos"}`, "digital_settings_failed", (*App).updateDigitalSettings},
		{"smtp load", "GET", "", "smtp_settings_failed", (*App).getSMTPSettings},
		{"smtp save", "PUT", `{"enabled":false}`, "smtp_settings_failed", (*App).updateSMTPSettings},
		{"smtp test", "POST", `{"recipient":"test@example.test"}`, "smtp_settings_failed", (*App).testSMTPSettings},
		{"two factor status", "GET", "", "two_factor_status_failed", (*App).twoFactorStatus},
		{"two factor setup", "POST", "", "two_factor_setup_failed", (*App).setupTwoFactor},
		{"two factor enable", "POST", `{"code":"invalid"}`, "two_factor_enable_failed", (*App).enableTwoFactor},
		{"two factor disable", "POST", `{"currentPassword":"synthetic-contract-password"}`, "two_factor_disable_failed", (*App).disableTwoFactor},
		{"vehicles", "GET", "", "vehicle_list_failed", (*App).listVehicles},
		{"vehicle", "GET", "", "vehicle_get_failed", (*App).getVehicle},
		{"maintenance list", "GET", "", "maintenance_list_failed", (*App).listVehicleMaintenance},
		{"maintenance create", "POST", `{"kind":"Wartung","status":"geplant"}`, "maintenance_create_failed", (*App).createVehicleMaintenance},
		{"maintenance update", "PUT", `{"kind":"Wartung","status":"geplant"}`, "maintenance_update_failed", (*App).updateVehicleMaintenance},
		{"maintenance delete", "DELETE", "", "maintenance_delete_failed", (*App).deleteVehicleMaintenance},
		{"spare parts list", "GET", "", "spare_part_list_failed", (*App).listVehicleSpareParts},
		{"spare parts suggestions", "GET", "", "spare_part_suggestions_failed", (*App).suggestVehicleSpareParts},
		{"spare parts create", "POST", `{"articleNumber":"ET-1"}`, "spare_part_create_failed", (*App).createVehicleSparePart},
		{"spare parts update", "PUT", `{"articleNumber":"ET-1"}`, "spare_part_update_failed", (*App).updateVehicleSparePart},
		{"spare parts delete", "DELETE", "", "spare_part_delete_failed", (*App).deleteVehicleSparePart},
		{"functions list", "GET", "", "function_list_failed", (*App).listVehicleFunctions},
		{"functions save", "PUT", `{"name":"Light"}`, "function_save_failed", (*App).upsertVehicleFunction},
		{"functions delete", "DELETE", "", "function_delete_failed", (*App).deleteVehicleFunction},
		{"CV list", "GET", "", "cv_value_list_failed", (*App).listVehicleCVValues},
		{"CV create", "POST", `{"cvNumber":1,"value":3}`, "cv_value_create_failed", (*App).createVehicleCVValue},
		{"CV delete", "DELETE", "", "cv_value_delete_failed", (*App).deleteVehicleCVValue},
		{"CV file delete", "DELETE", "", "cv_file_delete_failed", (*App).deleteVehicleCVFile},
		{"CV file download", "GET", "", "cv_file_download_failed", (*App).downloadVehicleCVFile},
		{"image download", "GET", "", "image_download_failed", (*App).downloadVehicleImage},
		{"image thumbnail", "GET", "", "image_thumbnail_failed", (*App).downloadVehicleImageThumbnail},
		{"image delete", "DELETE", "", "image_delete_failed", (*App).deleteVehicleImage},
		{"attachment download", "GET", "", "attachment_download_failed", (*App).downloadVehicleAttachment},
		{"attachment delete", "DELETE", "", "attachment_delete_failed", (*App).deleteVehicleAttachment},
		{"inventory list", "GET", "", "inventory_number_scheme_list_failed", (*App).listInventoryNumberSchemes},
		{"inventory create", "POST", `{"category":"Lokomotive","prefix":"L-","nextNumber":1,"padding":4,"active":true}`, "inventory_number_scheme_create_failed", (*App).createInventoryNumberScheme},
		{"inventory update", "PUT", `{"prefix":"L-","nextNumber":1,"padding":4,"active":true}`, "inventory_number_scheme_update_failed", (*App).updateInventoryNumberScheme},
		{"master data list", "GET", "", "master_data_list_failed", (*App).listMasterData},
		{"master data all", "GET", "", "master_data_list_failed", (*App).listAllMasterData},
		{"master data relations", "GET", "", "master_data_relations_failed", (*App).listMasterDataRelations},
		{"storage optimize", "POST", "", "storage_optimize_failed", (*App).optimizeSystemStorage},
		{"vehicle create", "POST", `{"name":"Retained vehicle","manufacturer":"Test","gauge":"H0","category":"Lokomotive","gattung":"Diesellok"}`, "vehicle_create_failed", (*App).createVehicle},
		{"vehicle update", "PUT", `{"name":"Retained vehicle","manufacturer":"Test","gauge":"H0","category":"Lokomotive","gattung":"Diesellok"}`, "vehicle_update_failed", (*App).updateVehicle},
		{"vehicle delete", "DELETE", "", "vehicle_delete_failed", (*App).deleteVehicle},
		{"vehicle set get", "GET", "", "vehicle_set_get_failed", (*App).getVehicleSet},
		{"vehicle mapping", "POST", `{"provider":"ecos","externalId":"123"}`, "external_mapping_failed", (*App).upsertVehicleExternalMapping},
		{"master data export", "GET", "", "master_data_export_failed", (*App).exportMasterData},
		{"master data create", "POST", `{"label":"Original","key":"original","active":true}`, "master_data_create_failed", (*App).createMasterData},
		{"master data update", "PUT", `{"label":"Original","active":true}`, "master_data_update_failed", (*App).updateMasterData},
		{"master data delete", "DELETE", "", "master_data_delete_failed", (*App).deleteMasterData},
		{"backup export", "GET", "", "backup_export_failed", (*App).exportBackup},
		{"password change", "PUT", `{"currentPassword":"synthetic-contract-password","newPassword":"synthetic-new-password"}`, "password_change_failed", (*App).changePassword},
		{"exhibition list", "GET", "", "exhibition_list_failed", (*App).listExhibitionLists},
		{"exhibition get", "GET", "", "exhibition_get_failed", (*App).getExhibitionList},
		{"exhibition create", "POST", `{"designation":"Retained event","date":"2026-11-09"}`, "exhibition_create_failed", (*App).createExhibitionList},
		{"exhibition update", "PUT", `{"designation":"Retained event","date":"2026-11-09"}`, "exhibition_update_failed", (*App).updateExhibitionList},
		{"exhibition delete", "DELETE", "", "exhibition_delete_failed", (*App).deleteExhibitionList},
		{"exhibition lock", "PUT", `{"locked":true}`, "exhibition_lock_failed", (*App).setExhibitionListLocked},
		{"exhibition entries", "GET", "", "exhibition_entries_failed", (*App).listExhibitionEntries},
		{"exhibition check", "POST", "", "exhibition_workspace_failed", (*App).checkExhibitionConflicts},
	} {
		t.Run(contract.name, func(t *testing.T) {
			r := httptest.NewRequest(contract.method, "/contract", strings.NewReader(contract.body))
			r = withActorUserID(r, f.adminID)
			for _, name := range []string{"id", "imageID", "attachmentID", "cvFileID", "cvValueID", "maintenanceID", "sparePartID"} {
				r.SetPathValue(name, strings.Repeat("1", 32))
			}
			r.SetPathValue("type", "manufacturer")
			r.SetPathValue("key", "synthetic")
			r.SetPathValue("category", "Lokomotive")
			r.SetPathValue("functionKey", "F1")
			w := httptest.NewRecorder()
			contract.handler(a, w, r)
			assertProblem(t, w, 500, contract.code)
		})
	}
}
