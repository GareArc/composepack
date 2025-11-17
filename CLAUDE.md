# ComposePack — Project Technical Description (with Extensible Layout)

*A Helm-style templating and packaging tool for Docker Compose*

---

## 1. Overview

ComposePack is a **Go-based CLI** that provides Helm-style templating, packaging, and release management for Docker Compose applications. It turns a chart (templated Compose fragments + assets) into a **single merged `docker-compose.yaml` per release** and manages a per-release runtime directory.

Core ideas:

* Charts: directories (or `.cpack.tgz` archives) containing:

  * Templated Compose fragments under `templates/compose/`
  * Templated and static files under `files/`
  * `Chart.yaml`, `values.yaml`, optional `values.schema.json`
* Templating:

  * Go templates + Sprig + Helm-style helper functions
  * Context: `.Values`, `.Env`, `.Release`, `.Chart`, `.Files`
* Runtime:

  * `.cpack-releases/<releaseName>/` with `docker-compose.yaml`, `files/`, `release.json`
* Operations:

  * Wrap `docker compose` for `up`, `down`, `logs`, `ps`, etc.
  * Never reimplement Compose merge rules: always rely on `docker compose config`.

---

## 2. Language, Stack, and Dependencies

### 2.1 Language

* **Go ≥ 1.22**
* Distribute as single binaries per platform (static when possible).

### 2.2 CLI & Config

* CLI: `github.com/spf13/cobra`
* Optional: `github.com/spf13/pflag` (implicitly via Cobra)
* Global configuration is minimal; most configuration is via chart values/env.

### 2.3 Templating

* `text/template`
* Sprig v3 (`github.com/Masterminds/sprig/v3`)
* Custom function map:

  * Helm-like:

    * `include`, `tpl`, `required`, `quote`, `squote`, `nindent`, `indent`
    * `toYaml`, `fromYaml`, `toJson`, `fromJson`
    * `semver`, `semverCompare`, and other non-K8s helpers
  * ComposePack-specific:

    * `env "VAR_NAME"`: read process env
* Template context:

  * `.Values` → merged values
  * `.Env` → map[string]string, process env snapshot
  * `.Release` → `{ Name, Service }`
  * `.Chart` → `{ Name, Version }`
  * `.Files` → chart file helper

### 2.4 YAML, JSON, Values, Schema

* YAML:

  * `gopkg.in/yaml.v3` or `sigs.k8s.io/yaml`
* JSON:

  * `encoding/json`
* Deep Merge:

  * `github.com/imdario/mergo`
  * Semantics:

    * Map keys: recursively merged; later override earlier
    * Scalars: later override earlier
    * Arrays: later override earlier (no deep merge)
* Schema:

  * `github.com/xeipuuv/gojsonschema` or `github.com/qri-io/jsonschema`

### 2.5 Packaging

* `archive/tar`
* `compress/gzip`
* Optional: `crypto/sha256` for chart digest

### 2.6 External Integration — Docker

* No Docker client SDK; **compose semantics remain in the binary**.
* Use `os/exec` to shell out to:

  * Preferred: `docker compose`
  * Fallback: `docker-compose`
* All fragment merging done via `docker compose config ...`.

---

## 3. High-Level Architecture

Conceptual layers:

1. **CLI Layer**

   * Parses commands/flags.
   * Delegates to application services.

2. **Application Layer**

   * Orchestrates workflows: `install`, `up`, `template`, `down`, `logs`, `ps`.
   * Combines chart loading, values merging, templating, runtime writing, docker calls.

3. **Domain/Core Layer**

   * Chart model & loader
   * Values + schema
   * Templating engine
   * Runtime directory management
   * Release metadata
   * Docker compose integration

4. **Infrastructure Layer**

   * Filesystem utilities
   * Process execution
   * Logging
   * Config, environment access

---

## 4. Folder Layout (Extensible, `internal/`-centric)

Use **`cmd/` for entrypoints**, **`internal/` for all project-specific logic** (not exported as a public Go API), and optionally later **`pkg/` for stable reusable libraries** if needed.

### 4.1 Top-Level Layout

```text
composepack/
  cmd/
    composepack/
      main.go

  internal/
    app/              # High-level use cases / orchestration
    cli/              # Cobra commands wiring
    core/             # Domain logic
      chart/
      values/
      templating/
      runtime/
      dockercompose/
      release/
    infra/            # Infrastructure concerns
      fs/
      process/
      logging/
      config/
    future/           # Placeholder for extensibility
      registry/
      hooks/
      plugins/

  # (optional for later)
  pkg/
    # small, generic utilities if we ever want third-party reuse
```

This layout is intentionally **extensible**:

* `internal/core` contains core “business logic” modules.
* `internal/app` orchestrates flows like `InstallRelease`, `RenderRelease`, `UpRelease`.
* `internal/cli` binds CLI commands to `internal/app`.
* `internal/infra` contains adapters to the OS / environment, making it easier to mock or swap implementations.
* `internal/future` names (registry/hooks/plugins) are reserved for future extensions without reshuffling core code.

### 4.2 Core Modules (under `internal/core`)

#### `internal/core/chart`

Responsible for chart representation and loading.

* `chart.go`:

  * `type ChartMetadata`
  * `type Chart`
* `loader.go`:

  * `LoadChartFromDir(path string) (*Chart, error)`
  * `LoadChartFromArchive(r io.Reader) (*Chart, error)`
* `packager.go` (later):

  * Build `.cpack.tgz` from a directory.
* `digest.go`:

  * Compute stable digest `ChartDigest(chart *Chart) string`

#### `internal/core/values`

Responsible for loading and merging values, and schema validation.

* `loader.go`:

  * Parse YAML into `map[string]any`.
  * Support `-f file1.yaml -f file2.yaml`.
* `merge.go`:

  * `MergeValues(base map[string]any, layers ...map[string]any) (map[string]any, error)`
  * Implements map/array override semantics.
* `schema.go`:

  * `ValidateValues(values map[string]any, schemaJSON []byte) error`

#### `internal/core/templating`

Helm-style template engine tailored for ComposePack.

* `context.go`:

  * Defines `RenderContext`, `ReleaseInfo`, `ChartInfo`.
* `functions.go`:

  * Builds FuncMap from Sprig + custom functions.
* `engine.go`:

  * Core engine API:

    ```go
    type Engine interface {
        RenderComposeTemplates(chart *Chart, ctx RenderContext) (map[string]string, error)
        RenderFiles(chart *Chart, ctx RenderContext) (map[string][]byte, error)
    }
    ```

#### `internal/core/runtime`

Managing per-release runtime folders.

* `runtime.go`:

  * `type RuntimeManager`
  * `PreparePath(releaseName string) (string, error)` for base path.
* `writer.go`:

  * `WriteRuntime(releaseName string, compose []byte, files map[string][]byte, meta ReleaseMetadata) error`
  * Handles atomic write (tmp dir + rename).
* `layout.go`:

  * Defines folder layout (where `docker-compose.yaml`, `files/`, `release.json` go).

#### `internal/core/release`

Encapsulates release metadata and lifecycle.

* `metadata.go`:

  * `type ReleaseMetadata` (JSON struct)
* `loader.go`:

  * `LoadMetadata(releaseName string) (*ReleaseMetadata, error)`
  * `WriteMetadata(releaseName string, meta *ReleaseMetadata) error`
* Optionally: `status.go` for derived release status.

#### `internal/core/dockercompose`

Wraps the Docker Compose CLI.

* `detect.go`:

  * Detect `docker` + `compose` subcommand vs `docker-compose`.
* `exec.go`:

  * Generic execution helper that runs compose commands in a given directory.
* `merge.go`:

  * `MergeFragments(runtimeDir string, fragmentPaths []string) ([]byte, error)`
  * Uses `docker compose config`.

---

## 5. Application Layer (`internal/app`)

This layer wires together core components into high-level operations.

* `install.go`:

  * `InstallRelease(chartSource ChartSource, opts InstallOptions) error`
* `up.go`:

  * `UpRelease(releaseName string, opts UpOptions) error`
* `template.go`:

  * `TemplateRelease(releaseName string, opts TemplateOptions) error`
* `down.go`, `logs.go`, `ps.go`:

  * Wrap Docker Compose operations.

`app` depends on:

* `core/chart`
* `core/values`
* `core/templating`
* `core/runtime`
* `core/dockercompose`
* `core/release`
* `util/fsutil`, `infra/process`, `infra/config`

This separation makes it easier to unit test logic without touching Cobra directly.

---

## 6. CLI Layer (`internal/cli` + `cmd/composepack`)

### `cmd/composepack/main.go`

* Creates root Cobra command.
* Initializes logging/config if needed.
* Calls into `internal/cli` to build the command tree.

### `internal/cli`

* `root.go`:

  * Defines `NewRootCmd() *cobra.Command`
* `install.go`, `up.go`, `template.go`, `down.go`, `logs.go`, `ps.go`:

  * Each defines a Cobra subcommand.
  * Parse flags.
  * Build appropriate `*Options`.
  * Call corresponding `internal/app` functions.

This layout keeps CLI-specific logic (flags, UX) isolated from business logic.

---

## 7. Infrastructure Layer (`internal/infra`)

Extensible area for “adapters” and low-level utilities.

### `internal/util/fsutil`

* Filesystem helpers:

  * Ensure directories exist
  * Safe temp dir creation
  * Atomic rename utilities

### `internal/infra/process`

* `RunCmd(cmd *exec.Cmd) (stdout, stderr []byte, err error)`
* Environment / PATH helpers (used by dockercompose).

### `internal/infra/logging`

* Simple logging facade wrapping `log/slog` or similar.
* Lets you swap logging implementation later.

### `internal/infra/config`

* Optional global configuration (e.g., base path for `.cpack-releases`).

---

## 8. Core Data Structures (unchanged conceptually, declared under `internal/core`)

### Chart

```go
type ChartMetadata struct {
    Name        string `yaml:"name"`
    Version     string `yaml:"version"`
    Description string `yaml:"description,omitempty"`
}

type Chart struct {
    Metadata      ChartMetadata
    BaseDir       string
    Values        map[string]any
    ValuesSchema  []byte                 // raw schema JSON if present
    ComposeTpls   map[string]string      // templates/compose/*.tpl.yaml
    FileTemplates map[string]string      // files/**/*.tpl
    StaticFiles   map[string][]byte      // non-templated files
}
```

### RenderContext

```go
type RenderContext struct {
    Values  map[string]any
    Env     map[string]string
    Release ReleaseInfo
    Chart   ChartInfo
}

type ReleaseInfo struct {
    Name    string
    Service string
}

type ChartInfo struct {
    Name    string
    Version string
}
```

### ReleaseMetadata

```go
type ReleaseMetadata struct {
    ReleaseName   string                 `json:"releaseName"`
    ChartName     string                 `json:"chartName"`
    ChartVersion  string                 `json:"chartVersion"`
    ChartDigest   string                 `json:"chartDigest"`
    RuntimePath   string                 `json:"runtimePath"`
    CreatedAt     time.Time              `json:"createdAt"`
    Values        map[string]any         `json:"values"`
    ValuesSources []string               `json:"valuesSources"`
    ComposeFiles  []string               `json:"composeFiles"`
}
```

---

## 9. Critical End-to-End Flow (MVP)

**For `composepack up <release> -f overrides.yaml`:**

1. `cmd/composepack/main.go` → root Cobra → `internal/cli.UpCmd`.
2. `UpCmd` parses flags, constructs `UpOptions`.
3. `UpCmd` calls `internal/app.UpRelease`.
4. `UpRelease`:

   * Loads chart (from archive/path stored in metadata or from flag).
   * Loads base values (`values.yaml` + previous metadata).
   * Loads additional values from `-f`.
   * Applies `--set` overrides.
   * Uses `internal/core/values` to merge and validate.
   * Builds `RenderContext` (including `.Env`, `.Release`, `.Chart`).
   * Uses `internal/core/templating.Engine` to render:

     * Compose fragments → fragment files in temp dir.
     * Files → `files/**` tree in temp dir.
   * Uses `internal/core/dockercompose.MergeFragments` to run `docker compose config` and get merged YAML.
   * Uses `internal/core/runtime.WriteRuntime` to atomically write runtime dir.
   * Writes/updates `release.json` via `internal/core/release`.
   * Uses `internal/core/dockercompose` to run `up` in the runtime dir.

This flow runs entirely inside `internal/`, and only the very thin `cmd/` layer interacts with Cobra.

---

## 10. Notes on Extensibility

Because core logic is under `internal/` and grouped logically, future additions are easy:

* Chart Registry:

  * `internal/core/registry`
  * `internal/app/registry.go`
  * `internal/cli/registry.go`
* Hooks / Plugins:

  * `internal/core/hooks`
  * `internal/core/plugins`
* Diagnostic tooling (e.g., `composepack diff`):

  * `internal/app/diff.go`
  * `internal/core/diff`
