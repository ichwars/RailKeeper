package api

import (
	"encoding/binary"
	"encoding/xml"
	"errors"
	"fmt"
	"path/filepath"
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
