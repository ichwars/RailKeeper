package application

import (
	"context"
	"strings"
)

func (s *LayoutService) ListVariants(ctx context.Context, unitID string) ([]PlanVariant, error) {
	return s.repository.ListVariants(ctx, strings.TrimSpace(unitID))
}

func (s *LayoutService) CreateVariant(ctx context.Context, unitID string, input CreatePlanVariantInput, actor string) (*PlanVariant, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.Description = strings.TrimSpace(input.Description)
	if strings.TrimSpace(unitID) == "" || input.Name == "" {
		return nil, ErrLayoutValidation
	}
	return s.repository.CreateVariant(ctx, strings.TrimSpace(unitID), input, actor)
}

func (s *LayoutService) CreateDraft(ctx context.Context, variantID string, input CreatePlanRevisionInput, actor string) (*PlanRevision, error) {
	input.BaseRevisionID = strings.TrimSpace(input.BaseRevisionID)
	if strings.TrimSpace(variantID) == "" {
		return nil, ErrLayoutValidation
	}
	return s.repository.CreateDraft(ctx, strings.TrimSpace(variantID), input, actor)
}

func (s *LayoutService) SubmitRevision(ctx context.Context, revisionID string, expectedVersion int, actor string) (*PlanRevision, error) {
	if strings.TrimSpace(revisionID) == "" || expectedVersion < 1 {
		return nil, ErrLayoutValidation
	}
	return s.repository.SubmitRevision(ctx, strings.TrimSpace(revisionID), expectedVersion, actor)
}

func (s *LayoutService) PublishRevision(ctx context.Context, revisionID string, expectedVersion int, actor string) (*PlanRevision, error) {
	if strings.TrimSpace(revisionID) == "" || expectedVersion < 1 {
		return nil, ErrLayoutValidation
	}
	return s.repository.PublishRevision(ctx, strings.TrimSpace(revisionID), expectedVersion, actor)
}
