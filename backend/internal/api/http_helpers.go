package api

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"railkeeper/backend/internal/application"
)

type rateLimitStore interface {
	Allow(ctx context.Context, scope, key string, limit int, window time.Duration) (bool, error)
}

type rateLimiter struct {
	mu       sync.Mutex
	attempts map[string][]time.Time
}

func newRateLimiter() *rateLimiter {
	return &rateLimiter{attempts: map[string][]time.Time{}}
}

func (r *rateLimiter) Allow(ctx context.Context, scope, key string, limit int, window time.Duration) (bool, error) {
	if r == nil {
		return true, nil
	}
	now := time.Now()
	cutoff := now.Add(-window)
	r.mu.Lock()
	defer r.mu.Unlock()

	compoundKey := scope + ":" + key
	current := r.attempts[compoundKey]
	filtered := current[:0]
	for _, attempt := range current {
		if attempt.After(cutoff) {
			filtered = append(filtered, attempt)
		}
	}
	if len(filtered) >= limit {
		r.attempts[compoundKey] = filtered
		return false, nil
	}
	r.attempts[compoundKey] = append(filtered, now)
	return true, nil
}

func (a *App) allowRequest(w http.ResponseWriter, r *http.Request, scope, key string, limit int, window time.Duration) (bool, bool) {
	allowed, err := a.rateLimits.Allow(r.Context(), scope, key, limit, window)
	if err != nil {
		a.logger.Error("rate limit check failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "rate_limit_failed", "Could not verify request rate limit.")
		return false, false
	}
	return allowed, true
}

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil && host != "" {
		return host
	}
	if r.RemoteAddr == "" {
		return "unknown"
	}
	return r.RemoteAddr
}

func respondJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func respondProblem(w http.ResponseWriter, status int, code, message string) {
	respondJSON(w, status, map[string]string{
		"error":   code,
		"message": message,
	})
}

func vehicleECoSMappingForSync(vehicle *application.Vehicle, objectID int) *application.VehicleExternalMap {
	if vehicle == nil {
		return nil
	}
	for index := range vehicle.ExternalMappings {
		mapping := &vehicle.ExternalMappings[index]
		if mapping.Provider != "ecos" {
			continue
		}
		if objectID <= 0 || mapping.ExternalID == strconv.Itoa(objectID) {
			return mapping
		}
	}
	return nil
}

func parsePositiveIntText(value string) int {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return 0
	}
	return parsed
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "same-origin")
		w.Header().Set("Cross-Origin-Opener-Policy", "same-origin")
		w.Header().Set("X-Frame-Options", "SAMEORIGIN")
		w.Header().Set("Permissions-Policy", "camera=(self), microphone=(), geolocation=()")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; connect-src 'self'; img-src 'self' data: blob: http: https:; style-src 'self' 'unsafe-inline'; script-src 'self'; object-src 'none'; frame-ancestors 'self'; base-uri 'self'")
		next.ServeHTTP(w, r)
	})
}

func (a *App) csrf(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}
		if r.URL.Path == "/api/v1/setup/admin" ||
			r.URL.Path == "/api/v1/auth/login" ||
			r.URL.Path == "/api/v1/auth/password-reset" ||
			r.URL.Path == "/api/v1/auth/password-reset/confirm" {
			next.ServeHTTP(w, r)
			return
		}
		if !strings.HasPrefix(r.URL.Path, "/api/") {
			next.ServeHTTP(w, r)
			return
		}

		if err := a.authService.ValidateCSRF(r.Context(), cookieValue(r, "rk_session"), r.Header.Get("X-CSRF-Token")); err != nil {
			respondProblem(w, http.StatusForbidden, "csrf_required", "CSRF token is missing or invalid.")
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (a *App) require(role string, next http.HandlerFunc) http.HandlerFunc {
	return a.requireAny([]string{role}, next)
}

func (a *App) requireMasterDataRead(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		roles := []string{"Viewer"}
		if r.PathValue("type") == "symbols" {
			roles = append(roles, "Messe")
		}
		a.requireAny(roles, next)(w, r)
	}
}

func (a *App) requireAny(roles []string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := a.authService.RequireAnyRole(r.Context(), cookieValue(r, "rk_session"), roles...)
		if err != nil {
			if errors.Is(err, application.ErrUnauthorized) {
				respondProblem(w, http.StatusUnauthorized, "unauthorized", "Not logged in.")
				return
			}
			if errors.Is(err, application.ErrForbidden) {
				respondProblem(w, http.StatusForbidden, "forbidden", "Insufficient role.")
				return
			}
			a.logger.Error("role check failed", "error", err)
			respondProblem(w, http.StatusInternalServerError, "role_check_failed", "Could not verify permissions.")
			return
		}

		next.ServeHTTP(w, withActorUserID(r, userID))
	}
}

func setCookie(w http.ResponseWriter, name, value string, maxAge int, httpOnly, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: httpOnly,
		SameSite: http.SameSiteLaxMode,
		Secure:   secure,
	})
}

func clearCookie(w http.ResponseWriter, name string, httpOnly, secure bool) {
	setCookie(w, name, "", -1, httpOnly, secure)
}

func cookieValue(r *http.Request, name string) string {
	cookie, err := r.Cookie(name)
	if err != nil {
		return ""
	}
	return cookie.Value
}

func timeUntil(t time.Time) time.Duration {
	duration := time.Until(t)
	if duration < time.Second {
		return time.Second
	}
	return duration
}

type actorUserIDKey struct{}

func withActorUserID(r *http.Request, userID string) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), actorUserIDKey{}, userID))
}

func actorUserID(r *http.Request) string {
	value, _ := r.Context().Value(actorUserIDKey{}).(string)
	return value
}

func staticHandler(staticDir string) http.Handler {
	fileServer := http.FileServer(http.Dir(staticDir))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			respondJSON(w, http.StatusNotFound, map[string]string{
				"error":   "not_found",
				"message": "API route not found",
			})
			return
		}

		path := filepath.Join(staticDir, filepath.Clean(r.URL.Path))
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			if strings.HasPrefix(r.URL.Path, "/assets/") {
				w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
			} else {
				w.Header().Set("Cache-Control", "no-cache")
			}
			fileServer.ServeHTTP(w, r)
			return
		}

		w.Header().Set("Cache-Control", "no-store")
		http.ServeFile(w, r, filepath.Join(staticDir, "index.html"))
	})
}
