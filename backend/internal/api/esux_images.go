package api

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"image"
	"io"
	"path/filepath"
	"strings"
)

func previewDecoderImage(filename string, data []byte) *decoderPreviewImage {
	extension := strings.ToLower(filepath.Ext(filename))
	switch extension {
	case ".esux", ".esu", ".lokprogrammer", ".zip":
	default:
		return nil
	}
	if len(data) == 0 || len(data) > 25*1024*1024 {
		return nil
	}
	if extension == ".zip" || bytes.HasPrefix(data, []byte("PK\x03\x04")) {
		if preview := previewImageFromZip(data); preview != nil {
			return preview
		}
	}
	return previewEmbeddedImage(data)
}

func previewImageFromZip(data []byte) *decoderPreviewImage {
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil
	}
	for _, file := range reader.File {
		if file.FileInfo().IsDir() || !isPreviewImageName(file.Name) || file.UncompressedSize64 > maxDecoderPreviewImageBytes {
			continue
		}
		handle, err := file.Open()
		if err != nil {
			continue
		}
		imageData, readErr := io.ReadAll(io.LimitReader(handle, maxDecoderPreviewImageBytes+1))
		_ = handle.Close()
		if readErr != nil || len(imageData) > maxDecoderPreviewImageBytes {
			continue
		}
		if preview := validatePreviewImage(imageData); preview != nil {
			return preview
		}
	}
	return nil
}

func isPreviewImageName(name string) bool {
	switch strings.ToLower(filepath.Ext(name)) {
	case ".jpg", ".jpeg", ".png", ".webp":
		return true
	default:
		return false
	}
}

func previewEmbeddedImage(data []byte) *decoderPreviewImage {
	for _, imageData := range candidateEmbeddedImages(data) {
		if preview := validatePreviewImage(imageData); preview != nil {
			return preview
		}
	}
	return nil
}

func candidateEmbeddedImages(data []byte) [][]byte {
	candidates := [][]byte{}
	candidates = append(candidates, candidatePNGImages(data)...)
	candidates = append(candidates, candidateJPEGImages(data)...)
	candidates = append(candidates, candidateWebPImages(data)...)
	return candidates
}

func candidatePNGImages(data []byte) [][]byte {
	candidates := [][]byte{}
	for offset := 0; offset < len(data); {
		index := bytes.Index(data[offset:], pngSignature)
		if index < 0 {
			break
		}
		start := offset + index
		endMarker := bytes.Index(data[start:], []byte("IEND"))
		if endMarker >= 0 {
			end := start + endMarker + 8
			if end <= len(data) && end-start <= maxDecoderPreviewImageBytes {
				candidates = append(candidates, data[start:end])
			}
		}
		offset = start + len(pngSignature)
	}
	return candidates
}

func candidateJPEGImages(data []byte) [][]byte {
	candidates := [][]byte{}
	for offset := 0; offset < len(data); {
		index := bytes.Index(data[offset:], jpegSignature)
		if index < 0 {
			break
		}
		start := offset + index
		endMarker := bytes.Index(data[start+len(jpegSignature):], []byte{0xff, 0xd9})
		if endMarker >= 0 {
			end := start + len(jpegSignature) + endMarker + 2
			if end <= len(data) && end-start <= maxDecoderPreviewImageBytes {
				candidates = append(candidates, data[start:end])
			}
		}
		offset = start + len(jpegSignature)
	}
	return candidates
}

func candidateWebPImages(data []byte) [][]byte {
	candidates := [][]byte{}
	for offset := 0; offset+12 <= len(data); {
		index := bytes.Index(data[offset:], []byte("RIFF"))
		if index < 0 {
			break
		}
		start := offset + index
		if start+12 <= len(data) && string(data[start+8:start+12]) == "WEBP" {
			size := int(binary.LittleEndian.Uint32(data[start+4 : start+8]))
			end := start + 8 + size
			if end <= len(data) && end-start <= maxDecoderPreviewImageBytes {
				candidates = append(candidates, data[start:end])
			}
		}
		offset = start + 4
	}
	return candidates
}

func validatePreviewImage(data []byte) *decoderPreviewImage {
	if len(data) == 0 || len(data) > maxDecoderPreviewImageBytes {
		return nil
	}
	config, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil || config.Width <= 0 || config.Height <= 0 {
		return nil
	}
	if config.Width > maxDecoderPreviewImageDimension || config.Height > maxDecoderPreviewImageDimension {
		return nil
	}
	mimeType := previewImageMimeType(format)
	if mimeType == "" {
		return nil
	}
	return &decoderPreviewImage{
		MimeType: mimeType,
		Width:    config.Width,
		Height:   config.Height,
		DataURL:  "data:" + mimeType + ";base64," + base64.StdEncoding.EncodeToString(data),
	}
}

func previewImageMimeType(format string) string {
	switch format {
	case "jpeg":
		return "image/jpeg"
	case "png":
		return "image/png"
	case "webp":
		return "image/webp"
	default:
		return ""
	}
}
