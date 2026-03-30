package template

// PostInit represents a command to run after scaffolding
type PostInit struct {
	Command string `yaml:"command" validate:"required"`
	WorkDir string `yaml:"workdir,omitempty"`
}

// AllPostInit recursively collects all post-init commands from the tree.
func (n *TemplateNode) AllPostInit() []PostInit {
	var cmds []PostInit
	n.collectPostInit(&cmds)
	return cmds
}

func (n *TemplateNode) collectPostInit(cmds *[]PostInit) {
	*cmds = append(*cmds, n.Template.PostInit...)
	for _, child := range n.Children {
		child.collectPostInit(cmds)
	}
}
