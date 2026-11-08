package vehicle

import "context"

type ExpiryRepository interface {
	ResetDirectAssignments(context.Context, string) error
	ResetLegacyMatches(context.Context, string) error
}

func ResetExpiredExhibitionFlags(ctx context.Context, repository ExpiryRepository, now string) error {
	if err := repository.ResetDirectAssignments(ctx, now); err != nil {
		return err
	}
	return repository.ResetLegacyMatches(ctx, now)
}
