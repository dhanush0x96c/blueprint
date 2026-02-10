package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

func (l *Loader) applyDefaults(cfg *Config) error {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return fmt.Errorf("resolve user config directory: %w", err)
	}
	templatesDir := filepath.Join(configDir, "blueprint", "templates")

	cfg.TemplatesDir = templatesDir

	return nil
}

func (l *Loader) applyConfigFile(cfg *Config) error {
	if l.ConfigFile == "" {
		path, err := DefaultPath()
		if err != nil {
			return fmt.Errorf("could not detect default config path: %w", err)
		}
		l.ConfigFile = path
	}

	data, err := os.ReadFile(l.ConfigFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	return yaml.Unmarshal(data, cfg)
}

func (l *Loader) applyEnv(cfg *Config) error {
	// TODO: Apply the environment variables
	return nil
}

func (l *Loader) applyCLI(cfg *Config) error {
	// TODO: Apply CLI options
	return nil
}
