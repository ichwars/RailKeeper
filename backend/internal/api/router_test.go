package api

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"railkeeper/backend/internal/application"
	"railkeeper/backend/internal/infrastructure"
)

type capturePasswordResetMailer struct {
	to        string
	resetURL  string
	expiresAt string
}

func (m *capturePasswordResetMailer) SendPasswordReset(_ context.Context, toEmail, resetURL, expiresAt string) error {
	m.to = toEmail
	m.resetURL = resetURL
	m.expiresAt = expiresAt
	return nil
}

func TestSecurityHeaders(t *testing.T) {
	handler := securityHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/health", nil))

	headers := recorder.Result().Header
	if headers.Get("X-Content-Type-Options") != "nosniff" {
		t.Fatalf("missing nosniff header")
	}
	if headers.Get("X-Frame-Options") != "SAMEORIGIN" {
		t.Fatalf("missing same-origin frame protection header")
	}
	if !strings.Contains(headers.Get("Content-Security-Policy"), "frame-ancestors 'self'") {
		t.Fatalf("missing same-origin frame ancestor policy")
	}
	permissionsPolicy := headers.Get("Permissions-Policy")
	if !strings.Contains(permissionsPolicy, "camera=(self)") {
		t.Fatalf("camera must be available to same-origin barcode scanner, got %q", permissionsPolicy)
	}
	if !strings.Contains(permissionsPolicy, "microphone=()") || !strings.Contains(permissionsPolicy, "geolocation=()") {
		t.Fatalf("expected microphone and geolocation to stay disabled, got %q", permissionsPolicy)
	}
}

func TestConfinedDataPathRejectsEscapes(t *testing.T) {
	dataDir := t.TempDir()
	if _, err := confinedDataPath(dataDir, "uploads/vehicle/manual.pdf"); err != nil {
		t.Fatalf("expected valid confined path: %v", err)
	}
	if _, err := confinedDataPath(dataDir, "../outside.pdf"); err == nil {
		t.Fatalf("expected path escape to be rejected")
	}
}

func TestAttachmentSafetyHelpers(t *testing.T) {
	if !isBlockedAttachmentName("setup.exe") {
		t.Fatalf("expected executable extension to be blocked")
	}
	if isBlockedAttachmentName("manual.pdf") {
		t.Fatalf("expected pdf extension to be allowed")
	}
	if !isBlockedAttachmentMime("application/x-msdownload") {
		t.Fatalf("expected executable mime type to be blocked")
	}
	if isBlockedAttachmentMime("application/pdf") {
		t.Fatalf("expected pdf mime type to be allowed")
	}
	if cleanOriginalFileName(`C:\Users\daniel\Downloads\manual.pdf`) != "manual.pdf" {
		t.Fatalf("expected original filename cleanup to strip client path")
	}
	if !isAllowedAttachmentUpload("manual.pdf", "application/pdf") {
		t.Fatalf("expected pdf upload to be allowed")
	}
	if !isAllowedAttachmentUpload("daten.json", "text/plain; charset=utf-8") {
		t.Fatalf("expected text-like json upload to be allowed")
	}
	if isAllowedAttachmentUpload("manual.pdf", "text/plain; charset=utf-8") {
		t.Fatalf("expected mismatched pdf mime type to be rejected")
	}
	if isAllowedAttachmentUpload("script.js", "text/plain; charset=utf-8") {
		t.Fatalf("expected blocked executable-like extension to be rejected")
	}
	if isAllowedAttachmentUpload("unknown.bin", "application/octet-stream") {
		t.Fatalf("expected unknown attachment type to be rejected")
	}
	onlyPDF := map[string]struct{}{".pdf": {}}
	if !isAllowedAttachmentUploadWithExtensions("manual.pdf", "application/pdf", onlyPDF) {
		t.Fatalf("expected configured pdf extension to be allowed")
	}
	if isAllowedAttachmentUploadWithExtensions("daten.json", "application/json", onlyPDF) {
		t.Fatalf("expected non-configured json extension to be rejected")
	}
	unsafeOnly := effectiveAttachmentExtensions(map[string]struct{}{".exe": {}})
	if _, ok := unsafeOnly[".exe"]; ok {
		t.Fatalf("expected unsafe configured extension to be ignored")
	}
}

func TestRemoteDocumentHTTPClientDoesNotFollowPrivateRedirect(t *testing.T) {
	client := remoteDocumentHTTPClient(t.Context())
	initial := httptest.NewRequest(http.MethodGet, "https://example.com/manual.pdf", nil)
	redirect := httptest.NewRequest(http.MethodGet, "http://127.0.0.1/private.pdf", nil)

	err := client.CheckRedirect(redirect, []*http.Request{initial})
	if !errors.Is(err, http.ErrUseLastResponse) {
		t.Fatalf("expected private redirect to be returned without following, got %v", err)
	}
}

func TestRemoteDocumentHTTPClientRejectsPrivateInitialURLBeforeRequest(t *testing.T) {
	requested := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requested = true
		_, _ = w.Write([]byte("private document"))
	}))
	defer server.Close()

	client := remoteDocumentHTTPClient(t.Context())
	resp, err := client.Get(server.URL)
	if resp != nil {
		defer func() { _ = resp.Body.Close() }()
	}
	if err == nil {
		t.Fatal("expected private document URL to be rejected")
	}
	if requested {
		t.Fatal("private document URL should not be requested")
	}
}

func TestCreateVehicleImageThumbnailSupportsWebP(t *testing.T) {
	data, err := base64.StdEncoding.DecodeString("UklGRiIAAABXRUJQVlA4IBYAAAAwAQCdASoBAAEADsD+JaQAA3AAAAAA")
	if err != nil {
		t.Fatal(err)
	}
	dataDir := t.TempDir()
	db := testRouterDBWithDataDir(t, dataDir)
	blobs := application.NewFileBlobService(db, dataDir)
	app := &App{dataDir: dataDir, fileBlobs: blobs}

	thumbnailBlobID, err := app.createVehicleImageThumbnail(t.Context(), data, "side.webp")
	if err != nil {
		t.Fatalf("expected webp thumbnail: %v", err)
	}
	if thumbnailBlobID == "" {
		t.Fatalf("expected jpeg thumbnail blob id")
	}
	thumbnailData, err := blobs.Load(t.Context(), thumbnailBlobID)
	if err != nil {
		t.Fatalf("expected thumbnail blob: %v", err)
	}
	if contentType := http.DetectContentType(thumbnailData); contentType != "image/jpeg" {
		t.Fatalf("expected jpeg thumbnail, got %q", contentType)
	}
}

func TestRateLimiterBlocksAfterLimit(t *testing.T) {
	limiter := newRateLimiter()
	allowed, err := limiter.Allow(t.Context(), "login", "127.0.0.1", 2, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if !allowed {
		t.Fatalf("first attempt should be allowed")
	}
	allowed, err = limiter.Allow(t.Context(), "login", "127.0.0.1", 2, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if !allowed {
		t.Fatalf("second attempt should be allowed")
	}
	allowed, err = limiter.Allow(t.Context(), "login", "127.0.0.1", 2, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if allowed {
		t.Fatalf("third attempt should be blocked")
	}
}

func TestPrinterHelpers(t *testing.T) {
	if got := parseLPStatDefault("system default destination: Office_Printer"); got != "Office_Printer" {
		t.Fatalf("unexpected default printer %q", got)
	}
	printers := printersFromNames([]string{"Office Printer", "Office Printer", "Label"}, "Label")
	if len(printers) != 2 {
		t.Fatalf("expected deduplicated printers, got %#v", printers)
	}
	if printers[0].ID != "office-printer" || printers[1].ID != "label" || !printers[1].IsDefault {
		t.Fatalf("unexpected printers: %#v", printers)
	}
}

func TestAuditLimit(t *testing.T) {
	cases := map[string]int{
		"":    50,
		"-1":  50,
		"25":  25,
		"500": 200,
	}
	for input, want := range cases {
		if got := auditLimit(input); got != want {
			t.Fatalf("auditLimit(%q) = %d, want %d", input, got, want)
		}
	}
}

func TestSessionLimit(t *testing.T) {
	cases := map[string]int{
		"":    200,
		"-1":  200,
		"5":   5,
		"500": 200,
	}
	for input, want := range cases {
		if got := sessionLimit(input); got != want {
			t.Fatalf("sessionLimit(%q) = %d, want %d", input, got, want)
		}
	}
}

func TestChangePasswordEndpoint(t *testing.T) {
	db := testRouterDB(t)
	setup := application.NewSetupService(db)
	auth := application.NewAuthService(db)
	if err := setup.CreateAdmin(t.Context(), application.CreateAdminInput{
		Username: "admin",
		Email:    "admin@example.test",
		Password: "very-secure-password",
	}); err != nil {
		t.Fatal(err)
	}

	mailer := &capturePasswordResetMailer{}
	router := NewRouter(Config{SetupService: setup, AuthService: auth, PasswordResetMailer: mailer})
	loginBody := bytes.NewBufferString(`{"username":"admin","password":"very-secure-password"}`)
	loginRequest := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", loginBody)
	loginRequest.Header.Set("Content-Type", "application/json")
	loginResponse := httptest.NewRecorder()
	router.ServeHTTP(loginResponse, loginRequest)
	if loginResponse.Code != http.StatusOK {
		t.Fatalf("expected login success, got %d", loginResponse.Code)
	}
	var session application.SessionView
	if err := json.NewDecoder(loginResponse.Body).Decode(&session); err != nil {
		t.Fatal(err)
	}

	changeBody := bytes.NewBufferString(`{"currentPassword":"very-secure-password","newPassword":"new-secure-password"}`)
	changeRequest := httptest.NewRequest(http.MethodPut, "/api/v1/auth/password", changeBody)
	changeRequest.Header.Set("Content-Type", "application/json")
	changeRequest.Header.Set("X-CSRF-Token", session.CSRFToken)
	for _, cookie := range loginResponse.Result().Cookies() {
		changeRequest.AddCookie(cookie)
	}
	changeResponse := httptest.NewRecorder()
	router.ServeHTTP(changeResponse, changeRequest)
	if changeResponse.Code != http.StatusNoContent {
		t.Fatalf("expected password change success, got %d: %s", changeResponse.Code, changeResponse.Body.String())
	}

	oldLogin := httptest.NewRecorder()
	router.ServeHTTP(oldLogin, httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(`{"username":"admin","password":"very-secure-password"}`)))
	if oldLogin.Code != http.StatusUnauthorized {
		t.Fatalf("expected old password to fail, got %d", oldLogin.Code)
	}
	newLogin := httptest.NewRecorder()
	router.ServeHTTP(newLogin, httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(`{"username":"admin","password":"new-secure-password"}`)))
	if newLogin.Code != http.StatusOK {
		t.Fatalf("expected new password to work, got %d", newLogin.Code)
	}
}

func TestPasswordResetEndpointCompletesPasswordChange(t *testing.T) {
	db := testRouterDB(t)
	setup := application.NewSetupService(db)
	auth := application.NewAuthService(db)
	if err := setup.CreateAdmin(t.Context(), application.CreateAdminInput{
		Username: "admin",
		Email:    "admin@example.test",
		Password: "very-secure-password",
	}); err != nil {
		t.Fatal(err)
	}

	mailer := &capturePasswordResetMailer{}
	router := NewRouter(Config{SetupService: setup, AuthService: auth, PasswordResetMailer: mailer})
	resetRequest := httptest.NewRequest(http.MethodPost, "/api/v1/auth/password-reset", bytes.NewBufferString(`{"email":"admin@example.test"}`))
	resetRequest.Host = "railkeeper.test"
	resetResponse := httptest.NewRecorder()
	router.ServeHTTP(resetResponse, resetRequest)
	if resetResponse.Code != http.StatusAccepted {
		t.Fatalf("expected reset request success, got %d: %s", resetResponse.Code, resetResponse.Body.String())
	}
	var resetResult application.PasswordResetRequestResult
	if err := json.NewDecoder(resetResponse.Body).Decode(&resetResult); err != nil {
		t.Fatal(err)
	}
	if resetResult.ResetURL != "" {
		t.Fatalf("reset url must not be returned to the browser, got %#v", resetResult.ResetURL)
	}
	if mailer.to != "admin@example.test" || mailer.resetURL == "" || mailer.expiresAt == "" {
		t.Fatalf("expected reset link to be sent by mailer, got %#v", mailer)
	}
	resetURL, err := url.Parse(mailer.resetURL)
	if err != nil {
		t.Fatal(err)
	}
	token := resetURL.Query().Get("token")
	if token == "" {
		t.Fatalf("expected reset token URL, got %#v", resetResult)
	}

	confirmBody := bytes.NewBufferString(`{"token":"` + token + `","newPassword":"new-secure-password"}`)
	confirmRequest := httptest.NewRequest(http.MethodPost, "/api/v1/auth/password-reset/confirm", confirmBody)
	confirmResponse := httptest.NewRecorder()
	router.ServeHTTP(confirmResponse, confirmRequest)
	if confirmResponse.Code != http.StatusNoContent {
		t.Fatalf("expected reset confirmation success, got %d: %s", confirmResponse.Code, confirmResponse.Body.String())
	}

	oldLogin := httptest.NewRecorder()
	router.ServeHTTP(oldLogin, httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(`{"username":"admin","password":"very-secure-password"}`)))
	if oldLogin.Code != http.StatusUnauthorized {
		t.Fatalf("expected old password to fail, got %d", oldLogin.Code)
	}
	newLogin := httptest.NewRecorder()
	router.ServeHTTP(newLogin, httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(`{"username":"admin","password":"new-secure-password"}`)))
	if newLogin.Code != http.StatusOK {
		t.Fatalf("expected new password to work, got %d", newLogin.Code)
	}
}

func TestSessionListAndRevokeEndpoints(t *testing.T) {
	db := testRouterDB(t)
	setup := application.NewSetupService(db)
	auth := application.NewAuthService(db)
	if err := setup.CreateAdmin(t.Context(), application.CreateAdminInput{
		Username: "admin",
		Email:    "admin@example.test",
		Password: "very-secure-password",
	}); err != nil {
		t.Fatal(err)
	}

	router := NewRouter(Config{SetupService: setup, AuthService: auth})
	loginBody := bytes.NewBufferString(`{"username":"admin","password":"very-secure-password"}`)
	loginRequest := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", loginBody)
	loginResponse := httptest.NewRecorder()
	router.ServeHTTP(loginResponse, loginRequest)
	if loginResponse.Code != http.StatusOK {
		t.Fatalf("expected login success, got %d", loginResponse.Code)
	}
	var session application.SessionView
	if err := json.NewDecoder(loginResponse.Body).Decode(&session); err != nil {
		t.Fatal(err)
	}

	listRequest := httptest.NewRequest(http.MethodGet, "/api/v1/sessions", nil)
	for _, cookie := range loginResponse.Result().Cookies() {
		listRequest.AddCookie(cookie)
	}
	listResponse := httptest.NewRecorder()
	router.ServeHTTP(listResponse, listRequest)
	if listResponse.Code != http.StatusOK {
		t.Fatalf("expected session list success, got %d: %s", listResponse.Code, listResponse.Body.String())
	}
	var sessions []application.SessionRecord
	if err := json.NewDecoder(listResponse.Body).Decode(&sessions); err != nil {
		t.Fatal(err)
	}
	if len(sessions) != 1 || !sessions[0].Active {
		t.Fatalf("expected one active session, got %#v", sessions)
	}

	revokeRequest := httptest.NewRequest(http.MethodPut, "/api/v1/sessions/"+sessions[0].ID+"/revoke", nil)
	revokeRequest.Header.Set("X-CSRF-Token", session.CSRFToken)
	for _, cookie := range loginResponse.Result().Cookies() {
		revokeRequest.AddCookie(cookie)
	}
	revokeResponse := httptest.NewRecorder()
	router.ServeHTTP(revokeResponse, revokeRequest)
	if revokeResponse.Code != http.StatusNoContent {
		t.Fatalf("expected revoke success, got %d: %s", revokeResponse.Code, revokeResponse.Body.String())
	}

	currentRequest := httptest.NewRequest(http.MethodGet, "/api/v1/auth/session", nil)
	for _, cookie := range loginResponse.Result().Cookies() {
		currentRequest.AddCookie(cookie)
	}
	currentResponse := httptest.NewRecorder()
	router.ServeHTTP(currentResponse, currentRequest)
	if currentResponse.Code != http.StatusUnauthorized {
		t.Fatalf("expected revoked session to be unauthorized, got %d", currentResponse.Code)
	}
}

func TestSessionListEndpointHonorsLimit(t *testing.T) {
	db := testRouterDB(t)
	setup := application.NewSetupService(db)
	auth := application.NewAuthService(db)
	if err := setup.CreateAdmin(t.Context(), application.CreateAdminInput{
		Username: "admin",
		Email:    "admin@example.test",
		Password: "very-secure-password",
	}); err != nil {
		t.Fatal(err)
	}

	router := NewRouter(Config{SetupService: setup, AuthService: auth})
	var cookies []*http.Cookie
	for range 4 {
		_, cookies = loginTestUser(t, router, "admin", "very-secure-password")
	}

	request := httptest.NewRequest(http.MethodGet, "/api/v1/sessions?limit=2", nil)
	for _, cookie := range cookies {
		request.AddCookie(cookie)
	}
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected limited session list success, got %d: %s", response.Code, response.Body.String())
	}
	var sessions []application.SessionRecord
	if err := json.NewDecoder(response.Body).Decode(&sessions); err != nil {
		t.Fatal(err)
	}
	if len(sessions) != 2 {
		t.Fatalf("expected two sessions from limited endpoint, got %#v", sessions)
	}
}

func TestBackupValidateEndpoint(t *testing.T) {
	dataDir := t.TempDir()
	db := testRouterDBWithDataDir(t, dataDir)
	setup := application.NewSetupService(db)
	auth := application.NewAuthService(db)
	backupService := application.NewBackupService(db, dataDir)
	fileBlobs := application.NewFileBlobService(db, dataDir)
	if err := setup.CreateAdmin(t.Context(), application.CreateAdminInput{
		Username: "admin",
		Email:    "admin@example.test",
		Password: "very-secure-password",
	}); err != nil {
		t.Fatal(err)
	}
	backup, err := backupService.Export(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	backupData, err := json.Marshal(backup)
	if err != nil {
		t.Fatal(err)
	}

	router := NewRouter(Config{SetupService: setup, AuthService: auth, BackupService: backupService, FileBlobService: fileBlobs, DataDir: dataDir})
	session, cookies := loginTestUser(t, router, "admin", "very-secure-password")
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "railkeeper-backup.json")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write(backupData); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}

	request := httptest.NewRequest(http.MethodPost, "/api/v1/backup/validate", body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	request.Header.Set("X-CSRF-Token", session.CSRFToken)
	for _, cookie := range cookies {
		request.AddCookie(cookie)
	}
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected backup validation success, got %d: %s", response.Code, response.Body.String())
	}
	var validation application.BackupValidationResult
	if err := json.NewDecoder(response.Body).Decode(&validation); err != nil {
		t.Fatal(err)
	}
	if !validation.Compatible || len(validation.Errors) > 0 || validation.TableCount == 0 {
		t.Fatalf("expected compatible backup validation response, got %#v", validation)
	}
}

func TestExhibitionEndpointsAllowMesseRole(t *testing.T) {
	db := testRouterDB(t)
	setup := application.NewSetupService(db)
	auth := application.NewAuthService(db)
	exhibition := application.NewExhibitionService(db)
	masterData := application.NewMasterDataService(db)
	if err := setup.CreateAdmin(t.Context(), application.CreateAdminInput{
		Username: "admin",
		Email:    "admin@example.test",
		Password: "very-secure-password",
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := auth.CreateUser(t.Context(), "", application.CreateUserInput{
		Username: "messe",
		Email:    "messe@example.test",
		Password: "messe-secure-password",
		Roles:    []string{"Messe"},
	}); err != nil {
		t.Fatal(err)
	}
	list, err := exhibition.Create(t.Context(), application.ExhibitionListInput{
		Designation: "Leipzig 2026",
		Date:        "2026-05-12",
	})
	if err != nil {
		t.Fatal(err)
	}

	router := NewRouter(Config{SetupService: setup, AuthService: auth, ExhibitionService: exhibition, MasterDataService: masterData})
	session, cookies := loginTestUser(t, router, "messe", "messe-secure-password")

	listRequest := httptest.NewRequest(http.MethodGet, "/api/v1/exhibition-lists", nil)
	for _, cookie := range cookies {
		listRequest.AddCookie(cookie)
	}
	listResponse := httptest.NewRecorder()
	router.ServeHTTP(listResponse, listRequest)
	if listResponse.Code != http.StatusOK {
		t.Fatalf("expected messe user to list exhibition lists, got %d: %s", listResponse.Code, listResponse.Body.String())
	}

	createRequest := httptest.NewRequest(http.MethodPost, "/api/v1/exhibition-lists", bytes.NewBufferString(`{"designation":"Leipzig","date":"2026-05-12"}`))
	createRequest.Header.Set("Content-Type", "application/json")
	createRequest.Header.Set("X-CSRF-Token", session.CSRFToken)
	for _, cookie := range cookies {
		createRequest.AddCookie(cookie)
	}
	createResponse := httptest.NewRecorder()
	router.ServeHTTP(createResponse, createRequest)
	if createResponse.Code != http.StatusForbidden {
		t.Fatalf("expected messe user to be forbidden from creating lists, got %d", createResponse.Code)
	}

	entryCreateRequest := httptest.NewRequest(http.MethodPost, "/api/v1/exhibition-lists/"+list.ID+"/entries", bytes.NewBufferString(`{"owner":"Daniel","locomotiveName":"V180","dtDecoder":true,"decoderNumber":"1001"}`))
	entryCreateRequest.Header.Set("Content-Type", "application/json")
	entryCreateRequest.Header.Set("X-CSRF-Token", session.CSRFToken)
	for _, cookie := range cookies {
		entryCreateRequest.AddCookie(cookie)
	}
	entryCreateResponse := httptest.NewRecorder()
	router.ServeHTTP(entryCreateResponse, entryCreateRequest)
	if entryCreateResponse.Code != http.StatusCreated {
		t.Fatalf("expected messe user to create entries, got %d: %s", entryCreateResponse.Code, entryCreateResponse.Body.String())
	}
	var entry application.ExhibitionEntry
	if err := json.NewDecoder(entryCreateResponse.Body).Decode(&entry); err != nil {
		t.Fatal(err)
	}

	entryUpdateRequest := httptest.NewRequest(http.MethodPut, "/api/v1/exhibition-lists/"+list.ID+"/entries/"+entry.ID, bytes.NewBufferString(`{"owner":"Daniel","locomotiveName":"V180 DR","dtDecoder":true,"decoderNumber":"1001"}`))
	entryUpdateRequest.Header.Set("Content-Type", "application/json")
	entryUpdateRequest.Header.Set("X-CSRF-Token", session.CSRFToken)
	for _, cookie := range cookies {
		entryUpdateRequest.AddCookie(cookie)
	}
	entryUpdateResponse := httptest.NewRecorder()
	router.ServeHTTP(entryUpdateResponse, entryUpdateRequest)
	if entryUpdateResponse.Code != http.StatusOK {
		t.Fatalf("expected messe user to update entries, got %d: %s", entryUpdateResponse.Code, entryUpdateResponse.Body.String())
	}

	detailRequest := httptest.NewRequest(http.MethodGet, "/api/v1/exhibition-lists/"+list.ID, nil)
	for _, cookie := range cookies {
		detailRequest.AddCookie(cookie)
	}
	detailResponse := httptest.NewRecorder()
	router.ServeHTTP(detailResponse, detailRequest)
	if detailResponse.Code != http.StatusOK {
		t.Fatalf("expected messe user to read list details, got %d: %s", detailResponse.Code, detailResponse.Body.String())
	}
	var detail application.ExhibitionList
	if err := json.NewDecoder(detailResponse.Body).Decode(&detail); err != nil {
		t.Fatal(err)
	}
	if len(detail.Entries) != 1 || detail.Entries[0].ID != entry.ID {
		t.Fatalf("expected list detail response to include entry, got %#v", detail)
	}

	entryDeleteRequest := httptest.NewRequest(http.MethodDelete, "/api/v1/exhibition-lists/"+list.ID+"/entries/"+entry.ID, nil)
	entryDeleteRequest.Header.Set("X-CSRF-Token", session.CSRFToken)
	for _, cookie := range cookies {
		entryDeleteRequest.AddCookie(cookie)
	}
	entryDeleteResponse := httptest.NewRecorder()
	router.ServeHTTP(entryDeleteResponse, entryDeleteRequest)
	if entryDeleteResponse.Code != http.StatusForbidden {
		t.Fatalf("expected messe user to be forbidden from deleting entries, got %d", entryDeleteResponse.Code)
	}

	vehicleRequest := httptest.NewRequest(http.MethodGet, "/api/v1/vehicles", nil)
	for _, cookie := range cookies {
		vehicleRequest.AddCookie(cookie)
	}
	vehicleResponse := httptest.NewRecorder()
	router.ServeHTTP(vehicleResponse, vehicleRequest)
	if vehicleResponse.Code != http.StatusForbidden {
		t.Fatalf("expected messe user to be forbidden from viewer inventory endpoints, got %d", vehicleResponse.Code)
	}

	symbolRequest := httptest.NewRequest(http.MethodGet, "/api/v1/master-data/symbols?active=true", nil)
	for _, cookie := range cookies {
		symbolRequest.AddCookie(cookie)
	}
	symbolResponse := httptest.NewRecorder()
	router.ServeHTTP(symbolResponse, symbolRequest)
	if symbolResponse.Code != http.StatusOK {
		t.Fatalf("expected messe user to read symbols for exhibition picker, got %d: %s", symbolResponse.Code, symbolResponse.Body.String())
	}
}

func TestExhibitionLockedListRejectsEntryWrites(t *testing.T) {
	db := testRouterDB(t)
	setup := application.NewSetupService(db)
	auth := application.NewAuthService(db)
	exhibition := application.NewExhibitionService(db)
	if err := setup.CreateAdmin(t.Context(), application.CreateAdminInput{
		Username: "admin",
		Email:    "admin@example.test",
		Password: "very-secure-password",
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := auth.CreateUser(t.Context(), "", application.CreateUserInput{
		Username: "messe",
		Email:    "messe@example.test",
		Password: "messe-secure-password",
		Roles:    []string{"Messe"},
	}); err != nil {
		t.Fatal(err)
	}
	list, err := exhibition.Create(t.Context(), application.ExhibitionListInput{
		Designation: "Leipzig 2026",
		Date:        "2026-05-12",
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := exhibition.SetLocked(t.Context(), list.ID, true); err != nil {
		t.Fatal(err)
	}

	router := NewRouter(Config{SetupService: setup, AuthService: auth, ExhibitionService: exhibition})
	session, cookies := loginTestUser(t, router, "messe", "messe-secure-password")
	body := bytes.NewBufferString(`{"owner":"Daniel","locomotiveName":"V180","dtDecoder":true,"decoderNumber":"1001"}`)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/exhibition-lists/"+list.ID+"/entries", body)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-CSRF-Token", session.CSRFToken)
	for _, cookie := range cookies {
		request.AddCookie(cookie)
	}
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusConflict {
		t.Fatalf("expected locked list conflict, got %d: %s", response.Code, response.Body.String())
	}
}

func TestExhibitionListCreateWritesAuditLog(t *testing.T) {
	db := testRouterDB(t)
	setup := application.NewSetupService(db)
	auth := application.NewAuthService(db)
	exhibition := application.NewExhibitionService(db)
	if err := setup.CreateAdmin(t.Context(), application.CreateAdminInput{
		Username: "admin",
		Email:    "admin@example.test",
		Password: "very-secure-password",
	}); err != nil {
		t.Fatal(err)
	}

	router := NewRouter(Config{SetupService: setup, AuthService: auth, ExhibitionService: exhibition})
	session, cookies := loginTestUser(t, router, "admin", "very-secure-password")
	request := httptest.NewRequest(http.MethodPost, "/api/v1/exhibition-lists", bytes.NewBufferString(`{"designation":"Leipzig","date":"2026-05-12"}`))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-CSRF-Token", session.CSRFToken)
	for _, cookie := range cookies {
		request.AddCookie(cookie)
	}
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusCreated {
		t.Fatalf("expected list create success, got %d: %s", response.Code, response.Body.String())
	}

	entries, err := auth.ListAuditLog(t.Context(), 10)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.Action == "ExhibitionListCreated" && entry.TargetType == "exhibition_list" && entry.ActorUsername == "admin" {
			return
		}
	}
	t.Fatalf("expected exhibition audit entry, got %#v", entries)
}

func TestExhibitionListChangesWriteAuditLog(t *testing.T) {
	db := testRouterDB(t)
	setup := application.NewSetupService(db)
	auth := application.NewAuthService(db)
	exhibition := application.NewExhibitionService(db)
	if err := setup.CreateAdmin(t.Context(), application.CreateAdminInput{
		Username: "admin",
		Email:    "admin@example.test",
		Password: "very-secure-password",
	}); err != nil {
		t.Fatal(err)
	}
	list, err := exhibition.Create(t.Context(), application.ExhibitionListInput{
		Designation: "Leipzig 2026",
		Date:        "2026-05-12",
	})
	if err != nil {
		t.Fatal(err)
	}

	router := NewRouter(Config{SetupService: setup, AuthService: auth, ExhibitionService: exhibition})
	session, cookies := loginTestUser(t, router, "admin", "very-secure-password")
	updateRequest := httptest.NewRequest(http.MethodPut, "/api/v1/exhibition-lists/"+list.ID, bytes.NewBufferString(`{"designation":"Leipzig 2026 aktualisiert","date":"2026-05-13"}`))
	updateRequest.Header.Set("Content-Type", "application/json")
	updateRequest.Header.Set("X-CSRF-Token", session.CSRFToken)
	for _, cookie := range cookies {
		updateRequest.AddCookie(cookie)
	}
	updateResponse := httptest.NewRecorder()
	router.ServeHTTP(updateResponse, updateRequest)
	if updateResponse.Code != http.StatusOK {
		t.Fatalf("expected list update success, got %d: %s", updateResponse.Code, updateResponse.Body.String())
	}

	lockRequest := httptest.NewRequest(http.MethodPut, "/api/v1/exhibition-lists/"+list.ID+"/lock", bytes.NewBufferString(`{"locked":true}`))
	lockRequest.Header.Set("Content-Type", "application/json")
	lockRequest.Header.Set("X-CSRF-Token", session.CSRFToken)
	for _, cookie := range cookies {
		lockRequest.AddCookie(cookie)
	}
	lockResponse := httptest.NewRecorder()
	router.ServeHTTP(lockResponse, lockRequest)
	if lockResponse.Code != http.StatusOK {
		t.Fatalf("expected list lock success, got %d: %s", lockResponse.Code, lockResponse.Body.String())
	}

	unlockRequest := httptest.NewRequest(http.MethodPut, "/api/v1/exhibition-lists/"+list.ID+"/lock", bytes.NewBufferString(`{"locked":false}`))
	unlockRequest.Header.Set("Content-Type", "application/json")
	unlockRequest.Header.Set("X-CSRF-Token", session.CSRFToken)
	for _, cookie := range cookies {
		unlockRequest.AddCookie(cookie)
	}
	unlockResponse := httptest.NewRecorder()
	router.ServeHTTP(unlockResponse, unlockRequest)
	if unlockResponse.Code != http.StatusOK {
		t.Fatalf("expected list unlock success, got %d: %s", unlockResponse.Code, unlockResponse.Body.String())
	}

	deleteRequest := httptest.NewRequest(http.MethodDelete, "/api/v1/exhibition-lists/"+list.ID, nil)
	deleteRequest.Header.Set("X-CSRF-Token", session.CSRFToken)
	for _, cookie := range cookies {
		deleteRequest.AddCookie(cookie)
	}
	deleteResponse := httptest.NewRecorder()
	router.ServeHTTP(deleteResponse, deleteRequest)
	if deleteResponse.Code != http.StatusNoContent {
		t.Fatalf("expected list delete success, got %d: %s", deleteResponse.Code, deleteResponse.Body.String())
	}

	entries, err := auth.ListAuditLog(t.Context(), 20)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]bool{
		"ExhibitionListUpdated":  false,
		"ExhibitionListLocked":   false,
		"ExhibitionListUnlocked": false,
		"ExhibitionListDeleted":  false,
	}
	for _, auditEntry := range entries {
		if _, ok := want[auditEntry.Action]; ok && auditEntry.TargetType == "exhibition_list" && auditEntry.TargetID == list.ID && auditEntry.ActorUsername == "admin" {
			want[auditEntry.Action] = true
		}
	}
	for action, seen := range want {
		if !seen {
			t.Fatalf("expected audit action %s for list %s, got %#v", action, list.ID, entries)
		}
	}
}

func TestExhibitionEntryChangesWriteAuditLog(t *testing.T) {
	db := testRouterDB(t)
	setup := application.NewSetupService(db)
	auth := application.NewAuthService(db)
	exhibition := application.NewExhibitionService(db)
	if err := setup.CreateAdmin(t.Context(), application.CreateAdminInput{
		Username: "admin",
		Email:    "admin@example.test",
		Password: "very-secure-password",
	}); err != nil {
		t.Fatal(err)
	}
	list, err := exhibition.Create(t.Context(), application.ExhibitionListInput{
		Designation: "Leipzig 2026",
		Date:        "2026-05-12",
	})
	if err != nil {
		t.Fatal(err)
	}

	router := NewRouter(Config{SetupService: setup, AuthService: auth, ExhibitionService: exhibition})
	session, cookies := loginTestUser(t, router, "admin", "very-secure-password")
	createRequest := httptest.NewRequest(http.MethodPost, "/api/v1/exhibition-lists/"+list.ID+"/entries", bytes.NewBufferString(`{"owner":"Daniel","locomotiveName":"V180","dtDecoder":true,"decoderNumber":"1001","functionKeys":"F0 Licht"}`))
	createRequest.Header.Set("Content-Type", "application/json")
	createRequest.Header.Set("X-CSRF-Token", session.CSRFToken)
	for _, cookie := range cookies {
		createRequest.AddCookie(cookie)
	}
	createResponse := httptest.NewRecorder()
	router.ServeHTTP(createResponse, createRequest)
	if createResponse.Code != http.StatusCreated {
		t.Fatalf("expected entry create success, got %d: %s", createResponse.Code, createResponse.Body.String())
	}
	var entry application.ExhibitionEntry
	if err := json.NewDecoder(createResponse.Body).Decode(&entry); err != nil {
		t.Fatal(err)
	}

	updateRequest := httptest.NewRequest(http.MethodPut, "/api/v1/exhibition-lists/"+list.ID+"/entries/"+entry.ID, bytes.NewBufferString(`{"owner":"Daniel","locomotiveName":"V180 DR","dtDecoder":true,"decoderNumber":"1001","functionKeys":"F0 Licht, F1 Sound"}`))
	updateRequest.Header.Set("Content-Type", "application/json")
	updateRequest.Header.Set("X-CSRF-Token", session.CSRFToken)
	for _, cookie := range cookies {
		updateRequest.AddCookie(cookie)
	}
	updateResponse := httptest.NewRecorder()
	router.ServeHTTP(updateResponse, updateRequest)
	if updateResponse.Code != http.StatusOK {
		t.Fatalf("expected entry update success, got %d: %s", updateResponse.Code, updateResponse.Body.String())
	}

	deleteRequest := httptest.NewRequest(http.MethodDelete, "/api/v1/exhibition-lists/"+list.ID+"/entries/"+entry.ID, nil)
	deleteRequest.Header.Set("X-CSRF-Token", session.CSRFToken)
	for _, cookie := range cookies {
		deleteRequest.AddCookie(cookie)
	}
	deleteResponse := httptest.NewRecorder()
	router.ServeHTTP(deleteResponse, deleteRequest)
	if deleteResponse.Code != http.StatusNoContent {
		t.Fatalf("expected entry delete success, got %d: %s", deleteResponse.Code, deleteResponse.Body.String())
	}

	entries, err := auth.ListAuditLog(t.Context(), 20)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]bool{
		"ExhibitionEntryCreated": false,
		"ExhibitionEntryUpdated": false,
		"ExhibitionEntryDeleted": false,
	}
	for _, auditEntry := range entries {
		if _, ok := want[auditEntry.Action]; ok && auditEntry.TargetType == "exhibition_entry" && auditEntry.TargetID == entry.ID && auditEntry.ActorUsername == "admin" {
			want[auditEntry.Action] = true
		}
	}
	for action, seen := range want {
		if !seen {
			t.Fatalf("expected audit action %s for entry %s, got %#v", action, entry.ID, entries)
		}
	}
}

func loginTestUser(t *testing.T, router http.Handler, username, password string) (application.SessionView, []*http.Cookie) {
	t.Helper()
	loginBody := bytes.NewBufferString(`{"username":"` + username + `","password":"` + password + `"}`)
	loginRequest := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", loginBody)
	loginRequest.Header.Set("Content-Type", "application/json")
	loginResponse := httptest.NewRecorder()
	router.ServeHTTP(loginResponse, loginRequest)
	if loginResponse.Code != http.StatusOK {
		t.Fatalf("expected login success, got %d: %s", loginResponse.Code, loginResponse.Body.String())
	}
	var session application.SessionView
	if err := json.NewDecoder(loginResponse.Body).Decode(&session); err != nil {
		t.Fatal(err)
	}
	return session, loginResponse.Result().Cookies()
}

func testRouterDB(t *testing.T) *sql.DB {
	t.Helper()
	return testRouterDBWithDataDir(t, t.TempDir())
}

func testRouterDBWithDataDir(t *testing.T, dataDir string) *sql.DB {
	t.Helper()
	db, err := infrastructure.OpenSQLite(dataDir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := infrastructure.Migrate(db, filepath.Join("..", "..", "migrations")); err != nil {
		t.Fatal(err)
	}
	if err := infrastructure.SeedRoles(db); err != nil {
		t.Fatal(err)
	}
	return db
}
