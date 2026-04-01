package scaffold

import (
	"fmt"
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

// NewScaffolder creates a new scaffolder with the given template resolver
func NewScaffolder(resolver template.Resolver) *Scaffolder {
	engine := template.NewEngine(resolver)
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
	TemplateRef     template.TemplateRef // Template reference to scaffold
	OutputDir       string               // Output directory for scaffolded files
	Variables       map[string]any       // Pre-provided variables (skip prompts)
	EnabledIncludes map[string]bool      // Pre-selected includes (skip prompt)
	Interactive     bool                 // Whether to prompt for variables
	DryRun          bool                 // If true, don't write files
	Overwrite       bool                 // Whether to overwrite existing files
}

// Result contains the results of a scaffolding operation
type Result struct {
	FilesWritten  []string                // List of files written
	FilesSkipped  []string                // List of files skipped (already exist)
	Dependencies  []string                // Dependencies that need to be installed
	PostInitCmds  []template.PostInit     // Post-init commands to run
	RenderedFiles []template.RenderedFile // List of rendered files
}

// Scaffold performs the complete scaffolding operation
func (s *Scaffolder) Scaffold(opts Options) (*Result, error) {
	// Load the root template
	loaded, err := s.engine.LoadTemplate(opts.TemplateRef)
	if err != nil {
		return nil, fmt.Errorf("failed to load template: %w", err)
	}

	tree, err := s.compose(loaded, opts)
	if err != nil {
		return nil, err
	}

	contexts, err := s.collectVariables(tree, opts)
	if err != nil {
		return nil, err
	}

	outputDir := s.determineOutputDir(tree, contexts, opts)

	renderedFiles, err := s.render(tree, contexts)
	if err != nil {
		return nil, err
	}

	written, skipped, err := s.writeFiles(renderedFiles, outputDir, opts)
	if err != nil {
		return nil, err
	}

	return &Result{
		FilesWritten:  written,
		FilesSkipped:  skipped,
		Dependencies:  tree.AllDependencies(),
		PostInitCmds:  tree.AllPostInit(),
		RenderedFiles: renderedFiles,
	}, nil
}

func (s *Scaffolder) compose(loaded *template.LoadedTemplate, opts Options) (*template.TemplateNode, error) {
	var confirm template.ConfirmIncludes
	if opts.Interactive {
		confirm = s.collector.ConfirmIncludes
	} else {
		confirm = s.confirmIncludesFromOptions(opts)
	}

	tree, err := s.engine.Compose(loaded, confirm)
	if err != nil {
		return nil, fmt.Errorf("failed to compose template tree: %w", err)
	}

	if err := s.engine.ValidateTree(tree); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	return tree, nil
}

func (s *Scaffolder) confirmIncludesFromOptions(opts Options) template.ConfirmIncludes {
	return func(includes []template.Include) ([]template.Include, error) {
		var enabled []template.Include
		for _, inc := range includes {
			isEnabled := inc.EnabledByDefault
			if opts.EnabledIncludes != nil {
				if val, ok := opts.EnabledIncludes[inc.Name]; ok {
					isEnabled = val
				}
			}
			if isEnabled {
				enabled = append(enabled, inc)
			}
		}
		return enabled, nil
	}
}

func (s *Scaffolder) collectVariables(tree *template.TemplateNode, opts Options) (template.RenderContexts, error) {
	var contexts template.RenderContexts
	var err error

	if opts.Interactive {
		contexts, err = s.collector.CollectTreeVariables(tree)
		if err != nil {
			return nil, fmt.Errorf("failed to collect variables: %w", err)
		}

		// Merge with pre-provided variables
		if opts.Variables != nil {
			for _, ctx := range contexts {
				for k, v := range opts.Variables {
					ctx.Set(k, v)
				}
			}
		}
	} else {
		contexts = s.engine.BuildContext(tree, opts.Variables)
	}

	// Validate contexts before rendering
	if err := s.engine.ValidateContexts(tree, contexts); err != nil {
		return nil, fmt.Errorf("context validation failed: %w", err)
	}

	return contexts, nil
}

func (s *Scaffolder) determineOutputDir(tree *template.TemplateNode, contexts template.RenderContexts, opts Options) string {
	if opts.OutputDir != "" {
		return opts.OutputDir
	}

	// Use project name from root template if available
	rootCtx := contexts[tree.Template.Name]
	projectName, err := tree.Template.ProjectName(rootCtx)
	if err == nil {
		return projectName
	}

	return "."
}

func (s *Scaffolder) render(tree *template.TemplateNode, contexts template.RenderContexts) ([]template.RenderedFile, error) {
	renderedFiles, err := s.engine.RenderNode(tree, contexts)
	if err != nil {
		return nil, fmt.Errorf("failed to render template tree: %w", err)
	}
	return renderedFiles, nil
}

func (s *Scaffolder) writeFiles(renderedFiles []template.RenderedFile, outputDir string, opts Options) ([]string, []string, error) {
	written := make([]string, 0)
	skipped := make([]string, 0)

	if opts.DryRun {
		return written, skipped, nil
	}

	for _, file := range renderedFiles {
		fullPath := filepath.Join(outputDir, file.Path)

		// Check if file exists
		if _, err := os.Stat(fullPath); err == nil && !opts.Overwrite {
			skipped = append(skipped, file.Path)
			continue
		}

		// Write the file
		if err := s.writer.WriteFile(fullPath, file.Content); err != nil {
			return nil, nil, fmt.Errorf("failed to write file %s: %w", file.Path, err)
		}

		written = append(written, file.Path)
	}

	return written, skipped, nil
}
