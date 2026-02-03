package config

// Loader holds all configuration inputs.
// It has no behavior beyond loading.
type Loader struct {
	ConfigFile string
	EnvPrefix  string
	CLIArgs    map[string]string
}

// Load applies configuration in the following order:
// defaults → config file → env vars → cli args
func (l *Loader) Load() (*Config, error) {
	cfg := &Config{}

	if err := l.applyDefaults(cfg); err != nil {
		return nil, err
	}

	if err := l.applyConfigFile(cfg); err != nil {
		return nil, err
	}

	if err := l.applyEnv(cfg); err != nil {
		return nil, err
	}

	if err := l.applyCLI(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
