package management

import "strings"

func HasRole(roles []string, role string) bool {
	if role == "Viewer" {
		return HasExactRole(roles, "Viewer") || HasExactRole(roles, "Editor") ||
			HasExactRole(roles, "Admin")
	}
	return HasExactRole(roles, role)
}

func HasExactRole(roles []string, role string) bool {
	for _, current := range roles {
		if current == role {
			return true
		}
	}
	return false
}

func CleanRoleNames(values []string) []string {
	seen := map[string]struct{}{}
	out := []string{}
	for _, value := range values {
		role := strings.TrimSpace(value)
		if role == "" {
			continue
		}
		if _, ok := seen[role]; ok {
			continue
		}
		seen[role] = struct{}{}
		out = append(out, role)
	}
	return out
}

func ValidEmail(value string) bool {
	value = strings.TrimSpace(value)
	at := strings.Index(value, "@")
	return at > 0 && at < len(value)-3 && strings.Contains(value[at+1:], ".")
}

func RequireAnotherAdmin(count int) error {
	if count == 0 {
		return ErrLastAdmin
	}
	return nil
}
