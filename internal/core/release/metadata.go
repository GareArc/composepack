package release

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

const metadataFileName = "release.json"

// Metadata captures release.json contents in runtime directories.
type Metadata struct {
	ReleaseName   string         `json:"releaseName"`
	ChartName     string         `json:"chartName"`
	ChartVersion  string         `json:"chartVersion"`
	ChartDigest   string         `json:"chartDigest"`
	RuntimePath   string         `json:"runtimePath"`
	CreatedAt     time.Time      `json:"createdAt"`
	Values        map[string]any `json:"values"`
	ValuesSources []string       `json:"valuesSources"`
	ComposeFiles  []string       `json:"composeFiles"`
}

// Store persists release metadata inside runtime directories.
type Store struct{}

// Load reads release metadata from `<runtime>/release.json`.
func (s *Store) Load(ctx context.Context, runtimePath string) (*Metadata, error) {
	if runtimePath == "" {
		return nil, errors.New("runtime path is required")
	}
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
	}

	path := filepath.Join(runtimePath, metadataFileName)
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("read release metadata: %w", err)
	}

	var meta Metadata
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, fmt.Errorf("parse release metadata: %w", err)
	}
	return &meta, nil
}

// Save writes release metadata to `<runtime>/release.json`.
func (s *Store) Save(ctx context.Context, runtimePath string, meta *Metadata) error {
	if runtimePath == "" {
		return errors.New("runtime path is required")
	}
	if meta == nil {
		return errors.New("metadata must be provided")
	}
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return err
		}
	}

	meta.RuntimePath = runtimePath
	if meta.CreatedAt.IsZero() {
		meta.CreatedAt = time.Now().UTC()
	}

	if err := os.MkdirAll(runtimePath, 0o755); err != nil {
		return fmt.Errorf("ensure runtime directory: %w", err)
	}

	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return fmt.Errorf("serialize metadata: %w", err)
	}

	tempPath := filepath.Join(runtimePath, ".release.json.tmp")
	if err := os.WriteFile(tempPath, data, 0o644); err != nil {
		return fmt.Errorf("write temp metadata: %w", err)
	}
	if err := os.Rename(tempPath, filepath.Join(runtimePath, metadataFileName)); err != nil {
		return fmt.Errorf("rename metadata file: %w", err)
	}

	return nil
}
