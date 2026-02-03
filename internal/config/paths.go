package config

import (
	"fmt"
	"os"
	"path/filepath"
)

func DefaultPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve user config directory: %w", err)
	}

	// Keep the layout opinionated and predictable
	path := filepath.Join(configDir, "blueprint", "config.yaml")

	return filepath.Clean(path), nil
}

func DefaultPathHint() string {
	if configDir, err := os.UserConfigDir(); err == nil {
		return filepath.Join(configDir, "blueprint", "config.yaml")
	}

	// Fallback when the OS can’t determine a config dir
	return "~/.config/blueprint/config.yaml"
}
