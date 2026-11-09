package startup

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

func writeMigrationReceipt(root string, receipt *MigrationReceipt) error {
	data, err := json.MarshalIndent(receipt, "", "  ")
	if err != nil {
		return fmt.Errorf("encode legacy migration receipt: %w", err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(filepath.Join(root, legacyMigrationReceiptName), data, 0o600); err != nil {
		return fmt.Errorf("write legacy migration receipt: %w", err)
	}
	return nil
}

func readMigrationReceipt(root string) (*MigrationReceipt, error) {
	path := filepath.Join(root, legacyMigrationReceiptName)
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read legacy migration receipt: %w", err)
	}
	var receipt MigrationReceipt
	if err := json.Unmarshal(data, &receipt); err != nil {
		return nil, fmt.Errorf("decode legacy migration receipt: %w", err)
	}
	if filepath.Clean(receipt.TargetPath) != filepath.Clean(root) || receipt.SourcePath == "" {
		return nil, errors.New("legacy migration receipt paths are invalid")
	}
	return &receipt, nil
}

func AcknowledgeMigrationReceipt(ctx context.Context, root string) (*MigrationReceipt, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	root = filepath.Clean(root)
	if err := validateDataRoot(root, "safe"); err != nil {
		return nil, err
	}
	receipt, err := readMigrationReceipt(root)
	if err != nil {
		return nil, err
	}
	if receipt == nil {
		return nil, errors.New("legacy migration receipt does not exist")
	}
	if receipt.Acknowledged {
		return receipt, nil
	}

	updated := *receipt
	updated.Acknowledged = true
	data, err := json.MarshalIndent(&updated, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("encode acknowledged migration receipt: %w", err)
	}
	data = append(data, '\n')
	targetPath := filepath.Join(root, legacyMigrationReceiptName)
	temporaryPath := filepath.Join(root, legacyMigrationReceiptName+".tmp-"+randomLegacySuffix())
	if !pathWithin(root, temporaryPath) {
		return nil, errors.New("migration receipt temporary path escaped data directory")
	}
	temporary, err := os.OpenFile(temporaryPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return nil, fmt.Errorf("create migration receipt update: %w", err)
	}
	removeTemporary := true
	defer func() {
		if removeTemporary {
			_ = os.Remove(temporaryPath)
		}
	}()
	if _, err = temporary.Write(data); err != nil {
		_ = temporary.Close()
		return nil, fmt.Errorf("write migration receipt update: %w", err)
	}
	if err = temporary.Sync(); err != nil {
		_ = temporary.Close()
		return nil, fmt.Errorf("sync migration receipt update: %w", err)
	}
	if err = temporary.Close(); err != nil {
		return nil, fmt.Errorf("close migration receipt update: %w", err)
	}
	if err = ctx.Err(); err != nil {
		return nil, err
	}
	if err = replaceFileAtomically(temporaryPath, targetPath); err != nil {
		return nil, fmt.Errorf("activate migration receipt update: %w", err)
	}
	removeTemporary = false
	return &updated, nil
}
