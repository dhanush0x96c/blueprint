package vars

import "github.com/dhanush0x96c/blueprint/internal/template"

func ensureContext(contexts template.RenderContexts, nodeID string) *template.Context {
	if ctx, ok := contexts[nodeID]; ok {
		return ctx
	}

	ctx := template.NewTemplateContext(make(map[string]any))
	contexts[nodeID] = ctx
	return ctx
}

func walk(node *template.TemplateNode, fn func(*template.TemplateNode) error) error {
	if err := fn(node); err != nil {
		return err
	}

	for _, child := range node.Children {
		if err := walk(child, fn); err != nil {
			return err
		}
	}

	return nil
}
