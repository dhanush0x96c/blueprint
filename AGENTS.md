# Blueprint — Agent Guide

## Project

Go CLI tool for universal project scaffolding via composable YAML templates.  
Module: `github.com/dhanush0x96c/blueprint` | Go 1.26.3

## Entrypoint

`main.go` → `cmd.Execute()` → `cmd.NewRootCmd()` (Cobra)

## Key Packages

| Package                       | Role                                                                                             |
| ----------------------------- | ------------------------------------------------------------------------------------------------ |
| `cmd/`                        | Cobra commands: `init`, `list`, `version`. Alias: `bp`                                           |
| `internal/app/context.go`     | DI wiring: Config, Sources, ChainResolver                                                        |
| `internal/scaffold/`          | Orchestrates: resolve tree → collect vars → render → write                                       |
| `internal/template/`          | Engine: loader (YAML), composer (tree + cycle detection), renderer (Go text/template), validator |
| `internal/vars/`              | Variable collector pipeline: CLI flags → defaults → inheritance → prompts                        |
| `internal/prompt/`            | TUI via charmbracelet/huh                                                                        |
| `internal/resolver/`          | Template resolution: User source → Builtin source (chain)                                        |
| `internal/config/`            | Config loader (env prefix: `BLUEPRINT`), default path `~/.config/blueprint/templates`            |
| `internal/builtin/templates/` | `//go:embed all:*` — embedded project/feature templates                                          |

## Commands

```bash
go build -o blueprint ./main.go
go test ./...                    # testify (require + assert)
go test -run TestName ./pkg/     # focused test
go mod tidy
goreleaser release --clean       # CI only (tag v*.*.* triggers release.yml)
```

No Makefile / Taskfile / Justfile — use raw Go commands.

## Test Conventions

- Tests live in the **same package** (not `_test` external package)
- testify helpers: `require.NoError` / `assert.Equal` etc.
- Tests use `t.TempDir()` + `os.DirFS` for temp filesystem
- Fakes in test files (`fakeResolver`, `fakeLoader`) for mocking

## Template Format

- Template YAML must be named `template.yaml`
- Templates discovered by walking directories for `template.yaml` files
- Git-like `--var scope:key=value` and `--var #nodeID:key=value` syntax
- Scoping: `--var app_name=myapp` (global), `--var template-name:var=val` (per-template), `--var #0.1:var=val` (per-node)
- Optional includes selected interactively or via `--include`/`--exclude`
- `.tmpl` files: rendered via Go text/template, `.tmpl` stripped in output
- Non-`.tmpl` files: copied as-is
- Dest paths also processed as templates
- Builtin template funcs: `toLower`, `toUpper`, `toInt`, `default`, `coalesce`, `joinPath`, etc. (see `renderer.go:191`)

## Key Behaviors

- `blueprint init <template> [output-dir]` — scaffold a project
- `--yes` / `-y` for non-interactive mode (CI)
- `--dry-run` preview without writing (global flag)
- `--force` / `-f` to overwrite existing files
- Safe by default: refuses to write to non-empty dir unless `--force`
- Output dir defaults to the project_name variable value
- Template types (`project`/`feature`/`component`) are semantic only — all processed identically
- Project templates MUST have exactly 1 variable with `role: project_name`
- Features/components cannot include project templates
- No duplicate features/components at same tree level

## Architecture Notes

- Pipeline: CLI args → Resolve template → Build Node tree (composition) → Collect vars (CLI→defaults→inheritance→prompts) → Render → Write
- User templates in `~/.config/blueprint/templates` override builtins
- Version injected at build via ldflags (`-X github.com/dhanush0x96c/blueprint/internal/version.Version=...`)
- Config file at `~/.config/blueprint/config.yaml` (optional)
