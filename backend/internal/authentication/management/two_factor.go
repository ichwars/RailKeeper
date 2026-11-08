package management

import (
	"context"
	"net/url"

	"railkeeper/backend/internal/authentication"
	"railkeeper/backend/internal/credentials"
)

func status(state TwoFactorState) *TwoFactorStatus {
	return &TwoFactorStatus{Enabled: state.Enabled, Prepared: state.Secret != "", EnabledAt: state.EnabledAt, Username: state.Username}
}
func (s *Service) TwoFactorStatus(ctx context.Context, id string) (*TwoFactorStatus, error) {
	state, err := s.store.ReadTwoFactor(ctx, id)
	if err != nil {
		return nil, err
	}
	return status(state), nil
}
func (s *Service) PrepareTwoFactor(ctx context.Context, id string) (*TwoFactorStatus, error) {
	state, err := s.store.PrepareTwoFactor(ctx, id, credentials.GenerateTOTPSecret)
	if err != nil {
		return nil, err
	}
	result := status(state)
	if !state.Enabled {
		result.Secret = state.Secret
		result.OtpauthURL = totpProvisioningURI("RailKeeper", state.Username, state.Secret)
	}
	return result, nil
}
func (s *Service) EnableTwoFactor(ctx context.Context, id string, input TwoFactorEnableInput) (*TwoFactorStatus, error) {
	state, err := s.store.ReadTwoFactor(ctx, id)
	if err != nil {
		return nil, err
	}
	step, valid := credentials.MatchTOTP(state.Secret, input.Code, s.now().UTC())
	if !valid {
		_ = s.store.RecordAudit(ctx, id, "TwoFactorEnableFailed", "user", id, "{}")
		return nil, authentication.ErrInvalidLogin
	}
	if err := s.store.CommitTwoFactorEnable(ctx, id, state.Secret, step); err != nil {
		return nil, err
	}
	return s.TwoFactorStatus(ctx, id)
}
func (s *Service) DisableTwoFactor(ctx context.Context, id, sessionToken string, input TwoFactorDisableInput) error {
	account, err := s.store.ReadCredential(ctx, id)
	if err != nil {
		return err
	}
	if !credentials.Verify(input.CurrentPassword, account.PasswordHash) {
		_ = s.store.RecordAudit(ctx, id, "TwoFactorDisableFailed", "user", id, "{}")
		return authentication.ErrInvalidLogin
	}
	step, valid := credentials.MatchTOTP(account.TwoFactorSecret, input.Code, s.now().UTC())
	if account.TwoFactorEnabled && !valid {
		_ = s.store.RecordAudit(ctx, id, "TwoFactorDisableFailed", "user", id, "{}")
		return authentication.ErrInvalidLogin
	}
	return s.store.CommitTwoFactorDisable(ctx, id, credentials.HashToken(sessionToken), account.PasswordHash, account.TwoFactorSecret, account.TwoFactorEnabled, step)
}
func totpProvisioningURI(issuer, account, secret string) string {
	label := issuer + ":" + account
	query := url.Values{}
	query.Set("secret", secret)
	query.Set("issuer", issuer)
	query.Set("algorithm", "SHA1")
	query.Set("digits", "6")
	query.Set("period", "30")
	return "otpauth://totp/" + url.PathEscape(label) + "?" + query.Encode()
}
