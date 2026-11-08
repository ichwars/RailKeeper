package ecosmodel

import (
	"errors"
	"strings"
	"unicode"
)

func NormalizeProtocol(value string, maximumBytes int) (string, error) {
	compact := strings.Map(func(character rune) rune {
		if unicode.IsSpace(character) || strings.ContainsRune("-_/.+", character) {
			return -1
		}
		if !unicode.IsLetter(character) && !unicode.IsDigit(character) {
			return 0
		}
		return unicode.ToUpper(character)
	}, strings.TrimSpace(value))
	if compact == "" && strings.TrimSpace(value) != "" {
		return "", errors.New("invalid protocol")
	}
	if len(compact) > maximumBytes {
		return "", errors.New("protocol is too long")
	}
	for _, character := range compact {
		if !unicode.IsLetter(character) && !unicode.IsDigit(character) {
			return "", errors.New("invalid protocol")
		}
	}
	switch {
	case strings.HasPrefix(compact, "DCC"):
		return "DCC", nil
	case strings.HasPrefix(compact, "MOTOROLA"), strings.HasPrefix(compact, "MM"):
		return "MOTOROLA", nil
	case strings.HasPrefix(compact, "MFX"), strings.HasPrefix(compact, "M4"):
		return "MFX", nil
	case strings.HasPrefix(compact, "SELECTRIX"), strings.HasPrefix(compact, "SX"):
		return "SELECTRIX", nil
	default:
		return compact, nil
	}
}
