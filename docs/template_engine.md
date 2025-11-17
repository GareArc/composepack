# Template Engine

## 🔄 Code Trace #1: RenderComposeFragments Flow

**Entry Point:** `RenderComposeFragments(ctx, chart, renderContext)`

### Trace Walkthrough

```
┌─ RenderComposeFragments (line 84)
│  ├─ INPUT: ctx, chart.ComposeTpls, chart.HelperTpls, RenderContext
│  └─ DELEGATES TO: renderTemplates(ctx, "compose", ComposeTpls, HelperTpls, rc)
│
├─ renderTemplates (line 108)
│  │
│  ├─ STEP 1: Early return check (line 109-111)
│  │   └─ if len(templates) == 0 → return empty map
│  │
│  ├─ STEP 2: Initialize root template (line 113-115)
│  │   ├─ root = template.New("compose")
│  │   ├─ funcMap = buildFuncMap(rc, root) → line 159
│  │   └─ root.Funcs(funcMap)  // Attach functions
│  │
│  ├─ STEP 3: Register helper templates FIRST (line 117-119)
│  │   └─ registerHelperTemplates(root, helpers, funcMap) → line 197
│  │       ├─ Loop through helpers map
│  │       ├─ For each helper (e.g., "_helpers.tpl"):
│  │       │   ├─ root.New("_helpers.tpl")
│  │       │   ├─ Attach funcMap
│  │       │   └─ Parse helper body
│  │       └─ These become available for {{ include }} in main templates
│  │
│  ├─ STEP 4: Parse main templates (line 121-129)
│  │   ├─ Loop through templates map (e.g., "app.yaml", "db.yaml")
│  │   ├─ Check ctx.Err() for cancellation
│  │   ├─ For each template:
│  │   │   ├─ root.New("app.yaml")
│  │   │   ├─ Attach funcMap
│  │   │   └─ Parse template body
│  │   └─ All templates now in root's template tree
│  │
│  ├─ STEP 5: Build template data (line 131-132)
│  │   └─ buildTemplateData(rc) → line 148
│  │       └─ Creates map with .Values, .Env, .Release, .Chart, .Files
│  │
│  └─ STEP 6: Execute all templates (line 134-143)
│      ├─ Loop through each template name
│      ├─ Check ctx.Err() for cancellation
│      ├─ For each template:
│      │   ├─ Create buffer
│      │   ├─ root.ExecuteTemplate(&buf, "app.yaml", data)
│      │   │   └─ Template engine replaces {{ .Values.x }} with actual values
│      │   └─ Store buf.Bytes() in results["app.yaml"]
│      └─ Return results map
│
└─ OUTPUT: map[string][]byte
    ├─ "app.yaml" → rendered bytes
    └─ "db.yaml" → rendered bytes
```

---

## 🔄 Code Trace #2: RenderFiles Flow

**Entry Point:** `RenderFiles(ctx, chart, renderContext)`

### Trace Walkthrough

```
┌─ RenderFiles (line 89)
│  ├─ INPUT: ctx, chart.FileTemplates, chart.StaticFiles, RenderContext
│  │
│  ├─ STEP 1: Render templated files (line 90-93)
│  │   └─ DELEGATES TO: renderTemplates(ctx, "files", FileTemplates, HelperTpls, rc)
│  │       │
│  │       └─ [SAME AS COMPOSE FLOW above, but scope="files"]
│  │           ├─ Initialize root template
│  │           ├─ Build funcMap
│  │           ├─ Register helpers
│  │           ├─ Parse file templates (e.g., "init.sh.tpl", "config.yaml.tpl")
│  │           └─ Execute templates → rendered map
│  │
│  ├─ STEP 2: Check for nil result (line 95-97)
│  │   └─ if rendered == nil → create empty map
│  │
│  ├─ STEP 3: Add static files (line 99-103)
│  │   ├─ Loop through chart.StaticFiles
│  │   ├─ For each static file (e.g., "logo.png", "README.txt"):
│  │   │   ├─ Make a copy of the bytes (no template rendering)
│  │   │   └─ Add to rendered["logo.png"]
│  │   └─ Static files are NOT templated, just copied verbatim
│  │
│  └─ OUTPUT: map[string][]byte
│      ├─ "init.sh" → rendered from template
│      ├─ "config.yaml" → rendered from template
│      └─ "logo.png" → copied static file
└─ Return merged map
```

---

## 🔧 Helper Functions Deep Dive

### buildFuncMap (line 159)

```
┌─ buildFuncMap(rc, templateRoot)
│  │
│  ├─ STEP 1: Get Sprig functions (line 160)
│  │   └─ funcMap = sprig.TxtFuncMap()
│  │       └─ ~70 functions: default, required, toYaml, upper, trim, etc.
│  │
│  ├─ STEP 2: Add custom "env" function (line 162-169)
│  │   └─ Closure that captures rc.Env
│  │       ├─ First checks rc.Env map
│  │       └─ Falls back to os.Getenv()
│  │
│  ├─ STEP 3: Add "include" function (line 171-177)
│  │   └─ Closure that captures templateRoot
│  │       ├─ Allows {{ include "_helpers.labels" . }}
│  │       └─ Executes named template, returns string
│  │
│  ├─ STEP 4: Add "tpl" function (line 179-192)
│  │   └─ Dynamic template rendering
│  │       ├─ Takes a string like "{{ .Values.image }}"
│  │       ├─ Parses it as a new template
│  │       ├─ Executes with provided data
│  │       └─ Returns rendered string
│  │
│  └─ Return funcMap with all functions
```

### buildTemplateData (line 148)

```
┌─ buildTemplateData(rc)
│  │
│  └─ Creates map for template execution:
│      ├─ "Values"  → rc.Values  (user values.yaml)
│      ├─ "Env"     → rc.Env     (env variables)
│      ├─ "Release" → rc.Release (name, service)
│      ├─ "Chart"   → rc.Chart   (name, version)
│      └─ "Files"   → rc.Files   (file accessor)
│
└─ This becomes the "." (dot) in templates
```

---

## 📊 Data Flow Example

Let's trace a real example:

**Input:**

```go
RenderContext{
  Values: {database: {host: "localhost", port: 5432}},
  Env: {DB_PASSWORD: "secret123"},
  Release: {Name: "myapp", Service: "web"},
  Chart: {Name: "postgres", Version: "1.0.0"},
}
```

**Template (`app.yaml`):**

```yaml
services:
  {{ .Release.Name }}:
    image: postgres:latest
    environment:
      PGHOST: {{ .Values.database.host }}
      PGPASSWORD: {{ env "DB_PASSWORD" }}
    ports:
      - "{{ .Values.database.port }}:5432"
```

**Execution Trace:**

```
1. buildTemplateData() creates:
   {
     Values: {database: {...}},
     Env: {DB_PASSWORD: "secret123"},
     Release: {Name: "myapp", ...}
   }

2. Template engine walks through template:
   ├─ {{ .Release.Name }} → looks up data["Release"].Name → "myapp"
   ├─ {{ .Values.database.host }} → data["Values"]["database"]["host"] → "localhost"
   ├─ {{ env "DB_PASSWORD" }} → calls funcMap["env"]("DB_PASSWORD") → "secret123"
   └─ {{ .Values.database.port }} → "5432"

3. Output:
```

**Rendered Output:**

```yaml
services:
  myapp:
    image: postgres:latest
    environment:
      PGHOST: localhost
      PGPASSWORD: secret123
    ports:
      - "5432:5432"
```

---

## Key Differences Between Flows

| Aspect              | RenderComposeFragments        | RenderFiles                       |
| ------------------- | ----------------------------- | --------------------------------- |
| **Scope**           | "compose"                     | "files"                           |
| **Input Templates** | `chart.ComposeTpls`           | `chart.FileTemplates`             |
| **Output**          | Only rendered templates       | Rendered templates + static files |
| **Post-processing** | None                          | Merges static files (line 99-103) |
| **Use Case**        | Docker Compose YAML fragments | Scripts, configs, assets          |
