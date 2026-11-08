package numbering

import (
	"context"
	"errors"
	"fmt"
)

func Reserve(
	ctx context.Context,
	counter Counter,
	category string,
	fallbackCategory string,
	ensureAvailable Availability,
) (string, error) {
	category = CleanCategory(category)
	fallbackCategory = CleanCategory(fallbackCategory)
	scheme, err := activeScheme(ctx, counter, category, fallbackCategory)
	if err != nil {
		return "", err
	}

	next := scheme.NextNumber
	for attempts := 0; attempts < 500; attempts++ {
		candidate := Format(scheme.Prefix, next, scheme.Padding)
		if err := ensureAvailable(candidate); err == nil {

			if err := counter.Advance(ctx, scheme.Category, next+1); err != nil {
				return "", err
			}

			return candidate, nil
		} else if !errors.Is(err, ErrConflict) {
			return "", err
		}
		next++
	}

	return "", fmt.Errorf("next inventory number: exhausted attempts for %s", scheme.Category)
}
func activeScheme(
	ctx context.Context,
	counter Counter,
	category string,
	fallbackCategory string,
) (*Scheme, error) {
	if category != "" {
		if scheme, err := counter.Active(ctx, category); err == nil {
			return scheme, nil
		} else if !errors.Is(err, ErrNotFound) {
			return nil, err
		}
	}
	if fallbackCategory != "" && fallbackCategory != category {
		if scheme, err := counter.Active(ctx, fallbackCategory); err == nil {
			return scheme, nil
		} else if !errors.Is(err, ErrNotFound) {
			return nil, err
		}
	}
	if fallbackCategory != "Fahrzeug" && category != "Fahrzeug" && fallbackCategory != "" {
		return counter.Active(ctx, "Fahrzeug")
	}
	return nil, ErrNotFound
}
