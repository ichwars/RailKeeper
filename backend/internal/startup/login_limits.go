package startup

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"railkeeper/backend/internal/authentication"
)

func ResolveLoginWorkLimiter(lookup func(string) (string, bool)) (*authentication.WorkLimiter, error) {
	limits := authentication.DefaultWorkLimits()
	waitMS := int(limits.Wait / time.Millisecond)
	settings := []struct {
		name   string
		target *int
	}{
		{"RAILKEEPER_LOGIN_PARALLEL", &limits.Parallel},
		{"RAILKEEPER_LOGIN_QUEUE", &limits.Queue},
		{"RAILKEEPER_LOGIN_WAIT_MS", &waitMS},
		{"RAILKEEPER_LOGIN_MEMORY_MB", &limits.MemoryMiB},
	}
	for _, setting := range settings {
		value, _ := lookup(setting.name)
		if value = strings.TrimSpace(value); value == "" {
			continue
		}
		parsed, err := strconv.Atoi(value)
		if err != nil {
			return nil, fmt.Errorf("%s must be an integer", setting.name)
		}
		*setting.target = parsed
	}
	if waitMS < 0 || waitMS > 2000 {
		return nil, fmt.Errorf("RAILKEEPER_LOGIN_WAIT_MS must be between 0 and 2000")
	}
	limits.Wait = time.Duration(waitMS) * time.Millisecond
	return authentication.NewWorkLimiter(limits)
}
