package credentials

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"crypto/subtle"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"strings"
	"time"
)

func GenerateTOTPSecret() (string, error) {
	value := make([]byte, 20)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(value), nil
}

// MatchTOTP returns the actual matching 30-second step. Persistence must consume that step atomically.
func MatchTOTP(secret, code string, now time.Time) (int64, bool) {
	normalized := strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(code), " ", ""), "-", "")
	if len(normalized) != 6 || secret == "" || len(secret) > 256 || now.Unix() < 0 {
		return 0, false
	}
	for offset := int64(-1); offset <= 1; offset++ {
		step := now.Unix()/30 + offset
		if step < 0 {
			continue
		}
		expected, err := codeAt(secret, uint64(step))
		if err != nil {
			return 0, false
		}
		if subtle.ConstantTimeCompare([]byte(normalized), []byte(expected)) == 1 {
			return step, true
		}
	}
	return 0, false
}

func codeAt(secret string, step uint64) (string, error) {
	key, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(strings.ToUpper(strings.TrimSpace(secret)))
	if err != nil || len(key) == 0 {
		return "", fmt.Errorf("invalid TOTP secret")
	}
	var counter [8]byte
	binary.BigEndian.PutUint64(counter[:], step)
	mac := hmac.New(sha1.New, key)
	if _, err := mac.Write(counter[:]); err != nil {
		return "", err
	}
	sum := mac.Sum(nil)
	offset := sum[len(sum)-1] & 0x0f
	value := binary.BigEndian.Uint32(sum[offset:offset+4]) & 0x7fffffff
	return fmt.Sprintf("%06d", value%1_000_000), nil
}
