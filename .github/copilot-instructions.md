# Copilot Instructions for Blueprint

## High-Level Architecture

Blueprint is a Go CLI tool for universal project scaffolding. Its core architecture includes:

- **cmd/**: Entry point and CLI command definitions (uses Cobra)
- **internal/config**: Configuration loading (YAML, Viper)
- **internal/prompt**: Interactive prompt engine for collecting user input and template variables
- **internal/scaffold**: Orchestrates the scaffolding process, including template loading, variable collection, file writing, and feature injection
- **internal/template**: Template system (loading, composing, rendering, and variable management)
- **templates/**: Built-in templates for projects and features
- **main.go**: CLI entrypoint

The scaffolding process loads templates, prompts for variables (with support for includes/features), composes templates, renders files, and writes them to disk. Feature injection is handled via template includes and post-init commands.

## Key Conventions

- **Templates** are defined in YAML (`template.yaml`) and support variables, includes (for features), dependencies, and post-init commands.
- **Interactive prompts** use the `huh` library for a unified UX.
- **Feature injection** is implemented via template includes, allowing dynamic addition of features during scaffolding.
- **Dry-run** and preview modes are supported for safe scaffolding.

---

This file is intended to help Copilot and other AI tools understand the structure and conventions of this repository for more effective assistance.
