package app

import (
	"io/fs"
	"os"

	"github.com/dhanush0x96c/blueprint/internal/builtin/templates"
	"github.com/dhanush0x96c/blueprint/internal/config"
	"github.com/dhanush0x96c/blueprint/internal/resolver"
	"github.com/dhanush0x96c/blueprint/internal/template"
)

// Context holds runtime dependencies for the application.
type Context struct {
	Config    *config.Config
	BuiltinFS fs.FS
	LocalFS   fs.FS
	Resolver  template.Resolver
	Options   Options
}

// Options holds CLI flags and runtime options.
type Options struct {
	Verbose bool
	DryRun  bool
}

// NewContext creates a new application context.
func NewContext(cfg *config.Config, opts Options) *Context {
	localFS := os.DirFS(cfg.TemplatesDir)
	builtinFS := templates.Templates

	return &Context{
		Config:    cfg,
		LocalFS:   localFS,
		BuiltinFS: builtinFS,
		Options:   opts,
		Resolver: resolver.NewChainResolver(
			resolver.NewFSResolver(localFS),
			resolver.NewFSResolver(builtinFS),
		),
	}
}
