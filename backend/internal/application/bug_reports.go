package application

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/mail"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"filippo.io/age"
	"filippo.io/age/armor"
)

const bugReportMarkerVersion = "v1"

var bugReportUUIDPattern = regexp.MustCompile(
	`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[1-8][0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}$`,
)

var (
	bugReportVersionPattern  = regexp.MustCompile(`^\d{0,4}$`)
	bugReportLanguagePattern = regexp.MustCompile(`^[A-Za-z]{2,3}(?:-[A-Za-z0-9]{2,8})*$`)
	bugReportTimeZonePattern = regexp.MustCompile(`^[A-Za-z0-9_+.-]+(?:/[A-Za-z0-9_+.-]+)*$`)
	bugReportRoutePattern    = regexp.MustCompile(`^/[a-z][a-z-]{0,39}$`)
)

var bugReportProblemCodes = map[string]struct{}{
	"network_error": {}, "request_failed": {}, "request_timeout": {}, "unknown_error": {},
}

var bugReportBrowserFamilies = map[string]struct{}{
	"Chrome": {}, "Edge": {}, "Firefox": {}, "Safari": {}, "Unknown": {},
}

var bugReportOSFamilies = map[string]struct{}{
	"Android": {}, "iOS": {}, "Linux": {}, "macOS": {}, "Unknown": {}, "Windows": {},
}

var bugReportClientErrorClasses = map[string]struct{}{
	"AbortError": {}, "ApiError": {}, "DOMException": {}, "Error": {}, "RangeError": {},
	"ReferenceError": {}, "SyntaxError": {}, "TypeError": {}, "UnhandledRejection": {},
}

var bugReportClientErrorMessages = map[string]string{
	"window": "Unexpected client error", "promise": "Unhandled promise rejection",
	"api": "API request failed", "network": "Network request failed", "timeout": "Request timed out",
}

var bugReportHTTPMethods = map[string]struct{}{
	"DELETE": {}, "GET": {}, "HEAD": {}, "OPTIONS": {}, "PATCH": {}, "POST": {}, "PUT": {},
}

var bugReportDiagnosticRoutes = map[string]struct{}{
	"/":            {},
	"/accessories": {}, "/article-search": {}, "/auth": {}, "/backups": {}, "/bug-reports": {},
	"/data-transfer": {}, "/database": {}, "/digital-centers": {}, "/ecos": {}, "/exhibition": {},
	"/exhibitions": {}, "/import-export": {}, "/master-data": {}, "/overview": {},
	"/password-reset": {},
	"/profile":        {}, "/settings": {}, "/setup": {},
	"/users": {}, "/vehicle-sets": {}, "/vehicle-types": {}, "/vehicles": {}, "/version": {},
}

var bugReportAreaLabels = map[BugReportArea]string{
	BugReportAreaOverview:       "area:overview",
	BugReportAreaVehicles:       "area:vehicles",
	BugReportAreaAccessories:    "area:accessories",
	BugReportAreaExhibition:     "area:exhibition",
	BugReportAreaImportExport:   "area:import-export",
	BugReportAreaDigitalCenters: "area:digital-centers",
	BugReportAreaSettings:       "area:settings",
	BugReportAreaOther:          "area:other",
}

var bugReportGermanAreaNames = map[BugReportArea]string{
	BugReportAreaOverview:       "Übersicht",
	BugReportAreaVehicles:       "Fahrzeuge",
	BugReportAreaAccessories:    "Zubehör",
	BugReportAreaExhibition:     "Ausstellung",
	BugReportAreaImportExport:   "Import/Export",
	BugReportAreaDigitalCenters: "Digitalzentralen",
	BugReportAreaSettings:       "Einstellungen",
	BugReportAreaOther:          "Sonstiges",
}

var bugReportEnglishAreaNames = map[BugReportArea]string{
	BugReportAreaOverview:       "Overview",
	BugReportAreaVehicles:       "Vehicles",
	BugReportAreaAccessories:    "Accessories",
	BugReportAreaExhibition:     "Exhibition",
	BugReportAreaImportExport:   "Import/Export",
	BugReportAreaDigitalCenters: "Digital centers",
	BugReportAreaSettings:       "Settings",
	BugReportAreaOther:          "Other",
}

func ValidateBugReportInput(input BugReportInput, hasScreenshot bool) (ValidatedBugReport, error) {
	input.ReportID = strings.TrimSpace(input.ReportID)
	input.Description = strings.TrimSpace(input.Description)
	input.ExpectedBehavior = strings.TrimSpace(input.ExpectedBehavior)
	input.ReproductionSteps = strings.TrimSpace(input.ReproductionSteps)
	input.Email = strings.TrimSpace(input.Email)
	areaLabel, areaKnown := bugReportAreaLabels[input.Area]
	if !bugReportUUIDPattern.MatchString(input.ReportID) {
		return ValidatedBugReport{}, fmt.Errorf("%w: invalid report id", ErrBugReportValidation)
	}
	if !areaKnown {
		return ValidatedBugReport{}, fmt.Errorf("%w: unsupported area", ErrBugReportValidation)
	}
	if err := validateRuneLength("description", input.Description, 20, 10_000); err != nil {
		return ValidatedBugReport{}, err
	}
	if err := validateRuneLength("expected behavior", input.ExpectedBehavior, 0, 5_000); err != nil {
		return ValidatedBugReport{}, err
	}
	if err := validateRuneLength("reproduction steps", input.ReproductionSteps, 0, 10_000); err != nil {
		return ValidatedBugReport{}, err
	}
	if input.Email != "" {
		if len([]byte(input.Email)) > 254 {
			return ValidatedBugReport{}, fmt.Errorf("%w: email is too long", ErrBugReportValidation)
		}
		address, err := mail.ParseAddress(input.Email)
		if err != nil || address.Address != input.Email || strings.ContainsAny(input.Email, "\r\n") {
			return ValidatedBugReport{}, fmt.Errorf("%w: invalid email", ErrBugReportValidation)
		}
	}
	if hasScreenshot && !input.ScreenshotConsent {
		return ValidatedBugReport{}, fmt.Errorf("%w: screenshot consent is required", ErrBugReportValidation)
	}
	if input.OmitDiagnostics {
		input.ClientDiagnostics = BugReportClientDiagnostics{}
	} else if err := validateBugReportClientDiagnostics(input.Area, input.ClientDiagnostics); err != nil {
		return ValidatedBugReport{}, err
	}
	return ValidatedBugReport{BugReportInput: input, AreaLabel: areaLabel}, nil
}

func validateRuneLength(field string, value string, minimum int, maximum int) error {
	if !utf8.ValidString(value) {
		return fmt.Errorf("%w: %s is not valid UTF-8", ErrBugReportValidation, field)
	}
	length := utf8.RuneCountInString(value)
	if length < minimum || length > maximum {
		return fmt.Errorf("%w: invalid %s length", ErrBugReportValidation, field)
	}
	return nil
}

func validateBugReportClientDiagnostics(area BugReportArea, diagnostics BugReportClientDiagnostics) error {
	if len(diagnostics.Errors) > 30 {
		return fmt.Errorf("%w: too many client errors", ErrBugReportValidation)
	}
	if diagnostics.ViewportWidth < 0 || diagnostics.ViewportWidth > 32_768 ||
		diagnostics.ViewportHeight < 0 || diagnostics.ViewportHeight > 32_768 {
		return fmt.Errorf("%w: invalid viewport", ErrBugReportValidation)
	}
	if diagnostics.Area != string(area) {
		return fmt.Errorf("%w: diagnostic area does not match report area", ErrBugReportValidation)
	}
	if _, valid := bugReportDiagnosticRoutes[diagnostics.Pathname]; !valid {
		return fmt.Errorf("%w: invalid pathname", ErrBugReportValidation)
	}
	if diagnostics.Theme != "" && diagnostics.Theme != "light" && diagnostics.Theme != "dark" {
		return fmt.Errorf("%w: invalid theme", ErrBugReportValidation)
	}
	if _, valid := bugReportBrowserFamilies[diagnostics.BrowserFamily]; !valid ||
		!bugReportVersionPattern.MatchString(diagnostics.BrowserMajor) {
		return fmt.Errorf("%w: invalid browser diagnostic", ErrBugReportValidation)
	}
	if _, valid := bugReportOSFamilies[diagnostics.OSFamily]; !valid ||
		!bugReportVersionPattern.MatchString(diagnostics.OSMajor) {
		return fmt.Errorf("%w: invalid operating system diagnostic", ErrBugReportValidation)
	}
	if !bugReportLanguagePattern.MatchString(diagnostics.Language) || len(diagnostics.Language) > 35 ||
		!bugReportTimeZonePattern.MatchString(diagnostics.TimeZone) || len(diagnostics.TimeZone) > 100 {
		return fmt.Errorf("%w: invalid locale diagnostic", ErrBugReportValidation)
	}
	if len(diagnostics.Roles) > 8 {
		return fmt.Errorf("%w: too many roles", ErrBugReportValidation)
	}
	seenRoles := make(map[string]struct{}, len(diagnostics.Roles))
	for _, role := range diagnostics.Roles {
		if role != "Admin" && role != "Editor" && role != "Viewer" && role != "Messe" {
			return fmt.Errorf("%w: invalid role", ErrBugReportValidation)
		}
		if _, exists := seenRoles[role]; exists {
			return fmt.Errorf("%w: duplicate role", ErrBugReportValidation)
		}
		seenRoles[role] = struct{}{}
	}
	for _, clientError := range diagnostics.Errors {
		if _, err := time.Parse(time.RFC3339, clientError.Timestamp); err != nil {
			return fmt.Errorf("%w: invalid client error timestamp", ErrBugReportValidation)
		}
		expectedMessage, validCategory := bugReportClientErrorMessages[clientError.Category]
		_, validClass := bugReportClientErrorClasses[clientError.ErrorClass]
		if !validCategory || !validClass || clientError.Message != expectedMessage {
			return fmt.Errorf("%w: invalid client error classification", ErrBugReportValidation)
		}
		if (clientError.Category == "api" || clientError.Category == "network" || clientError.Category == "timeout") &&
			clientError.ProblemCode == "" {
			return fmt.Errorf("%w: missing client error problem code", ErrBugReportValidation)
		}
		if clientError.ProblemCode != "" {
			if _, valid := bugReportProblemCodes[clientError.ProblemCode]; !valid {
				return fmt.Errorf("%w: invalid client error problem code", ErrBugReportValidation)
			}
		}
		if clientError.Method != "" {
			if _, valid := bugReportHTTPMethods[clientError.Method]; !valid {
				return fmt.Errorf("%w: invalid client error method", ErrBugReportValidation)
			}
		}
		if clientError.Route != "" {
			if !bugReportRoutePattern.MatchString(clientError.Route) {
				return fmt.Errorf("%w: invalid client error route", ErrBugReportValidation)
			}
			if _, valid := bugReportDiagnosticRoutes[clientError.Route]; !valid {
				return fmt.Errorf("%w: unknown client error route", ErrBugReportValidation)
			}
		}
		if clientError.Status != 0 && (clientError.Status < 100 || clientError.Status > 599) {
			return fmt.Errorf("%w: invalid client error", ErrBugReportValidation)
		}
	}
	return nil
}

func BuildBugReportIssue(input BugReportIssueInput) (BugReportIssue, error) {
	clientJSON, err := json.MarshalIndent(input.Report.ClientDiagnostics, "", "  ")
	if err != nil {
		return BugReportIssue{}, fmt.Errorf("encode client diagnostics: %w", err)
	}
	serverJSON, err := json.MarshalIndent(input.ServerDiagnostics, "", "  ")
	if err != nil {
		return BugReportIssue{}, fmt.Errorf("encode server diagnostics: %w", err)
	}
	areaNames := bugReportEnglishAreaNames
	if strings.HasPrefix(strings.ToLower(input.Report.ClientDiagnostics.Language), "de") {
		areaNames = bugReportGermanAreaNames
	}
	areaName := areaNames[input.Report.Area]
	titlePrefix := "[Bug][" + areaName + "] "
	summary := firstBugReportSentence(input.Report.Description)
	summary = truncateRunes(summary, 120-utf8.RuneCountInString(titlePrefix))

	var body strings.Builder
	body.WriteString("> [!IMPORTANT]\n> Dieser Bericht wurde automatisch aus RailKeeper erstellt. " +
		"This report was created automatically by RailKeeper. Sicherheitslücken und vertrauliche Daten " +
		"gehören nicht in öffentliche Issues. Vulnerabilities and confidential data must not be posted " +
		"in public issues. [Privat melden / Report privately](https://github.com/ichwars/RailKeeper/security/advisories/new).\n\n")
	writeBugReportTextSection(&body, "Fehlerbeschreibung / Bug description", input.Report.Description)
	writeBugReportTextSection(&body, "Erwartetes Verhalten / Expected behavior", input.Report.ExpectedBehavior)
	writeBugReportTextSection(&body, "Schritte zur Reproduktion / Reproduction steps", input.Report.ReproductionSteps)
	body.WriteString("## Produktbereich / Product area\n\n`")
	body.WriteString(string(input.Report.Area))
	body.WriteString("`\n\n")
	if input.Report.OmitDiagnostics {
		body.WriteString("Diagnosedaten wurden vom Benutzer deaktiviert. Diagnostics were disabled by the reporter.\n\n")
	} else {
		body.WriteString("## Automatisch erfasste Diagnosedaten / Automatically collected diagnostics\n\n")
		body.WriteString("### Server\n\n```json\n")
		body.Write(serverJSON)
		body.WriteString("\n```\n\n### Browser\n\n```json\n")
		body.Write(clientJSON)
		body.WriteString("\n```\n\n")
	}
	if input.EncryptedContact != "" {
		body.WriteString("<details><summary>Verschlüsselter Kontakt / Encrypted contact</summary>\n\n```text\n")
		body.WriteString(input.EncryptedContact)
		body.WriteString("\n```\n\n</details>\n\n")
	}
	if input.HasScreenshot {
		body.WriteString("## Öffentlicher Screenshot / Public screenshot\n\n")
		body.WriteString("Der separat hochgeladene Screenshot ist öffentlich sichtbar. " +
			"The separately uploaded screenshot is publicly visible.\n\n")
	}
	body.WriteString("<!-- railkeeper-bug-report:")
	body.WriteString(bugReportMarkerVersion)
	body.WriteString(":")
	body.WriteString(input.Report.ReportID)
	body.WriteString(" -->\n")

	labels := normalizeBugReportLabels(input.BaseLabels, input.Report.AreaLabel)
	return BugReportIssue{Title: titlePrefix + summary, Body: body.String(), Labels: labels}, nil
}

func writeBugReportTextSection(body *strings.Builder, heading string, value string) {
	if value == "" {
		return
	}
	body.WriteString("## ")
	body.WriteString(heading)
	body.WriteString("\n\n")
	fence := bugReportFence(value)
	body.WriteString(fence)
	body.WriteString("text\n")
	body.WriteString(value)
	body.WriteString("\n")
	body.WriteString(fence)
	body.WriteString("\n\n")
}

func bugReportFence(value string) string {
	longest := 2
	current := 0
	for _, valueRune := range value {
		if valueRune == '`' {
			current++
			if current > longest {
				longest = current
			}
			continue
		}
		current = 0
	}
	return strings.Repeat("`", longest+1)
}

func firstBugReportSentence(value string) string {
	value = strings.Map(func(valueRune rune) rune {
		if unicode.IsControl(valueRune) {
			return ' '
		}
		return valueRune
	}, value)
	value = strings.Join(strings.Fields(value), " ")
	for index, valueRune := range value {
		if valueRune == '.' || valueRune == '!' || valueRune == '?' {
			return strings.TrimSpace(value[:index])
		}
	}
	return value
}

func truncateRunes(value string, maximum int) string {
	if maximum <= 0 {
		return ""
	}
	runes := []rune(value)
	if len(runes) <= maximum {
		return value
	}
	if maximum == 1 {
		return "…"
	}
	return strings.TrimSpace(string(runes[:maximum-1])) + "…"
}

func normalizeBugReportLabels(baseLabels []string, areaLabel string) []string {
	seen := make(map[string]struct{}, len(baseLabels)+1)
	labels := make([]string, 0, len(baseLabels)+1)
	for _, label := range append(append([]string{}, baseLabels...), areaLabel) {
		label = strings.TrimSpace(label)
		if label == "" {
			continue
		}
		key := strings.ToLower(label)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		labels = append(labels, label)
	}
	return labels
}

type AgeContactEncryptor struct {
	recipient age.Recipient
}

func NewAgeContactEncryptor(recipientText string) (*AgeContactEncryptor, error) {
	recipient, err := age.ParseX25519Recipient(strings.TrimSpace(recipientText))
	if err != nil {
		return nil, fmt.Errorf("parse bug report contact recipient: %w", err)
	}
	return &AgeContactEncryptor{recipient: recipient}, nil
}

func (encryptor *AgeContactEncryptor) Encrypt(value string) (string, error) {
	var output bytes.Buffer
	armorWriter := armor.NewWriter(&output)
	writer, err := age.Encrypt(armorWriter, encryptor.recipient)
	if err != nil {
		return "", fmt.Errorf("start contact encryption: %w", err)
	}
	if _, err := writer.Write([]byte(value)); err != nil {
		return "", fmt.Errorf("encrypt contact: %w", err)
	}
	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("finish contact encryption: %w", err)
	}
	if err := armorWriter.Close(); err != nil {
		return "", fmt.Errorf("armor encrypted contact: %w", err)
	}
	return output.String(), nil
}

func BugReportSupportedAreas() []BugReportArea {
	areas := make([]BugReportArea, 0, len(bugReportAreaLabels))
	for area := range bugReportAreaLabels {
		areas = append(areas, area)
	}
	sort.Slice(areas, func(left int, right int) bool { return areas[left] < areas[right] })
	return areas
}
