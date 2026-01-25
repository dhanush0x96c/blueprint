package template

import (
	"fmt"
	"slices"
)

type Loader interface {
	Load(name string) (*Template, error)
}

// Composer handles resolving and merging template includes
type Composer struct {
	loader Loader
}

// NewComposer creates a new template composer with the given loader
func NewComposer(loader Loader) *Composer {
	return &Composer{
		loader: loader,
	}
}

// Compose resolves all includes for a template and returns a fully merged template
// It recursively loads included templates and merges them into a single template
func (c *Composer) Compose(tmpl *Template) (*Template, error) {
	return c.composeWithPath(tmpl, []string{tmpl.Name})
}

// composeWithPath is the internal recursive composition function that tracks the path
// to detect circular dependencies
func (c *Composer) composeWithPath(tmpl *Template, path []string) (*Template, error) {
	composed := &Template{
		Name:         tmpl.Name,
		Type:         tmpl.Type,
		Version:      tmpl.Version,
		Description:  tmpl.Description,
		Variables:    make([]Variable, len(tmpl.Variables)),
		Includes:     make([]Include, 0),
		Dependencies: make([]string, len(tmpl.Dependencies)),
		Files:        make([]File, len(tmpl.Files)),
		PostInit:     make([]PostInit, len(tmpl.PostInit)),
	}

	copy(composed.Variables, tmpl.Variables)
	copy(composed.Dependencies, tmpl.Dependencies)
	copy(composed.Files, tmpl.Files)
	copy(composed.PostInit, tmpl.PostInit)

	for _, inc := range tmpl.Includes {
		if slices.Contains(path, inc.Template) {
			return nil, fmt.Errorf("circular dependency detected: %v -> %s", path, inc.Template)
		}

		includedTmpl, err := c.loader.Load(inc.Template)
		if err != nil {
			return nil, fmt.Errorf("failed to load included template '%s': %w", inc.Template, err)
		}

		newPath := append(slices.Clone(path), inc.Template)
		resolvedInclude, err := c.composeWithPath(includedTmpl, newPath)
		if err != nil {
			return nil, err
		}

		c.mergeTemplate(composed, resolvedInclude)
	}

	return composed, nil
}

// mergeTemplate merges the source template into the destination template
func (c *Composer) mergeTemplate(dst, src *Template) {
	existingVars := make(map[string]bool)
	for _, v := range dst.Variables {
		existingVars[v.Name] = true
	}

	for _, v := range src.Variables {
		if !existingVars[v.Name] {
			dst.Variables = append(dst.Variables, v)
			existingVars[v.Name] = true
		}
	}

	dst.Dependencies = c.mergeDependencies(dst.Dependencies, src.Dependencies)

	existingDests := make(map[string]bool)
	for _, f := range dst.Files {
		existingDests[f.Dest] = true
	}

	for _, f := range src.Files {
		if !existingDests[f.Dest] {
			dst.Files = append(dst.Files, f)
			existingDests[f.Dest] = true
		}
	}

	dst.PostInit = append(dst.PostInit, src.PostInit...)
}

// mergeDependencies merges two dependency lists and deduplicates them
// Handles dependencies in the format "package@version" or just "package"
func (c *Composer) mergeDependencies(dst, src []string) []string {
	depMap := make(map[string]string)

	for _, dep := range dst {
		pkg, version := c.parseDependency(dep)
		depMap[pkg] = version
	}

	for _, dep := range src {
		pkg, version := c.parseDependency(dep)
		if existing, ok := depMap[pkg]; !ok || existing == "" {
			depMap[pkg] = version
		}
	}

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

// parseDependency parses a dependency string into package and version
// Returns (package, version) where version may be empty
func (c *Composer) parseDependency(dep string) (string, string) {
	for i, ch := range dep {
		if ch == '@' {
			return dep[:i], dep[i+1:]
		}
	}
	return dep, ""
}

// ComposeWithEnabledIncludes composes a template but allows filtering includes
// based on user selection (respecting enabled_by_default)
func (c *Composer) ComposeWithEnabledIncludes(tmpl *Template, enabledIncludes map[string]bool) (*Template, error) {
	filtered := &Template{
		Name:         tmpl.Name,
		Type:         tmpl.Type,
		Version:      tmpl.Version,
		Description:  tmpl.Description,
		Variables:    tmpl.Variables,
		Includes:     make([]Include, 0),
		Dependencies: tmpl.Dependencies,
		Files:        tmpl.Files,
		PostInit:     tmpl.PostInit,
	}

	// Filter includes based on enabled map
	for _, inc := range tmpl.Includes {
		enabled, exists := enabledIncludes[inc.Template]
		if exists && enabled {
			filtered.Includes = append(filtered.Includes, inc)
		} else if !exists && inc.EnabledByDefault {
			filtered.Includes = append(filtered.Includes, inc)
		}
	}

	return c.Compose(filtered)
}

// GetAllIncludes returns all includes (direct and transitive) for a template
// This is useful for prompting users about which features to enable
func (c *Composer) GetAllIncludes(tmpl *Template) ([]Include, error) {
	return c.getAllIncludesWithPath(tmpl, []string{tmpl.Name})
}

// getAllIncludesWithPath recursively collects all includes
func (c *Composer) getAllIncludesWithPath(tmpl *Template, path []string) ([]Include, error) {
	allIncludes := make([]Include, 0)
	seen := make(map[string]bool)

	for _, inc := range tmpl.Includes {
		if slices.Contains(path, inc.Template) {
			return nil, fmt.Errorf("circular dependency detected: %v -> %s", path, inc.Template)
		}

		// Add to result if not seen
		if !seen[inc.Template] {
			allIncludes = append(allIncludes, inc)
			seen[inc.Template] = true
		}

		includedTmpl, err := c.loader.Load(inc.Template)
		if err != nil {
			return nil, fmt.Errorf("failed to load included template '%s': %w", inc.Template, err)
		}

		newPath := append(slices.Clone(path), inc.Template)
		transitiveIncludes, err := c.getAllIncludesWithPath(includedTmpl, newPath)
		if err != nil {
			return nil, err
		}

		for _, transInc := range transitiveIncludes {
			if !seen[transInc.Template] {
				allIncludes = append(allIncludes, transInc)
				seen[transInc.Template] = true
			}
		}
	}

	return allIncludes, nil
}
