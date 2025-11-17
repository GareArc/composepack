package fileloader

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// FileSystemLoader provides helpers for reading directories and files from disk.
type FileSystemLoader struct{}

// NewFileSystemLoader creates a filesystem loader instance.
func NewFileSystemLoader() *FileSystemLoader {
	return &FileSystemLoader{}
}

// ResolveDir ensures the provided path exists and is a directory, returning its absolute path.
func (l *FileSystemLoader) ResolveDir(path string) (string, error) {
	if path == "" {
		return "", errors.New("path must be provided")
	}

	abs, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolve path: %w", err)
	}

	info, err := os.Stat(abs)
	if err != nil {
		return "", fmt.Errorf("stat path: %w", err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("path %q is not a directory", abs)
	}

	return abs, nil
}

// ReadFileIfExists reads a file returning (nil, nil) when the file is missing.
func (l *FileSystemLoader) ReadFileIfExists(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	return data, nil
}

// ReadFile reads a file and returns an error if it cannot be read.
func (l *FileSystemLoader) ReadFile(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// WalkFiles walks the directory tree rooted at dir, invoking visit for each file.
func (l *FileSystemLoader) WalkFiles(ctx context.Context, dir string, visit func(rel string, data []byte) error) error {
	info, err := os.Stat(dir)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("stat directory: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("path %q is not a directory", dir)
	}

	return filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}

		rel, relErr := filepath.Rel(dir, path)
		if relErr != nil {
			return relErr
		}

		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}

		return visit(filepath.ToSlash(rel), data)
	})
}
