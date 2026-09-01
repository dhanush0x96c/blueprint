# Blueprint Architecture

This document describes the internal architecture of Blueprint — how the system is structured, how data flows through
it, and how the major components interact.

## Table of Contents

- [1. High-Level Overview](#1-high-level-overview)
- [2. Directory Structure](#2-directory-structure)
- [3. Package Responsibilities](#3-package-responsibilities)
  - [3.1 `cmd`](#31-cmd)
  - [3.2 `internal/app`](#32-internalapp)
  - [3.3 `internal/resolver`](#33-internalresolver)
  - [3.4 `internal/template`](#34-internaltemplate)
  - [3.5 `internal/vars`](#35-internalvars)
  - [3.6 `internal/prompt`](#36-internalprompt)
  - [3.7 `internal/scaffold`](#37-internalscaffold)
  - [3.8 `internal/config`](#38-internalconfig)
  - [3.9 `internal/ui`](#39-internalui)
  - [3.10 `internal/builtin/templates`](#310-internalbuiltintemplates)
- [4. Core Data Flow](#4-core-data-flow)
  - [4.1 `blueprint init` Lifecycle](#41-blueprint-init-lifecycle)
  - [4.2 `blueprint list` Lifecycle](#42-blueprint-list-lifecycle)
- [5. Template Resolution](#5-template-resolution)
  - [5.1 Source Model](#51-source-model)
  - [5.2 Resolver Interface](#52-resolver-interface)
  - [5.3 Chain Resolver](#53-chain-resolver)
- [6. Template Engine](#6-template-engine)
  - [6.1 Loading](#61-loading)
  - [6.2 Tree-based Composition](#62-tree-based-composition)
  - [6.3 Rendering](#63-rendering)
- [7. Variable Context & Collection](#7-variable-context--collection)
- [8. Error Handling](#8-error-handling)
- [9. Design Principles](#9-design-principles)

---

## 1. High-Level Overview

Blueprint follows a pipeline architecture centered around a **Template Tree**. A user command enters through the CLI layer, resolves a template from a source, constructs a tree of templates (root + selected includes), collects variables through a multi-stage pipeline, renders all files in the tree, and writes them to disk.

```
CLI Command
    │
    ▼
App Context (config, sources, resolver)
    │
    ▼
Template Resolution (Chain: User → Builtin)
    │
    ▼
Template Tree Construction (Composition + Include Selection)
    │
    ▼
Variable Collection Pipeline (CLI → Default → Inheritance → Prompt)
    │
    ▼
Template Rendering (Recursive walk of the tree)
    │
    ▼
Writer (Files to disk)
    │
    ▼
UI Output (Result summary)
```

---

## 2. Directory Structure

```
blueprint/
├── main.go                          # Entry point → cmd.Execute()
├── cmd/                             # Cobra commands
│   ├── root.go                      # Context setup & global flags
│   ├── init.go                      # Scaffolding command
│   └── list.go                      # Discovery command
├── internal/
│   ├── app/                         # Runtime context & DI
│   │   └── context.go               # Context (Config, Sources, Resolver)
│   ├── resolver/                    # Template source resolution
│   │   ├── source.go                # Source model (FS + Metadata)
│   │   ├── source_resolver.go       # FS-backed resolver & discovery
│   │   └── chain.go                 # Chain-of-responsibility resolver
│   ├── template/                    # Core template domain models & interfaces
│   │   ├── model.go                 # Data structures (Template, Node, Context)
│   │   ├── dependency.go            # Node dependency methods
│   │   ├── post_init.go             # Node post-init methods
│   │   ├── resolver.go              # TemplateRef & Resolver interfaces
│   │   ├── errors.go                # TemplateNotFoundError
│   │   ├── validator/               # Struct tag & semantic validation
│   │   ├── loader/                  # YAML parsing & manifest loading
│   │   ├── composer/                # Tree construction & cycle detection
│   │   ├── renderer/                # text/template processing
│   │   └── engine/                  # Orchestrator facade
│   ├── vars/                        # Variable collection pipeline
│   │   ├── collector.go             # Collector interface
│   │   ├── cli.go                   # CLI flag collector
│   │   ├── default.go               # Default value collector
│   │   ├── inheritance.go           # Parent -> Child variable mapping
│   │   └── prompt.go                # Interactive TUI collector
│   ├── prompt/                      # TUI rendering (bubbletea/huh)
│   │   └── engine.go                # Input, Select, MultiSelect prompts
│   ├── scaffold/                    # Scaffolding orchestration
│   │   ├── scaffolder.go            # Workflow coordinator
│   │   └── writer.go                # Safe file writing
│   ├── config/                      # Configuration management
│   ├── ui/                          # Terminal output formatting
│   └── builtin/templates/           # Embedded templates (go:embed)
```

---

## 3. Package Responsibilities

### 3.1 `cmd`

The CLI layer built on Cobra. It translates flags and arguments into `app.Context` and `scaffold.Options`.

### 3.2 `internal/app`

Wires together configuration and template sources. It initializes the `ChainResolver` with user-defined and builtin sources.

### 3.3 `internal/resolver`

Handles the discovery and location of templates across different filesystems.

- **Source:** Abstraction of a template repository (name, type, filesystem).
- **SourceResolver:** Finds templates in a specific filesystem by walking directories for `template.yaml`.
- **ChainResolver:** Implements `template.Resolver` by trying multiple sources in order.

### 3.4 `internal/template`

The domain foundation and processing engine of Blueprint. It is organized into modular subpackages:

- **`template` (root):** Core domain models (`Template`, `Node`, `Context`, `Variable`), node methods (`AllDependencies`, `AllPostInit`), and resolver interfaces.
- **`template/validator`:** Struct tag and semantic validation for templates, includes, and contexts.
- **`template/loader`:** Manifest YAML parsing (`template.yaml`) and metadata loading.
- **`template/composer`:** Recursively resolves includes to build the `Node` tree. Supports `ConfirmIncludes` callback for interactive selection and cycle detection.
- **`template/renderer`:** Renders files using Go's `text/template` with a rich function map and dynamic path interpolation.
- **`template/engine`:** High-level orchestrator facade unifying loader, composer, renderer, and validator.

### 3.5 `internal/vars`

A dedicated package for populating variable contexts.

- **Collector:** Interface for mutation of `RenderContexts`.
- **Inheritance:** Automatically propagates variables from parent nodes to child nodes based on `inherits` mappings in `template.yaml`.
- **Pipeline:** Collects variables in a specific order: CLI flags → Template Defaults → Inheritance → Interactive Prompts.

### 3.6 `internal/prompt`

Low-level TUI interaction using [charmbracelet/huh](https://github.com/charmbracelet/huh). It is used by the `vars.PromptCollector` to interact with the user.

### 3.7 `internal/scaffold`

Orchestrates the high-level workflow:

1. Load root template metadata.
2. Construct template tree (composing includes).
3. Run the variable collection pipeline.
4. Render the tree into a list of files.
5. Write files to the destination directory.

---

## 4. Core Data Flow

### 4.1 `blueprint init` Lifecycle

```
1. Resolve template reference (Name → SourceResolver)
2. Load root template (YAML parsing + validation)
3. Compose Tree:
   - Walk includes recursively
   - If include is optional, call ConfirmIncludes (TUI Multi-select)
   - Create Node tree
4. Collect Variables (for each node in tree):
   - Apply --var flags (CLICollector)
   - Apply default values (DefaultCollector)
   - Propagate from parents (Inheritance)
   - Prompt for remaining missing values (PromptCollector)
5. Render Tree:
   - Walk tree recursively
   - Render destination paths
   - Render file content (.tmpl) or copy as-is
6. Write Files:
   - Create directories
   - Write content with safe-write semantics (don't overwrite unless forced)
7. Render Result Summary (UI)
```

---

## 5. Template Resolution

### 5.1 Source Model

A `Source` represents a location where templates are stored.

- **Builtin:** Templates embedded in the binary.
- **User:** Local filesystem directory (e.g., `~/.config/blueprint/templates`).

### 5.2 Resolver Interface

```go
type Resolver interface {
    Resolve(ref TemplateRef) (*ResolvedTemplate, error)
}
```

### 5.3 Chain Resolver

Templates are resolved by trying sources in sequence:

1. **User Source** (allows local overrides)
2. **Builtin Source** (default templates)

---

## 6. Template Engine

### 6.1 Loading

`FileLoader` parses `template.yaml` and validates it against a strict schema using `go-playground/validator`. It supports semantic types: `project`, `feature`, and `component`.

### 6.2 Tree-based Composition

Unlike many scaffolders that "flatten" templates, Blueprint preserves a tree structure. Each node in the tree (`Node`) has its own:

- Metadata & Files
- Isolated variable context
- Mount point (relative path in the output)
- Inheritance rules

### 6.3 Rendering

The `Renderer` processes the tree recursively.

- **Destination Rendering:** Paths themselves can contain variables (e.g., `cmd/{{.ProjectName}}/main.go`).
- **Content Rendering:** Files ending in `.tmpl` are processed via `text/template`.
- **Context Isolation:** Each node is rendered with its specific context, but can access inherited values.

---

## 7. Variable Context & Collection

`RenderContexts` is a map of `NodeID` to `Context`. This allows the same variable name to have different values in different parts of the tree, while `Inheritance` provides a mechanism for sharing values (e.g., `project_name`) down the tree.

---

## 8. Error Handling

Blueprint uses specialized error types:

- `TemplateNotFoundError`: When a reference cannot be resolved.
- `CircularDependencyError`: When composition detects a loop in includes.
- `ValidationError`: For invalid `template.yaml` files or missing required variables.

---

## 9. Design Principles

- **Tree over Flat:** Preserving the hierarchy allows for more complex composition and cleaner variable scoping.
- **Explicit over Implicit:** Variable inheritance must be explicitly defined in the template manifest.
- **Safe by Default:** Never overwrite existing files unless `--force` is used.
- **Pluggable Sources:** The resolver architecture allows for future addition of Git or Remote sources.
- **Minimal Magic:** Standard Go templates and standard YAML.
