package backup

import (
	"context"
	"errors"
)

func (s *Service) Import(ctx context.Context, doc *BackupDocument) (*BackupImportResult, error) {
	if !s.configured() {
		return nil, errors.New("backup service is not configured")
	}
	if doc == nil {
		return nil, ErrBackupInvalid
	}
	if err := ValidateBackupFiles(doc.Files); err != nil {
		return nil, err
	}
	validation, err := s.Validate(ctx, doc)
	if err != nil {
		return nil, err
	}
	if !validation.Compatible {
		return nil, ErrBackupInvalid
	}

	return s.store.Restore(ctx, doc)
}
