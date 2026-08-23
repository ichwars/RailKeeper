package application

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

type ECoSLocomotiveCreateInput struct {
	Host    string                    `json:"host"`
	Port    int                       `json:"port"`
	Desired ECoSLocomotiveSyncDesired `json:"desired"`
	Confirm bool                      `json:"confirm"`
}

type ECoSLocomotiveCreateResult struct {
	ObjectID int                       `json:"objectId"`
	Desired  ECoSLocomotiveSyncDesired `json:"desired"`
	Command  string                    `json:"command"`
	Applied  bool                      `json:"applied"`
	RawLines []string                  `json:"rawLines,omitempty"`
	Message  string                    `json:"message"`
}

func (s *ECoSService) CreateLocomotive(
	ctx context.Context,
	input ECoSLocomotiveCreateInput,
) (*ECoSLocomotiveCreateResult, error) {
	target, err := normalizeECoSInput(ECoSConnectionInput{Host: input.Host, Port: input.Port})
	if err != nil {
		return nil, err
	}
	desired := cleanECoSLocomotiveSyncDesired(input.Desired)
	command, err := buildECoSLocomotiveCreateCommand(desired)
	if err != nil {
		return nil, err
	}
	result := &ECoSLocomotiveCreateResult{
		Desired: desired,
		Command: command,
		Message: "ECoS-Lok kann angelegt werden.",
	}
	if !input.Confirm {
		return result, nil
	}
	lines, err := s.exchange(ctx, target.Host, target.Port, command)
	if err != nil {
		return nil, fmt.Errorf("%w: ECoS-Erstellungsantwort fehlt: %v", ErrECoSWriteStateUnknown, err)
	}
	result.RawLines = lines
	status, ok := parseECoSEndStatus(lines)
	if status == "" {
		return nil, fmt.Errorf("%w: ECoS-Erstellungsantwort enthält keinen Status", ErrECoSWriteStateUnknown)
	}
	if !ok {
		return nil, fmt.Errorf("ECoS-Erstellungsbefehl nicht bestätigt: %s", status)
	}
	objectID, err := strconv.Atoi(strings.TrimSpace(parseECoSFields(lines)["id"]))
	if err != nil || objectID < 1 || objectID > maxDigitalCenterObjectID {
		return nil, fmt.Errorf("%w: ECoS-Erstellungsantwort enthält keine gültige Objekt-ID",
			ErrECoSWriteStateUnknown)
	}
	result.ObjectID = objectID
	result.Applied = true
	result.Message = "ECoS-Lok wurde angelegt."
	return result, nil
}

func buildECoSLocomotiveCreateCommand(desired ECoSLocomotiveSyncDesired) (string, error) {
	if desired.Name == "" || desired.Address < 1 || desired.Address > maxDigitalCenterAddress ||
		desired.Protocol == "" {
		return "", errors.New("ECoS-Lok benötigt Name, Decoderadresse und Protokoll")
	}
	if strings.IndexFunc(desired.Name, unicode.IsControl) >= 0 {
		return "", errors.New("ECoS-Name enthielt unzulässige Zeichen")
	}
	protocol, err := eCoSLocomotiveCreateProtocol(desired.Protocol)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("create(10, addr[%d], name[%s], protocol[%s], append)",
		desired.Address, quoteECoSString(desired.Name), protocol), nil
}

func eCoSLocomotiveCreateProtocol(value string) (string, error) {
	token := strings.ToUpper(strings.TrimSpace(value))
	switch token {
	case "DCC14", "DCC28", "DCC128", "MM14", "MM27", "MM28", "SX32", "MMFKT":
		return token, nil
	}
	normalized, err := normalizeDigitalCenterProtocol(value)
	if err != nil {
		return "", err
	}
	switch normalized {
	case "DCC":
		return "DCC28", nil
	case "MOTOROLA":
		return "MM28", nil
	case "SELECTRIX":
		return "SX32", nil
	case "MFX":
		return "M4", nil
	default:
		return "", fmt.Errorf("ECoS-Protokoll %q wird beim Anlegen nicht unterstützt", normalized)
	}
}
