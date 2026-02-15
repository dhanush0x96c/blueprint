# Blueprint Architecture

This document describes the internal architecture of Blueprint — how the system is structured, how data flows through
it, and how the major components interact.

## Table of Contents

- [1. High-Level Overview](#1-high-level-overview)
- [2. Directory Structure](#2-directory-structure)
- [3. Package Responsibilities](#3-package-responsibilities)
  - [3.1 `cmd`](#31-cmd)
  - [3.2 `internal/app`](#32-internalapp)
  - [3.3 `internal/template`](#33-internaltemplate)
  - [3.4 `internal/prompt`](#34-internalprompt)
  - [3.5 `internal/scaffold`](#35-internalscaffold)
  - [3.6 `internal/config`](#36-internalconfig)
  - [3.7 `internal/ui`](#37-internalui)
  - [3.8 `internal/builtin/templates`](#38-internalbuiltintemplates)
  - [3.9 `internal/version`](#39-internalversion)
- [4. Core Data Flow](#4-core-data-flow)
  - [4.1 `blueprint init` Lifecycle](#41-blueprint-init-lifecycle)
  - [4.2 `blueprint list` Lifecycle](#42-blueprint-list-lifecycle)
- [5. Template Resolution](#5-template-resolution)
  - [5.1 Resolver Interface](#51-resolver-interface)
  - [5.2 Chain Resolver](#52-chain-resolver)
  - [5.3 FS Resolvers](#53-fs-resolvers)
- [6. Template Engine](#6-template-engine)
  - [6.1 Loading](#61-loading)
  - [6.2 Composition](#62-composition)
  - [6.3 Rendering](#63-rendering)
- [7. Variable Context](#7-variable-context)
- [8. Error Handling](#8-error-handling)
- [9. Design Principles](#9-design-principles)

---

## 1. High-Level Overview

Blueprint follows a pipeline architecture. A user command enters through the CLI layer, resolves a template from the
filesystem, collects user input via interactive prompts, composes the template with selected includes, renders all files,
and writes them to disk.

```
CLI Command
    │
    ▼
App Context (config, resolvers)
    │
    ▼
Template Resolution (local → builtin)
    │
    ▼
Scaffolder
    ├── Prompt Collector (variables + includes)
    ├── Template Engine (load → compose → render)
    └── Writer (files to disk)
    │
    ▼
UI Output (result summary)
```

Every template — project, feature, or component — is processed by the same engine. The `type` field is semantic only
and does not change processing behavior.

---

## 2. Directory Structure

```
blueprint/
├── main.go                          # Entry point → cmd.Execute()
├── cmd/
│   ├── root.go                      # Root command, global flags, context setup
│   ├── init.go                      # Template initialization command
│   ├── list.go                      # Template listing command
│   └── version.go                   # Version display command
├── internal/
│   ├── app/                         # Runtime context & dependency injection
│   │   ├── context.go               # App context (config, FS, resolver, options)
│   │   ├── resolver.go              # Resolver interface definition
│   │   ├── resolver_chain.go        # Chain-of-responsibility resolver
│   │   ├── resolver_fs.go           # Filesystem resolvers (local + builtin)
│   │   └── errors.go               # App-level error types
│   ├── template/                    # Template processing engine
│   │   ├── engine.go                # Orchestrator (load + compose + render)
│   │   ├── model.go                 # Data structures (Template, Variable, File, etc.)
│   │   ├── loader.go                # YAML parsing, validation, discovery
│   │   ├── composer.go              # Include resolution, merging, cycle detection
│   │   └── renderer.go              # Variable substitution via text/template
│   ├── prompt/                      # Interactive input collection
│   │   ├── engine.go                # TUI prompt rendering (charmbracelet/huh)
│   │   └── collector.go             # Variable/include collection & validation
│   ├── scaffold/                    # Scaffolding orchestration
│   │   ├── scaffolder.go            # Workflow coordinator
│   │   └── writer.go                # File writing & directory management
│   ├── config/                      # Configuration management
│   │   ├── config.go                # Config data model
│   │   ├── loader.go                # Hierarchical config loading
│   │   └── paths.go                 # Default config path resolution
│   ├── ui/                          # Terminal output
│   │   ├── errors.go                # Error rendering & dispatch
│   │   ├── render_template.go       # Scaffolding result display
│   │   ├── render_list.go           # Template list display
│   │   ├── exit_codes.go            # Error → exit code mapping
│   │   └── writer.go                # Output writing utilities
│   ├── version/                     # Build version info
│   │   └── version.go               # Version, commit, build date (ldflags)
│   └── builtin/templates/           # Embedded templates
│       ├── embed.go                 # go:embed directive
│       ├── projects/                # Project templates (go-cli, go-api, etc.)
│       └── features/                # Feature templates (go/testing, etc.)
└── docs/                            # Documentation
```

---

## 3. Package Responsibilities

### 3.1 `cmd`

The CLI layer built on [Cobra](https://github.com/spf13/cobra). Each file defines a single command.

- **root.go** — Creates the root command, initializes `app.Context` with config and resolvers, registers global flags
  (`--config`, `--verbose`, `--dry-run`), and attaches all subcommands.
- **init.go** — Resolves a template by name, constructs scaffolding options from flags (`--var`, `--include`,
  `--exclude`, `--yes`, `--force`), invokes the `Scaffolder`, and renders the result.
- **list.go** — Discovers templates from all sources and renders them as a table.
- **version.go** — Displays build version information.

### 3.2 `internal/app`

Runtime context and dependency injection. This package wires together configuration, filesystem access, and template
resolution.

**Key types:**

| Type               | Purpose                                                           |
| ------------------ | ----------------------------------------------------------------- |
| `Context`          | Holds Config, BuiltinFS, LocalFS, Resolver, and Options           |
| `Resolver`         | Interface: `Resolve(ctx, ref) → (ResolvedTemplate, error)`        |
| `ChainResolver`    | Tries resolvers in sequence until one succeeds                    |
| `ResolverLocal`    | Resolves from user template directory (`~/.config/blueprint/templates`) |
| `ResolverBuiltin`  | Resolves from embedded filesystem (`go:embed`)                    |
| `ResolvedTemplate` | Result tuple: filesystem handle + path within that filesystem     |
| `TemplateRef`      | Template reference: name + type                                   |

### 3.3 `internal/template`

The core template processing engine. Handles loading, composition, and rendering.

**Key types:**

| Type         | Purpose                                                              |
| ------------ | -------------------------------------------------------------------- |
| `Template`   | Full template definition (name, type, version, variables, includes, dependencies, files, post_init) |
| `Variable`   | User-input variable (name, prompt, type, role, default, options)      |
| `Include`    | Reference to another template with enabled_by_default flag           |
| `File`       | Source/destination mapping for template files                        |
| `PostInit`   | Post-scaffolding command (command string, optional workdir)          |
| `Context`    | Variable map (`map[string]any`) for template rendering               |
| `Engine`     | Orchestrator: wraps loader, composer, and renderer                   |
| `FileLoader` | Parses template.yaml, validates, discovers templates                 |
| `Composer`   | Resolves includes recursively, merges templates, detects cycles      |
| `Renderer`   | Processes files using Go `text/template` with custom function map    |

### 3.4 `internal/prompt`

Interactive input collection using [charmbracelet/huh](https://github.com/charmbracelet/huh).

**Key types:**

| Type        | Purpose                                                     |
| ----------- | ----------------------------------------------------------- |
| `Engine`    | Renders TUI prompts (Input, Confirm, Select, MultiSelect)   |
| `Collector` | Orchestrates variable collection, context merging, validation |

The `Collector` coordinates prompting for both template variables and include selection, skipping variables already
present in the context.

### 3.5 `internal/scaffold`

Orchestrates the full scaffolding workflow.

**Key types:**

| Type        | Purpose                                                     |
| ----------- | ----------------------------------------------------------- |
| `Scaffolder`| Coordinates: Engine + Collector + Writer                     |
| `Writer`    | File I/O with safe-write semantics (skip existing files)     |
| `Options`   | Configuration: template path, output dir, variables, dry-run |
| `Result`    | Output: files written/skipped, dependencies, post-init cmds  |

### 3.6 `internal/config`

Configuration loading with hierarchical precedence.

```
Defaults → Config File → Environment Variables → CLI Flags
```

- Default config path: `~/.config/blueprint/config.yaml`
- Override via `$BLUEPRINT_CONFIG` or `--config` flag

### 3.7 `internal/ui`

Terminal output formatting. Handles result rendering (files written, dependencies, post-init commands), template list
display, error presentation, and exit code mapping.

### 3.8 `internal/builtin/templates`

Embedded template resources using Go's `//go:embed` directive. Templates are compiled into the binary, requiring no
external files at runtime. Organized into `projects/` and `features/` subdirectories.

### 3.9 `internal/version`

Build-time version information injected via `ldflags`: version string, git commit SHA, and build date.

---

## 4. Core Data Flow

### 4.1 `blueprint init` Lifecycle

```
1. Parse CLI args & flags
         │
2. Load config (defaults → file → env → CLI)
         │
3. Build app.Context (config, BuiltinFS, LocalFS, ChainResolver)
         │
4. Resolve template reference
   ChainResolver: ResolverLocal → ResolverBuiltin
   Returns: ResolvedTemplate{FS, Path}
         │
5. Scaffolder.Scaffold(options)
   │
   ├─ 5a. Engine.LoadTemplate(path)
   │       Parse template.yaml → validate → resolve file paths
   │
   ├─ 5b. Engine.GetAllIncludes(template)
   │       Collect all transitive includes for prompting
   │
   ├─ 5c. Collector.CollectWithIncludes(template, includes)
   │       ├─ PromptIncludes → Multi-select feature picker
   │       └─ PromptVariables → Input/Confirm/Select forms
   │
   ├─ 5d. Engine.ComposeTemplateWithIncludes(template, enabledIncludes)
   │       ├─ Filter includes by user selection + enabled_by_default
   │       ├─ Recursively resolve and merge included templates
   │       └─ Deduplicate variables, files, dependencies
   │
   ├─ 5e. Collector.CollectMissing(composedTemplate, context)
   │       Prompt for variables introduced by selected includes
   │
   ├─ 5f. Collector.ValidateContext(template, context)
   │       Ensure all required variables are present
   │
   ├─ 5g. Engine.RenderTemplate(template, context)
   │       ├─ Process .tmpl files through text/template
   │       ├─ Copy non-.tmpl files as-is
   │       └─ Render destination paths with template variables
   │
   └─ 5h. Writer.SafeWriteFiles(renderedFiles)  [unless dry-run]
           ├─ Create directories (0755)
           ├─ Write files (0644)
           └─ Skip files that already exist
         │
6. UI.RenderResult(result)
   Display: files written ✓, skipped, dependencies, post-init commands
```

**Note:** Post-init commands are collected and displayed in the result but are NOT automatically executed.

### 4.2 `blueprint list` Lifecycle

```
1. Build app.Context
         │
2. Discover templates from all sources (builtin + local)
   FileLoader.DiscoverAll(fs) for each source
         │
3. UI.RenderTemplateList(templates)
   Display grouped by source in table format
```

---

## 5. Template Resolution

### 5.1 Resolver Interface

```go
type Resolver interface {
    Resolve(ctx *Context, ref TemplateRef) (*ResolvedTemplate, error)
}
```

A `TemplateRef` contains a template name and type. A `ResolvedTemplate` contains an `fs.FS` handle and the path to the
template directory within that filesystem.

### 5.2 Chain Resolver

The `ChainResolver` implements the chain-of-responsibility pattern. It holds an ordered list of resolvers and tries each
in sequence. The first resolver to return a successful result wins. If all resolvers fail, `ErrTemplateNotFound` is
returned.

Default chain order:

```
ResolverLocal → ResolverBuiltin
```

This allows user templates to override builtin templates by name.

### 5.3 FS Resolvers

Both `ResolverLocal` and `ResolverBuiltin` share the same resolution logic (`ResolveFromFS`):

1. Construct path based on template type directories (`projects/`, `features/`)
2. Look for `template.yaml` at that path in the given `fs.FS`
3. Return the filesystem handle + resolved path

The difference is the underlying filesystem:

- **ResolverLocal** — Uses `os.DirFS` pointing to the user's template directory
- **ResolverBuiltin** — Uses the `embed.FS` compiled into the binary

---

## 6. Template Engine

The `Engine` type is the unified orchestrator that wraps the loader, composer, and renderer behind a clean API.

### 6.1 Loading

`FileLoader.Load(path)`:

1. Resolve path to `template.yaml` (handles both directory and file paths)
2. Read YAML content from the filesystem
3. Unmarshal into `Template` struct
4. Validate required fields using `go-playground/validator` (name, type, version, files)
5. Resolve file source paths relative to the template directory

`FileLoader.Discover()` and `FileLoader.DiscoverAll()` walk the filesystem looking for `template.yaml` files and return
discovered templates organized by path.

### 6.2 Composition

The `Composer` resolves template includes recursively and merges them into a single composed template.

**Algorithm:**

```
Compose(template)
  └─ composeWithPath(template, [template.Name])
      ├─ Copy base template fields
      └─ For each include:
          ├─ Check for circular dependency (is name in path?)
          ├─ Load included template via resolver
          ├─ Recursively compose: composeWithPath(include, path + name)
          └─ mergeTemplate(destination, source)
```

**Merge rules:**

| Field          | Strategy                                                |
| -------------- | ------------------------------------------------------- |
| Variables      | Deduplicate by name — earliest definition wins          |
| Dependencies   | Deduplicate by package name — explicit version overrides |
| Files          | Deduplicate by destination path — earliest definition wins |
| Post-init      | Append all — execution order is preserved               |

**Selective composition:**

`ComposeWithEnabledIncludes(template, enabledMap)` filters includes before composing. An include is composed if the user
selected it OR it has `enabled_by_default: true`.

**Circular dependency detection:**

The composer tracks the resolution path as a list of template names. Before resolving an include, it checks whether that
template's name already appears in the path. If it does, composition fails with `ErrCircularDependency`.

### 6.3 Rendering

The `Renderer` processes files using Go's `text/template` package.

**Processing logic:**

```
RenderAll(template, context)
  └─ For each file in template.Files:
      ├─ Render destination path with template variables
      ├─ If source is a directory:
      │   └─ Recursively process all files within
      ├─ If file has .tmpl extension:
      │   ├─ Render content through text/template
      │   └─ Strip .tmpl from destination filename
      └─ Otherwise:
          └─ Copy file content as-is
```

**Template functions:**

| Category        | Functions                                             |
| --------------- | ----------------------------------------------------- |
| String          | `toLower`, `toUpper`, `title`, `trim`, `replace`, `contains`, `split`, `join` |
| Path            | `base`, `dir`, `ext`, `joinPath`                      |
| Type conversion | `toString`, `toInt`, `toBool`                         |
| Logic           | `default`, `empty`, `coalesce`                        |

---

## 7. Variable Context

Variables are collected into a `map[string]any` context that flows through the entire pipeline.

**Sources of variables (in merge order):**

1. CLI `--var` flags (pre-provided values)
2. Interactive prompts for base template variables
3. Interactive prompts for variables from selected includes

**Behavior:**

- Variables already present in the context are not prompted for
- `CollectMissing` fills gaps introduced by newly composed includes
- `ValidateContext` ensures all required variables have values before rendering
- The same context is shared across all rendered files and destination paths

---

## 8. Error Handling

Blueprint uses wrapped errors with context throughout the pipeline:

```go
return fmt.Errorf("failed to load template: %w", err)
```

**App-level sentinel errors:**

- `ErrTemplateNotFound` — No resolver could locate the template
- `ErrCircularDependency` — Cycle detected during template composition

**UI error dispatch:**

The `ui.RenderError` function matches error types and renders appropriate messages. For example,
`ErrTemplateNotFound` suggests running `blueprint list` to see available templates.

**Exit codes:**

| Code | Meaning             |
| ---- | ------------------- |
| 0    | Success             |
| 1    | General error       |
| 2    | Template not found  |

---

## 9. Design Principles

**Everything is a template.** Projects, features, and components share the same structure and are processed by the
same engine. The `type` field is semantic metadata — it does not change behavior.

**Composition over configuration.** Templates are composed through includes rather than configured through flags.
Features are injected by merging templates, not by toggling options.

**Resolution is pluggable.** The chain resolver pattern allows adding new template sources (git, HTTP, registries)
without changing the scaffolding pipeline.

**No silent side effects.** Files are never silently overwritten. Post-init commands are reported but not
automatically executed. Dry-run mode previews all operations.

**Hierarchical configuration.** Settings follow a clear precedence: defaults → config file → environment variables →
CLI flags. Each layer can override the previous one.

**Minimal magic.** Template rendering uses standard Go `text/template` syntax. File processing is determined by
extension (`.tmpl` or not). Directory structures are preserved as-is.
