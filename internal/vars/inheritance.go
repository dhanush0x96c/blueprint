package vars

import "github.com/dhanush0x96c/blueprint/internal/template"

func ApplyInheritance(tree *template.TemplateNode, contexts template.RenderContexts) {
	applyInheritance(tree, "", contexts)
}

func applyInheritance(node *template.TemplateNode, parentID string, contexts template.RenderContexts) {
	ctx := ensureContext(contexts, node.ID)
	if parentID != "" && len(node.Inherited) > 0 {
		if parentCtx, ok := contexts[parentID]; ok {
			for childVar, parentVar := range node.Inherited {
				if value, ok := parentCtx.Get(parentVar); ok {
					ctx.Set(childVar, value)
				}
			}
		}
	}

	for _, child := range node.Children {
		applyInheritance(child, node.ID, contexts)
	}
}
