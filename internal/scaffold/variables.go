package scaffold

import (
	"fmt"

	"github.com/dhanush0x96c/blueprint/internal/prompt"
	"github.com/dhanush0x96c/blueprint/internal/template"
	"github.com/dhanush0x96c/blueprint/internal/vars"
)

type variablePipeline struct {
	tree         *template.TemplateNode
	engine       *template.Engine
	promptEngine *prompt.Engine
	opts         Options
}

func newVariablePipeline(
	tree *template.TemplateNode,
	engine *template.Engine,
	promptEngine *prompt.Engine,
	opts Options,
) *variablePipeline {
	return &variablePipeline{
		tree:         tree,
		engine:       engine,
		promptEngine: promptEngine,
		opts:         opts,
	}
}

func (p *variablePipeline) Collect() (template.RenderContexts, error) {
	contexts := make(template.RenderContexts)

	for _, collector := range p.collectors() {
		if err := collector.Collect(contexts); err != nil {
			return nil, fmt.Errorf("failed to collect variables: %w", err)
		}
	}

	vars.ApplyInheritance(p.tree, contexts)

	if err := p.engine.ValidateContexts(p.tree, contexts); err != nil {
		return nil, fmt.Errorf("context validation failed: %w", err)
	}

	return contexts, nil
}

func (p *variablePipeline) collectors() []vars.Collector {
	collectors := []vars.Collector{
		vars.NewDefaultCollector(p.tree),
		vars.NewCLICollector(p.tree, p.opts.Variables),
	}

	if p.opts.Interactive {
		collectors = append(collectors, vars.NewPromptCollector(p.tree, p.promptEngine))
	}

	return collectors
}
