package application

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const maxCS3ResponseBytes = 8 * 1024 * 1024

const (
	cs3ErrorNetwork        = cs3ErrorKind("network")
	cs3ErrorAuthentication = cs3ErrorKind("authentication")
	cs3ErrorHTTP           = cs3ErrorKind("http")
	cs3ErrorRedirect       = cs3ErrorKind("redirect")
	cs3ErrorContentType    = cs3ErrorKind("content_type")
	cs3ErrorFormat         = cs3ErrorKind("format")
	cs3ErrorUnsupported    = cs3ErrorKind("unsupported_version")
	cs3ErrorDeviceOutput   = cs3ErrorKind("device_output")
	cs3ErrorNotFound       = cs3ErrorKind("not_found")
)

type cs3ErrorKind string

type cs3ReadError struct {
	kind   cs3ErrorKind
	status string
	cause  error
}

func (err *cs3ReadError) Error() string {
	if err == nil {
		return ""
	}
	if err.cause == nil {
		return "CS3-Antwort ist ungültig"
	}
	if err.kind == cs3ErrorRedirect {
		return fmt.Sprintf("CS3-Weiterleitung wurde abgelehnt: %v", err.cause)
	}
	return fmt.Sprintf("CS3-Antwort ist ungültig: %v", err.cause)
}

func (err *cs3ReadError) Unwrap() error {
	if err == nil {
		return nil
	}
	return err.cause
}

type CS3RosterMetadata struct {
	APIPath         string `json:"apiPath"`
	APIGeneration   string `json:"apiGeneration"`
	HTTPStatus      string `json:"httpStatus"`
	ContentType     string `json:"contentType"`
	LocomotiveCount int    `json:"locomotiveCount"`
}

type DigitalCenterLocomotive = ECoSRawLocomotive

type cs3LocomotiveRecord struct {
	UID          string `json:"uid"`
	Name         string `json:"name"`
	InternalName string `json:"internname"`
	Address      *int   `json:"address"`
	DecoderType  string `json:"dectyp"`
}

type cs3APIEndpoint struct {
	path       string
	generation string
}

var cs3APIEndpoints = []cs3APIEndpoint{
	{path: "/app/api/locos", generation: "2.6+"},
	{path: "/app/api/loks", generation: "pre-2.6"},
}

func (s *DigitalCenterService) ReadCS3Locomotives(
	ctx context.Context,
	input DigitalCenterConnectionInput,
) ([]DigitalCenterLocomotive, CS3RosterMetadata, error) {
	target, err := normalizeDigitalCenterInput(input, defaultCS3Port)
	if err != nil {
		return nil, CS3RosterMetadata{}, err
	}
	return s.readCS3Locomotives(ctx, target)
}

func (s *DigitalCenterService) readCS3Locomotives(
	ctx context.Context,
	target DigitalCenterConnectionInput,
) ([]DigitalCenterLocomotive, CS3RosterMetadata, error) {
	var lastMetadata CS3RosterMetadata
	for _, endpoint := range cs3APIEndpoints {
		locomotives, metadata, err := s.fetchCS3Locomotives(ctx, target, endpoint)
		lastMetadata = metadata
		if err == nil {
			return locomotives, metadata, nil
		}
		if cs3ErrorKindOf(err) != cs3ErrorNotFound {
			return nil, metadata, err
		}
	}
	return nil, lastMetadata, &cs3ReadError{
		kind:  cs3ErrorUnsupported,
		cause: errors.New("keine unterstützte Loklisten-API gefunden"),
	}
}

func (s *DigitalCenterService) ProbeCS3Connection(
	ctx context.Context,
	input DigitalCenterConnectionInput,
) (*DigitalCenterProbeResult, error) {
	target, err := normalizeDigitalCenterInput(input, defaultCS3Port)
	if err != nil {
		return nil, err
	}
	result := &DigitalCenterProbeResult{
		Provider: "cs3",
		Host:     target.Host,
		Port:     target.Port,
		Message:  "Keine unterstützte CS3-Loklisten-API gefunden.",
		Fields:   map[string]string{},
		Commands: []DigitalCenterProbeCommandResult{},
	}
	for _, endpoint := range cs3APIEndpoints {
		locomotives, metadata, readErr := s.fetchCS3Locomotives(ctx, target, endpoint)
		command := DigitalCenterProbeCommandResult{
			Name:        "CS3_LOCOMOTIVE_API",
			Description: "Read-only Loklisten-API",
			Request:     "GET " + endpoint.path,
			Fields:      cs3DiagnosticFields(metadata, len(locomotives)),
		}
		if readErr != nil {
			command.Error = cs3UserMessage(readErr)
			command.Fields["errorKind"] = string(cs3ErrorKindOf(readErr))
			result.Commands = append(result.Commands, command)
			if cs3ErrorKindOf(readErr) == cs3ErrorNotFound {
				continue
			}
			result.Message = command.Error
			result.Fields = command.Fields
			return result, nil
		}
		command.OK = true
		result.Commands = append(result.Commands, command)
		result.Connected = true
		result.Message = "CS3-Diagnose abgeschlossen. Kompatible read-only Loklisten-API erkannt."
		result.Fields = command.Fields
		return result, nil
	}
	result.Fields["errorKind"] = string(cs3ErrorUnsupported)
	return result, nil
}

func (s *DigitalCenterService) fetchCS3Locomotives(
	ctx context.Context,
	target DigitalCenterConnectionInput,
	endpoint cs3APIEndpoint,
) ([]DigitalCenterLocomotive, CS3RosterMetadata, error) {
	metadata := CS3RosterMetadata{APIPath: endpoint.path, APIGeneration: endpoint.generation}
	requestURL := url.URL{
		Scheme: "http",
		Host:   net.JoinHostPort(target.Host, strconv.Itoa(target.Port)),
		Path:   endpoint.path,
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL.String(), nil)
	if err != nil {
		return nil, metadata, &cs3ReadError{kind: cs3ErrorNetwork, cause: err}
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("User-Agent", "RailKeeper")

	client := *s.httpClient()
	client.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	response, err := client.Do(request)
	if err != nil {
		return nil, metadata, &cs3ReadError{kind: cs3ErrorNetwork, cause: err}
	}
	defer func() { _ = response.Body.Close() }()
	metadata.HTTPStatus = response.Status
	metadata.ContentType = response.Header.Get("Content-Type")

	switch {
	case response.StatusCode == http.StatusUnauthorized || response.StatusCode == http.StatusForbidden:
		return nil, metadata, &cs3ReadError{
			kind: cs3ErrorAuthentication, status: response.Status, cause: errors.New(response.Status),
		}
	case response.StatusCode == http.StatusNotFound:
		return nil, metadata, &cs3ReadError{
			kind: cs3ErrorNotFound, status: response.Status, cause: errors.New(response.Status),
		}
	case response.StatusCode >= 300 && response.StatusCode < 400:
		return nil, metadata, &cs3ReadError{
			kind: cs3ErrorRedirect, status: response.Status, cause: errors.New(response.Status),
		}
	case response.StatusCode < 200 || response.StatusCode >= 300:
		return nil, metadata, &cs3ReadError{
			kind: cs3ErrorHTTP, status: response.Status, cause: errors.New(response.Status),
		}
	}

	mediaType, _, parseErr := mime.ParseMediaType(metadata.ContentType)
	if parseErr != nil || (mediaType != "application/json" && !strings.HasSuffix(mediaType, "+json")) {
		return nil, metadata, &cs3ReadError{
			kind: cs3ErrorContentType, status: response.Status,
			cause: fmt.Errorf("unerwarteter Content-Type %q", metadata.ContentType),
		}
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, maxCS3ResponseBytes+1))
	if err != nil {
		return nil, metadata, &cs3ReadError{kind: cs3ErrorNetwork, status: response.Status, cause: err}
	}
	if len(data) > maxCS3ResponseBytes {
		return nil, metadata, &cs3ReadError{
			kind: cs3ErrorFormat, status: response.Status, cause: errors.New("Antwort ist zu groß"),
		}
	}
	var records []cs3LocomotiveRecord
	if err := json.Unmarshal(data, &records); err != nil {
		return nil, metadata, &cs3ReadError{
			kind: cs3ErrorFormat, status: response.Status, cause: fmt.Errorf("ungültiges JSON-Array: %w", err),
		}
	}
	locomotives, err := normalizeCS3Locomotives(records)
	if err != nil {
		return nil, metadata, err
	}
	metadata.LocomotiveCount = len(locomotives)
	return locomotives, metadata, nil
}

func normalizeCS3Locomotives(records []cs3LocomotiveRecord) ([]DigitalCenterLocomotive, error) {
	if len(records) > maxDigitalCenterLocomotives {
		return nil, &cs3ReadError{kind: cs3ErrorDeviceOutput, cause: errors.New("zu viele Lokomotiven")}
	}
	locomotives := make([]DigitalCenterLocomotive, 0, len(records))
	seen := make(map[int]struct{}, len(records))
	for _, record := range records {
		objectID64, err := strconv.ParseUint(strings.TrimSpace(record.UID), 0, 32)
		if err != nil || objectID64 < 1 || objectID64 > maxDigitalCenterObjectID {
			return nil, &cs3ReadError{kind: cs3ErrorDeviceOutput, cause: errors.New("ungültige Lok-UID")}
		}
		objectID := int(objectID64)
		if _, found := seen[objectID]; found {
			return nil, &cs3ReadError{kind: cs3ErrorDeviceOutput, cause: errors.New("doppelte Lok-UID")}
		}
		seen[objectID] = struct{}{}
		if record.Address == nil || *record.Address < 1 || *record.Address > maxDigitalCenterAddress {
			return nil, &cs3ReadError{kind: cs3ErrorDeviceOutput, cause: errors.New("ungültige Decoderadresse")}
		}
		name := strings.TrimSpace(record.Name)
		if name == "" {
			name = decodeCS3InternalName(record.InternalName)
		}
		name, err = normalizeDigitalCenterName(name)
		if err != nil || name == "" {
			return nil, &cs3ReadError{kind: cs3ErrorDeviceOutput, cause: errors.New("ungültiger Lokname")}
		}
		protocol, err := normalizeDigitalCenterProtocol(record.DecoderType)
		if err != nil || protocol == "" {
			return nil, &cs3ReadError{kind: cs3ErrorDeviceOutput, cause: errors.New("ungültiges Decoderprotokoll")}
		}
		locomotives = append(locomotives, DigitalCenterLocomotive{
			ObjectID: objectID,
			Name:     name,
			Address:  *record.Address,
			Protocol: protocol,
		})
	}
	return locomotives, nil
}

func decodeCS3InternalName(value string) string {
	value = strings.TrimSpace(value)
	if value == "" || !strings.Contains(value, "#") {
		return value
	}
	decoded, err := url.PathUnescape(strings.ReplaceAll(value, "#", "%"))
	if err != nil {
		return value
	}
	return decoded
}

func cs3DiagnosticFields(metadata CS3RosterMetadata, count int) map[string]string {
	fields := map[string]string{
		"apiPath":         metadata.APIPath,
		"apiGeneration":   metadata.APIGeneration,
		"locomotiveCount": strconv.Itoa(count),
		"readOnly":        "true",
	}
	if metadata.HTTPStatus != "" {
		fields["httpStatus"] = metadata.HTTPStatus
	}
	if metadata.ContentType != "" {
		fields["contentType"] = metadata.ContentType
	}
	return fields
}

func cs3ErrorKindOf(err error) cs3ErrorKind {
	var readErr *cs3ReadError
	if errors.As(err, &readErr) {
		return readErr.kind
	}
	return cs3ErrorNetwork
}

func cs3UserMessage(err error) string {
	switch cs3ErrorKindOf(err) {
	case cs3ErrorAuthentication:
		return "CS3-Authentifizierung fehlgeschlagen. Zugriff auf die read-only API prüfen."
	case cs3ErrorHTTP:
		return "CS3-HTTP-Antwort war nicht erfolgreich."
	case cs3ErrorRedirect:
		return "CS3-Weiterleitung wurde aus Sicherheitsgründen abgelehnt."
	case cs3ErrorContentType:
		return "CS3-API lieferte kein JSON. Möglicherweise wurde nur eine Webseite erreicht."
	case cs3ErrorFormat:
		return "CS3-API lieferte ungültiges JSON oder ein unbekanntes Format."
	case cs3ErrorUnsupported, cs3ErrorNotFound:
		return "Keine unterstützte CS3-Loklisten-API gefunden."
	case cs3ErrorDeviceOutput:
		return "CS3-Lokdaten sind ungültig und wurden nicht übernommen."
	default:
		return "CS3 nicht erreichbar: Netzwerk- oder Timeoutfehler."
	}
}
