package scaffold

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/dhanush0x96c/blueprint/internal/prompt"
	"github.com/dhanush0x96c/blueprint/internal/template"
)

// Scaffolder orchestrates the complete scaffolding process
type Scaffolder struct {
	engine    *template.Engine
	collector *prompt.Collector
	writer    *Writer
}

// NewScaffolder creates a new scaffolder with the given template base directory
func NewScaffolder(templatesFS fs.FS) *Scaffolder {
	engine := template.NewEngine(templatesFS)
	collector := prompt.NewCollector()
	writer := NewWriter()

	return &Scaffolder{
		engine:    engine,
		collector: collector,
		writer:    writer,
	}
}

// Options contains options for scaffolding
type Options struct {
	TemplatePath    string          // Path to the template
	OutputDir       string          // Output directory for scaffolded files
	Variables       map[string]any  // Pre-provided variables (skip prompts)
	EnabledIncludes map[string]bool // Pre-selected includes (skip prompt)
	Interactive     bool            // Whether to prompt for variables
	DryRun          bool            // If true, don't write files
	Overwrite       bool            // Whether to overwrite existing files
}

// Result contains the results of a scaffolding operation
type Result struct {
	FilesWritten  []string            // List of files written
	FilesSkipped  []string            // List of files skipped (already exist)
	Dependencies  []string            // Dependencies that need to be installed
	PostInitCmds  []template.PostInit // Post-init commands to run
	RenderedFiles map[string]string   // Map of file path -> content (for dry-run)
}

// Scaffold performs the complete scaffolding operation
func (s *Scaffolder) Scaffold(opts Options) (*Result, error) {
	// Load the template
	tmpl, err := s.engine.LoadTemplate(opts.TemplatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load template: %w", err)
	}

	// Collect variables and includes if interactive
	var ctx *template.Context
	var enabledIncludes map[string]bool

	if opts.Interactive {
		// Get all includes for prompting
		allIncludes, err := s.engine.GetAllIncludes(tmpl)
		if err != nil {
			return nil, fmt.Errorf("failed to get includes: %w", err)
		}

		// Collect interactively
		ctx, enabledIncludes, err = s.collector.CollectWithIncludes(tmpl, allIncludes)
		if err != nil {
			return nil, fmt.Errorf("failed to collect input: %w", err)
		}

		// Merge with pre-provided variables (pre-provided takes precedence)
		if opts.Variables != nil {
			providedCtx := template.NewTemplateContext(opts.Variables)
			ctx.Merge(providedCtx)
		}
	} else {
		// Use pre-provided variables
		if opts.Variables == nil {
			opts.Variables = make(map[string]any)
		}
		ctx = template.NewTemplateContext(opts.Variables)
		enabledIncludes = opts.EnabledIncludes
		if enabledIncludes == nil {
			enabledIncludes = make(map[string]bool)
		}
	}

	// Compose template with selected includes
	composedTmpl, err := s.engine.ComposeTemplateWithIncludes(tmpl, enabledIncludes)
	if err != nil {
		return nil, fmt.Errorf("failed to compose template: %w", err)
	}

	// Collect variables from enabled includes
	if opts.Interactive {
		if err := s.collector.CollectMissing(composedTmpl, ctx); err != nil {
			return nil, fmt.Errorf("failed to collect include variables: %w", err)
		}
	}

	// Validate that all required variables are present
	if err := s.collector.ValidateContext(composedTmpl, ctx); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Render all files
	renderedFiles, err := s.engine.RenderTemplate(composedTmpl, ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to render template: %w", err)
	}

	result := &Result{
		FilesWritten:  make([]string, 0),
		FilesSkipped:  make([]string, 0),
		Dependencies:  composedTmpl.Dependencies,
		PostInitCmds:  composedTmpl.PostInit,
		RenderedFiles: renderedFiles,
	}

	// Write files if not dry-run
	if !opts.DryRun {
		for destPath, content := range renderedFiles {
			fullPath := filepath.Join(opts.OutputDir, destPath)

			// Check if file exists
			if _, err := os.Stat(fullPath); err == nil && !opts.Overwrite {
				result.FilesSkipped = append(result.FilesSkipped, destPath)
				continue
			}

			// Write the file
			if err := s.writer.WriteFile(fullPath, content); err != nil {
				return nil, fmt.Errorf("failed to write file %s: %w", destPath, err)
			}

			result.FilesWritten = append(result.FilesWritten, destPath)
		}
	}

	return result, nil
}

// ScaffoldInteractive performs interactive scaffolding with prompts
func (s *Scaffolder) ScaffoldInteractive(templatePath, outputDir string) (*Result, error) {
	return s.Scaffold(Options{
		TemplatePath: templatePath,
		OutputDir:    outputDir,
		Interactive:  true,
		DryRun:       false,
		Overwrite:    false,
	})
}

// ScaffoldNonInteractive performs non-interactive scaffolding with pre-provided values
func (s *Scaffolder) ScaffoldNonInteractive(templatePath, outputDir string, variables map[string]any, includes map[string]bool) (*Result, error) {
	return s.Scaffold(Options{
		TemplatePath:    templatePath,
		OutputDir:       outputDir,
		Variables:       variables,
		EnabledIncludes: includes,
		Interactive:     false,
		DryRun:          false,
		Overwrite:       false,
	})
}

// Preview shows what would be scaffolded without writing files
func (s *Scaffolder) Preview(templatePath string, variables map[string]any, includes map[string]bool) (*Result, error) {
	return s.Scaffold(Options{
		TemplatePath:    templatePath,
		OutputDir:       ".",
		Variables:       variables,
		EnabledIncludes: includes,
		Interactive:     false,
		DryRun:          true,
		Overwrite:       false,
	})
}

// AddFeature adds a feature to an existing project
// This is similar to scaffolding but designed for adding features post-initialization
func (s *Scaffolder) AddFeature(featurePath, outputDir string, interactive bool) (*Result, error) {
	opts := Options{
		TemplatePath: featurePath,
		OutputDir:    outputDir,
		Interactive:  interactive,
		DryRun:       false,
		Overwrite:    false, // Don't overwrite existing files when adding features
	}

	return s.Scaffold(opts)
}
