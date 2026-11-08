package management

import (
	"context"
	"fmt"
	"strings"
	"time"

	"railkeeper/backend/internal/authentication"
	"railkeeper/backend/internal/credentials"
)

const passwordResetTTL = 30 * time.Minute

func (s *Service) ChangeOwnPassword(ctx context.Context, id, sessionToken string, input ChangePasswordInput) error {
	if len(input.NewPassword) < 12 {
		return ErrUserValidation
	}
	account, err := s.store.ReadCredential(ctx, id)
	if err != nil {
		return err
	}
	if !credentials.Verify(input.CurrentPassword, account.PasswordHash) {
		_ = s.store.RecordAudit(ctx, id, "PasswordChangeFailed", "user", id, "{}")
		return authentication.ErrInvalidLogin
	}
	hash, err := credentials.Hash(input.NewPassword)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	return s.store.ChangePassword(ctx, PasswordChange{UserID: id, SessionTokenHash: credentials.HashToken(sessionToken), ExpectedHash: account.PasswordHash, NextHash: hash})
}
func (s *Service) RequestPasswordReset(ctx context.Context, input PasswordResetRequestInput) (*PasswordResetRequestResult, error) {
	email := strings.TrimSpace(strings.ToLower(input.Email))
	if !ValidEmail(email) {
		return nil, ErrUserValidation
	}
	id, err := s.store.FindResetUser(ctx, email)
	if err != nil {
		return nil, err
	}
	record := ResetRequest{UserID: id, Email: email}
	token := ""
	if id != "" {
		token = credentials.RandomToken()
		record.TokenHash = credentials.HashToken(token)
		record.ExpiresAt = s.now().UTC().Add(passwordResetTTL).Format(time.RFC3339)
	}
	if err := s.store.CreateResetRequest(ctx, record); err != nil {
		return nil, err
	}
	return &PasswordResetRequestResult{Status: "ready", Message: "Wenn die E-Mail-Adresse bekannt ist, wurde ein Passwort-Reset-Link versendet.", ResetToken: token, ExpiresAt: record.ExpiresAt}, nil
}
func (s *Service) ResetPassword(ctx context.Context, input PasswordResetConfirmInput) error {
	token := strings.TrimSpace(input.Token)
	if token == "" || len(input.NewPassword) < 12 {
		return ErrUserValidation
	}
	proof, err := s.store.ReadResetRequest(ctx, credentials.HashToken(token))
	if err != nil {
		return err
	}
	if !ValidResetProof(proof, s.now().UTC()) {
		return ErrPasswordResetInvalid
	}
	hash, err := credentials.Hash(input.NewPassword)
	if err != nil {
		return fmt.Errorf("hash password reset password: %w", err)
	}
	return s.store.CommitPasswordReset(ctx, ResetCommit{RequestID: proof.ID, UserID: proof.UserID, TokenHash: credentials.HashToken(token), NextHash: hash})
}
func ValidResetProof(proof ResetProof, now time.Time) bool {
	if proof.UserID == "" || proof.UsedAt != "" || proof.HandledAt != "" || proof.ExpiresAt == "" {
		return false
	}
	expires, err := time.Parse(time.RFC3339, proof.ExpiresAt)
	return err == nil && now.Before(expires)
}
