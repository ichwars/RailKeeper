package startup

import "testing"

func TestCookieSecurityUsesOnlyExplicitPolicyAndConfiguredURL(t *testing.T) {
	for _, test := range []struct {
		mode, url       string
		secure, invalid bool
	}{
		{"auto", "", false, false},
		{"", "http://127.0.0.1:8080", false, false},
		{"auto", "http://192.0.2.1:8080", false, false},
		{"auto", "https://railkeeper.example.test/base", true, false},
		{"true", "http://127.0.0.1:8080", true, false},
		{"false", "https://railkeeper.example.test", false, false},
		{"auto", "//railkeeper.example.test", false, true},
		{"auto", "https://user:password@railkeeper.example.test", false, true},
		{"auto", "https:///missing", false, true},
		{"auto", "javascript:alert(1)", false, true},
		{"auto", "https://railkeeper.example.test/#fragment", false, true},
		{"sometimes", "", false, true},
	} {
		t.Run(test.mode+"/"+test.url, func(t *testing.T) {
			secure, err := ResolveCookieSecurity(test.mode, test.url)
			if (err != nil) != test.invalid || secure != test.secure {
				t.Fatalf("secure=%v err=%v", secure, err)
			}
		})
	}
}
