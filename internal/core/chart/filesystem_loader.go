package chart

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"sigs.k8s.io/yaml"

	"composepack/internal/util/fileloader"
)

// FileSystemChartLoader loads charts from directories on disk.
type FileSystemChartLoader struct {
	files *fileloader.FileSystemLoader
}

// NewFileSystemChartLoader constructs a loader that reads charts from the local filesystem.
func NewFileSystemChartLoader(files *fileloader.FileSystemLoader) *FileSystemChartLoader {
	if files == nil {
		files = fileloader.NewFileSystemLoader()
	}
	return &FileSystemChartLoader{files: files}
}

// Load satisfies the Loader interface.
func (l *FileSystemChartLoader) Load(ctx context.Context, source string) (*Chart, error) {
	baseDir, err := l.files.ResolveDir(source)
	if err != nil {
		return nil, err
	}

	ch := &Chart{
		BaseDir:       baseDir,
		ComposeTpls:   map[string]string{},
		FileTemplates: map[string]string{},
		HelperTpls:    map[string]string{},
		StaticFiles:   map[string][]byte{},
	}

	if err := l.loadMetadata(ch); err != nil {
		return nil, err
	}

	if err := l.loadValues(ch); err != nil {
		return nil, err
	}

	if err := l.loadSchema(ch); err != nil {
		return nil, err
	}

	if err := l.loadComposeTemplates(ctx, ch); err != nil {
		return nil, err
	}

	if err := l.loadFileTemplates(ctx, ch); err != nil {
		return nil, err
	}

	if err := l.loadHelperTemplates(ctx, ch); err != nil {
		return nil, err
	}

	if err := l.loadStaticFiles(ctx, ch); err != nil {
		return nil, err
	}

	return ch, nil
}

func (l *FileSystemChartLoader) loadMetadata(ch *Chart) error {
	metadataPath := filepath.Join(ch.BaseDir, MetadataFile)
	data, err := l.files.ReadFile(metadataPath)
	if err != nil {
		return fmt.Errorf("read %s: %w", MetadataFile, err)
	}

	var meta ChartMetadata
	if err := yaml.Unmarshal(data, &meta); err != nil {
		return fmt.Errorf("parse %s: %w", MetadataFile, err)
	}
	if meta.Name == "" || meta.Version == "" {
		return fmt.Errorf("chart metadata must include name and version")
	}

	ch.Metadata = meta
	return nil
}

func (l *FileSystemChartLoader) loadValues(ch *Chart) error {
	valuesPath := filepath.Join(ch.BaseDir, ValuesFile)
	data, err := l.files.ReadFileIfExists(valuesPath)
	if err != nil {
		return fmt.Errorf("read %s: %w", ValuesFile, err)
	}
	if len(data) == 0 {
		return nil
	}

	var vals map[string]any
	if err := yaml.Unmarshal(data, &vals); err != nil {
		return fmt.Errorf("parse %s: %w", ValuesFile, err)
	}
	ch.Values = vals
	return nil
}

func (l *FileSystemChartLoader) loadSchema(ch *Chart) error {
	schemaPath := filepath.Join(ch.BaseDir, ValuesSchemaFile)
	data, err := l.files.ReadFileIfExists(schemaPath)
	if err != nil {
		return fmt.Errorf("read %s: %w", ValuesSchemaFile, err)
	}
	ch.ValuesSchema = data
	return nil
}

func (l *FileSystemChartLoader) loadComposeTemplates(ctx context.Context, ch *Chart) error {
	dir := filepath.Join(ch.BaseDir, TemplatesCompose)
	return l.files.WalkFiles(ctx, dir, func(rel string, data []byte) error {
		ch.ComposeTpls[rel] = string(data)
		return nil
	})
}

func (l *FileSystemChartLoader) loadFileTemplates(ctx context.Context, ch *Chart) error {
	dir := filepath.Join(ch.BaseDir, TemplatesFiles)
	return l.files.WalkFiles(ctx, dir, func(rel string, data []byte) error {
		if !strings.HasSuffix(rel, TemplateFileSuffix) {
			return fmt.Errorf("file template %s must end with %s", rel, TemplateFileSuffix)
		}
		renderedName := strings.TrimSuffix(rel, TemplateFileSuffix)
		ch.FileTemplates[renderedName] = string(data)
		return nil
	})
}

func (l *FileSystemChartLoader) loadHelperTemplates(ctx context.Context, ch *Chart) error {
	dir := filepath.Join(ch.BaseDir, TemplatesHelpers)
	return l.files.WalkFiles(ctx, dir, func(rel string, data []byte) error {
		ch.HelperTpls[rel] = string(data)
		return nil
	})
}

func (l *FileSystemChartLoader) loadStaticFiles(ctx context.Context, ch *Chart) error {
	dir := filepath.Join(ch.BaseDir, FilesDir)
	return l.files.WalkFiles(ctx, dir, func(rel string, data []byte) error {
		ch.StaticFiles[rel] = data
		return nil
	})
}
