# Blueprint

> Universal project scaffolding — because copy-pasting is not a strategy

Blueprint is a CLI tool written in Go that streamlines project initialization through an intelligent template system with interactive prompts, feature composition, and dependency management.

![Blueprint Demo](examples/demo/demo.gif)

## Features

- **Universal Templates** — Scaffold projects in any language or framework using a single YAML-based template format
- **Interactive Prompts** — Guided setup with text inputs, selects, multi-selects, and confirmations
- **Feature Composition** — Compose templates together via includes to add optional capabilities (testing, logging, etc.) during initialization
- **Custom Templates** — Use the built-in templates or create your own in `~/.config/blueprint/templates`
- **Dry Run** — Preview what Blueprint will generate before writing any files
- **Non-Interactive Mode** — Pass variables via `--var` flags and skip prompts with `--yes` for CI/scripting

## Installation

### From Source

```bash
go install github.com/dhanush0x96c/blueprint@latest
```

Make sure `$GOPATH/bin` (or `$GOBIN`) is in your `PATH`.

### Pre-built Binaries

Download the latest release from the [Releases](https://github.com/dhanush0x96c/blueprint/releases/latest) page.

## Quick Start

```bash
# Initialize a new Go CLI project
blueprint init go-cli
```

Blueprint will prompt you for variables like the application name and module path, then scaffold the project with all the necessary files.

### Non-Interactive Mode

```bash
blueprint init go-cli --yes \
  --var app_name=my-app \
  --var module_path=github.com/user/my-app
```

### Preview Without Writing

```bash
blueprint init go-api --dry-run
```

### Include Optional Features

```bash
blueprint init go-cli --include features/go/testing
```

## Custom Templates

Create your own templates in `~/.config/blueprint/templates/`. A template is a directory containing a `template.yaml` file and any source files to scaffold.

```yaml
name: my-template
type: project
version: 0.1.0

variables:
  - name: app_name
    prompt: "Application name?"
    type: string
    role: project_name

files:
  - src: main.go.tmpl
    dest: main.go
```

For the full template format, see the [Template Specification](docs/template-spec.md).

## Documentation

| Document | Description |
|----------|-------------|
| [CLI Reference](docs/cli.md) | Complete command-line reference and usage examples |
| [Template Specification](docs/template-spec.md) | Authoritative spec for the template format |
| [Template Naming Conventions](docs/template-naming.md) | Naming rules for templates |
| [Architecture](docs/architecture.md) | Internal architecture and data flow |

## Project Status

🚧 **Blueprint is in active development.** Some features documented in the CLI reference (such as `blueprint add`, `blueprint search`, and remote template sources) have not yet been implemented. The core scaffolding workflow — `blueprint init` with interactive prompts, template composition, and file rendering — is fully functional.

If the project interests you, consider starring or watching the repository to follow progress.

## License

This project is licensed under the MIT License – see the [LICENSE](LICENSE) file for details.
