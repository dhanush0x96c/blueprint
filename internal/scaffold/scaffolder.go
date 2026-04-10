package scaffold

import (
	"fmt"
	"path/filepath"

	"github.com/dhanush0x96c/blueprint/internal/prompt"
	"github.com/dhanush0x96c/blueprint/internal/template"
	"github.com/dhanush0x96c/blueprint/internal/vars"
)

// Scaffolder orchestrates the complete scaffolding process
type Scaffolder struct {
	engine       *template.Engine
	promptEngine *prompt.Engine
	writer       *Writer
}

// NewScaffolder creates a new scaffolder with the given template resolver.
func NewScaffolder(resolver template.Resolver) *Scaffolder {
	return &Scaffolder{
		engine:       template.NewEngine(resolver),
		promptEngine: prompt.NewEngine(),
		writer:       NewWriter(),
	}
}

// Options contains options for scaffolding
type Options struct {
	TemplateRef     template.TemplateRef // Template reference to scaffold
	OutputDir       string               // Output directory for scaffolded files
	Variables       vars.Variables       // Pre-provided variables
	EnabledIncludes map[string]bool      // Pre-selected includes (skip prompt)
	Interactive     bool                 // Whether to prompt for variables
	DryRun          bool                 // If true, don't write files
	Overwrite       bool                 // Whether to overwrite existing files
}

// Result contains the results of a scaffolding operation
type Result struct {
	FilesWritten []string            // List of files written
	FilesSkipped []string            // List of files skipped (already exist)
	Dependencies []string            // Dependencies that need to be installed
	PostInitCmds []template.PostInit // Post-init commands to run
}

// Scaffold performs the complete scaffolding operation
func (s *Scaffolder) Scaffold(opts Options) (*Result, error) {
	tree, err := s.resolveTemplateTree(opts)
	if err != nil {
		return nil, err
	}

	contexts, err := s.collectVariables(tree, opts)
	if err != nil {
		return nil, err
	}

	outputDir := s.determineOutputDir(opts)

	renderResult, err := s.render(tree, contexts)
	if err != nil {
		return nil, err
	}

	written, skipped, err := s.writeFiles(tree, renderResult, contexts, outputDir, opts)
	if err != nil {
		return nil, err
	}

	return &Result{
		FilesWritten: written,
		FilesSkipped: skipped,
		Dependencies: tree.AllDependencies(),
		PostInitCmds: tree.AllPostInit(),
	}, nil
}

func (s *Scaffolder) resolveTemplateTree(opts Options) (*template.TemplateNode, error) {
	var confirm template.ConfirmIncludes
	if opts.Interactive {
		confirm = s.promptEngine.PromptIncludes
	} else {
		confirm = s.confirmIncludesFromOptions(opts.EnabledIncludes)
	}

	tree, err := s.engine.GetFullTree(opts.TemplateRef, confirm)
	if err != nil {
		return nil, err
	}

	return tree, nil
}

func (s *Scaffolder) confirmIncludesFromOptions(enabledIncludes map[string]bool) template.ConfirmIncludes {
	return func(includes []template.Include) ([]template.Include, error) {
		var enabled []template.Include
		for _, inc := range includes {
			isEnabled := inc.EnabledByDefault
			if enabledIncludes != nil {
				if val, ok := enabledIncludes[inc.Name]; ok {
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
	pipeline := newVariablePipeline(tree, s.engine, s.promptEngine, opts)
	return pipeline.Collect()
}

func (s *Scaffolder) determineOutputDir(opts Options) string {
	if opts.OutputDir != "" {
		return opts.OutputDir
	}

	return "."
}

func (s *Scaffolder) render(tree *template.TemplateNode, contexts template.RenderContexts) (*template.RenderResult, error) {
	renderResult, err := s.engine.RenderNode(tree, contexts)
	if err != nil {
		return nil, fmt.Errorf("failed to render template tree: %w", err)
	}
	return renderResult, nil
}

func (s *Scaffolder) writeFiles(
	tree *template.TemplateNode,
	renderResult *template.RenderResult,
	contexts template.RenderContexts,
	outputDir string,
	opts Options,
) ([]string, []string, error) {
	written := make([]string, 0)
	skipped := make([]string, 0)

	if opts.DryRun {
		return written, skipped, nil
	}

	if err := s.writeNode(tree, renderResult, contexts, outputDir, opts, &written, &skipped); err != nil {
		return nil, nil, err
	}

	return written, skipped, nil
}

func (s *Scaffolder) writeNode(
	node *template.TemplateNode,
	renderResult *template.RenderResult,
	contexts template.RenderContexts,
	outputDir string,
	opts Options,
	written *[]string,
	skipped *[]string,
) error {
	nodeOutputDir := s.resolveNodeOutputDir(node, contexts, outputDir)

	files, ok := renderResult.Files[node.ID]
	if ok {
		writeResult, err := s.writer.WriteFiles(nodeOutputDir, files, opts.Overwrite)
		if err != nil {
			return err
		}
		*written = append(*written, writeResult.Written...)
		*skipped = append(*skipped, writeResult.Skipped...)
	}

	for _, child := range node.Children {
		if err := s.writeNode(child, renderResult, contexts, nodeOutputDir, opts, written, skipped); err != nil {
			return err
		}
	}
	return nil
}

func (s *Scaffolder) resolveNodeOutputDir(
	node *template.TemplateNode,
	contexts template.RenderContexts,
	parentDir string,
) string {
	mount := node.Mount
	if mount == "" && node.Template.Type == template.TypeProject {
		ctx := contexts[node.ID]
		projectName, err := node.Template.ProjectName(ctx)
		if err == nil {
			mount = projectName
		}
	}

	if mount != "" {
		return filepath.Join(parentDir, mount)
	}

	return parentDir
}
