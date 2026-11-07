package application

import "time"

type AuthOption func(*AuthService)

// WithAuthClock supplies a trusted clock, making expiry and TOTP boundary checks deterministic.
func WithAuthClock(clock func() time.Time) AuthOption {
	return func(service *AuthService) {
		if clock != nil {
			service.now = clock
		}
	}
}
