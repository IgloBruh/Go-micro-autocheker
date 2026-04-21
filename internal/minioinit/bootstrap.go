package minioinit

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"autocheck-microservices/internal/storage"
)

func BootstrapTests(ctx context.Context, store *storage.Store, bucket, root string) error {
	if root == "" {
		return nil
	}
	info, err := os.Stat(root)
	if err != nil || !info.IsDir() {
		return nil
	}
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		key := filepath.ToSlash(rel)
		if err = store.PutObject(ctx, bucket, key, data, detectContentType(path)); err != nil {
			return fmt.Errorf("upload %s: %w", key, err)
		}
		return nil
	})
}

func detectContentType(path string) string {
	if filepath.Ext(path) == ".txt" {
		return "text/plain"
	}
	return "application/octet-stream"
}
