// Package credentials implements bounded, self-describing password hashing without storage dependencies.
package credentials

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/crypto/argon2"
)

type parameters struct {
	memory     uint32
	iterations uint32
	parallel   uint8
}

type decoded struct {
	parameters
	salt []byte
	key  []byte
}

const MaxPasswordMemoryMiB = 256

var policy = parameters{memory: 64 * 1024, iterations: 1, parallel: 4}

var dummy = derive("RailKeeper-unknown-user", []byte("RailKeeper-dummy!"), policy, 32)

// DummyHash has the current verification cost, but is never stored as an account credential.
func DummyHash() string { return dummy }

func Hash(password string) (string, error) { return generate(password, policy, 32) }

func Verify(password, encoded string) bool {
	value, err := parse(encoded)
	if err != nil {
		return false
	}
	actual := argon2.IDKey([]byte(password), value.salt, value.iterations, value.memory,
		value.parallel, uint32(len(value.key)))
	return subtle.ConstantTimeCompare(actual, value.key) == 1
}

// Upgrade returns an empty string when no upgrade is needed. Existing stronger parameters are retained.
// Call it only after the complete authentication, including any second factor, has succeeded.
func Upgrade(password, encoded string) (string, error) {
	value, err := parse(encoded)
	if err != nil {
		return "", err
	}
	if value.memory >= policy.memory && value.iterations >= policy.iterations && len(value.key) >= 32 {
		return "", nil
	}
	next := parameters{memory: max(value.memory, policy.memory), iterations: max(value.iterations, policy.iterations),
		parallel: max(value.parallel, policy.parallel)}
	return generate(password, next, uint32(max(len(value.key), 32)))
}

func generate(password string, p parameters, length uint32) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generate password salt: %w", err)
	}
	return derive(password, salt, p, length), nil
}

func derive(password string, salt []byte, p parameters, length uint32) string {
	key := argon2.IDKey([]byte(password), salt, p.iterations, p.memory, p.parallel, length)
	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s", argon2.Version, p.memory, p.iterations,
		p.parallel, base64.RawStdEncoding.EncodeToString(salt), base64.RawStdEncoding.EncodeToString(key))
}

func parse(encoded string) (decoded, error) {
	invalid := errors.New("invalid or unsupported password hash")
	if len(encoded) > 512 {
		return decoded{}, invalid
	}
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[0] != "" || parts[1] != "argon2id" || parts[2] != "v=19" {
		return decoded{}, invalid
	}
	options := strings.Split(parts[3], ",")
	if len(options) != 3 {
		return decoded{}, invalid
	}
	values := make(map[string]uint64, 3)
	for _, option := range options {
		key, text, found := strings.Cut(option, "=")
		if !found || (key != "m" && key != "t" && key != "p") {
			return decoded{}, invalid
		}
		if _, duplicate := values[key]; duplicate {
			return decoded{}, invalid
		}
		value, err := strconv.ParseUint(text, 10, 32)
		if err != nil {
			return decoded{}, invalid
		}
		values[key] = value
	}
	memory, iterations, parallel := values["m"], values["t"], values["p"]
	if parallel == 0 || parallel > 16 || iterations == 0 || iterations > 10 ||
		memory < 8*parallel || memory > MaxPasswordMemoryMiB*1024 {
		return decoded{}, invalid
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil || len(salt) < 8 || len(salt) > 64 {
		return decoded{}, invalid
	}
	key, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil || len(key) < 16 || len(key) > 64 {
		return decoded{}, invalid
	}
	return decoded{parameters: parameters{uint32(memory), uint32(iterations), uint8(parallel)},
		salt: salt, key: key}, nil
}
