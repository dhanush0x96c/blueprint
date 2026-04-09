package vars

import "github.com/dhanush0x96c/blueprint/internal/template"

// Collector mutates render contexts for a template tree.
type Collector interface {
	Collect(template.RenderContexts) error
}
