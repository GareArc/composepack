package templating

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"text/template"

	"github.com/Masterminds/sprig/v3"

	"composepack/internal/core/chart"
)

// Engine encapsulates the Go template rendering stack (text/template + Sprig + helpers).
type Engine struct{}

// NewEngine constructs an Engine.
func NewEngine() *Engine {
	return &Engine{}
}

// RenderContext contains the data exposed to templates at runtime.
type RenderContext struct {
	Values  map[string]any
	Env     map[string]string
	Release ReleaseInfo
	Chart   chart.ChartMetadata
	Files   FilesAccessor
}

// ReleaseInfo mirrors the fields surfaced via `.Release` in templates.
type ReleaseInfo struct {
	Name    string
	Service string
}

// FilesAccessor allows templates to read embedded file contents via `.Files`.
type FilesAccessor struct {
	files map[string][]byte
}

// NewFilesAccessor creates a FilesAccessor from chart static files.
func NewFilesAccessor(files map[string][]byte) FilesAccessor {
	copied := make(map[string][]byte, len(files))
	for k, v := range files {
		dup := make([]byte, len(v))
		copy(dup, v)
		copied[k] = dup
	}
	return FilesAccessor{files: copied}
}

// Get returns the file contents as string.
func (f FilesAccessor) Get(name string) string {
	data := f.files[name]
	return string(data)
}

// GetBytes returns a copy of the file bytes.
func (f FilesAccessor) GetBytes(name string) []byte {
	data := f.files[name]
	if data == nil {
		return nil
	}
	dup := make([]byte, len(data))
	copy(dup, data)
	return dup
}

// Exists reports whether the file exists.
func (f FilesAccessor) Exists(name string) bool {
	_, ok := f.files[name]
	return ok
}

// RenderComposeFragments renders templates/compose/* to concrete fragments, parsing helper `.tpl` snippets first.
func (e *Engine) RenderComposeFragments(ctx context.Context, ch *chart.Chart, rc RenderContext) (map[string][]byte, error) {
	return e.renderTemplates(ctx, "compose", ch.ComposeTpls, ch.HelperTpls, rc)
}

// RenderFiles renders chart file assets (scripts/config) into a runtime tree.
func (e *Engine) RenderFiles(ctx context.Context, ch *chart.Chart, rc RenderContext) (map[string][]byte, error) {
	rendered, err := e.renderTemplates(ctx, "files", ch.FileTemplates, ch.HelperTpls, rc)
	if err != nil {
		return nil, err
	}

	if rendered == nil {
		rendered = map[string][]byte{}
	}

	for name, data := range ch.StaticFiles {
		dup := make([]byte, len(data))
		copy(dup, data)
		rendered[name] = dup
	}

	return rendered, nil
}

func (e *Engine) renderTemplates(ctx context.Context, scope string, templates map[string]string, helpers map[string]string, rc RenderContext) (map[string][]byte, error) {
	if len(templates) == 0 {
		return map[string][]byte{}, nil
	}

	root := template.New(scope)
	funcMap := e.buildFuncMap(rc, root)
	root.Funcs(funcMap)

	if err := e.registerHelperTemplates(root, helpers, funcMap); err != nil {
		return nil, err
	}

	for name, body := range templates {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		tpl := root.New(name).Funcs(funcMap)
		if _, err := tpl.Parse(body); err != nil {
			return nil, fmt.Errorf("parse template %s: %w", name, err)
		}
	}

	results := make(map[string][]byte, len(templates))
	data := e.buildTemplateData(rc)

	for name := range templates {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		var buf bytes.Buffer
		if err := root.ExecuteTemplate(&buf, name, data); err != nil {
			return nil, fmt.Errorf("render template %s: %w", name, err)
		}
		results[name] = buf.Bytes()
	}

	return results, nil
}

func (e *Engine) buildTemplateData(rc RenderContext) map[string]any {
	data := map[string]any{
		"Values":  rc.Values,
		"Env":     rc.Env,
		"Release": rc.Release,
		"Chart":   rc.Chart,
		"Files":   rc.Files,
	}
	return data
}

func (e *Engine) buildFuncMap(rc RenderContext, t *template.Template) template.FuncMap {
	funcMap := sprig.TxtFuncMap()

	funcMap["env"] = func(key string) string {
		if rc.Env != nil {
			if v, ok := rc.Env[key]; ok {
				return v
			}
		}
		return os.Getenv(key)
	}

	funcMap["include"] = func(name string, data any) (string, error) {
		var buf bytes.Buffer
		if err := t.ExecuteTemplate(&buf, name, data); err != nil {
			return "", err
		}
		return buf.String(), nil
	}

	funcMap["tpl"] = func(text string, data any) (string, error) {
		if text == "" {
			return "", nil
		}
		tmp, err := template.New("tpl").Funcs(funcMap).Parse(text)
		if err != nil {
			return "", err
		}
		var buf bytes.Buffer
		if err := tmp.Execute(&buf, data); err != nil {
			return "", err
		}
		return buf.String(), nil
	}

	return funcMap
}

func (e *Engine) registerHelperTemplates(root *template.Template, helpers map[string]string, funcMap template.FuncMap) error {
	if len(helpers) == 0 {
		return nil
	}
	for name, body := range helpers {
		tpl := root.New(name).Funcs(funcMap)
		if _, err := tpl.Parse(body); err != nil {
			return fmt.Errorf("parse helper template %s: %w", name, err)
		}
	}
	return nil
}
