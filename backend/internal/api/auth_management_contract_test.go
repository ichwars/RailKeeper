package api

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"strings"
	"testing"
	"time"

	"railkeeper/backend/internal/application"
)

func TestUserManagementHTTPContractsKeepLastAdminAndSessionBoundaries(t *testing.T) {
	f := newContractFixture(t)
	path := "/api/v1/users"
	assertStatus(t, f.call(t, "GET", path, nil), 200)
	input := map[string]any{"username": "new-user", "email": "new@example.test", "password": "synthetic-new-user-password", "roles": []string{"Viewer"}}
	created := f.call(t, "POST", path, input)
	assertStatus(t, created, 201)
	id := responseID(t, created)
	assertProblem(t, f.call(t, "POST", path, input), 409, "user_duplicate")
	assertProblem(t, f.raw(t, "POST", path, "{"), 400, "invalid_json")
	assertProblem(t, f.call(t, "POST", path, map[string]any{"username": "invalid"}), 400, "user_validation")
	input["username"] = "renamed-user"
	input["password"] = ""
	input["roles"] = []string{"Editor"}
	response := f.call(t, "PUT", path+"/"+id, input)
	assertStatus(t, response, 200)
	var user application.UserView
	decodeResponse(t, response, &user)
	if user.Username != "renamed-user" || len(user.Roles) != 1 || user.Roles[0] != "Editor" {
		t.Fatalf("user changes lost: %#v", user)
	}
	assertProblem(t, f.raw(t, "PUT", path+"/"+id, "{"), 400, "invalid_json")
	assertProblem(t, f.call(t, "PUT", path+"/"+id, map[string]any{}), 400, "user_validation")
	assertProblem(t, f.call(t, "PUT", path+"/missing", input), 404, "user_not_found")
	assertProblem(t, f.call(t, "DELETE", path+"/"+f.adminID, nil), 409, "last_admin")
	assertStatus(t, f.call(t, "DELETE", path+"/"+id, nil), 204)
	assertProblem(t, f.call(t, "DELETE", path+"/"+id, nil), 404, "user_not_found")
	assertStatus(t, authedJSONRequest(t, f.router, f.viewer, "GET", path, nil, true), 403)
	for _, limit := range []string{"invalid", "0", "999", "1"} {
		assertStatus(t, f.call(t, "GET", "/api/v1/sessions?limit="+limit, nil), 200)
		response := f.call(t, "GET", "/api/v1/system/audit-log?limit="+limit, nil)
		assertStatus(t, response, 200)
		if strings.Contains(response.Body.String(), "synthetic-new-user-password") {
			t.Fatal("audit leaked credentials")
		}
	}
	assertProblem(t, f.call(t, "PUT", "/api/v1/sessions/missing/revoke", nil), 404, "session_not_found")
}

func TestTwoFactorHTTPContractsConsumeCodesAndHideSecrets(t *testing.T) {
	f := newContractFixture(t)
	path := "/api/v1/auth/two-factor"
	response := f.call(t, "GET", path, nil)
	assertStatus(t, response, 200)
	var status application.TwoFactorStatus
	decodeResponse(t, response, &status)
	if status.Enabled || status.Secret != "" {
		t.Fatalf("initial state: %#v", status)
	}
	response = f.call(t, "POST", path+"/setup", nil)
	assertStatus(t, response, 200)
	decodeResponse(t, response, &status)
	if status.Secret == "" || !strings.HasPrefix(status.OtpauthURL, "otpauth://totp/") {
		t.Fatal("setup did not provision a local authenticator")
	}
	secret := status.Secret
	assertProblem(t, f.raw(t, "POST", path+"/enable", "{"), 400, "invalid_json")
	assertProblem(t, f.raw(t, "POST", path+"/disable", "{"), 400, "invalid_json")
	assertProblem(t, f.call(t, "POST", path+"/enable", map[string]string{"code": "invalid"}), 401, "invalid_two_factor")
	now := time.Now().UTC()
	usedCode := contractTOTP(t, secret, now.Add(-30*time.Second))
	response = f.call(t, "POST", path+"/enable", map[string]string{"code": usedCode})
	assertStatus(t, response, 200)
	if strings.Contains(response.Body.String(), secret) {
		t.Fatal("status disclosed TOTP secret")
	}
	decodeResponse(t, response, &status)
	if !status.Enabled {
		t.Fatal("2FA was not enabled")
	}
	response = f.call(t, "POST", path+"/setup", nil)
	assertStatus(t, response, 200)
	if strings.Contains(response.Body.String(), secret) {
		t.Fatal("enabled setup disclosed TOTP secret")
	}
	assertProblem(t, f.call(t, "POST", path+"/disable", map[string]string{"currentPassword": "wrong", "code": usedCode}), 401, "invalid_two_factor")
	assertProblem(t, f.call(t, "POST", path+"/disable", map[string]string{"currentPassword": "synthetic-contract-password", "code": usedCode}), 401, "invalid_two_factor")
	assertStatus(t, f.call(t, "POST", path+"/disable", map[string]string{"currentPassword": "synthetic-contract-password", "code": contractTOTP(t, secret, time.Now().UTC())}), 204)
	response = f.call(t, "GET", path, nil)
	assertStatus(t, response, 200)
	status = application.TwoFactorStatus{}
	decodeResponse(t, response, &status)
	if status.Enabled || status.Secret != "" {
		t.Fatal("2FA remained enabled or secret was returned")
	}
}

func contractTOTP(t *testing.T, secret string, at time.Time) string {
	t.Helper()
	key, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(secret)
	if err != nil {
		t.Fatal(err)
	}
	var counter [8]byte
	binary.BigEndian.PutUint64(counter[:], uint64(at.Unix()/30))
	mac := hmac.New(sha1.New, key)
	if _, err := mac.Write(counter[:]); err != nil {
		t.Fatal(err)
	}
	sum := mac.Sum(nil)
	offset := sum[len(sum)-1] & 15
	return fmt.Sprintf("%06d", (binary.BigEndian.Uint32(sum[offset:offset+4])&0x7fffffff)%1000000)
}
