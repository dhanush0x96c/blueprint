# Blueprint Template Specification

This document defines the authoritative specification for Blueprint templates. All templates — whether projects,
features, or components — MUST follow this format.

Blueprint is built around a single core principle:

> Everything is a template.

Projects, features, and components share the exact same structure and are processed by the same engine.

## Table of Contents

- [1. Template File Location](#1-template-file-location)
- [2. Top-Level Fields](#2-top-level-fields)
  - [2.1 `name`](#21-name)
  - [2.2 `type`](#22-type)
  - [2.3 `version`](#23-version)
  - [2.4 `description`](#24-description)
  - [2.5 `tags`](#25-tags)
- [3. Variables](#3-variables)
  - [3.1 Variable Fields](#31-variable-fields)
  - [3.2 Roles](#32-roles)
- [4. Includes (Template Composition)](#4-includes-template-composition)
  - [4.1 Fields](#41-fields)
  - [4.2 Resolution Rules](#42-resolution-rules)
- [5. Dependencies](#5-dependencies)
- [6. Files](#6-files)
  - [6.1 Fields](#61-fields)
  - [6.2 File Processing](#62-file-processing)
  - [6.3 Directory Processing](#63-directory-processing)
  - [6.4 Rendering Context](#64-rendering-context)
- [7. Post-Init Commands](#7-post-init-commands)
- [8. Validation Rules](#8-validation-rules)
- [9. Execution Pipeline](#9-execution-pipeline)
- [10. Design Principles](#10-design-principles)

---

## 1. Template File Location

Each template MUST live in its own directory and contain a `template.yaml` file at its root.

Example:

```
templates/
  projects/
    go-cli/
      template.yaml
      main.go.tmpl
```

The canonical template system structure is documented in the repository README and reflected in the reference directory
layout.

---

## 2. Top-Level Fields

Every `template.yaml` MUST define the following fields:

```yaml
name: go-cli
type: project|feature|component
version: 1.0.0
description: "Short human-readable description"
tags: ["web", "api", "cli"] # optional
```

### 2.1 `name`

- Unique within its namespace.
- Lowercase, kebab-case recommended.
- Must match directory name.

### 2.2 `type`

Semantic only. Does NOT change behavior.

Allowed values:

- `project`
- `feature`
- `component`

All types are processed identically by the engine.

### 2.3 `version`

- Semantic version string.
- Used for template registry and compatibility in the future.

### 2.4 `description`

- Short explanation displayed in `blueprint list` and `blueprint search`.

### 2.5 `tags`

- **Optional** list of tags for categorization and filtering.
- Lowercase, kebab-case recommended.
- Used for discovery and search.
- Examples: `["web", "api", "cli", "microservice", "testing"]`

---

## 3. Variables

Templates may declare interactive variables.

```yaml
variables:
  - name: app_name
    prompt: "What is your application name?"
    type: string
    role: project_name
    default: my-app
```

### 3.1 Variable Fields

| Field     | Required | Description                                      |
| --------- | -------- | ------------------------------------------------ |
| `name`    | Yes      | Unique identifier                                |
| `prompt`  | Yes      | Question shown to user                           |
| `type`    | Yes      | `string`, `int`, `bool`, `select`, `multiselect` |
| `default` | No       | Default value                                    |
| `role`    | No       | Special semantic meaning                         |
| `options` | Cond.    | List of options for `select` and `multiselect`   |

### 3.2 Roles

Roles provide semantic meaning to variables.

Currently supported roles:

#### `project_name`

This role defines the canonical name of the generated project.

**STRICT RULES:**

- Only templates of type `project` MUST have exactly one variable with `role: project_name`.
- Templates of type `feature` or `component` SHOULD NOT have a `project_name` role.
- The variable with this role MUST be of type `string`.

This guarantees:

- Deterministic output directory naming
- Predictable module name resolution
- Clear ownership of the root project identity

---

## 4. Includes (Template Composition)

Templates may include other templates.

```yaml
includes:
  - template: go-testing
    enabled_by_default: true
```

### 4.1 Fields

| Field                | Required | Description                                                 |
| -------------------- | -------- | ----------------------------------------------------------- |
| `template`           | Yes      | Template name                                               |
| `enabled_by_default` | No       | Default inclusion state (default: `false`)                  |
| `mount`              | No       | Subdirectory to place the included template (projects only) |
| `inherits`           | No       | Mapping of parent variables to child variables              |

### 4.2 Resolution Rules

- Includes are resolved recursively.
- Cycles MUST be detected and rejected.
- Variables from all included templates are merged.
- Dependency lists are merged and deduplicated.
- File lists are concatenated.

### 4.3 Mount and Inheritance

#### `mount`

When a `project` template includes another `project` template, the `mount` field specifies the subdirectory where the included project will be scaffolded. If `mount` is not provided, the included project's `project_name` variable is used as the directory name.

`mount` has no effect when including `feature` or `component` templates.

#### `inherits`

The `inherits` field allows a parent template to pass its variable values to an included template. This is useful for sharing configuration (like `app_name`) without prompting the user twice.

```yaml
includes:
  - template: go-api
    inherits:
      api_name: project_name # child variable 'api_name' takes value of parent 'project_name'
```

---

## 5. Dependencies

Templates may declare external dependencies.

```yaml
dependencies:
  - "github.com/spf13/cobra@v1.10.2"
  - "github.com/spf13/viper@v1.21.0"
```

Rules:

- Treated as opaque strings.
- Merged across composed templates.
- Duplicates removed.
- Installer strategy depends on project language.

Dependency resolution must be deterministic.

---

## 6. Files

Templates define files to be rendered or copied.

```yaml
files:
  - src: "main.go.tmpl"
    dest: "main.go"
  - src: "static/"
    dest: "static/"
```

### 6.1 Fields

| Field  | Required | Description                                                 |
| ------ | -------- | ----------------------------------------------------------- |
| `src`  | Yes      | Source file or directory relative to template root          |
| `dest` | Yes      | Output path relative to project root (can be a Go template) |

### 6.2 File Processing

Files are processed based on their extension:

- **Template files (`.tmpl`)**: Rendered using Go `text/template` with all collected variables. The `.tmpl` extension is stripped from the output.
- **Non-template files**: Copied as-is without any processing.

The `dest` field itself is also processed as a template, allowing for dynamic file paths:

```yaml
files:
  - src: "handler.go.tmpl"
    dest: "internal/handlers/{{ .handler_name }}.go"
```

### 6.3 Directory Processing

When `src` is a directory, Blueprint recursively processes all files within:

- Each file with `.tmpl` extension is rendered and the `.tmpl` extension is automatically stripped from the output filename.
- All other files are copied without modification.
- The directory structure is preserved in the destination.

### 6.4 Rendering Context

- Uses Go `text/template`.
- All collected variables available in root context.
- Includes share the same render context.

#### Template Functions

Blueprint provides a set of built-in functions for use in templates:

- **Strings:** `toLower`, `toUpper`, `title`, `trim`, `trimLeft`, `trimRight`, `replace`, `contains`, `hasPrefix`, `hasSuffix`, `split`, `join`.
- **Paths:** `base`, `dir`, `ext`, `joinPath`.
- **Conversions:** `toString`, `toInt`, `toBool`.
- **Utilities:** `default`, `empty`, `coalesce`.

Example usage:
`{{ .app_name | toLower }}`
`{{ default "default_value" .optional_var }}`

---

## 7. Post-Init Commands

Templates may define commands to execute after scaffolding.

```yaml
post_init:
  - command: "go mod tidy"
    workdir: "./"
```

### 7.1 Fields

| Field     | Required | Description                                          |
| --------- | -------- | ---------------------------------------------------- |
| `command` | Yes      | Command to execute                                   |
| `workdir` | No       | Working directory for the command (relative to root) |

### 7.2 Execution Rules

- Executed after all files are written.
- Run in project root directory (unless `workdir` is specified).
- Executed sequentially.
- Failure MUST stop execution and return error.

Post-init commands from composed templates are appended in resolution order.

---

## 8. Validation Rules

A valid template MUST satisfy:

- Required top-level fields present (`name`, `type`, `version`).
- `type` is one of `project`, `feature`, `component`.
- `project` templates MUST have exactly one `project_name` role variable.
- Features and components MUST NOT include project templates.
- Features and components CANNOT be included twice at the same level.
- No cyclic includes.
- All referenced template paths exist.
- All referenced `src` files exist.
- Variable `options` are required for `select` and `multiselect` types.
- Default values must match the variable `type`.

Validation occurs before any filesystem writes.

---

## 9. Execution Pipeline

Blueprint processes templates as follows:

1. Load root template
2. Resolve includes recursively
3. Validate composition
4. Collect variables
5. Prompt user
6. Merge dependencies
7. Render files
8. Write filesystem
9. Execute post-init

This unified pipeline applies identically to projects, features, and components.

---

## 10. Design Principles

The specification enforces:

- Single responsibility per template
- Infinite composition
- Deterministic output
- Explicit validation
- Zero hidden magic

Blueprint does not distinguish between project, feature, and component at engine level — only at semantic level.

The result is a minimal, composable, and predictable scaffolding system aligned with the core design philosophy
documented in the project.
