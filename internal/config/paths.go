package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// DefaultPath returns the default path to the config file.
func DefaultPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve user config directory: %w", err)
	}

	path := filepath.Join(configDir, "blueprint", "config.yaml")

	return filepath.Clean(path), nil
}

// DefaultPathDisplay returns the default config path in a human-readable form.
func DefaultPathDisplay() (string, error) {
	path, err := DefaultPath()
	if err != nil {
		return "", err
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve user home directory: %w", err)
	}

	path = filepath.Clean(path)
	homeDir = filepath.Clean(homeDir)

	// Replace home directory prefix with a symbolic form
	if rel, err := filepath.Rel(homeDir, path); err == nil && !strings.HasPrefix(rel, "..") {
		if runtime.GOOS == "windows" {
			return filepath.Join("%USERPROFILE%", rel), nil
		}
		return filepath.Join("~", rel), nil
	}

	return path, nil
}

// DefaultPathUsage returns the CLI usage string describing the config flag and its default path.
func DefaultPathUsage() string {
	path, err := DefaultPathDisplay()
	if err != nil {
		return "config file"
	}
	return fmt.Sprintf("config file (default is %s)", path)
}
