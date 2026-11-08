package management

import (
	"context"
	"time"
)

type Service struct {
	store Store
	now   func() time.Time
}

func NewService(store Store, now func() time.Time) *Service {
	if now == nil {
		now = time.Now
	}
	return &Service{store: store, now: now}
}
func (s *Service) ListUsers(ctx context.Context) ([]UserView, error) { return s.store.ListUsers(ctx) }
func (s *Service) ListRoles(ctx context.Context) ([]RoleView, error) { return s.store.ListRoles(ctx) }
func (s *Service) GetUser(ctx context.Context, id string) (*UserView, error) {
	return s.store.GetUser(ctx, id)
}
func (s *Service) DeleteUser(ctx context.Context, actor, id string) error {
	return s.store.DeleteUser(ctx, actor, id)
}
func (s *Service) ListAuditLog(ctx context.Context, limit int) ([]AuditLogEntry, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	return s.store.ListAuditLog(ctx, limit)
}
func (s *Service) ListSessions(ctx context.Context, limit int) ([]SessionRecord, error) {
	if limit <= 0 || limit > 200 {
		limit = 200
	}
	return s.store.ListSessions(ctx, limit)
}
func (s *Service) RecordAudit(ctx context.Context, actor, action, targetType, targetID, details string) error {
	return s.store.RecordAudit(ctx, actor, action, targetType, targetID, details)
}
func (s *Service) RevokeSession(ctx context.Context, actor, id string) error {
	return s.store.RevokeSession(ctx, actor, id)
}
