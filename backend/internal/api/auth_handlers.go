package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"railkeeper/backend/internal/application"
)

func (a *App) setupStatus(w http.ResponseWriter, r *http.Request) {
	required, err := a.setupService.SetupRequired(r.Context())
	if err != nil {
		a.logger.Error("setup status failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "setup_status_failed", "Could not read setup state.")
		return
	}

	respondJSON(w, http.StatusOK, map[string]bool{"setupRequired": required})
}

func (a *App) createAdmin(w http.ResponseWriter, r *http.Request) {
	allowed, ok := a.allowRequest(w, r, "setup", clientIP(r), 5, 10*time.Minute)
	if !ok {
		return
	}
	if !allowed {
		respondProblem(w, http.StatusTooManyRequests, "rate_limited", "Too many setup attempts. Please try again later.")
		return
	}

	var input application.CreateAdminInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondProblem(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}

	if err := a.setupService.CreateAdmin(r.Context(), input); err != nil {
		switch {
		case errors.Is(err, application.ErrWeakSetup):
			respondProblem(w, http.StatusBadRequest, "weak_setup", "Username, valid email and password with at least 12 characters are required.")
		case errors.Is(err, application.ErrAlreadySetup):
			respondProblem(w, http.StatusConflict, "already_setup", "Setup has already been completed.")
		default:
			a.logger.Error("admin setup failed", "error", err)
			respondProblem(w, http.StatusInternalServerError, "setup_failed", "Could not create admin user.")
		}
		return
	}

	respondJSON(w, http.StatusCreated, map[string]string{"status": "created"})
}

func (a *App) login(w http.ResponseWriter, r *http.Request) {
	allowed, ok := a.allowRequest(w, r, "login", clientIP(r), 10, 5*time.Minute)
	if !ok {
		return
	}
	if !allowed {
		respondProblem(w, http.StatusTooManyRequests, "rate_limited", "Too many login attempts. Please try again later.")
		return
	}

	var input application.LoginInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondProblem(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}

	result, err := a.authService.Login(r.Context(), input)
	if err != nil {
		if errors.Is(err, application.ErrTwoFactorRequired) {
			respondProblem(w, http.StatusUnauthorized, "two_factor_required", "Zwei-Faktor-Code erforderlich.")
			return
		}
		if errors.Is(err, application.ErrInvalidLogin) {
			respondProblem(w, http.StatusUnauthorized, "invalid_login", "Invalid username or password.")
			return
		}
		a.logger.Error("login failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "login_failed", "Could not create session.")
		return
	}

	setCookie(w, "rk_session", result.SessionToken, int(timeUntil(result.ExpiresAt).Seconds()), true, a.cookieSecure)
	setCookie(w, "rk_csrf", result.CSRFToken, int(timeUntil(result.ExpiresAt).Seconds()), false, a.cookieSecure)
	respondJSON(w, http.StatusOK, result.Session)
}

func (a *App) requestPasswordReset(w http.ResponseWriter, r *http.Request) {
	allowed, ok := a.allowRequest(w, r, "password-reset", clientIP(r), 5, 10*time.Minute)
	if !ok {
		return
	}
	if !allowed {
		respondProblem(w, http.StatusTooManyRequests, "rate_limited", "Too many reset attempts. Please try again later.")
		return
	}

	var input application.PasswordResetRequestInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondProblem(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}
	result, err := a.authService.RequestPasswordReset(r.Context(), input)
	if err != nil {
		if errors.Is(err, application.ErrUserValidation) {
			respondProblem(w, http.StatusBadRequest, "invalid_email", "Bitte eine g?ltige E-Mail-Adresse angeben.")
			return
		}
		a.logger.Error("password reset request failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "password_reset_failed", "Passwort-Zurücksetzung konnte nicht vorbereitet werden.")
		return
	}
	if result.ResetToken != "" {
		mailer := a.passwordResetMailer
		mailerBaseURL := a.publicURL
		if a.smtpSettingsService != nil {
			settingsMailer, publicURL, err := a.smtpSettingsService.EffectiveMailer(r.Context())
			if err != nil {
				a.logger.Error("smtp settings invalid", "error", err)
			} else if settingsMailer != nil {
				mailer = settingsMailer
				mailerBaseURL = publicURL
			}
		}
		if mailer != nil {
			resetURL, err := configuredPasswordResetURL(result.ResetToken, mailerBaseURL)
			if err != nil {
				a.logger.Error("password reset email disabled because public URL is invalid", "error", err)
			} else if err := mailer.SendPasswordReset(r.Context(), input.Email, resetURL, result.ExpiresAt); err != nil {
				a.logger.Error("password reset email failed", "error", err)
			}
		} else {
			resetURL := a.passwordResetURL(r, result.ResetToken)
			a.logger.Warn("password reset email disabled; link is available in server log for local recovery only", "reset_url", resetURL)
		}
	}
	result.ResetToken = ""
	result.ResetURL = ""
	respondJSON(w, http.StatusAccepted, result)
}

func (a *App) confirmPasswordReset(w http.ResponseWriter, r *http.Request) {
	allowed, ok := a.allowRequest(w, r, "password-reset-confirm", clientIP(r), 10, 10*time.Minute)
	if !ok {
		return
	}
	if !allowed {
		respondProblem(w, http.StatusTooManyRequests, "rate_limited", "Too many reset attempts. Please try again later.")
		return
	}

	var input application.PasswordResetConfirmInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondProblem(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}
	if err := a.authService.ResetPassword(r.Context(), input); err != nil {
		if errors.Is(err, application.ErrUserValidation) {
			respondProblem(w, http.StatusBadRequest, "invalid_password_reset", "Reset-Link und neues Passwort m?ssen g?ltig sein.")
			return
		}
		if errors.Is(err, application.ErrPasswordResetInvalid) {
			respondProblem(w, http.StatusBadRequest, "invalid_reset_token", "Reset-Link ist ung?ltig oder abgelaufen.")
			return
		}
		a.logger.Error("password reset confirmation failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "password_reset_failed", "Passwort konnte nicht zur?ckgesetzt werden.")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *App) passwordResetURL(r *http.Request, token string) string {
	if resetURL, err := configuredPasswordResetURL(token, a.publicURL); err == nil {
		return resetURL
	}

	scheme := "http"
	if forwarded := strings.TrimSpace(r.Header.Get("X-Forwarded-Proto")); forwarded != "" {
		if strings.EqualFold(strings.Split(forwarded, ",")[0], "https") {
			scheme = "https"
		}
	} else if r.TLS != nil {
		scheme = "https"
	}
	u := url.URL{
		Scheme: scheme,
		Host:   r.Host,
		Path:   "/password-reset",
	}
	query := u.Query()
	query.Set("token", token)
	u.RawQuery = query.Encode()
	return u.String()
}

func configuredPasswordResetURL(token, baseURL string) (string, error) {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		return "", errors.New("public URL is required for password reset email")
	}
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("parse public URL: %w", err)
	}
	if (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" || u.User != nil {
		return "", errors.New("public URL must be an HTTP(S) origin without credentials")
	}
	u.Path = "/password-reset"
	u.RawPath = ""
	u.RawQuery = ""
	u.Fragment = ""
	query := u.Query()
	query.Set("token", token)
	u.RawQuery = query.Encode()
	return u.String(), nil
}

type smtpTestRequest struct {
	Recipient string `json:"recipient"`
}

func (a *App) getSMTPSettings(w http.ResponseWriter, r *http.Request) {
	if a.smtpSettingsService == nil {
		respondJSON(w, http.StatusOK, application.SMTPSettings{TLSMode: "starttls", Port: "587"})
		return
	}
	settings, err := a.smtpSettingsService.Get(r.Context())
	if err != nil {
		a.logger.Error("smtp settings load failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "smtp_settings_failed", "SMTP-Einstellungen konnten nicht geladen werden.")
		return
	}
	respondJSON(w, http.StatusOK, settings)
}

func (a *App) updateSMTPSettings(w http.ResponseWriter, r *http.Request) {
	if a.smtpSettingsService == nil {
		respondProblem(w, http.StatusServiceUnavailable, "smtp_settings_unavailable", "SMTP-Einstellungen sind nicht verfügbar.")
		return
	}
	var input application.SMTPSettingsInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondProblem(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}
	settings, err := a.smtpSettingsService.Update(r.Context(), input)
	if err != nil {
		if errors.Is(err, application.ErrSMTPSettingsValidation) {
			respondProblem(w, http.StatusBadRequest, "smtp_settings_validation", err.Error())
			return
		}
		a.logger.Error("smtp settings update failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "smtp_settings_failed", "SMTP-Einstellungen konnten nicht gespeichert werden.")
		return
	}
	respondJSON(w, http.StatusOK, settings)
}

func (a *App) testSMTPSettings(w http.ResponseWriter, r *http.Request) {
	if a.smtpSettingsService == nil {
		respondProblem(w, http.StatusServiceUnavailable, "smtp_settings_unavailable", "SMTP-Einstellungen sind nicht verfügbar.")
		return
	}
	var input smtpTestRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondProblem(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}
	mailer, _, err := a.smtpSettingsService.EffectiveMailer(r.Context())
	if err != nil {
		if errors.Is(err, application.ErrSMTPSettingsValidation) {
			respondProblem(w, http.StatusBadRequest, "smtp_settings_validation", err.Error())
			return
		}
		a.logger.Error("smtp settings load failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "smtp_settings_failed", "SMTP-Einstellungen konnten nicht geladen werden.")
		return
	}
	if mailer == nil {
		respondProblem(w, http.StatusBadRequest, "smtp_disabled", "SMTP ist nicht aktiviert oder unvollständig konfiguriert.")
		return
	}
	if err := mailer.SendTest(r.Context(), input.Recipient); err != nil {
		a.logger.Error("smtp test email failed", "error", err)
		respondProblem(w, http.StatusBadGateway, "smtp_test_failed", err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "sent"})
}

func (a *App) logout(w http.ResponseWriter, r *http.Request) {
	sessionToken := cookieValue(r, "rk_session")
	if err := a.authService.Logout(r.Context(), sessionToken); err != nil {
		a.logger.Error("logout failed", "error", err)
	}

	clearCookie(w, "rk_session", true, a.cookieSecure)
	clearCookie(w, "rk_csrf", false, a.cookieSecure)
	w.WriteHeader(http.StatusNoContent)
}

func (a *App) session(w http.ResponseWriter, r *http.Request) {
	sessionToken := cookieValue(r, "rk_session")
	session, err := a.authService.CurrentSession(r.Context(), sessionToken)
	if err != nil {
		if errors.Is(err, application.ErrUnauthorized) {
			respondProblem(w, http.StatusUnauthorized, "unauthorized", "Not logged in.")
			return
		}
		a.logger.Error("session lookup failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "session_failed", "Could not read current session.")
		return
	}

	respondJSON(w, http.StatusOK, session)
}

func (a *App) changePassword(w http.ResponseWriter, r *http.Request) {
	var input application.ChangePasswordInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondProblem(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}

	if err := a.authService.ChangeOwnPassword(r.Context(), actorUserID(r), cookieValue(r, "rk_session"), input); err != nil {
		switch {
		case errors.Is(err, application.ErrUserValidation):
			respondProblem(w, http.StatusBadRequest, "weak_password", "Das neue Passwort muss mindestens 12 Zeichen lang sein.")
		case errors.Is(err, application.ErrInvalidLogin):
			respondProblem(w, http.StatusUnauthorized, "invalid_password", "Das aktuelle Passwort ist nicht korrekt.")
		case errors.Is(err, application.ErrUserNotFound), errors.Is(err, application.ErrUnauthorized):
			respondProblem(w, http.StatusUnauthorized, "unauthorized", "Not logged in.")
		default:
			a.logger.Error("password change failed", "error", err)
			respondProblem(w, http.StatusInternalServerError, "password_change_failed", "Passwort konnte nicht geändert werden.")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
