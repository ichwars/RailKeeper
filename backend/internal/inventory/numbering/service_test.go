package numbering

import (
	"context"
	"errors"
	"testing"
)

type schemeStore struct {
	category string
	input    SchemeInput
	Schemes
}

func (s *schemeStore) Create(_ context.Context, category string, input SchemeInput) (*Scheme, error) {
	s.category, s.input = category, input
	return &Scheme{Category: category, Prefix: input.Prefix}, nil
}

func TestCreateSchemeKeepsLegacyDefaultsAndCategoryIdentity(t *testing.T) {
	store := &schemeStore{}
	value, err := New(store).Create(t.Context(), CreateInput{Category: " Custom ", InventoryNumberSchemeInput: SchemeInput{Prefix: " old prefix "}})
	if err != nil || value.Prefix != "OLD-PREFIX" || store.category != "Custom" ||
		store.input.NextNumber != 1 || store.input.Padding != 6 || store.input.Active {
		t.Fatal("normalization changed legacy scheme defaults or category identity", value, err, store)
	}
}

func TestInvalidSchemesAreRejectedBeforeStorage(t *testing.T) {
	for _, input := range []SchemeInput{
		{}, {Prefix: "RK", NextNumber: -1}, {Prefix: "RK", Padding: -1}, {Prefix: "RK", Padding: 13},
	} {
		service := New(nil)
		if _, err := service.Create(t.Context(), CreateInput{Category: "Custom", InventoryNumberSchemeInput: input}); !errors.Is(err, ErrValidation) {
			t.Fatal("invalid create reached storage", err)
		}
		if _, err := service.Update(t.Context(), "Custom", input); !errors.Is(err, ErrValidation) {
			t.Fatal("invalid update reached storage", err)
		}
	}
}
