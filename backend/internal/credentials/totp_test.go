package credentials

import (
	"encoding/base32"
	"strings"
	"testing"
	"time"
)

func TestTOTPRFC6238SHA1VectorsWithSixDigitTruncation(t *testing.T) {
	secret := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString([]byte("12345678901234567890"))
	// RFC 6238 Appendix B gives eight digits; RailKeeper retains the last six for its existing accounts.
	for _, value := range []struct {
		seconds int64
		code    string
	}{{59, "287082"}, {1111111109, "081804"}, {1111111111, "050471"},
		{1234567890, "005924"}, {2000000000, "279037"}, {20000000000, "353130"}} {
		at := time.Unix(value.seconds, 0)
		for _, skew := range []time.Duration{-30 * time.Second, 0, 30 * time.Second} {
			step, ok := MatchTOTP(secret, value.code, at.Add(skew))
			if !ok || step != value.seconds/30 {
				t.Fatalf("vector %d/skew %v: step=%d accepted=%v", value.seconds, skew, step, ok)
			}
		}
		if _, ok := MatchTOTP(secret, value.code, at.Add(60*time.Second)); ok {
			t.Fatalf("expired code accepted at %d", value.seconds)
		}
	}
}

func TestTOTPRejectsInvalidInputsAndRetainsFormattedCodes(t *testing.T) {
	secret := "GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ"
	at := time.Unix(59, 0)
	if _, ok := MatchTOTP(strings.ToLower(secret), "287-082", at); !ok {
		t.Fatal("existing normalized secret/code forms must remain supported")
	}
	for _, input := range []struct{ secret, code string }{
		{"", "287082"}, {"invalid!", "287082"}, {strings.Repeat("A", 257), "287082"},
		{secret, "123"}, {secret, "ABCDEF"}, {secret, "2870820"},
	} {
		if _, ok := MatchTOTP(input.secret, input.code, at); ok {
			t.Fatal("invalid secret/code accepted")
		}
	}
	if _, ok := MatchTOTP(secret, "287082", time.Unix(-1, 0)); ok {
		t.Fatal("negative Unix counter accepted")
	}
}
