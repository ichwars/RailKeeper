package api

import (
	"encoding/binary"
	"strings"
	"testing"
)

func TestApplyESUXMetadata(t *testing.T) {
	xml := `<?xml version="1.0" encoding="utf-8"?><meta xmlns="http://www.esu.eu/2010/LOKPROGRAMMER/Metadata"><name>V180</name><address>1002</address><type>diesel</type><decoder>LokPilot micro V4.0 DCC</decoder><manufacturer>Piko</manufacturer><manid></manid><lokprogrammer>5.2.15</lokprogrammer></meta>`
	data := make([]byte, 18+len(xml))
	copy(data[:4], []byte("ESU "))
	data[6] = 2
	data[10] = 1
	data[12] = 16
	binary.LittleEndian.PutUint32(data[14:18], uint32(len(xml)))
	copy(data[18:], []byte(xml))

	profile, description := applyESUXMetadata("V180.esux", data, "", "")
	if profile != "LokPilot micro V4.0 DCC" {
		t.Fatalf("unexpected profile %q", profile)
	}
	if description == "" || !containsAll(description, "ESU LokProgrammer-Projekt", "Name: V180", "Adresse: 1002", "Hersteller: Piko") {
		t.Fatalf("unexpected description %q", description)
	}
}

func TestApplyESUXMetadataPreservesUserInput(t *testing.T) {
	xml := `<meta><name>Projekt</name><decoder>Decoder</decoder></meta>`
	data := make([]byte, 18+len(xml))
	copy(data[:4], []byte("ESU "))
	binary.LittleEndian.PutUint32(data[14:18], uint32(len(xml)))
	copy(data[18:], []byte(xml))

	profile, description := applyESUXMetadata("projekt.esux", data, "Manuell", "Eigene Notiz")
	if profile != "Manuell" || description != "Eigene Notiz" {
		t.Fatalf("user input was not preserved: %q %q", profile, description)
	}
}

func TestESUXPreviewResponse(t *testing.T) {
	xml := `<meta><name>BR106</name><address>768</address><type>diesel</type><decoder>LokPilot 5 micro DCC</decoder><manufacturer>Tillig</manufacturer><lokprogrammer>5.2.15</lokprogrammer></meta>`
	data := make([]byte, 18+len(xml))
	copy(data[:4], []byte("ESU "))
	binary.LittleEndian.PutUint32(data[14:18], uint32(len(xml)))
	copy(data[18:], []byte(xml))

	preview := esuxPreviewResponse("BR106.esux", int64(len(data)), "application/octet-stream", data)
	if !preview.HasMetadata {
		t.Fatal("expected metadata in preview")
	}
	if preview.ProjectName != "BR106" || preview.Address != "768" || preview.Decoder != "LokPilot 5 micro DCC" {
		t.Fatalf("unexpected preview metadata: %+v", preview)
	}
	if preview.SuggestedDecoderProfile != "LokPilot 5 micro DCC" {
		t.Fatalf("unexpected decoder profile suggestion %q", preview.SuggestedDecoderProfile)
	}
	if !containsAll(preview.SuggestedDescription, "ESU LokProgrammer-Projekt", "Name: BR106", "Adresse: 768", "Hersteller: Tillig") {
		t.Fatalf("unexpected description suggestion %q", preview.SuggestedDescription)
	}
}

func TestESUXPreviewResponseWithoutMetadata(t *testing.T) {
	data := []byte("plain cv file")

	preview := esuxPreviewResponse("cv.txt", int64(len(data)), "text/plain", data)
	if preview.HasMetadata {
		t.Fatalf("expected no metadata: %+v", preview)
	}
	if preview.FileName != "cv.txt" || preview.SizeBytes != int64(len(data)) || preview.MimeType != "text/plain" {
		t.Fatalf("unexpected preview base data: %+v", preview)
	}
}

func containsAll(value string, parts ...string) bool {
	for _, part := range parts {
		if !strings.Contains(value, part) {
			return false
		}
	}
	return true
}
