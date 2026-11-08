package media

import (
	"bytes"
	"compress/zlib"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"railkeeper/backend/internal/credentials"
)

type Blob struct {
	ID             string
	OriginalSize   int64
	CompressedSize int64
	Compression    string
	SHA256         string
	Data           []byte
	CreatedAt      string
}

func NewBlob(data []byte) (Blob, error) {
	if len(data) == 0 {
		return Blob{}, errors.New("file blob is empty")
	}
	compressed, err := Encode(data)
	if err != nil {
		return Blob{}, err
	}
	sum := sha256.Sum256(data)
	blob := Blob{
		ID:             credentials.RandomID(),
		OriginalSize:   int64(len(data)),
		CompressedSize: int64(len(compressed)),
		Compression:    "zlib",
		SHA256:         hex.EncodeToString(sum[:]),
		Data:           compressed,
		CreatedAt:      time.Now().UTC().Format(time.RFC3339),
	}
	return blob, nil
}
func Encode(data []byte) ([]byte, error) {
	var buffer bytes.Buffer
	writer := zlib.NewWriter(&buffer)
	if _, err := writer.Write(data); err != nil {
		_ = writer.Close()
		return nil, err
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func Decode(data []byte, compression string) ([]byte, error) {
	switch strings.ToLower(strings.TrimSpace(compression)) {
	case "", "none":
		return data, nil
	case "zlib":
		reader, err := zlib.NewReader(bytes.NewReader(data))
		if err != nil {
			return nil, err
		}
		defer func() { _ = reader.Close() }()
		return io.ReadAll(reader)
	default:
		return nil, fmt.Errorf("unsupported blob compression %q", compression)
	}
}
