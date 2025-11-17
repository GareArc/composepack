# PRD — ComposePack

*A Helm-style templating and packaging tool for Docker Compose*

---

## 1. Problem & Vision

### 1.1 Problem

Docker Compose is widely used, but it’s missing critical things for product-style distribution:

* No **templating**: `docker-compose.yaml` must be static.
* No separation between **system defaults** and **user overrides**.
* No **chart/package** concept.
* **Volumes must use relative paths**, which makes directory structure brittle.
* **Environment variables** often need to be injected at deploy time, but there is no first-class mechanism to turn them into config.

This leads to:

* Huge, hand-edited Compose files.
* Copy-paste deployments that are hard to upgrade.
* Frequent issues around paths, environment differences, and misconfigured variables.

### 1.2 Vision

ComposePack is a **Go CLI** that acts like Helm, but for Docker Compose.

* Apps are shipped as **charts**: templated Compose fragments + assets.

* Customers **install ComposePack** and run:

  ```bash
  composepack up my-release -f my-values.yaml
  ```

* At runtime, on the client machine, ComposePack:

  1. Loads chart defaults, user values, and environment variables.
  2. Renders templates (Compose YAML + scripts/configs).
  3. Merges all Compose fragments into a **single `docker-compose.yaml`**.
  4. Runs `docker compose` against that merged file from a well-defined runtime directory.

Dynamic templating and env resolution always happen **on the client**, just like Helm templates render on the cluster side.

---

## 2. Goals & Non-Goals

### 2.1 Goals

* Define a **chart format** for Compose-based applications.
* Provide a **templating system** for Compose YAML and auxiliary files.
* Support **values layering** (defaults, user files, CLI overrides, env).
* Standardize a **runtime directory layout** so relative volume paths are safe.
* Always produce a **single merged `docker-compose.yaml`** at runtime.
* Wrap Docker Compose with a small set of ergonomic commands (`install`, `up`, `down`, `logs`, etc.).

### 2.2 Non-Goals

* Do not replace Docker Compose (it’s still the engine).
* Do not implement orchestration features (scaling, scheduling, etc.).
* Do not encode “optional services” or “profiles” in the tool itself — those are implemented via templates and values.
* Do not target Kubernetes (Helm already covers that).

---

## 3. Users & Use Cases

### 3.1 Users

* **Product teams / platform engineers**: package services into reusable charts.
* **DevOps / infra**: maintain internal charts for common stacks.
* **Customers**: deploy a product via ComposePack + Docker Compose.
* **Developers**: maintain multi-environment Compose setups.

### 3.2 Use Cases

* Ship a product as a **chart** that can be configured via `values.yaml` and env vars.
* Allow customers to provision secrets, ports, and endpoints via values and environment.
* Simplify local and on-prem deployments with a consistent workflow.
* Eliminate fragile path handling for volume mounts by standardizing the runtime layout.

---

## 4. Architecture

### 4.1 Chart Structure (Source Format)

A chart is a directory that looks like:

```text
myapp/
  Chart.yaml             # name, version, description, metadata
  values.yaml            # system default values
  values.schema.json     # optional JSON Schema for validation

  templates/
    compose/
      00-core.tpl.yaml
      10-db.tpl.yaml
      20-api.tpl.yaml

  files/
    scripts/
      api-entrypoint.sh.tpl     # optional templated script
    config/
      app-config.yaml.tpl       # optional templated config
```

* **`templates/compose/`**: all templated Compose fragments (Go templates).
* **`files/`**: scripts/config/assets that may be templated or static.
* Charts are packaged into `*.cpack.tgz` archives for distribution.

> The tool does **not** understand “optional services” or “profiles” as a concept; that is done via template logic inside these files.

---

### 4.2 Runtime Directory (Per Release)

For each named release, ComposePack manages a runtime directory on the client machine:

```text
.cpack-releases/<releaseName>/
  docker-compose.yaml     # single merged compose file
  files/
    scripts/
    config/
  release.json            # metadata (chart, version, values used, etc.)
```

Key properties:

* **Exactly one** Compose file: `docker-compose.yaml`.
* All rendered assets live under `files/`.
* `docker compose` is always run **from this directory**.

---

### 4.3 Runtime Generation Pipeline

Whenever a runtime is (re)generated (e.g., `install`, `up`, `template`):

1. **Load Values & Env**

   * Start with `values.yaml` in the chart.
   * Merge additional values files (`-f` flags).
   * Apply `--set key=value` overrides.
   * Capture environment variables into `.Env`.

   Result: a single merged context for `.Values` and `.Env`.

2. **Render Templates**

   * Render all `templates/compose/*.tpl.yaml` → temporary directory as plain YAML fragments.
   * Render all files under `files/`:

     * `*.tpl` → rendered output without `.tpl` suffix.
     * Non-templated files are copied directly.

3. **Merge Compose Fragments into One File**

   Using Docker Compose to preserve behavior:

   ```bash
   docker compose \
     -f fragment1.yaml \
     -f fragment2.yaml \
     ... \
     config
   ```

   * Capture stdout → this is the merged `docker-compose.yaml`.
   * This ensures alignment with actual Compose semantics.

4. **Write/Refresh Runtime Directory**

   * Replace (or create) `.cpack-releases/<releaseName>/` with:

     * `docker-compose.yaml` (merged result).
     * `files/` tree (rendered assets).
     * `release.json` metadata.

5. **Invoke Docker Compose (if applicable)**

   For commands like `up`:

   ```bash
   (cd .cpack-releases/<releaseName>/ && docker compose -f docker-compose.yaml up -d)
   ```

---

### 4.4 Templating Model

* **Language**: Go templates.

* Template context includes:

  * `.Values` — merged values from all sources.
  * `.Env` — environment variables at runtime.
  * `.Release` — name, chart, version, etc.

* Functions:

  * Basic string/number/list/map utilities.
  * `env "VAR_NAME"` for direct access to environment variables.
  * `default`, `required`, etc., similar to Helm (scope to be defined).

Example:

```yaml
# templates/compose/20-api.tpl.yaml
services:
  api:
    image: "{{ .Values.api.image }}:{{ .Values.api.tag }}"
    environment:
      DB_HOST: "{{ .Values.db.host }}"
      DB_PASSWORD: "{{ env "DB_PASSWORD" | default .Values.db.password }}"
    volumes:
      - ./files/config/app-config.yaml:/app/config.yaml:ro
      - ./files/scripts/api-entrypoint.sh:/docker-entrypoint.d/api.sh:ro
```

---

### 4.5 Volumes & Paths

* All volume paths in templates are written assuming the **runtime directory** as the working directory.

* Example:

  ```yaml
  volumes:
    - ./files/scripts/api-entrypoint.sh:/docker-entrypoint.d/api.sh:ro
  ```

* ComposePack guarantees:

  * All files referenced under `./files/...` exist in `.cpack-releases/<releaseName>/files/...`.
  * All `docker compose` commands are executed from `.cpack-releases/<releaseName>/`.

This makes relative mounts reliable and predictable.

---

## 5. CLI & User Flows

### 5.1 Core Commands

#### `composepack install`

Install a chart into a named release and render its runtime:

```bash
composepack install myapp-1.0.0.cpack.tgz \
  --name my-release \
  -f my-values.yaml
```

* Creates/updates `.cpack-releases/my-release/`.
* Renders templates and writes `docker-compose.yaml` + `files/`.
* Optionally can run `docker compose up -d` (via flag).

#### `composepack up`

Re-render and bring containers up for an existing release:

```bash
composepack up my-release -f my-values.yaml
```

* Rebuilds `docker-compose.yaml` based on current values and env.
* Runs `docker compose up` in the release directory.

#### `composepack template`

Render only (no containers started):

```bash
composepack template my-release -f my-values.yaml
```

* Rebuilds runtime directory.
* Useful for inspection and debugging.

#### `composepack down`

Stop containers for a release:

```bash
composepack down my-release
```

* Runs `docker compose down` in the release directory.

#### `composepack logs`, `composepack ps`, etc

Thin wrappers:

```bash
composepack logs my-release --follow
composepack ps my-release
```

* Forwarded to `docker compose` in the release directory.

---

## 6. Implementation Phases

### Phase 1 — MVP

* Chart structure & packaging.
* Values loading and merging.
* Go template rendering for Compose YAML.
* Runtime directory creation.
* Single merged `docker-compose.yaml` using `docker compose config`.
* `template` and `up` commands.

### Phase 2 — Releases & Operations

* `install`, `down`, `logs`, `ps` commands.
* `release.json` metadata.
* Basic validation (schema for values, sanity checks).

### Phase 3 — Ecosystem

* Chart registry & discovery.
* Plugin/hooks system (optional).
* Diff/dry-run (`composepack diff`).
* Additional helpers around env files and secrets.

---

## 7. Success Criteria

* Internal products are consistently packaged as ComposePack charts.
* Customers deploy by running **ComposePack + Docker Compose** instead of editing YAML manually.
* Environment variable and volume path issues drop significantly.
* Single-file `docker-compose.yaml` per release makes support and debugging straightforward.
