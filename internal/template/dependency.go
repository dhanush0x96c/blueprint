package template

// AllDependencies recursively collects and merges all dependencies from the tree.
func (n *TemplateNode) AllDependencies() []string {
	depMap := make(map[string]string)
	n.collectDependencies(depMap)

	result := make([]string, 0, len(depMap))
	for pkg, version := range depMap {
		if version != "" {
			result = append(result, pkg+"@"+version)
		} else {
			result = append(result, pkg)
		}
	}
	return result
}

func (n *TemplateNode) collectDependencies(depMap map[string]string) {
	for _, dep := range n.Template.Dependencies {
		pkg, version := parseDependency(dep)
		if existing, ok := depMap[pkg]; !ok || existing == "" {
			depMap[pkg] = version
		}
	}

	for _, child := range n.Children {
		child.collectDependencies(depMap)
	}
}

func parseDependency(dep string) (string, string) {
	for i, ch := range dep {
		if ch == '@' {
			return dep[:i], dep[i+1:]
		}
	}
	return dep, ""
}
