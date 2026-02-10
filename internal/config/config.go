package config

// Config is the root configuration model for the application.
type Config struct {
	TemplatesDir string `yaml:"templates_dir"`
}
