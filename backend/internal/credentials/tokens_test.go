package credentials

import "testing"

func TestTokenDigestPreservesStoredSHA256Encoding(t *testing.T) {
	// Existing session and rate-limit rows use unpadded URL-safe SHA-256 digests.
	const emptyDigest = "47DEQpj8HBSa-_TImW-5JCeuQeRkm5NMpJWZG3hSuFU"
	if HashToken("") != emptyDigest {
		t.Fatal("stored token digest encoding changed")
	}
}
