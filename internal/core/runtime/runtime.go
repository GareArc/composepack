package runtime

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"composepack/internal/util/fsutil"
)

const (
	composeFileName = "docker-compose.yaml"
	filesDirName    = "files"
)

// Writer is responsible for materializing runtime directories per release.
type Writer struct{}

// WriteOptions captures the artifacts that need to be written into the runtime directory.
type WriteOptions struct {
	ReleaseName string
	BaseDir     string
	ComposeYAML []byte
	Files       map[string][]byte
}

// Write commits the rendered artifacts to `.cpack-releases/<release>`.
func (w *Writer) Write(ctx context.Context, opts WriteOptions) (string, error) {
	if opts.ReleaseName == "" {
		return "", errors.New("release name is required")
	}
	if opts.BaseDir == "" {
		return "", errors.New("base directory is required")
	}
	if len(opts.ComposeYAML) == 0 {
		return "", errors.New("compose YAML cannot be empty")
	}
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return "", err
		}
	}

	runtimeDir := filepath.Join(opts.BaseDir, opts.ReleaseName)
	if err := fsutil.EnsureDir(runtimeDir); err != nil {
		return "", fmt.Errorf("ensure runtime dir: %w", err)
	}

	composePath := filepath.Join(runtimeDir, composeFileName)
	if err := fsutil.WriteFileAtomic(ctx, composePath, opts.ComposeYAML, 0o644); err != nil {
		return "", fmt.Errorf("write compose file: %w", err)
	}

	filesRoot := filepath.Join(runtimeDir, filesDirName)
	if err := os.RemoveAll(filesRoot); err != nil {
		return "", fmt.Errorf("clean files dir: %w", err)
	}
	if err := fsutil.EnsureDir(filesRoot); err != nil {
		return "", fmt.Errorf("ensure files dir: %w", err)
	}

	if len(opts.Files) > 0 {
		if err := w.writeFiles(ctx, filesRoot, opts.Files); err != nil {
			return "", err
		}
	}

	return runtimeDir, nil
}

func (w *Writer) writeFiles(ctx context.Context, root string, files map[string][]byte) error {
	keys := make([]string, 0, len(files))
	for rel := range files {
		keys = append(keys, rel)
	}
	sort.Strings(keys)

	for _, rel := range keys {
		if ctx != nil {
			if err := ctx.Err(); err != nil {
				return err
			}
		}

		clean := filepath.Clean(rel)
		if clean == "." || clean == "" || strings.HasPrefix(clean, "..") || filepath.IsAbs(clean) {
			return fmt.Errorf("invalid file path %q", rel)
		}

		dest := filepath.Join(root, clean)
		data := files[rel]
		if err := fsutil.WriteFileAtomic(ctx, dest, data, 0o644); err != nil {
			return fmt.Errorf("write file %s: %w", rel, err)
		}
	}

	return nil
}
