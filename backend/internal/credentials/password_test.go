package credentials

import (
	"encoding/base64"
	"strings"
	"testing"

	"golang.org/x/crypto/argon2"
)

func independentHash(password string) string {
	salt := []byte("independent-salt")
	key := argon2.IDKey([]byte(password), salt, 2, 32, 1, 24)
	return "$argon2id$v=19$m=32,t=2,p=1$" + base64.RawStdEncoding.EncodeToString(salt) + "$" +
		base64.RawStdEncoding.EncodeToString(key)
}

func TestVerificationUsesEncodedParametersAndKeyLength(t *testing.T) {
	hash := independentHash("a sufficiently long password")
	if !Verify("a sufficiently long password", hash) || Verify("wrong", hash) {
		t.Fatal("encoded, non-default parameters must determine verification")
	}
	if Verify("a sufficiently long password", strings.Replace(hash, "t=2", "t=1", 1)) {
		t.Fatal("modified work factor must change the derived key")
	}
}

func TestHashRetainsCurrentPolicyAndIndependentSalts(t *testing.T) {
	one, err := Hash("a sufficiently long password")
	if err != nil {
		t.Fatal(err)
	}
	two, err := Hash("a sufficiently long password")
	if err != nil || one == two || !Verify("a sufficiently long password", one) {
		t.Fatalf("hashing must retain random salt and valid verification: %v", err)
	}
	if !strings.HasPrefix(one, "$argon2id$v=19$m=65536,t=1,p=4$") {
		t.Fatal("existing hash policy unexpectedly changed")
	}
	if upgraded, err := Upgrade("a sufficiently long password", one); err != nil || upgraded != "" {
		t.Fatalf("current policy must not rewrite existing hashes: %v", err)
	}
}

func TestUpgradeNeverLowersExistingWorkFactor(t *testing.T) {
	next, err := Upgrade("a sufficiently long password", independentHash("a sufficiently long password"))
	if err != nil || !Verify("a sufficiently long password", next) {
		t.Fatalf("upgraded hash does not verify: %v", err)
	}
	value, err := parse(next)
	if err != nil || value.memory != 65536 || value.iterations != 2 || len(value.key) != 32 {
		t.Fatalf("upgrade lost stronger iterations or did not upgrade memory/key: %+v %v", value.parameters, err)
	}
}

func TestMalformedOrUnboundedHashRejectedBeforeDerivation(t *testing.T) {
	valid := independentHash("password")
	for _, text := range []string{
		"", strings.Repeat("x", 513), strings.Replace(valid, "argon2id", "argon2i", 1),
		strings.Replace(valid, "v=19", "v=16", 1), strings.Replace(valid, "m=32", "m=0", 1),
		strings.Replace(valid, "m=32", "m=4294967296", 1), strings.Replace(valid, "m=32", "m=300000", 1),
		strings.Replace(valid, "t=2", "t=0", 1), strings.Replace(valid, "t=2", "t=11", 1),
		strings.Replace(valid, "p=1", "p=0", 1), strings.Replace(valid, "p=1", "p=17", 1),
		strings.Replace(valid, "p=1", "p=256", 1), strings.Replace(valid, "p=1", "p=4294967295", 1),
		strings.Replace(valid, "p=1", "m=32", 1), strings.Replace(valid, "m=32", "unknown=32", 1),
		valid + "$extra", strings.Replace(valid, "m=32", "m=-32", 1),
	} {
		if _, err := parse(text); err == nil || Verify("password", text) {
			t.Errorf("malformed or excessive hash accepted: %q", text)
		}
		if _, err := Upgrade("password", text); err == nil {
			t.Error("cannot upgrade an invalid hash")
		}
	}
}

func FuzzPasswordHashParser(f *testing.F) {
	f.Add(independentHash("password"))
	f.Add("$argon2id$v=19$m=4294967295,t=1,p=4$bad$bad")
	f.Fuzz(func(t *testing.T, encoded string) {
		value, err := parse(encoded)
		if err == nil && (value.memory > 256*1024 || value.iterations > 10 || value.parallel > 16 ||
			len(value.salt) < 8 || len(value.salt) > 64 || len(value.key) < 16 || len(value.key) > 64) {
			t.Fatal("parser accepted unbounded derivation parameters")
		}
	})
}
