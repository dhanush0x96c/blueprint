# Blueprint Template Specification

This document defines the authoritative specification for Blueprint templates. All templates — whether projects,
features, or components — MUST follow this format.

Blueprint is built around a single core principle:

> Everything is a template.

Projects, features, and components share the exact same structure and are processed by the same engine.

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

### 3.2 Roles

Roles provide semantic meaning to variables.

Currently supported roles:

#### `project_name`

This role defines the canonical name of the generated project.

**STRICT RULE:**

- Exactly ONE variable across the entire composed template tree MUST have `role: project_name`.
- Zero is invalid.
- More than one is invalid.
- Validation MUST fail if the constraint is violated.

This guarantees:

- Deterministic output directory naming
- Predictable module name resolution
- Clear ownership of the root project identity

If a feature defines `project_name`, it MUST only be usable in isolation OR validation must fail during composition.

Future roles may include:

- `module_path`
- `package_name`
- `service_name`

But only `project_name` is currently reserved and enforced.

---

## 4. Includes (Template Composition)

Templates may include other templates.

```yaml
includes:
  - template: features/go/testing
    enabled_by_default: true
```

### 4.1 Fields

| Field                | Required | Description             |
| -------------------- | -------- | ----------------------- |
| `template`           | Yes      | Template path           |
| `enabled_by_default` | No       | Default inclusion state |

### 4.2 Resolution Rules

- Includes are resolved recursively.
- Cycles MUST be detected and rejected.
- Variables from all included templates are merged.
- Dependency lists are merged and deduplicated.
- File lists are concatenated.

Composition order:

1. Load root template
2. Resolve includes depth-first
3. Merge results

This enables infinite composition as described in the core design.

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

| Field  | Required | Description                                       |
| ------ | -------- | ------------------------------------------------- |
| `src`  | Yes      | Source file or directory relative to template root |
| `dest` | Yes      | Output path relative to project root              |

### 6.2 File Processing

Files are processed based on their extension:

- **Template files (`.tmpl`)**: Rendered using Go `text/template` with all collected variables.
- **Non-template files**: Copied as-is without any processing.

Although the `.tmpl` extension is stripped during rendering,
explicitly listed files should specify the destination path directly (without `.tmpl`).

### 6.3 Directory Processing

When `src` is a directory, Blueprint recursively processes all files within:

- Each file with `.tmpl` extension is rendered and the `.tmpl` extension is automatically stripped from the output filename.
- All other files are copied without modification.
- The directory structure is preserved in the destination.

Example directory structure:
```
src/
  config.go.tmpl  → rendered and written as config.go
  utils.go.tmpl   → rendered and written as utils.go
  data.json       → copied as-is to data.json
```

### 6.4 Rendering Context

- Uses Go `text/template`.
- All collected variables available in root context.
- Includes share the same render context.

Files are processed in composition order.

If multiple templates write to the same destination:

- Behavior MUST be explicitly defined (error or override strategy).
- Silent overwrites are forbidden.

---

## 7. Post-Init Commands

Templates may define commands to execute after scaffolding.

```yaml
post_init:
  - command: "go mod tidy"
  - command: "go fmt ./..."
```

Rules:

- Executed after all files are written.
- Run in project root directory.
- Executed sequentially.
- Failure MUST stop execution and return error.

Post-init commands from composed templates are appended in resolution order.

---

## 8. Validation Rules

A valid template MUST satisfy:

- Required top-level fields present
- No duplicate variable names in composed tree
- Exactly one `project_name` role in full composition
- No cyclic includes
- All referenced template paths exist
- All referenced `src` files exist

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
