package application

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"math"
	"strings"

	"golang.org/x/image/draw"
	_ "golang.org/x/image/webp"
)

const (
	MaxBugReportScreenshotBytes     = 5 * 1024 * 1024
	MaxBugReportScreenshotPixels    = 16_000_000
	MaxBugReportScreenshotDimension = 4_096
)

var (
	ErrBugReportScreenshotInvalid  = errors.New("bug report screenshot is invalid")
	ErrBugReportScreenshotTooLarge = errors.New("bug report screenshot is too large")
)

type BugReportScreenshot struct {
	Bytes     []byte
	Extension string
	MediaType string
}

func SanitizeBugReportScreenshot(source io.Reader, declaredMediaType string) (BugReportScreenshot, error) {
	input, err := io.ReadAll(io.LimitReader(source, MaxBugReportScreenshotBytes+1))
	if err != nil {
		return BugReportScreenshot{}, fmt.Errorf("%w: read failed", ErrBugReportScreenshotInvalid)
	}
	if len(input) > MaxBugReportScreenshotBytes {
		return BugReportScreenshot{}, ErrBugReportScreenshotTooLarge
	}
	if len(input) == 0 || !validBugReportDeclaredType(declaredMediaType) {
		return BugReportScreenshot{}, ErrBugReportScreenshotInvalid
	}
	if declaredMediaType == "image/webp" && animatedWebP(input) {
		return BugReportScreenshot{}, fmt.Errorf("%w: animated WebP is not supported", ErrBugReportScreenshotInvalid)
	}
	if declaredMediaType == "image/png" && animatedPNG(input) {
		return BugReportScreenshot{}, fmt.Errorf("%w: animated PNG is not supported", ErrBugReportScreenshotInvalid)
	}

	config, format, err := image.DecodeConfig(bytes.NewReader(input))
	if err != nil || mediaTypeForImageFormat(format) != declaredMediaType {
		return BugReportScreenshot{}, fmt.Errorf("%w: format mismatch", ErrBugReportScreenshotInvalid)
	}
	if config.Width <= 0 || config.Height <= 0 ||
		int64(config.Width)*int64(config.Height) > MaxBugReportScreenshotPixels {
		return BugReportScreenshot{}, fmt.Errorf("%w: decoded dimensions", ErrBugReportScreenshotTooLarge)
	}
	decoded, decodedFormat, err := image.Decode(bytes.NewReader(input))
	if err != nil || decodedFormat != format {
		return BugReportScreenshot{}, fmt.Errorf("%w: decode failed", ErrBugReportScreenshotInvalid)
	}

	normalized := renderBugReportImage(decoded, normalizedBugReportBounds(decoded.Bounds()))
	preservePNG := format == "png" || !bugReportImageIsOpaque(normalized)
	for attempt := 0; attempt < 8; attempt++ {
		encoded, encodeErr := encodeBugReportImage(normalized, preservePNG)
		if encodeErr != nil {
			return BugReportScreenshot{}, fmt.Errorf("%w: encode failed", ErrBugReportScreenshotInvalid)
		}
		if len(encoded) <= MaxBugReportScreenshotBytes {
			if preservePNG {
				return BugReportScreenshot{Bytes: encoded, Extension: "png", MediaType: "image/png"}, nil
			}
			return BugReportScreenshot{Bytes: encoded, Extension: "jpg", MediaType: "image/jpeg"}, nil
		}
		ratio := math.Sqrt(float64(MaxBugReportScreenshotBytes)/float64(len(encoded))) * 0.9
		if ratio >= 1 {
			ratio = 0.9
		}
		bounds := normalized.Bounds()
		width := max(1, int(float64(bounds.Dx())*ratio))
		height := max(1, int(float64(bounds.Dy())*ratio))
		if width == bounds.Dx() && height == bounds.Dy() {
			break
		}
		normalized = renderBugReportImage(normalized, image.Rect(0, 0, width, height))
	}
	return BugReportScreenshot{}, ErrBugReportScreenshotTooLarge
}

func validBugReportDeclaredType(mediaType string) bool {
	return mediaType == "image/png" || mediaType == "image/jpeg" || mediaType == "image/webp"
}

func mediaTypeForImageFormat(format string) string {
	switch format {
	case "png":
		return "image/png"
	case "jpeg":
		return "image/jpeg"
	case "webp":
		return "image/webp"
	default:
		return ""
	}
}

func normalizedBugReportBounds(bounds image.Rectangle) image.Rectangle {
	width := bounds.Dx()
	height := bounds.Dy()
	if width <= MaxBugReportScreenshotDimension && height <= MaxBugReportScreenshotDimension {
		return image.Rect(0, 0, width, height)
	}
	ratio := math.Min(
		float64(MaxBugReportScreenshotDimension)/float64(width),
		float64(MaxBugReportScreenshotDimension)/float64(height),
	)
	return image.Rect(0, 0, max(1, int(float64(width)*ratio)), max(1, int(float64(height)*ratio)))
}

func renderBugReportImage(source image.Image, bounds image.Rectangle) *image.NRGBA {
	destination := image.NewNRGBA(bounds)
	draw.CatmullRom.Scale(destination, destination.Bounds(), source, source.Bounds(), draw.Src, nil)
	return destination
}

func bugReportImageIsOpaque(source image.Image) bool {
	if opaque, ok := source.(interface{ Opaque() bool }); ok {
		return opaque.Opaque()
	}
	bounds := source.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			_, _, _, alpha := source.At(x, y).RGBA()
			if alpha != 0xffff {
				return false
			}
		}
	}
	return true
}

func encodeBugReportImage(source image.Image, preservePNG bool) ([]byte, error) {
	var output bytes.Buffer
	if preservePNG {
		if err := png.Encode(&output, source); err != nil {
			return nil, err
		}
	} else if err := jpeg.Encode(&output, source, &jpeg.Options{Quality: 88}); err != nil {
		return nil, err
	}
	return output.Bytes(), nil
}

func animatedWebP(input []byte) bool {
	if len(input) < 12 || string(input[:4]) != "RIFF" || string(input[8:12]) != "WEBP" {
		return false
	}
	return strings.Contains(string(input[12:]), "ANIM") || strings.Contains(string(input[12:]), "ANMF")
}

func animatedPNG(input []byte) bool {
	if len(input) < 8 || !bytes.Equal(input[:8], []byte("\x89PNG\r\n\x1a\n")) {
		return false
	}
	for offset := 8; offset+12 <= len(input); {
		chunkLength := int(uint32(input[offset])<<24 | uint32(input[offset+1])<<16 |
			uint32(input[offset+2])<<8 | uint32(input[offset+3]))
		if chunkLength < 0 || chunkLength > len(input)-offset-12 {
			return false
		}
		if string(input[offset+4:offset+8]) == "acTL" {
			return true
		}
		offset += 12 + chunkLength
	}
	return false
}
