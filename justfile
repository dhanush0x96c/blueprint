# Blueprint Justfile
# Run `just` or `just --list` to view all available commands.

# Variables
binary_name := "blueprint"
version := `git describe --tags --always --dirty 2>/dev/null || echo "dev"`
commit := `git rev-parse --short HEAD 2>/dev/null || echo "unknown"`
date := `date -u +%Y-%m-%dT%H:%M:%SZ`
ldflags := "-s -w -X github.com/dhanush0x96c/blueprint/internal/version.Version=" + version + " -X github.com/dhanush0x96c/blueprint/internal/version.GitCommit=" + commit + " -X github.com/dhanush0x96c/blueprint/internal/version.BuildDate=" + date

# Default recipe: list available recipes
[group('help')]
default:
    @just --list

# -----------------------------------------------------------------------------
# Development & Build
# -----------------------------------------------------------------------------

# Build the blueprint binary with version ldflags
[group('build')]
build:
    go build -ldflags "{{ ldflags }}" -o {{ binary_name }} ./main.go

# Verify compilation across all packages without producing binaries
[group('build')]
verify-build:
    go build ./...

# Install the blueprint binary to GOBIN/GOPATH
[group('build')]
install:
    go install -ldflags "{{ ldflags }}" .

# Run the CLI directly with arguments (e.g. `just run init go-cli`)
[group('build')]
run *args:
    go run -ldflags "{{ ldflags }}" ./main.go {{ args }}

# Remove build artifacts and temporary test files
[group('build')]
clean:
    rm -f {{ binary_name }} coverage.out coverage.html result

# -----------------------------------------------------------------------------
# Testing & Quality
# -----------------------------------------------------------------------------

# Run all unit tests
[group('test')]
test *flags:
    go test ./... {{ flags }}

# Run tests with code coverage summary
[group('test')]
test-cover:
    go test -coverprofile=coverage.out ./...
    go tool cover -func=coverage.out

# Generate HTML code coverage report
[group('test')]
test-cover-html:
    go test -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out -o coverage.html
    @echo "Coverage report written to coverage.html"

# Run golangci-lint
[group('lint')]
lint:
    golangci-lint run

# Format all Go source files
[group('lint')]
fmt:
    gofmt -s -w .

# Run go vet
[group('lint')]
vet:
    go vet ./...

# Tidy and verify go.mod and go.sum dependencies
[group('lint')]
tidy:
    go mod tidy

# Run all checks: formatting, vetting, linting, build verification, and tests
[group('test')]
check: fmt vet lint verify-build test

# -----------------------------------------------------------------------------
# Nix
# -----------------------------------------------------------------------------

# Build package using Nix
[group('nix')]
nix-build:
    nix build

# Run Nix flake checks
[group('nix')]
nix-check:
    nix flake check

# -----------------------------------------------------------------------------
# Release & Demo
# -----------------------------------------------------------------------------

# Generate changelog preview using git-cliff
[group('release')]
changelog *args:
    git cliff {{ args }}

# Run GoReleaser in snapshot mode (dry-run without publishing)
[group('release')]
snapshot:
    goreleaser release --snapshot --clean

# Record demo GIF using VHS
[group('demo')]
demo:
    vhs examples/demo/demo.tape
