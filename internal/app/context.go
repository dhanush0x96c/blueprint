package app

import (
	"github.com/dhanush0x96c/blueprint/internal/config"
)

// Context holds runtime dependencies for the application.
type Context struct {
	Config *config.Config
}

func NewContext(cfg *config.Config) *Context {
	return &Context{
		Config: cfg,
	}
}
