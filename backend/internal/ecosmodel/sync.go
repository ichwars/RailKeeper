package ecosmodel

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

func CleanECoSLocomotiveSyncDesired(input ECoSLocomotiveSyncDesired) ECoSLocomotiveSyncDesired {
	input.Name = strings.TrimSpace(input.Name)
	input.Protocol = strings.TrimSpace(input.Protocol)
	if input.Address < 0 {
		input.Address = 0
	}
	return input
}

func BuildECoSLocomotiveSyncCommands(objectID int, current *ECoSLocomotive, desired ECoSLocomotiveSyncDesired) ([]ECoSLocomotiveSyncChange, []string, error) {
	if current == nil {
		return nil, nil, errors.New("ECoS-Istwerte fehlen")
	}
	changes := []ECoSLocomotiveSyncChange{}
	commands := []string{}
	if desired.Name != "" && desired.Name != current.Name {
		if strings.IndexFunc(desired.Name, unicode.IsControl) >= 0 {
			return nil, nil, errors.New("ECoS-Name enthielt unzulässige Zeichen")
		}
		changes = append(changes, ECoSLocomotiveSyncChange{Field: "name", Current: current.Name, Desired: desired.Name})
		commands = append(commands, fmt.Sprintf("set(%d, name[%s])", objectID, QuoteECoSString(desired.Name)))
	}
	if desired.Address > 0 && desired.Address != current.Address {
		changes = append(changes, ECoSLocomotiveSyncChange{Field: "address", Current: strconv.Itoa(current.Address), Desired: strconv.Itoa(desired.Address)})
		commands = append(commands, fmt.Sprintf("set(%d, addr[%d])", objectID, desired.Address))
	}
	if desired.Protocol != "" && desired.Protocol != current.Protocol {
		if !IsSafeECoSToken(desired.Protocol) {
			return nil, nil, errors.New("ECoS-Protokoll enthielt unzulässige Zeichen")
		}
		changes = append(changes, ECoSLocomotiveSyncChange{Field: "protocol", Current: current.Protocol, Desired: desired.Protocol})
		commands = append(commands, fmt.Sprintf("set(%d, protocol[%s])", objectID, desired.Protocol))
	}
	return changes, commands, nil
}

func ECoSLocomotiveSyncFields(changes []ECoSLocomotiveSyncChange) []string {
	fields := make([]string, 0, len(changes))
	for _, change := range changes {
		switch change.Field {
		case "address":
			fields = append(fields, "addr")
		default:
			fields = append(fields, change.Field)
		}
	}
	return fields
}

func QuoteECoSString(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, `"`, `\"`)
	return `"` + value + `"`
}

func IsSafeECoSToken(value string) bool {
	if value == "" {
		return false
	}
	for _, char := range value {
		if (char >= 'A' && char <= 'Z') || (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '_' || char == '-' || char == '.' || char == ':' {
			continue
		}
		return false
	}
	return true
}

func FirstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
