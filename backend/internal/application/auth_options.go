package application

import (
	"time"

	"railkeeper/backend/internal/authentication"
)

type AuthOption func(*AuthService)

func WithLoginWorkLimiter(limiter *authentication.WorkLimiter) AuthOption {
	return func(service *AuthService) { service.loginWork = limiter }
}

func WithLoginStore(store authentication.Store) AuthOption {
	return func(service *AuthService) { service.loginStore = store }
}

// WithAuthClock supplies a trusted clock, making expiry and TOTP boundary checks deterministic.
func WithAuthClock(clock func() time.Time) AuthOption {
	return func(service *AuthService) {
		if clock != nil {
			service.now = clock
		}
	}
}
