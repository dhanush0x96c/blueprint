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
	tmpl, err := s.engine.LoadTemplate(opts.TemplateRef)
	if err != nil {
		return nil, fmt.Errorf("failed to load template: %w", err)
	}

	// 1. Composition Stage
	var confirm template.ConfirmIncludes
	if opts.Interactive {
		confirm = s.collector.ConfirmIncludes
	} else {
		// Non-interactive: use pre-provided enabled includes or defaults
		confirm = func(includes []template.Include) ([]template.Include, error) {
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

	tree, err := s.engine.Compose(tmpl, confirm)
	if err != nil {
		return nil, fmt.Errorf("failed to compose template tree: %w", err)
	}

	// Validate tree before prompting
	if err := s.engine.ValidateTree(tree); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// 2. Prompt Stage
	var contexts template.RenderContexts
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
		// Non-interactive: use pre-provided variables for ALL nodes
		contexts = make(template.RenderContexts)
		s.fillContextsWithDefaults(tree, contexts, opts.Variables)
	}

	// Validate contexts before rendering
	if err := s.engine.ValidateContexts(tree, contexts); err != nil {
		return nil, fmt.Errorf("context validation failed: %w", err)
	}

	// Determine output directory
	if opts.OutputDir == "" {
		// Use project name from root template if available
		rootCtx := contexts[tree.Template.Name]
		projectName, err := tree.Template.ProjectName(rootCtx)
		if err != nil {
			// Fallback to name from variables or current directory
			if name, ok := opts.Variables["project_name"].(string); ok {
				opts.OutputDir = name
			} else {
				opts.OutputDir = "."
			}
		} else {
			opts.OutputDir = projectName
		}
	}

	// 3. Render Stage
	renderedFiles, err := s.engine.RenderNode(tree, contexts)
	if err != nil {
		return nil, fmt.Errorf("failed to render template tree: %w", err)
	}

	result := &Result{
		FilesWritten:  make([]string, 0),
		FilesSkipped:  make([]string, 0),
		Dependencies:  tree.AllDependencies(),
		PostInitCmds:  tree.AllPostInit(),
		RenderedFiles: renderedFiles,
	}

	// 4. Write Stage
	if !opts.DryRun {
		for _, file := range renderedFiles {
			fullPath := filepath.Join(opts.OutputDir, file.Path)

			// Check if file exists
			if _, err := os.Stat(fullPath); err == nil && !opts.Overwrite {
				result.FilesSkipped = append(result.FilesSkipped, file.Path)
				continue
			}

			// Write the file
			if err := s.writer.WriteFile(fullPath, file.Content); err != nil {
				return nil, fmt.Errorf("failed to write file %s: %w", file.Path, err)
			}

			result.FilesWritten = append(result.FilesWritten, file.Path)
		}
	}

	return result, nil
}

// TODO: parse variable keys with template prefix e.g. template_name:var_name
func (s *Scaffolder) fillContextsWithDefaults(node *template.TemplateNode, contexts template.RenderContexts, vars map[string]any) {
	if _, ok := contexts[node.Template.Name]; !ok {
		ctx := template.NewTemplateContext(make(map[string]any))
		// Set defaults from template
		for _, v := range node.Template.Variables {
			if v.Default != nil {
				ctx.Set(v.Name, v.Default)
			}
		}
		// Overwrite with provided variables
		for k, v := range vars {
			ctx.Set(k, v)
		}
		contexts[node.Template.Name] = ctx
	}

	for _, child := range node.Children {
		s.fillContextsWithDefaults(child, contexts, vars)
	}
}
