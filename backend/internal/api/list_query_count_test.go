package api

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"net/http"
	"path/filepath"
	"slices"
	"testing"

	"modernc.org/sqlite"

	"railkeeper/backend/internal/application"
	"railkeeper/backend/internal/infrastructure"
	"railkeeper/backend/internal/transfer"
)

func TestListQueryCounts(t *testing.T) {
	for _, size := range []int{1, 25, 100} {
		t.Run(fmt.Sprint(size), func(t *testing.T) {
			testListQueryCounts(t, size)
		})
	}
}

func testListQueryCounts(t *testing.T, size int) {
	db, counts := newCountedListDB(t)
	setup := application.NewSetupService(db)
	auth := application.NewAuthService(db)
	password := "query-count-fixture-password"
	if err := setup.CreateAdmin(t.Context(), application.CreateAdminInput{
		Username: "admin", Email: "admin@example.test", Password: password,
	}); err != nil {
		t.Fatal(err)
	}
	ids := seedListQueryRows(t, db, size)
	vehicles := application.NewVehicleService(db)
	masterData := application.NewMasterDataService(db)
	settings := application.NewSettingsService(db)
	centers := infrastructure.NewDigitalCenterWorkspaceRepository(db)
	if _, err := settings.UpdateDigitalSettings(t.Context(), application.DigitalCenterSettings{
		Provider:    "ecos",
		ECoS:        application.DigitalProviderSettings{Enabled: true, Host: "ecos.test", Port: "15471"},
		Z21:         application.DigitalProviderSettings{Host: "z21.test", Port: "21105"},
		Intellibox3: application.DigitalProviderSettings{Host: "ib.test", Port: "21105"},
		CS3:         application.DigitalProviderSettings{Host: "cs3.test", Port: "15731"},
	}); err != nil {
		t.Fatal(err)
	}
	for _, center := range []application.DigitalCenterReadSession{
		{ID: "ecos-ready", Provider: "ecos", Host: "ecos.test", Port: 15471},
		{ID: "z21-ready", Provider: "z21", Host: "z21.test", Port: 21105},
	} {
		center.State = application.DigitalCenterSessionReady
		center.ReadCompletedAt = "2026-01-01T00:00:00Z"
		center.Capabilities.ReadLocomotives = true
		if _, err := centers.CreateSession(t.Context(), center); err != nil {
			t.Fatal(err)
		}
	}

	router := NewRouter(Config{
		SetupService: setup, AuthService: auth, VehicleService: vehicles, MasterDataService: masterData,
		DigitalCenterWorkspace: application.NewDigitalCenterWorkspaceService(centers, settings, nil, nil, nil, nil),
	})
	session, cookies := loginTestUser(t, router, "admin", password)
	login := &application.LoginResult{CSRFToken: session.CSRFToken}
	for _, cookie := range cookies {
		if cookie.Name == "rk_session" {
			login.SessionToken = cookie.Value
		}
	}
	request := func(method, path string, body any, target any) {
		t.Helper()
		response := authedJSONRequest(t, router, login, method, "/api/v1"+path, body, true)
		assertStatus(t, response, http.StatusOK)
		decodeResponse(t, response, target)
	}
	assertCounts := func(name string, wantReads, wantWrites int) {
		t.Helper()
		t.Logf("%s: %d reads, %d writes", name, counts.reads, counts.writes)
		if *counts != (listStatementCounts{wantReads, wantWrites}) {
			t.Errorf("%s: got %+v, want %d reads, %d writes", name, *counts, wantReads, wantWrites)
		}
		*counts = listStatementCounts{}
	}

	*counts = listStatementCounts{}
	var inventory []application.Vehicle
	request(http.MethodGet, "/vehicles", nil, &inventory)
	assertCounts("vehicle inventory", 12, 2)
	if len(inventory) != size || inventory[0].ID != ids[0] || len(inventory[0].Images) != 1 ||
		len(inventory[0].CVValues) != size || len(inventory[0].CVValues[0].History) != 1 {
		t.Fatalf("wrong inventory or missing vehicle relations: %d vehicles", len(inventory))
	}

	var options map[string][]application.MasterDataEntry
	var relations []application.MasterDataRelation
	request(http.MethodGet, "/master-data-all", nil, &options)
	request(http.MethodGet, "/master-data-relations?parentType=vehicle_category&childType=vehicle_gattung", nil, &relations)
	assertCounts("vehicle inventory filter data (cold cache)", 6, 0)
	masterData.InvalidateCache()

	var manufacturer []application.MasterDataEntry
	request(http.MethodGet, "/master-data/manufacturer?management=true", nil, &manufacturer)
	assertCounts("managed manufacturers", 4, 0)
	checked := 0
	for _, item := range manufacturer {
		if item.Type == "manufacturer" && item.Key == fmt.Sprintf("maker-%03d", checked) {
			if item.Capabilities == nil || item.Capabilities.CanDelete != (checked != 0) {
				t.Fatalf("incorrect deletion capability for %s: %+v", item.Key, item.Capabilities)
			}
			checked++
		}
	}
	if checked != size {
		t.Fatalf("checked %d manufacturers, want %d", checked, size)
	}

	var allMasterData map[string][]application.MasterDataEntry
	request(http.MethodGet, "/master-data-all?management=true", nil, &allMasterData)
	assertCounts("all managed master data (cold cache)", 4, 0)
	request(http.MethodGet, "/master-data-all?management=true", nil, &allMasterData)
	assertCounts("all managed master data (warm cache)", 3, 0)

	var users []struct {
		Username string
		Roles    []string
	}
	request(http.MethodGet, "/users", nil, &users)
	assertCounts("user list", 3, 0)
	if len(users) != size+1 || !slices.Equal(users[0].Roles, []string{"Admin", "Editor", "Viewer"}) ||
		!slices.Equal(users[1].Roles, []string{"Editor", "Viewer"}) ||
		(size > 1 && (users[2].Roles == nil || len(users[2].Roles) != 0)) {
		t.Fatalf("incorrect user count or role grouping: %d users", len(users))
	}

	var configuration application.DecoderConfigurationContext
	request(http.MethodGet, "/vehicles/"+ids[0]+"/decoder-configurations", nil, &configuration)
	assertCounts("decoder configuration", 10, 0)
	if len(configuration.Workspaces) != size || configuration.Workspaces[0].ValueCount != 1 {
		t.Fatalf("wrong decoder workspaces: %d", len(configuration.Workspaces))
	}
	t.Logf("decoder revision %s", configuration.Revision)

	var workspace struct {
		Centers []application.DigitalCenterSummary
	}
	request(http.MethodGet, "/digital-centers/workspace", nil, &workspace)
	assertCounts("digital center workspace", 4, 0)
	if len(workspace.Centers) != 4 || workspace.Centers[0].LastReadSession == nil ||
		workspace.Centers[0].LastReadSession.ID != "ecos-ready" ||
		workspace.Centers[1].LastReadSession == nil || workspace.Centers[1].LastReadSession.ID != "z21-ready" ||
		workspace.Centers[2].LastReadSession != nil || workspace.Centers[3].LastReadSession != nil {
		t.Fatalf("wrong saved digital center sessions: %+v", workspace.Centers)
	}

	slices.Reverse(ids)
	report := loadCountedVehicleReport(request, ids)
	assertCounts("PDF report data", 13, 2)
	if len(report) != size || report[0].ID != ids[0] || report[size-1].ID != ids[size-1] ||
		report[size-1].StorageDetails != "Drawer" || len(report[size-1].Images) != 1 {
		t.Fatalf("wrong report order or fields: %d vehicles", len(report))
	}

	repository := infrastructure.NewDataTransferRepository(db)
	accessory, err := repository.Snapshot(t.Context(), []transfer.TransferArea{transfer.TransferAccessories})
	if err != nil {
		t.Fatal(err)
	}
	assertCounts("data transfer accessory snapshot", 6, 0)
	if len(accessory.Accessories) != size || len(accessory.Accessories[0].Stock) != 1 || accessory.Accessories[0].Stock[0].Quantity != 2 {
		t.Fatalf("wrong accessory snapshot: %d products", len(accessory.Accessories))
	}
	exhibition, err := repository.Snapshot(t.Context(), []transfer.TransferArea{transfer.TransferExhibitionLists})
	if err != nil {
		t.Fatal(err)
	}
	assertCounts("data transfer exhibition snapshot", 2, 0)
	if len(exhibition.ExhibitionLists) != size || len(exhibition.ExhibitionLists[0].Entries) != 1 {
		t.Fatalf("wrong exhibition snapshot: %d lists", len(exhibition.ExhibitionLists))
	}
	testListQueryCountPreviews(t, db, counts, centers, repository, ids)

}

func loadCountedVehicleReport(request func(string, string, any, any), ids []string) []application.Vehicle {
	var vehicles []application.Vehicle
	request(http.MethodPost, "/vehicles/report-data", map[string][]string{"vehicleIds": ids}, &vehicles)
	return vehicles
}

func testListQueryCountPreviews(
	t *testing.T, db *sql.DB, counts *listStatementCounts,
	centers *infrastructure.DigitalCenterWorkspaceRepository, repository *infrastructure.DataTransferRepository,
	ids []string,
) {
	t.Helper()
	objectIDs := make([]int, len(ids))
	items := make([]application.DigitalCenterWorkItem, len(ids))
	for i, id := range ids {
		objectIDs[i] = i + 1
		externalID := fmt.Sprint(objectIDs[i])
		items[i] = application.DigitalCenterWorkItem{ID: "item-" + externalID, CenterObjectID: externalID}
		if _, err := db.ExecContext(t.Context(), `
INSERT INTO vehicle_external_mappings(id, vehicle_id, provider, external_id, sync_status, created_at, updated_at)
VALUES(?,?,'ecos',?,'linked','2026-01-01','2026-01-01')`, "mapping-"+externalID, id, externalID); err != nil {
			t.Fatal(err)
		}
	}
	if err := centers.ReplaceWorkItems(t.Context(), "ecos-ready", items); err != nil {
		t.Fatal(err)
	}
	*counts = listStatementCounts{}
	contexts, err := centers.LoadECoSBackupContexts(t.Context(), "ecos.test", 15471, objectIDs)
	if err != nil || len(contexts) != len(ids) || contexts[objectIDs[0]].VehicleID != ids[0] {
		t.Fatalf("wrong ECoS backup contexts: %d entries, %v", len(contexts), err)
	}
	t.Logf("ECoS backup context: %d reads, %d writes", counts.reads, counts.writes)
	if *counts != (listStatementCounts{reads: 2}) {
		t.Fatalf("ECoS backup context statements: %+v", *counts)
	}

	checks := make([]transfer.AccessoryReferenceCheck, len(ids))
	for i := range checks {
		checks[i] = transfer.AccessoryReferenceCheck{Accessory: transfer.TransferAccessory{
			ArticleType: "other", Subtype: "other:other",
		}, TargetID: fmt.Sprintf("product-%03d", i)}
	}
	checks[0].Accessory.Subtype = "other:missing"
	*counts = listStatementCounts{}
	valid, err := repository.ValidateTransferAccessoryReferences(t.Context(), checks)
	if err != nil || len(valid) != len(checks) || valid[0] || (len(valid) > 1 && !valid[1]) {
		t.Fatalf("wrong reference checks: %v, %v", valid, err)
	}
	t.Logf("transfer preview references: %d reads, %d writes", counts.reads, counts.writes)
	if *counts != (listStatementCounts{reads: 1}) {
		t.Fatalf("transfer preview reference statements: %+v", *counts)
	}
}

func seedListQueryRows(t *testing.T, db *sql.DB, size int) []string {
	t.Helper()
	run := func(statement string, args ...any) {
		t.Helper()
		if _, err := db.ExecContext(t.Context(), statement, args...); err != nil {
			t.Fatal(err)
		}
	}
	const now = "'2026-01-01T00:00:00Z'"
	run(`INSERT INTO storage_locations(id,name,created_at,updated_at) VALUES('shelf','Shelf',` + now + `,` + now + `)`)
	var ids []string
	for i := range size {
		id := fmt.Sprintf("vehicle-%03d", i)
		key := fmt.Sprintf("maker-%03d", i)
		profile := fmt.Sprintf("profile-%03d", i)
		user := fmt.Sprintf("user-%03d", i)
		product := fmt.Sprintf("product-%03d", i)
		event := fmt.Sprintf("event-%03d", i)
		ids = append(ids, id)
		run(`INSERT INTO vehicles(id,inventory_number,name,manufacturer,gauge,storage_details,decoder_type,created_at,updated_at)
VALUES(?,?,?,'MAKER-000','TT','Drawer','DCC',`+now+`,`+now+`)`, id, id, id)
		run(`INSERT INTO master_data_entries(id,type,key,label,origin,created_at,updated_at)
VALUES(?,'manufacturer',?,?,'custom',`+now+`,`+now+`)`, key, key, key)
		run(`INSERT INTO users(id,username,password_hash,created_at) VALUES(?,?,'unused',`+now+`)`, user, user)
		run(`INSERT INTO decoder_cv_workspaces(id,vehicle_id,profile_text,origin,created_at,updated_at)
VALUES(?,'vehicle-000',?,'manual',`+now+`,`+now+`)`, profile, profile)
		run(`INSERT INTO vehicle_cv_values(id,vehicle_id,decoder_profile,cv_number,value,created_at,updated_at)
VALUES(?,'vehicle-000',?,1,2,`+now+`,`+now+`)`, profile, profile)
		run(`INSERT INTO vehicle_cv_value_history(id,cv_value_id,vehicle_id,old_value,new_value,changed_at)
VALUES(?,?,'vehicle-000',1,2,`+now+`)`, profile, profile)
		run(`INSERT INTO accessory_products(id,inventory_number,manufacturer,name,category,tracking_mode,created_at,updated_at)
VALUES(?,?,'Example','Example','other','quantity',`+now+`,`+now+`)`, product, product)
		run(`INSERT INTO accessory_stock(product_id,location_id,quantity,updated_at) VALUES(?,'shelf',2,`+now+`)`, product)
		run(`INSERT INTO exhibition_lists(id,designation,list_date,created_at,updated_at)
VALUES(?,?,'2026-01-01',`+now+`,`+now+`)`, event, event)
		run(`INSERT INTO exhibition_entries(id,list_id,owner,locomotive_name,created_at,updated_at)
VALUES(?,?,'Owner','Vehicle',`+now+`,`+now+`)`, event, event)
	}
	run(`INSERT INTO vehicle_images(id,vehicle_id,url,is_primary,created_at)
VALUES('image','vehicle-000','https://example.test/image.png',1,` + now + `)`)
	run(`INSERT INTO user_roles(user_id,role_id) SELECT 'user-000',id FROM roles WHERE name IN ('Viewer','Editor')`)
	return ids
}

type listStatementCounts struct{ reads, writes int }

type countedListConnector struct {
	driver.Connector
	counts *listStatementCounts
}

func (c countedListConnector) Connect(ctx context.Context) (driver.Conn, error) {
	conn, err := c.Connector.Connect(ctx)
	if err != nil {
		return nil, err
	}
	return countedListConn{Conn: conn, counts: c.counts}, nil
}

type countedListConn struct {
	driver.Conn
	counts *listStatementCounts
}

func (c countedListConn) QueryContext(ctx context.Context, statement string, args []driver.NamedValue) (driver.Rows, error) {
	c.counts.reads++
	return c.Conn.(driver.QueryerContext).QueryContext(ctx, statement, args)
}

func (c countedListConn) ExecContext(ctx context.Context, statement string, args []driver.NamedValue) (driver.Result, error) {
	c.counts.writes++
	return c.Conn.(driver.ExecerContext).ExecContext(ctx, statement, args)
}

func (c countedListConn) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	return c.Conn.(driver.ConnBeginTx).BeginTx(ctx, opts)
}

func newCountedListDB(t *testing.T) (*sql.DB, *listStatementCounts) {
	t.Helper()
	connector, err := sqlite.NewConnector(filepath.Join(t.TempDir(), "query-count.db") + "?_pragma=foreign_keys(1)")
	if err != nil {
		t.Fatal(err)
	}
	counts := &listStatementCounts{}
	db := sql.OpenDB(countedListConnector{Connector: connector, counts: counts})
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = db.Close() })
	if err := infrastructure.Migrate(db, filepath.Join("..", "..", "migrations")); err != nil {
		t.Fatal(err)
	}
	if err := infrastructure.SeedRoles(db); err != nil {
		t.Fatal(err)
	}
	return db, counts
}
