package numbering

import "context"

type Service struct{ schemes Schemes }

func New(schemes Schemes) *Service                            { return &Service{schemes: schemes} }
func (s *Service) List(ctx context.Context) ([]Scheme, error) { return s.schemes.List(ctx) }
func (s *Service) Create(ctx context.Context, input CreateInput) (*Scheme, error) {
	category := CleanCategory(input.Category)
	scheme := CleanSchemeInput(input.InventoryNumberSchemeInput)
	if err := validateScheme(category, scheme); err != nil {
		return nil, err
	}
	return s.schemes.Create(ctx, category, scheme)
}
func (s *Service) Update(ctx context.Context, category string, input SchemeInput) (*Scheme, error) {
	category = CleanCategory(category)
	input = CleanSchemeInput(input)
	if err := validateScheme(category, input); err != nil {
		return nil, err
	}
	return s.schemes.Update(ctx, category, input)
}
func validateScheme(category string, input SchemeInput) error {
	if category == "" || input.Prefix == "" || input.NextNumber < 1 || input.Padding < 1 || input.Padding > 12 {
		return ErrValidation
	}
	return nil
}
