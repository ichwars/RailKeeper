package startup

import "testing"

func TestLoginWorkConfiguration(t *testing.T) {
	for _, values := range []map[string]string{
		nil, {"RAILKEEPER_LOGIN_PARALLEL": "1", "RAILKEEPER_LOGIN_MEMORY_MB": "256"},
		{"RAILKEEPER_LOGIN_PARALLEL": "4", "RAILKEEPER_LOGIN_MEMORY_MB": "1024", "RAILKEEPER_LOGIN_QUEUE": "64", "RAILKEEPER_LOGIN_WAIT_MS": "2000"},
		{"RAILKEEPER_LOGIN_QUEUE": "0", "RAILKEEPER_LOGIN_WAIT_MS": "0"},
		{"RAILKEEPER_LOGIN_PARALLEL": "  "},
	} {
		if _, err := ResolveLoginWorkLimiter(func(key string) (string, bool) { value, ok := values[key]; return value, ok }); err != nil {
			t.Fatal("valid login configuration rejected", err)
		}
	}
	for _, values := range []map[string]string{
		{"RAILKEEPER_LOGIN_PARALLEL": "bad"}, {"RAILKEEPER_LOGIN_QUEUE": "bad"},
		{"RAILKEEPER_LOGIN_MEMORY_MB": "bad"}, {"RAILKEEPER_LOGIN_WAIT_MS": "bad"},
		{"RAILKEEPER_LOGIN_WAIT_MS": "-1"}, {"RAILKEEPER_LOGIN_WAIT_MS": "2001"},
		{"RAILKEEPER_LOGIN_PARALLEL": "0"}, {"RAILKEEPER_LOGIN_PARALLEL": "3"},
	} {
		if _, err := ResolveLoginWorkLimiter(func(key string) (string, bool) { value, ok := values[key]; return value, ok }); err == nil {
			t.Fatal("unsafe login configuration accepted")
		}
	}
}
