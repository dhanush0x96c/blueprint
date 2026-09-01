# Blueprint — Agent Guide

## Project

Go CLI tool for universal project scaffolding via composable YAML templates.  
Module: `github.com/dhanush0x96c/blueprint` | Go 1.26.3

## Documentation

- Start with `README.md` for setup and quick usage.
- See `docs/cli.md` for command and flag behavior.
- See `docs/template-spec.md` for template schema and semantics.
- See `docs/architecture.md` for system design and flow.
- See `docs/template-naming.md` for naming conventions.

## Entrypoint

`main.go` → `cmd.Execute()` → `cmd.NewRootCmd()` (Cobra)

## Key Packages

| Package                       | Role                                                                                             |
| ----------------------------- | ------------------------------------------------------------------------------------------------ |
| `cmd/`                        | Cobra commands: `init`, `list`, `version`                                                        |
| `internal/app/context.go`     | DI wiring: Config, Sources, ChainResolver                                                        |
| `internal/scaffold/`          | Orchestrates: resolve tree → collect vars → render → write                                       |
| `internal/template/`          | Engine: loader (YAML), composer (tree + cycle detection), renderer (Go text/template), validator |
| `internal/vars/`              | Variable collector pipeline: CLI flags → defaults → inheritance → prompts                        |
| `internal/prompt/`            | TUI via charmbracelet/huh                                                                        |
| `internal/resolver/`          | Template resolution: User source → Builtin source (chain)                                        |
| `internal/config/`            | Config loader (env prefix: `BLUEPRINT`), default path `~/.config/blueprint/templates`            |
| `internal/builtin/templates/` | `//go:embed all:*` — embedded project/feature templates                                          |
| `internal/testutil/`          | Shared test fakes, helpers, and YAML fixtures reused across package tests                        |

## Commands

Task runner: `just` (see `justfile` or run `just --list`)

```bash
just build                       # build blueprint binary with version ldflags
just test                        # run all tests
just check                       # format, vet, lint, build, and test
just lint                        # run golangci-lint
just tidy                        # go mod tidy
just snapshot                    # test goreleaser release locally
```

- **Verification**: Always and only run `just check` to verify changes.

## Version Control

This is a `jj` repo — always use `jj` commands instead of `git` for version control.

## Dependency Management & Nix

- Whenever Go dependencies change (`go.mod` or `go.sum` modified via `go get` / `go mod tidy`), the `vendorHash` in `flake.nix` **must** be updated.
- To update `vendorHash`:
  1. Set `vendorHash = pkgs.lib.fakeHash;` (or empty `""`) in `flake.nix`.
  2. Run `nix build` and copy the expected hash from the failure message (`got: sha256-...`).
  3. Update `vendorHash` in `flake.nix` with the new hash.
  4. Run `nix build` to verify the build succeeds.

## Test Conventions

- Package tests use external `_test` packages (`template_test`, `resolver_test`) for black-box coverage
- testify helpers: `require.NoError` / `assert.Equal` etc.
- Tests use `t.TempDir()` + `os.DirFS` for temp filesystem
- Shared fakes/helpers/fixtures live in `internal/testutil` (instead of being duplicated per test file)

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
- Test architecture (latest): package tests were moved to external `_test` packages and now share common test infrastructure from `internal/testutil`
