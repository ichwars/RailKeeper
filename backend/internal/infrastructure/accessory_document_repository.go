package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"railkeeper/backend/internal/application"
)

func (r *AccessoryRepository) ListDocuments(
	ctx context.Context,
	productID string,
) ([]application.AccessoryDocument, error) {
	rows, err := r.db.QueryContext(ctx, accessoryDocumentSelect+`
WHERE product_id=? ORDER BY is_primary DESC, created_at DESC, id`, productID)
	if err != nil {
		return nil, fmt.Errorf("list accessory documents: %w", err)
	}
	defer func() { _ = rows.Close() }()
	documents := []application.AccessoryDocument{}
	for rows.Next() {
		document, err := scanAccessoryDocument(rows)
		if err != nil {
			return nil, fmt.Errorf("scan accessory document: %w", err)
		}
		documents = append(documents, *document)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate accessory documents: %w", err)
	}
	return documents, nil
}

func (r *AccessoryRepository) GetDocument(
	ctx context.Context,
	id string,
) (*application.AccessoryDocument, error) {
	return getAccessoryDocumentWith(ctx, r.db, id)
}

func (r *AccessoryRepository) CreateDocument(
	ctx context.Context,
	input application.CreateAccessoryDocumentInput,
	actor string,
) (*application.AccessoryDocument, error) {
	documentID := randomID()
	now := timestamp()
	err := r.withTx(ctx, func(tx *sql.Tx) error {
		if err := requireAccessoryDocumentReferences(ctx, tx, input.ProductID, input.FileBlobID); err != nil {
			return err
		}
		if input.IsPrimary {
			if err := clearAccessoryPrimaryDocument(ctx, tx, input.ProductID, now); err != nil {
				return err
			}
		}
		if _, err := tx.ExecContext(ctx, `
INSERT INTO accessory_documents(
  id, product_id, file_blob_id, file_name, original_name, description, category, mime_type,
  size_bytes, is_primary, created_by, created_at, updated_at
) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, documentID, input.ProductID, input.FileBlobID,
			input.FileName, input.OriginalName, input.Description, input.Category, input.MimeType,
			input.SizeBytes, boolToInt(input.IsPrimary), actor, now, now); err != nil {
			if isSQLiteConstraint(err) {
				return application.ErrAccessoryConflict
			}
			return fmt.Errorf("insert accessory document: %w", err)
		}
		return writeAccessoryAudit(ctx, tx, "AccessoryDocumentCreated", "accessory_document",
			documentID, actor, now, "{}")
	})
	if err != nil {
		return nil, err
	}
	return r.GetDocument(ctx, documentID)
}

func (r *AccessoryRepository) UpdateDocument(
	ctx context.Context,
	id string,
	input application.UpdateAccessoryDocumentInput,
	actor string,
) (*application.AccessoryDocument, error) {
	now := timestamp()
	err := r.withTx(ctx, func(tx *sql.Tx) error {
		document, err := getAccessoryDocumentWith(ctx, tx, id)
		if err != nil {
			return err
		}
		if input.IsPrimary {
			if err := clearAccessoryPrimaryDocument(ctx, tx, document.ProductID, now); err != nil {
				return err
			}
		}
		result, err := tx.ExecContext(ctx, `
UPDATE accessory_documents
SET description=?, category=?, is_primary=?, updated_at=? WHERE id=?`, input.Description,
			input.Category, boolToInt(input.IsPrimary), now, id)
		if err != nil {
			if isSQLiteConstraint(err) {
				return application.ErrAccessoryConflict
			}
			return fmt.Errorf("update accessory document: %w", err)
		}
		if err := requireAccessoryUpdated(result); err != nil {
			return err
		}
		return writeAccessoryAudit(ctx, tx, "AccessoryDocumentUpdated", "accessory_document",
			id, actor, now, "{}")
	})
	if err != nil {
		return nil, err
	}
	return r.GetDocument(ctx, id)
}

func (r *AccessoryRepository) DeleteDocument(
	ctx context.Context,
	id string,
	actor string,
) (string, error) {
	var blobID string
	now := timestamp()
	err := r.withTx(ctx, func(tx *sql.Tx) error {
		if err := tx.QueryRowContext(ctx, `SELECT file_blob_id FROM accessory_documents WHERE id=?`, id).
			Scan(&blobID); errors.Is(err, sql.ErrNoRows) {
			return application.ErrAccessoryNotFound
		} else if err != nil {
			return fmt.Errorf("read accessory document blob: %w", err)
		}
		result, err := tx.ExecContext(ctx, `DELETE FROM accessory_documents WHERE id=?`, id)
		if err != nil {
			return fmt.Errorf("delete accessory document: %w", err)
		}
		if err := requireAccessoryUpdated(result); err != nil {
			return err
		}
		return writeAccessoryAudit(ctx, tx, "AccessoryDocumentDeleted", "accessory_document",
			id, actor, now, "{}")
	})
	return blobID, err
}

func requireAccessoryDocumentReferences(
	ctx context.Context,
	tx *sql.Tx,
	productID, blobID string,
) error {
	for _, reference := range []struct {
		query string
		id    string
	}{
		{`SELECT COUNT(*) FROM accessory_products WHERE id=?`, productID},
		{`SELECT COUNT(*) FROM file_blobs WHERE id=?`, blobID},
	} {
		exists, err := accessoryRecordExists(ctx, tx, reference.query, reference.id)
		if err != nil {
			return err
		}
		if !exists {
			return application.ErrAccessoryNotFound
		}
	}
	return nil
}

func clearAccessoryPrimaryDocument(ctx context.Context, tx *sql.Tx, productID, now string) error {
	if _, err := tx.ExecContext(ctx, `
UPDATE accessory_documents SET is_primary=0, updated_at=?
WHERE product_id=? AND category='image' AND is_primary=1`, now, productID); err != nil {
		return fmt.Errorf("clear accessory primary document: %w", err)
	}
	return nil
}

const accessoryDocumentSelect = `SELECT
  id, product_id, file_blob_id, file_name, original_name, description, category, mime_type,
  size_bytes, is_primary, created_by, created_at, updated_at
FROM accessory_documents`

func scanAccessoryDocument(scanner rowScanner) (*application.AccessoryDocument, error) {
	document := &application.AccessoryDocument{}
	var primary int
	err := scanner.Scan(&document.ID, &document.ProductID, &document.FileBlobID, &document.FileName,
		&document.OriginalName, &document.Description, &document.Category, &document.MimeType,
		&document.SizeBytes, &primary, &document.CreatedBy, &document.CreatedAt, &document.UpdatedAt)
	document.IsPrimary = primary != 0
	return document, err
}

func getAccessoryDocumentWith(
	ctx context.Context,
	queryer accessoryQueryer,
	id string,
) (*application.AccessoryDocument, error) {
	document, err := scanAccessoryDocument(queryer.QueryRowContext(ctx, accessoryDocumentSelect+` WHERE id=?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, application.ErrAccessoryNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get accessory document: %w", err)
	}
	return document, nil
}

var _ application.AccessoryDocumentRepository = (*AccessoryRepository)(nil)
