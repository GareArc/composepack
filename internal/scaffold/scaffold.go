package scaffold

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Options controls chart scaffolding behavior.
type Options struct {
	Path    string
	Name    string
	Version string
	Force   bool
}

// CreateChart scaffolds a ComposePack chart directory with starter files.
func CreateChart(opts Options) error {
	if opts.Path == "" {
		return errors.New("path is required")
	}
	if opts.Name == "" {
		return errors.New("chart name is required")
	}
	if opts.Version == "" {
		opts.Version = "0.1.0"
	}

	path := opts.Path
	if err := ensureDirReady(path, opts.Force); err != nil {
		return err
	}

	dirs := []string{
		filepath.Join(path, "templates", "compose"),
		filepath.Join(path, "templates", "files"),
		filepath.Join(path, "templates", "helpers"),
		filepath.Join(path, "files"),
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create directory %s: %w", dir, err)
		}
	}

	files := map[string]string{
		filepath.Join(path, "Chart.yaml"):                              chartYAML(opts.Name, opts.Version),
		filepath.Join(path, "values.yaml"):                             defaultValues,
		filepath.Join(path, "templates", "compose", "00-app.tpl.yaml"): composeTemplate,
		filepath.Join(path, "templates", "files", "config.yml.tpl"):    fileTemplate,
		filepath.Join(path, "templates", "helpers", "_helpers.tpl"):    helperTemplate,
		filepath.Join(path, "files", ".gitkeep"):                       "",
	}

	for file, content := range files {
		if err := os.WriteFile(file, []byte(content), 0o644); err != nil {
			return fmt.Errorf("write %s: %w", file, err)
		}
	}

	return nil
}

func ensureDirReady(path string, force bool) error {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return os.MkdirAll(path, 0o755)
	}
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("%s exists and is not a directory", path)
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return err
	}
	if len(entries) > 0 && !force {
		return fmt.Errorf("directory %s is not empty (use --force to overwrite)", path)
	}
	return nil
}

func chartYAML(name, version string) string {
	return fmt.Sprintf(`name: %s
version: %s
description: Starter ComposePack chart
`, name, version)
}

const defaultValues = `app:
  image: my-app
  tag: latest
  env:
    EXAMPLE: "true"
`

const composeTemplate = `services:
  {{ .Release.Name }}-app:
    image: "{{ .Values.app.image }}:{{ .Values.app.tag }}"
    env_file:
      - ./files/config.yml
`

const fileTemplate = `example:
  env: {{ .Values.app.env.EXAMPLE | default "true" }}
`

const helperTemplate = `{{- define "example.fullname" -}}
{{ printf "%s-app" .Release.Name }}
{{- end -}}
`
