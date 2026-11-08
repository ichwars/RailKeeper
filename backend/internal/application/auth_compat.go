package application

import (
	"strings"

	"railkeeper/backend/internal/credentials"
)

func isValidEmail(value string) bool {
	value = strings.TrimSpace(value)
	at := strings.Index(value, "@")
	return at > 0 && at < len(value)-3 && strings.Contains(value[at+1:], ".")
}

func hashToken(token string) string {
	return credentials.HashToken(token)
}
