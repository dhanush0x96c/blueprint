# Blueprint

> Universal project scaffolding — because copy-pasting is not a strategy

Blueprint is a powerful command-line tool written in Go that streamlines project initialization through an intelligent template system with interactive prompts and feature injection capabilities.

## Features

- **Universal Support** - Scaffold projects across any programming language or framework
- **Interactive Prompts** - Guided setup process for seamless project configuration
- **Feature Injection** - Dynamically add features like validation libraries, testing frameworks, and more to your scaffolded projects
- **Template System** - Use built-in templates or create your own custom templates
- **Fast & Lightweight** - Built in Go for maximum performance

## Why Blueprint?

Unlike traditional scaffolding tools, Blueprint's feature injection system allows you to add capabilities to your projects on-the-fly during initialization, giving you the perfect starting point without bloating your codebase with unused dependencies.

## Installation

```bash
go install github.com/dhanush0x96c/blueprint@latest
```

Make sure `$GOPATH/bin` (or `$GOBIN`) is in your `PATH`.

## Project Status

🚧 **Blueprint is in active development.**

Core scaffolding functionality is implemented and working. The template system, interactive prompts, and basic project initialization are functional. Additional commands and built-in templates are in progress.
If the idea interests you, consider starring or watching the repository to follow progress and upcoming releases.

## Roadmap

- [x] Core CLI command structure
- [x] Template definition format
- [x] Interactive prompt engine
- [x] Template composition system
- [ ] Feature injection system (`blueprint add` command)
- [ ] Initial built-in templates
- [ ] Template search and discovery
- [ ] Documentation and examples

## Quick Start

```bash
# Initialize a new project from a template
blueprint init <template-name>

# Example
blueprint init go-cli
```

For detailed information on creating and using templates, see the [Template Specification](docs/template-spec.md) and [Template Naming Conventions](docs/template-naming.md).

## License

This project is licensed under the MIT License – see the [LICENSE](LICENSE) file for details.
