package startup

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

const SetupCodeFile = "setup-code.txt"

// ReadOrCreateSetupCode is local-only. Call it only for an installation without users.
func ReadOrCreateSetupCode(dataDir string) (string, error) {
	root, err := os.OpenRoot(dataDir)
	if err != nil {
		return "", fmt.Errorf("open setup directory: %w", err)
	}
	defer func() { _ = root.Close() }()
	file, err := root.OpenFile(SetupCodeFile, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if errors.Is(err, os.ErrExist) {
		return readSetupCode(root)
	}
	if err != nil {
		return "", fmt.Errorf("create private setup code file: %w", err)
	}
	defer func() { _ = file.Close() }()
	var random [32]byte
	if _, err := rand.Read(random[:]); err != nil {
		return "", fmt.Errorf("generate setup code: %w", err)
	}
	code := base64.RawURLEncoding.EncodeToString(random[:])
	if _, err := io.WriteString(file, code+"\n"); err != nil {
		return "", fmt.Errorf("write setup code file: %w", err)
	}
	if err := file.Sync(); err != nil {
		return "", fmt.Errorf("flush setup code file: %w", err)
	}
	return code, nil
}

func readSetupCode(root *os.Root) (string, error) {
	info, err := root.Lstat(SetupCodeFile)
	if err != nil || !info.Mode().IsRegular() || info.Size() > 256 {
		return "", errors.New("setup code file must be a small regular file")
	}
	file, err := root.Open(SetupCodeFile)
	if err != nil {
		return "", fmt.Errorf("open private setup code file: %w", err)
	}
	defer func() { _ = file.Close() }()
	if err := file.Chmod(0600); err != nil {
		return "", fmt.Errorf("protect setup code file: %w", err)
	}
	content, err := io.ReadAll(io.LimitReader(file, 257))
	if err != nil {
		return "", fmt.Errorf("read setup code file: %w", err)
	}
	code := strings.TrimSpace(string(content))
	decoded, err := base64.RawURLEncoding.Strict().DecodeString(code)
	if err != nil || len(decoded) != 32 {
		return "", errors.New("setup code file contains an invalid code")
	}
	return code, nil
}
