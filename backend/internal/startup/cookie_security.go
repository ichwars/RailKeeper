package startup

import (
	"fmt"
	"net/url"
	"strings"
)

// ResolveCookieSecurity uses configured public URLs, never forwarded request headers.
func ResolveCookieSecurity(mode, publicURL string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "true":
		return true, nil
	case "false":
		return false, nil
	case "", "auto":
		if strings.TrimSpace(publicURL) == "" {
			return false, nil
		}
		parsed, err := url.Parse(strings.TrimSpace(publicURL))
		if err != nil || parsed.Hostname() == "" || parsed.User != nil || parsed.Fragment != "" ||
			(parsed.Scheme != "http" && parsed.Scheme != "https") {
			return false, fmt.Errorf("cookie auto mode requires an absolute HTTP(S) public URL")
		}
		return parsed.Scheme == "https", nil
	default:
		return false, fmt.Errorf("RAILKEEPER_COOKIE_SECURE must be auto, true or false")
	}
}
