package scaffold

// Variables holds variable mappings at different scopes for template rendering.
type Variables struct {
	// Global holds variables that apply to all templates
	Global map[string]string

	// NameSpecific holds variables specific to a named template
	NameSpecific map[string]map[string]string

	// NodeSpecific holds variables specific to a node in the template tree
	NodeSpecific map[string]map[string]string
}
