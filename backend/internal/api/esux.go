package api

import (
	"encoding/binary"
	"encoding/xml"
	"errors"
	"fmt"
	"path/filepath"
	"railkeeper/backend/internal/decoderlibrary"
	"regexp"
	"strings"
)

type esuxMetadata struct {
	Name           string `xml:"name"`
	Address        string `xml:"address"`
	Type           string `xml:"type"`
	Decoder        string `xml:"decoder"`
	Manufacturer   string `xml:"manufacturer"`
	ManufacturerID string `xml:"manid"`
	LokProgrammer  string `xml:"lokprogrammer"`
}

type vehicleCVFilePreviewResponse struct {
	DecoderLibraryMatches   map[string]decoderlibrary.Match `json:"decoderLibraryMatches,omitempty"`
	FileName                string                          `json:"fileName"`
	SizeBytes               int64                           `json:"sizeBytes"`
	MimeType                string                          `json:"mimeType"`
	HasMetadata             bool                            `json:"hasMetadata"`
	ProjectName             string                          `json:"projectName,omitempty"`
	Address                 string                          `json:"address,omitempty"`
	Type                    string                          `json:"type,omitempty"`
	Decoder                 string                          `json:"decoder,omitempty"`
	Manufacturer            string                          `json:"manufacturer,omitempty"`
	ManufacturerID          string                          `json:"manufacturerId,omitempty"`
	LokProgrammer           string                          `json:"lokProgrammer,omitempty"`
	SuggestedDecoderProfile string                          `json:"suggestedDecoderProfile,omitempty"`
	SuggestedDescription    string                          `json:"suggestedDescription,omitempty"`
	CVImportError           string                          `json:"cvImportError,omitempty"`
	SuggestedCVValues       []cvValuePreview                `json:"suggestedCvValues,omitempty"`
	SuggestedFunctions      []functionMappingPreview        `json:"suggestedFunctions,omitempty"`
	SuggestedPreviewImage   *decoderPreviewImage            `json:"suggestedPreviewImage,omitempty"`
}

type cvValuePreview struct {
	SourceLine       *int   `json:"sourceLine,omitempty"`
	SourceExportedAt string `json:"sourceExportedAt,omitempty"`
	DecoderProfile   string `json:"decoderProfile,omitempty"`
	Protocol         string `json:"protocol,omitempty"`
	IndexCV31        *int   `json:"indexCV31,omitempty"`
	IndexCV32        *int   `json:"indexCV32,omitempty"`
	CVNumber         int    `json:"cvNumber"`
	Value            int    `json:"value"`
	Description      string `json:"description,omitempty"`
	Category         string `json:"category,omitempty"`
}

type functionMappingPreview struct {
	FunctionKey  string `json:"functionKey"`
	Name         string `json:"name"`
	FunctionType string `json:"functionType,omitempty"`
}

type decoderPreviewImage struct {
	MimeType string `json:"mimeType"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
	DataURL  string `json:"dataUrl"`
}

var errESUXMetadataUnavailable = errors.New("esux metadata unavailable")

func parseESUXMetadata(filename string, data []byte) (*esuxMetadata, error) {
	extension := strings.ToLower(filepath.Ext(filename))
	if extension != ".esux" && extension != ".esu" && extension != ".lokprogrammer" {
		return nil, errESUXMetadataUnavailable
	}
	if len(data) < 18 || string(data[:4]) != "ESU " {
		return nil, errESUXMetadataUnavailable
	}
	metadataLength := int(binary.LittleEndian.Uint32(data[14:18]))
	if metadataLength <= 0 || metadataLength > len(data)-18 || metadataLength > 64*1024 {
		return nil, errESUXMetadataUnavailable
	}
	var metadata esuxMetadata
	if err := xml.Unmarshal(data[18:18+metadataLength], &metadata); err != nil {
		return nil, err
	}
	metadata.Name = strings.TrimSpace(metadata.Name)
	metadata.Address = strings.TrimSpace(metadata.Address)
	metadata.Type = strings.TrimSpace(metadata.Type)
	metadata.Decoder = strings.TrimSpace(metadata.Decoder)
	metadata.Manufacturer = strings.TrimSpace(metadata.Manufacturer)
	metadata.ManufacturerID = strings.TrimSpace(metadata.ManufacturerID)
	metadata.LokProgrammer = strings.TrimSpace(metadata.LokProgrammer)
	if metadata.Name == "" && metadata.Address == "" && metadata.Decoder == "" && metadata.LokProgrammer == "" {
		return nil, errESUXMetadataUnavailable
	}
	return &metadata, nil
}

func applyESUXMetadata(filename string, data []byte, decoderProfile string, description string) (string, string) {
	metadata, err := parseESUXMetadata(filename, data)
	if err != nil {
		if strings.TrimSpace(decoderProfile) == "" {
			if text, ok := decoderPreviewText(filename, data); ok {
				if _, profile, recognized, parseErr := parseLokProgrammerCVList(text); recognized && parseErr == nil {
					decoderProfile = profile
				}
			}
		}
		return decoderProfile, description
	}
	if strings.TrimSpace(decoderProfile) == "" {
		decoderProfile = firstNonEmpty(metadata.Decoder, metadata.Name, "ESU LokProgrammer")
	}
	if strings.TrimSpace(description) == "" {
		description = metadata.Description()
	}
	return decoderProfile, description
}

func esuxPreviewResponse(filename string, sizeBytes int64, mimeType string, data []byte) vehicleCVFilePreviewResponse {
	response := vehicleCVFilePreviewResponse{
		FileName:  filename,
		SizeBytes: sizeBytes,
		MimeType:  mimeType,
	}
	response.SuggestedCVValues = previewCVValues(filename, data)
	if text, ok := decoderPreviewText(filename, data); ok {
		if rows, profile, recognized, err := parseLokProgrammerCVList(text); recognized {
			response.SuggestedCVValues = rows
			response.SuggestedDecoderProfile = profile
			if err != nil {
				response.CVImportError = err.Error()
			} else if profile != "" {
				response.HasMetadata = true
				response.Decoder = profile
			}
		}
	}
	if strings.EqualFold(filepath.Ext(filename), ".json") && len(data) <= 2*1024*1024 {
		rows, err := previewCVJSON(data)
		response.SuggestedCVValues = rows
		if err != nil {
			response.CVImportError = err.Error()
		}
	}
	response.SuggestedFunctions = previewFunctionMappings(filename, data)
	response.SuggestedPreviewImage = previewDecoderImage(filename, data)
	metadata, err := parseESUXMetadata(filename, data)
	if err != nil {
		return response
	}
	response.HasMetadata = true
	response.ProjectName = metadata.Name
	response.Address = metadata.Address
	response.Type = metadata.Type
	response.Decoder = metadata.Decoder
	response.Manufacturer = metadata.Manufacturer
	response.ManufacturerID = metadata.ManufacturerID
	response.LokProgrammer = metadata.LokProgrammer
	response.SuggestedDecoderProfile = firstNonEmpty(metadata.Decoder, metadata.Name, "ESU LokProgrammer")
	response.SuggestedDescription = metadata.Description()
	return response
}

func (m esuxMetadata) Description() string {
	parts := []string{"ESU LokProgrammer-Projekt"}
	if m.Name != "" {
		parts = append(parts, fmt.Sprintf("Name: %s", m.Name))
	}
	if m.Decoder != "" && m.Decoder != m.Name {
		parts = append(parts, fmt.Sprintf("Decoder: %s", m.Decoder))
	}
	if m.Address != "" {
		parts = append(parts, fmt.Sprintf("Adresse: %s", m.Address))
	}
	if m.Type != "" {
		parts = append(parts, fmt.Sprintf("Typ: %s", m.Type))
	}
	if m.Manufacturer != "" {
		parts = append(parts, fmt.Sprintf("Hersteller: %s", m.Manufacturer))
	}
	if m.LokProgrammer != "" {
		parts = append(parts, fmt.Sprintf("LokProgrammer: %s", m.LokProgrammer))
	}
	return strings.Join(parts, " | ")
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

const maxDecoderPreviewRows = 32

const maxDecoderPreviewImageBytes = 512 * 1024

const maxDecoderPreviewImageDimension = 4096

var (
	cvTextPattern       = regexp.MustCompile(`(?i)\bcv\s*([0-9]{1,4})\D+([0-9]{1,4})`)
	functionTextPattern = regexp.MustCompile(`(?i)\bF\s*([0-9]{1,2})\s*[:=;-]\s*(.+)`)
	pngSignature        = []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}
	jpegSignature       = []byte{0xff, 0xd8, 0xff}
)
