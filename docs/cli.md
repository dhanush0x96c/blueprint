# Blueprint CLI Reference

Complete command-line reference for Blueprint, the universal project scaffolding tool.

## Table of Contents

- [Global Options](#global-options)
- [Commands](#commands)
  - [blueprint init](#blueprint-init)
  - [blueprint add](#blueprint-add)
  - [blueprint list](#blueprint-list)
  - [blueprint search](#blueprint-search)
  - [blueprint version](#blueprint-version)
  - [blueprint completion](#blueprint-completion)
- [Configuration](#configuration)
- [Template Paths](#template-paths)
- [Examples](#examples)

---

## Global Options

These flags are available for all commands:

```
--config string         Config file path (default: ~/.config/blueprint/config.yaml)
--template-dir string   Override default template directory
--dry-run               Preview actions without writing files
--verbose               Enable verbose logging
--help, -h              Show help for any command
```

**Environment Variables:**

- `BLUEPRINT_CONFIG` - Path to configuration file
- `BLUEPRINT_TEMPLATE_DIR` - Custom template directory location

---

## Commands

### blueprint init

Initialize a new project from a template.

```bash
blueprint init <template-name> [output-dir] [flags]
```

**Arguments:**

- `<template-name>` - Template identifier (e.g., `go-cli`, `node-api-express`)
- `[output-dir]` - Output directory (optional, default: derived from project name)

**Flags:**

```
--var stringArray         Set template variable (format: key=value)
--yes, -y                 Skip interactive prompts, use defaults
--include stringArray     Force-enable optional features
--exclude stringArray     Force-disable default features
--force                   Overwrite existing files
```

**Examples:**

```bash
# Interactive initialization
blueprint init go-cli

# Non-interactive with variables
blueprint init go-api --yes \
  --var app_name=my-service \
  --var port=8080

# Custom output directory
blueprint init python-api-fastapi ./backend

# Force-enable specific features
blueprint init go-api --include features/go/database/postgres

# Dry run to preview
blueprint init node-api-express --dry-run

# Skip confirmation on overwrite
blueprint init go-cli existing-dir --force
```

**Interactive Prompts:**

When run without `--yes`, Blueprint will:
1. Prompt for required variables
2. Offer optional features (from `enabled_by_default: false` includes)
3. Confirm before writing files

Press `Ctrl+C` at any prompt to cancel safely.

---

### blueprint add

> **Status: Not yet implemented**

Add features or components to an existing project.

```bash
blueprint add <template-name> [flags]
```

**Arguments:**

- `<template-name>` - Feature or component template to add

**Flags:**

```
--target string          Target directory (default: current directory)
--var stringArray        Set template variable (format: key=value)
--yes, -y                Skip interactive prompts
--force                  Overwrite existing files
--merge                  Attempt to merge conflicts intelligently
```

**Examples:**

```bash
# Add testing to current project
blueprint add features/go/testing

# Add Docker configuration
blueprint add components/docker

# Add multiple features
blueprint add features/go/logging
blueprint add features/go/database/postgres

# Add to specific directory
blueprint add features/node/linting --target ./backend

# Non-interactive mode
blueprint add components/ci-cd/github-actions --yes

# Preview changes
blueprint add features/go/config --dry-run
```

**Conflict Resolution:**

When files already exist:
- Blueprint will prompt for each conflict
- Options: `[o]verwrite`, `[s]kip`, `[m]erge`, `[a]bort`
- Use `--force` to overwrite all automatically
- Use `--merge` to attempt intelligent merging

**Dependencies:**

`blueprint add` will:
1. Install new dependencies automatically
2. Update existing dependency files (go.mod, package.json, etc.)
3. Run post-init commands defined in the template

---

### blueprint list

List available templates.

```bash
blueprint list [projects|features|components] [flags]
```

**Subcommands:**

- `projects` - List only project templates
- `features` - List only feature templates
- `components` - List only component templates

**Flags:**

```
--source string          Filter by source: builtin, user (default: all)
--short                  Show compact output (name only)
```

**Examples:**

```bash
# List only project templates
blueprint list projects

# List only builtin templates
blueprint list components --source builtin

# List user-defined features
blueprint list features --source user

# Short output for scripting
blueprint list projects --short

# Combine subcommand and filters
blueprint list components --source builtin
```

**Output Format:**

Templates are grouped by source:

```
BUILTIN TEMPLATES
─────────────────────────────────────────────────────────────────────
NAME                     DESCRIPTION
─────────────────────────────────────────────────────────────────────
go-cli                   Command-line application
go-api                   HTTP API service
node-api-express         Express.js REST API

USER TEMPLATES
─────────────────────────────────────────────────────────────────────
NAME                     DESCRIPTION
─────────────────────────────────────────────────────────────────────
company-api              Company API template
```

For features:
```
BUILTIN TEMPLATES
─────────────────────────────────────────────────────────────────────
NAME                     DESCRIPTION
─────────────────────────────────────────────────────────────────────
features/go/testing      Testing framework setup
features/go/logging      Structured logging setup
features/go/config       Configuration management

USER TEMPLATES
─────────────────────────────────────────────────────────────────────
NAME                     DESCRIPTION
─────────────────────────────────────────────────────────────────────
features/auth            Authentication module
```

**Short Output:**

```bash
$ blueprint list projects --short
go-cli
go-api
node-api-express
company-api
```

---

### blueprint search

> **Status: Not yet implemented**

Search templates by name, tags, or language.

```bash
blueprint search <query> [flags]
```

**Arguments:**

- `<query>` - Search term (matches name, description, tags)

**Flags:**

```
--language string        Filter by language
--type string            Filter by type (project, feature, component)
--tags stringArray       Filter by tags
--format string          Output format: table, json (default: table)
```

**Examples:**

```bash
# Search for API templates
blueprint search api

# Search with language filter
blueprint search database --language go

# Search by tags
blueprint search --tags rest,http

# Search Go testing features
blueprint search test --language go --type feature

# Fuzzy matching (automatically applied)
blueprint search "web api"
```

**Search Algorithm:**

- Fuzzy matching on template names
- Full-text search in descriptions
- Tag matching (exact)
- Language filtering (exact)
- Results ranked by relevance

---

### blueprint version

Display version information.

```bash
blueprint version
```

The version command uses the global `--verbose` flag to control output detail.

**Examples:**

```bash
# Basic version
blueprint version

# Detailed version with build info
blueprint version --verbose
```

**Output:**

Basic output:
```
Blueprint v0.1.0
```

Verbose output:
```
Blueprint v0.1.0
Git Commit: a1b2c3d
Build Date: 2024-02-15T10:30:00Z
```

---

### blueprint completion

Generate shell completion scripts.

```bash
blueprint completion <shell>
```

**Arguments:**

- `<shell>` - Target shell: `bash`, `zsh`, `fish`, `powershell`

**Examples:**

```bash
# Bash
blueprint completion bash > /etc/bash_completion.d/blueprint

# Zsh
blueprint completion zsh > "${fpath[1]}/_blueprint"

# Fish
blueprint completion fish > ~/.config/fish/completions/blueprint.fish

# PowerShell
blueprint completion powershell > blueprint.ps1
```

**Setup Instructions:**

**Bash:**
```bash
echo 'source <(blueprint completion bash)' >> ~/.bashrc
```

**Zsh:**
```bash
echo 'source <(blueprint completion zsh)' >> ~/.zshrc
# Or for Oh My Zsh users:
mkdir -p ~/.oh-my-zsh/completions
blueprint completion zsh > ~/.oh-my-zsh/completions/_blueprint
```

**Fish:**
```bash
blueprint completion fish > ~/.config/fish/completions/blueprint.fish
```

---

## Configuration

Blueprint looks for configuration in the following locations (in order):

1. `--config` flag
2. `$BLUEPRINT_CONFIG` environment variable
3. `$HOME/.config/blueprint/config.yaml`
4. Current directory `.blueprint.yaml` (project-specific overrides)

**Configuration File Format:**

```yaml
# ~/.config/blueprint/config.yaml

# Default template directory
template_dir: ~/.config/blueprint/templates

# Custom template sources
sources:
  - name: official
    url: https://github.com/dhanush0x96c/blueprint-templates
    branch: main
  
  - name: company
    url: git@github.com:company/blueprint-templates.git
    branch: main

# Default variables (override in templates)
defaults:
  author: "Your Name"
  license: mit
  go_version: "1.22"

# Prompt preferences
prompts:
  confirm_before_write: true
  show_preview: true

# Output preferences
output:
  verbose: false
```

**Template Sources:**

Blueprint can pull templates from multiple sources:
- Local filesystem (default)
- Git repositories (coming soon)
- HTTP endpoints (coming soon)

---

## Template Paths

Templates are referenced using hierarchical paths:

**Projects:**
```
go-cli
node-api-express
python-api-fastapi
```

**Features:**
```
features/go/testing
features/go/database/postgres
features/node/linting
```

**Components:**
```
components/docker
components/ci-cd/github-actions
components/monitoring/prometheus
```

**Path Resolution:**

1. Check `--template-dir` flag
2. Check `template_dir` in config
3. Check `$BLUEPRINT_TEMPLATE_DIR` environment variable
4. Default to `~/.config/blueprint/templates`
5. Fall back to embedded templates

---

## Examples

### Complete Workflows

**Create a new Go CLI application:**

```bash
# Initialize with interactive prompts
blueprint init go-cli

# Add testing framework
cd my-app
blueprint add features/go/testing

# Add configuration management
blueprint add features/go/config

# Add logging
blueprint add features/go/logging
```

**Create a Go API with database:**

```bash
# Initialize API project
blueprint init go-api user-service \
  --var port=8080

cd user-service

# Add PostgreSQL support
blueprint add features/go/database/postgres

# Add Docker configuration
blueprint add components/docker

# Add GitHub Actions CI
blueprint add components/ci-cd/github-actions

# Add monitoring
blueprint add components/monitoring/prometheus
```

**Non-interactive scripting:**

```bash
#!/bin/bash
# Automated project setup

blueprint init go-api ./services/inventory \
  --yes \
  --var port=3000 \
  --include features/go/testing \
  --include features/go/logging

cd services/inventory

blueprint add features/go/database/postgres --yes
blueprint add components/docker --yes --force
blueprint add components/ci-cd/github-actions --yes

go mod tidy
git init
git add .
git commit -m "Initial commit via Blueprint"
```

**Search and explore:**

```bash
# Find all Go API templates
blueprint search api --language go

# List all database features
blueprint search database --type feature

# Preview what a template includes
blueprint init go-api --dry-run

# See all project templates
blueprint list projects
```

---

## Exit Codes

Blueprint uses standard exit codes:

- `0` - Success
- `1` - General error
- `2` - Misuse of command (invalid arguments)
- `3` - Template not found
- `4` - Validation failed
- `5` - Filesystem error (permission denied, disk full)
- `130` - Interrupted by user (Ctrl+C)

Use exit codes in scripts:

```bash
if ! blueprint init go-cli --yes; then
    echo "Failed to initialize project"
    exit 1
fi
```

---

## Getting Help

- `blueprint --help` - Show available commands
- `blueprint <command> --help` - Show command-specific help
- `blueprint <command> <subcommand> --help` - Show subcommand help

**Quick References:**

```bash
blueprint init --help          # Init command help
blueprint add -h               # Short help flag
blueprint search --help        # Search syntax
```

---

## Tips & Tricks

**Shell Aliases:**

```bash
# ~/.bashrc or ~/.zshrc
alias bpi='blueprint init'
alias bpa='blueprint add'
alias bpl='blueprint list'
alias bps='blueprint search'
```

**Find Template Before Init:**

```bash
# Search, then use fzf to select
blueprint search api | fzf | awk '{print $2}' | xargs blueprint init
```

**Preview Multiple Templates:**

```bash
for template in go-cli go-api node-api-express; do
    echo "=== $template ==="
    blueprint init $template --dry-run
done
```

**Batch Add Features:**

```bash
features=(
    "features/go/testing"
    "features/go/logging"
    "features/go/config"
)

for feature in "${features[@]}"; do
    blueprint add "$feature" --yes
done
```

---

## Troubleshooting

**Template not found:**
```bash
# List available project templates
blueprint list projects

# Or list all types
blueprint list features
blueprint list components

# Search by name
blueprint search <name>

# Check template directory
echo $BLUEPRINT_TEMPLATE_DIR
ls -la ~/.config/blueprint/templates/
```

**Permission denied:**
```bash
# Check output directory permissions
ls -ld ./output-dir

# Use sudo only if necessary
sudo blueprint init go-cli /opt/my-app
```

**Conflict resolution:**
```bash
# Preview changes first
blueprint add features/go/testing --dry-run

# Force overwrite if confident
blueprint add features/go/testing --force

# Or handle conflicts interactively
blueprint add features/go/testing  # prompts for each conflict
```

**Variables not set:**
```bash
# Use --var flag
blueprint init go-cli --var app_name=myapp

# Or set defaults in config
cat >> ~/.config/blueprint/config.yaml <<EOF
defaults:
  author: "Your Name"
EOF
```

---

## Further Reading

- [Template Specification](../docs/template-spec.md) - How to write templates
- [Template Naming](../docs/template-naming.md) - Naming conventions
- [Architecture](../docs/architecture.md) - How Blueprint works internally
- [Examples](../docs/examples/) - Template creation examples
