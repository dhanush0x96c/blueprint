# Copilot Instructions for Blueprint

## Project Overview

**Blueprint** is a universal project scaffolding CLI tool written in Go that enables intelligent template-based project initialization with interactive prompts, feature composition, and dependency management.

**Stack**: Go 1.25+, Cobra (CLI framework), Charmbracelet Huh (interactive UI), go-playground/validator, YAML parsing

---

## Architecture

### Complete Directory Structure

```
blueprint/
├── main.go                          # Application entry point
├── cmd/
│   ├── root.go                      # Root command & global CLI setup
│   └── init.go                      # Template initialization command (blueprint init)
├── internal/
│   ├── app/                         # Application runtime & dependency injection
│   │   ├── context.go               # App context with resolver injection
│   │   ├── resolver.go              # Template resolution interface
│   │   ├── resolver_chain.go        # Chain-of-responsibility resolver pattern
│   │   ├── resolver_fs.go           # FS-based resolvers (Builtin/Local)
│   │   └── errors.go                # App-level error types
│   ├── template/                    # Template processing engine
│   │   ├── engine.go                # Unified template orchestration (load + compose + render)
│   │   ├── model.go                 # Template data structures (Template, Variable, File, etc.)
│   │   ├── loader.go                # YAML parsing, validation, template discovery
│   │   ├── composer.go              # Include resolution, merging, circular dependency detection
│   │   ├── renderer.go              # Variable substitution & file processing (text/template)
│   │   └── *_test.go                # Unit tests (table-driven with testify)
│   ├── prompt/                      # Interactive input collection
│   │   ├── engine.go                # Prompt rendering (via Huh forms/inputs)
│   │   └── collector.go             # Variable/include collection, validation, context merging
│   ├── scaffold/                    # Scaffolding orchestration
│   │   ├── scaffolder.go            # Main workflow coordinator (prompt → compose → render → write)
│   │   └── writer.go                # File writing, directory management
│   ├── config/                      # Configuration management
│   │   ├── config.go                # Config data model
│   │   ├── loader.go                # Hierarchical loading (defaults→file→env→CLI)
│   │   └── paths.go                 # Config path resolution (~/.blueprint/config.yaml)
│   ├── ui/                          # User interface output
│   │   ├── errors.go                # Error rendering & dispatch
│   │   ├── result.go                # Scaffolding result display
│   │   ├── exit_codes.go            # Error → exit code mapping
│   │   ├── writer.go                # Output writing utilities
│   │   └── render_template.go       # Template output rendering
│   └── builtin/templates/           # Embedded templates (go:embed)
│       ├── embed.go                 # Embedding directive
│       ├── projects/                # Project templates (go-api, go-cli, python-api-fastapi)
│       └── features/                # Feature templates (go/testing)
├── docs/                            # Documentation (template-spec.md)
└── go.mod, go.sum                   # Dependency management
```

### Package Responsibilities

| Package | Purpose | Key Types |
|---------|---------|-----------|
| **app** | Runtime context, dependency injection, template resolution via chain pattern | `Context`, `Resolver`, `ChainResolver`, `ResolverLocal`, `ResolverBuiltin` |
| **template** | Core template processing: loading, composing (includes), rendering (variables) | `Template`, `TemplateEngine`, `FileLoader`, `Composer`, `Renderer` |
| **prompt** | Interactive CLI prompts for variables & feature selection | `PromptEngine`, `Collector`, `PromptOptions` |
| **scaffold** | Orchestrates full scaffolding workflow (prompt → compose → render → write) | `Scaffolder`, `Writer`, `Result` |
| **config** | Configuration loading with precedence: defaults → file → env → CLI | `Config`, `Loader` |
| **ui** | Terminal output, error handling, result rendering | `ErrorRenderer`, `ResultRenderer` |
| **builtin/templates** | Embedded template resources (projects: go-api, go-cli, python-api-fastapi; features: go/testing) | Organized by type/name |

---

## Core Concepts

### 1. Template System

#### Template Definition (template.yaml)

```yaml
name: go-cli                          # Required: Template identifier
type: project                         # Required: project|feature|component
version: 0.0.0                        # Required: Semantic version
description: "Go CLI application"     # Optional: Human-readable description

variables:                            # Optional: User-input variables
  - name: app_name                    # Variable identifier
    prompt: "Application name?"       # UI prompt text
    type: string                      # string|int|bool|select|multiselect
    role: project_name                # Semantic role (e.g., determines output directory)
    default: my-app                   # Optional: Pre-filled value
    options: [opt1, opt2]            # Required for select/multiselect types

includes:                             # Optional: Compose other templates (feature injection)
  - template: features/go/testing     # Path to included template (relative or absolute)
    enabled_by_default: false         # Pre-selection flag for interactive prompts

dependencies:                         # Optional: Packages to install
  - "github.com/spf13/cobra@v1.10.2" # Format: package@version or just package

files:                                # Required: Files to scaffold
  - src: cmd/                         # Source path (relative to template directory)
    dest: cmd/                        # Destination path (supports template vars: {{.var}})
  - src: main.go.tmpl                 # Files ending in .tmpl are processed with text/template
    dest: main.go                     # .tmpl extension stripped after rendering

post_init:                            # Optional: Post-scaffolding commands
  - command: "go mod tidy"            # Shell command to run
    workdir: optional/workdir         # Optional: Working directory (default: project root)
```

#### Template Loading Flow

```
FileLoader.Load(path)
  ├─ Resolve path to template.yaml (handles both dir and file paths)
  ├─ fs.ReadFile → Read YAML content
  ├─ yaml.Unmarshal → Parse into Template struct
  ├─ validator.Struct → Validate required fields (name, type, version, files)
  └─ Resolve file paths: prepend template directory to all src paths
```

**Discovery**: `Discover()` walks the FS looking for `template.yaml` files, organized by folder structure.

#### Template Composition (Include Resolution)

**Key Features**:
- Circular dependency detection (tracks template path during composition)
- Selective composition based on enabled includes
- Recursive merging with deduplication

```
Composer.Compose(template)
  └─ composeWithPath(template, [template.Name])  # Track path for circular check
      ├─ Copy base template fields
      ├─ For each include:
      │   ├─ Check if template.Name in path → circular dependency error
      │   ├─ Load included template via resolver
      │   ├─ Recursively compose it: composeWithPath(include, append(path, template.Name))
      │   └─ mergeTemplate(dst, src)
      │       ├─ Merge variables (by name, no duplicates)
      │       ├─ Merge dependencies (dedupe by package name, version override if explicit)
      │       ├─ Merge files (by destination path, earliest wins)
      │       └─ Append post_init commands (execution order preserved)
      └─ Return fully composed template
```

**Selective Composition**:
```go
ComposeWithEnabledIncludes(template, enabledMap)
  └─ Filters includes by: enabledMap[template] || include.EnabledByDefault
```

#### Template Rendering

**Template Processing**: Uses Go `text/template` with custom function map.

```
Renderer.RenderAll(template, context)
  └─ For each file in template.Files:
      └─ processPath(src, dest)
          ├─ If directory: recursively process all files
          └─ If file:
              ├─ RenderPath(dest) → Process destination path with template vars
              ├─ If .tmpl file:
              │   ├─ Render(src) → Process file content with variables
              │   └─ Strip .tmpl extension from destination
              └─ Else: Copy() → Return file content as-is
```

**Available Template Functions**:
- **String**: `toLower`, `toUpper`, `title`, `trim`, `replace`, `contains`, `split`, `join`
- **Path**: `base`, `dir`, `ext`, `joinPath`
- **Type Conversion**: `toString`, `toInt`, `toBool`
- **Logic**: `default`, `empty`, `coalesce`

**Template Syntax**:
```go
// main.go.tmpl
package main
import "{{ .module_path }}/cmd"

// Destination path template
dest: "{{ .module_path }}/src/main.go"

// Function chaining
{{ .app_name | toLower | title }}

// Default values
{{ default "8080" .port }}
```

### 2. Interactive Prompts

#### Variable Types & UI Components

```go
VariableType: string | int | bool | select | multiselect

Prompt Methods (via Charmbracelet Huh):
  ├─ string/int → Input field (with validation for int)
  ├─ bool → Confirm dialog
  ├─ select → Single-select menu
  └─ multiselect → Multi-select checkboxes
```

#### Collector Workflow

```
Collector.CollectWithIncludes(template, includes)
  ├─ PromptIncludes(includes) → Multi-select UI (respects enabled_by_default)
  ├─ CollectFromTemplate(template) → Unified form or individual prompts
  │   └─ Skips variables already provided in context
  └─ Return: variables map + selected includes

Collector.CollectMissing(template, context)
  └─ Only prompts for variables not yet in context (for include variables)

Collector.ValidateContext(template, context)
  └─ Ensures all required variables are present (returns error if missing)
```

### 3. Scaffolding Workflow

**End-to-End Execution** (`blueprint init <template>`):

```
cmd/init.go → RunE handler
  ├─ appCtx.Resolver.Resolve(TemplateRef)
  │   └─ ChainResolver tries: Local FS (~/.blueprint/templates) → Builtin FS (embed)
  │
  ├─ scaffolder.Scaffold(Options)
  │   ├─ engine.LoadTemplate(path)                  # Load base template
  │   ├─ engine.GetAllIncludes(template)            # Discover all includes (transitive)
  │   │
  │   ├─ collector.CollectWithIncludes()            # [IF Interactive]
  │   │   ├─ PromptIncludes → User selects features
  │   │   └─ PromptVariables → User provides values
  │   │
  │   ├─ engine.ComposeTemplateWithIncludes()       # Merge selected includes
  │   ├─ collector.CollectMissing()                 # Prompt for include variables
  │   │
  │   ├─ engine.RenderTemplate(template, context)   # Substitute variables
  │   │
  │   ├─ writer.WriteFile() for each rendered file  # [UNLESS DryRun]
  │   │   └─ SafeWrite: skip if file exists
  │   │
  │   └─ Return Result{
  │       FilesWritten: []string,
  │       FilesSkipped: []string,
  │       Dependencies: []string,
  │       PostInitCommands: []Command
  │     }
  │
  └─ ui.RenderResult(result) → Print summary (files, deps, commands)
```

**Important**: Post-init commands are returned in `Result` but NOT executed automatically. Caller is responsible for running them.

### 4. Configuration Management

**Hierarchical Loading** (precedence order):

```
Defaults → ConfigFile → EnvVars → CLIFlags
```

**Paths**:
- Config file: `~/.blueprint/config.yaml` or `$BLUEPRINT_CONFIG`
- Templates directory: `~/.blueprint/templates` or `Config.TemplatesDir`

**Config Structure**:
```yaml
templates_dir: /path/to/custom/templates  # User template directory
```

---

## Key Conventions

### Error Handling

```go
// Wrapped error pattern (all errors include context)
return fmt.Errorf("failed to load template: %w", err)

// App-level custom errors
var (
    ErrTemplateNotFound = errors.New("template not found")
    ErrCircularDependency = errors.New("circular dependency detected")
)

// UI dispatch (errors.go)
switch {
    case errors.Is(err, app.ErrTemplateNotFound):
        renderTemplateNotFound(err)
    default:
        renderDefault(err)
}

// Exit code mapping
ExitCode(err) → int  // 0 = success, non-zero = failure (custom mapping)
```

### Template Resolution (Resolver Chain)

```go
Context.Resolver = ChainResolver([ResolverLocal, ResolverBuiltin])
  ├─ ResolverLocal: checks ~/.blueprint/templates/<name>
  └─ ResolverBuiltin: checks embedded templates (go:embed)
```

**Resolution Algorithm**: First resolver to successfully resolve the template wins. If none resolve, return `ErrTemplateNotFound`.

### Variable Context Management

```go
// Context is map[string]any (dynamic typing)
context := map[string]any{
    "app_name": "my-app",
    "module_path": "github.com/user/my-app",
    "port": 8080,  // Can be int, string, bool, etc.
}

// Merging (later values override earlier)
context.Merge(newVars)  // newVars take precedence

// Validation
collector.ValidateContext(template, context)  // Ensures all required variables present
```

### Template Merging Rules (Composer)

```
Deduplication Strategy:
  ├─ Variables: By name (earliest definition wins)
  ├─ Dependencies: By package name (explicit version overrides)
  ├─ Files: By destination path (earliest definition wins)
  └─ PostInit: All appended (execution order preserved)
```

### File Writing

```go
Writer provides:
  ├─ WriteFile(path, content) → Create with 0644 permissions
  ├─ WriteFileWithPerm(path, content, perm) → Custom permissions
  ├─ SafeWrite(path, content) → Skip if file exists (no overwrite)
  ├─ SafeWriteFiles(files map) → Batch write with skip tracking
  ├─ EnsureDir(path) → Create parent directories (0755)
  └─ FileExists, DirExists, IsEmpty → File system helpers
```

### Testing Patterns

- Unit tests use `testify/assert` and `testify/require`
- Temporary directories for FS operations (`t.TempDir()`)
- Fixture templates defined as YAML strings in tests
- Table-driven tests for multiple scenarios
- Focus on error cases, edge conditions, and circular dependencies

---

## Example Templates

### go-cli Project Template

```yaml
# Location: internal/builtin/templates/projects/go-cli/
name: go-cli
type: project
version: 0.0.0
description: "Go CLI application with Cobra"

variables:
  - name: app_name
    prompt: "Application name?"
    type: string
    role: project_name
  - name: module_path
    prompt: "Module path? (e.g., github.com/user/app)"
    type: string
  - name: description
    type: string
    default: "A CLI application written in Go"

includes:
  - template: features/go/testing
    enabled_by_default: false

dependencies:
  - "github.com/spf13/cobra@v1.10.2"

files:
  - src: main.go.tmpl
    dest: main.go
  - src: cmd/root.go.tmpl
    dest: cmd/root.go
  - src: go.mod.tmpl
    dest: go.mod
  - src: README.md.tmpl
    dest: README.md

post_init:
  - command: "go mod tidy"
  - command: "go fmt ./..."
```

**Rendered Files**:
```go
// main.go.tmpl
package main
import "{{ .module_path }}/cmd"

func main() {
    cmd.Execute()
}

// go.mod.tmpl
module {{ .module_path }}
go 1.25
```

### go/testing Feature Template

```yaml
# Location: internal/builtin/templates/features/go/testing/
name: go-testing
type: feature
version: 0.0.0
description: "Testing setup with testify"

variables:
  - name: use_testify
    prompt: "Use testify for assertions?"
    type: bool
    default: true

dependencies:
  - "github.com/stretchr/testify@v1.9.0"

files:
  - src: main_test.go.tmpl
    dest: main_test.go
```

**Usage**: Referenced via `includes: [{ template: "features/go/testing" }]` in project templates.

---

## Important Implementation Notes

1. **No automatic post-init execution**: Commands are collected and returned; CLI doesn't run them automatically
2. **Variable typing is dynamic**: Context uses `map[string]any` with type conversion helpers (`toInt`, `toString`, etc.)
3. **Includes enable feature composition**: Powerful pattern for adding optional features without duplication
4. **File sources can be directories**: Recursively processed with same rendering logic
5. **Config follows 12-factor app principles**: Environment variables override file config override defaults
6. **All errors wrapped with context**: Enables better debugging through the pipeline
7. **Templates are version-aware**: `version` field in metadata (currently informational, not enforced)
8. **Circular dependency detection**: Prevents infinite loops during template composition
9. **Destination paths support template syntax**: `dest: "{{ .module_path }}/src/"` is valid
10. **Dry-run support**: `DryRun` flag skips all `writer.WriteFile` calls

---

## Common Development Tasks

### Adding a New Template
1. Create directory: `internal/builtin/templates/{projects|features}/<name>/`
2. Add `template.yaml` with required fields (name, type, version, files)
3. Add template files (`.tmpl` for rendered, plain for copied)
4. Test with `blueprint init <name>`

### Adding Template Functions
1. Add function to `template/renderer.go` in `createFuncMap()`
2. Update this documentation with function signature
3. Add unit test in `renderer_test.go`

### Modifying Prompt Behavior
1. Edit `prompt/engine.go` for UI rendering (Huh integration)
2. Edit `prompt/collector.go` for variable collection logic
3. Test interactive mode: `blueprint init <template>` (without `--var` flags)

### Adding New Variable Types
1. Update `VariableType` in `template/model.go`
2. Add handling in `prompt/engine.go` → `PromptVariable()`
3. Update validation in `prompt/collector.go`

---

## Development Workflow Preferences

### Testing and Building
- **Do NOT automatically run tests or builds after making changes** unless explicitly requested
- Only run tests/builds when the user specifically asks (e.g., "run tests", "test this", "build the project")
- When you need to verify syntax, use `go build` with minimal output or rely on the language server
- Trust that changes are correct based on code review rather than always verifying with test runs

---

This file is intended to help Copilot and other AI tools understand the structure, architecture, and conventions of this repository for more effective assistance.
