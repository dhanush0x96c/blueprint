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

func (l *Loader) selectConfigFile() error {
	if l.ConfigFile != "" {
		return nil
	}

	if envPath := os.Getenv(l.envKey("CONFIG")); envPath != "" {
		l.ConfigFile = envPath
		return nil
	}

	path, err := DefaultPath()
	if err != nil {
		return fmt.Errorf("could not detect default config path: %w", err)
	}
	l.ConfigFile = path

	return nil
}

func (l *Loader) applyConfigFile(cfg *Config) error {
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
	if val, ok := os.LookupEnv(l.envKey("TEMPLATES_DIR")); ok && val != "" {
		cfg.TemplatesDir = val
	}

	return nil
}

func (l *Loader) applyCLI(_ *Config) error {
	return nil
}

func (l *Loader) envKey(key string) string {
	if l.EnvPrefix == "" {
		return key
	}
	return fmt.Sprintf("%s_%s", l.EnvPrefix, key)
}
