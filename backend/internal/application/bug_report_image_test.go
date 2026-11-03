package application_test

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"hash/crc32"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"strings"
	"testing"

	"railkeeper/backend/internal/application"
)

func TestBugReportImageReencodesPNGAndPreservesTransparency(t *testing.T) {
	source := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	source.SetNRGBA(0, 0, color.NRGBA{R: 255, A: 80})
	var encoded bytes.Buffer
	if err := png.Encode(&encoded, source); err != nil {
		t.Fatal(err)
	}
	original := append(encoded.Bytes(), []byte("EXIF-SECRET")...)

	result, err := application.SanitizeBugReportScreenshot(bytes.NewReader(original), "image/png")
	if err != nil {
		t.Fatal(err)
	}
	if result.MediaType != "image/png" || result.Extension != "png" {
		t.Fatalf("unexpected normalized format: %#v", result)
	}
	if bytes.Contains(result.Bytes, []byte("EXIF-SECRET")) {
		t.Fatal("re-encoded image retained appended metadata")
	}
	decoded, err := png.Decode(bytes.NewReader(result.Bytes))
	if err != nil {
		t.Fatal(err)
	}
	_, _, _, alpha := decoded.At(0, 0).RGBA()
	if alpha == 0xffff {
		t.Fatal("PNG transparency was lost")
	}
}

func TestBugReportImageReencodesJPEG(t *testing.T) {
	source := image.NewRGBA(image.Rect(0, 0, 20, 10))
	source.Set(5, 5, color.RGBA{R: 180, G: 20, B: 30, A: 255})
	var encoded bytes.Buffer
	if err := jpeg.Encode(&encoded, source, &jpeg.Options{Quality: 95}); err != nil {
		t.Fatal(err)
	}
	result, err := application.SanitizeBugReportScreenshot(bytes.NewReader(encoded.Bytes()), "image/jpeg")
	if err != nil {
		t.Fatal(err)
	}
	if result.MediaType != "image/jpeg" || result.Extension != "jpg" {
		t.Fatalf("unexpected normalized format: %#v", result)
	}
	if _, err := jpeg.Decode(bytes.NewReader(result.Bytes)); err != nil {
		t.Fatalf("normalized JPEG is invalid: %v", err)
	}
}

func TestBugReportImageAcceptsStaticWebP(t *testing.T) {
	encoded, err := base64.StdEncoding.DecodeString(
		"UklGRiIAAABXRUJQVlA4IBYAAAAwAQCdASoBAAEADsD+JaQAA3AAAAAA",
	)
	if err != nil {
		t.Fatal(err)
	}
	result, err := application.SanitizeBugReportScreenshot(bytes.NewReader(encoded), "image/webp")
	if err != nil {
		t.Fatal(err)
	}
	if result.MediaType != "image/jpeg" || result.Extension != "jpg" {
		t.Fatalf("opaque WebP normalized as %#v", result)
	}
}

func TestBugReportImageRejectsInvalidInputBeforeDecode(t *testing.T) {
	pngBytes := encodedPNG(t, image.NewRGBA(image.Rect(0, 0, 2, 2)))
	var animatedPNG bytes.Buffer
	animatedPNG.Write(pngBytes[:len(pngBytes)-12])
	writePNGChunk(&animatedPNG, "acTL", []byte{0, 0, 0, 1, 0, 0, 0, 0})
	animatedPNG.Write(pngBytes[len(pngBytes)-12:])
	tests := []struct {
		name      string
		data      []byte
		mediaType string
	}{
		{name: "mismatched type", data: pngBytes, mediaType: "image/jpeg"},
		{name: "unsupported type", data: pngBytes, mediaType: "image/gif"},
		{name: "malformed", data: []byte("not an image"), mediaType: "image/png"},
		{name: "animated webp", data: []byte("RIFF\x10\x00\x00\x00WEBPANIM\x00\x00\x00\x00"), mediaType: "image/webp"},
		{name: "animated png", data: animatedPNG.Bytes(), mediaType: "image/png"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := application.SanitizeBugReportScreenshot(bytes.NewReader(test.data), test.mediaType); err == nil {
				t.Fatal("expected invalid screenshot error")
			}
		})
	}
}

func TestBugReportImageRejectsInputOverFiveMiB(t *testing.T) {
	data := bytes.Repeat([]byte{'x'}, application.MaxBugReportScreenshotBytes+1)
	if _, err := application.SanitizeBugReportScreenshot(bytes.NewReader(data), "image/png"); err == nil {
		t.Fatal("expected size error")
	}
}

func TestBugReportImageRejectsMoreThanSixteenMegapixelsFromHeader(t *testing.T) {
	data := minimalPNGHeader(4_097, 4_097)
	if _, err := application.SanitizeBugReportScreenshot(bytes.NewReader(data), "image/png"); err == nil {
		t.Fatal("expected dimension error")
	}
}

func TestBugReportImageDownscalesDimensionAbove4096(t *testing.T) {
	data := encodedPNG(t, image.NewRGBA(image.Rect(0, 0, 4_097, 1)))
	result, err := application.SanitizeBugReportScreenshot(bytes.NewReader(data), "image/png")
	if err != nil {
		t.Fatal(err)
	}
	config, err := png.DecodeConfig(bytes.NewReader(result.Bytes))
	if err != nil {
		t.Fatal(err)
	}
	if config.Width > application.MaxBugReportScreenshotDimension ||
		config.Height > application.MaxBugReportScreenshotDimension {
		t.Fatalf("normalized dimensions = %dx%d", config.Width, config.Height)
	}
}

func encodedPNG(t *testing.T, source image.Image) []byte {
	t.Helper()
	var encoded bytes.Buffer
	if err := png.Encode(&encoded, source); err != nil {
		t.Fatal(err)
	}
	return encoded.Bytes()
}

func minimalPNGHeader(width uint32, height uint32) []byte {
	var output bytes.Buffer
	output.Write([]byte("\x89PNG\r\n\x1a\n"))
	ihdr := make([]byte, 13)
	binary.BigEndian.PutUint32(ihdr[0:4], width)
	binary.BigEndian.PutUint32(ihdr[4:8], height)
	ihdr[8] = 8
	ihdr[9] = 6
	writePNGChunk(&output, "IHDR", ihdr)
	writePNGChunk(&output, "IEND", nil)
	return output.Bytes()
}

func writePNGChunk(output *bytes.Buffer, name string, data []byte) {
	_ = binary.Write(output, binary.BigEndian, uint32(len(data)))
	output.WriteString(name)
	output.Write(data)
	checksum := crc32.NewIEEE()
	_, _ = checksum.Write([]byte(name))
	_, _ = checksum.Write(data)
	_ = binary.Write(output, binary.BigEndian, checksum.Sum32())
}

func TestBugReportImageErrorDoesNotExposeSubmittedBytes(t *testing.T) {
	secret := "screenshot-secret-content"
	_, err := application.SanitizeBugReportScreenshot(strings.NewReader(secret), "image/png")
	if err == nil || strings.Contains(err.Error(), secret) {
		t.Fatalf("unsafe error: %v", err)
	}
}
