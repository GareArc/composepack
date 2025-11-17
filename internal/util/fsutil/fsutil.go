package fsutil

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// EnsureDir ensures a directory exists before writes occur.
func EnsureDir(path string) error {
	if path == "" {
		return fmt.Errorf("directory path is empty")
	}
	return os.MkdirAll(path, 0o755)
}

// WriteFileAtomic writes a file via temporary path and rename for durability.
func WriteFileAtomic(ctx context.Context, path string, data []byte, perm uint32) error {
	if path == "" {
		return fmt.Errorf("file path is empty")
	}
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return err
		}
	}

	dir := filepath.Dir(path)
	if err := EnsureDir(dir); err != nil {
		return err
	}

	tmp, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpPath := tmp.Name()

	defer func() {
		_ = os.Remove(tmpPath)
	}()

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return fmt.Errorf("write temp file: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return fmt.Errorf("sync temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp file: %w", err)
	}

	if perm != 0 {
		if err := os.Chmod(tmpPath, os.FileMode(perm)); err != nil {
			return fmt.Errorf("chmod temp file: %w", err)
		}
	}

	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("rename temp file: %w", err)
	}

	return nil
}
