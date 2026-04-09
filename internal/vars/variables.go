package vars

// Variables holds variable mappings at different scopes for template rendering.
type Variables struct {
	Global map[string]string

	NameSpecific map[string]map[string]string

	NodeSpecific map[string]map[string]string
}
