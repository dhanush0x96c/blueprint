package resolver

import (
	"errors"

	"github.com/dhanush0x96c/blueprint/internal/template"
)

// ChainResolver is a resolver that chains multiple resolvers together.
type ChainResolver struct {
	resolvers []template.Resolver
}

// NewChainResolver creates a new chain resolver from the provided sources.
func NewChainResolver(sources ...Source) *ChainResolver {
	resolvers := make([]template.Resolver, 0, len(sources))
	for _, src := range sources {
		resolvers = append(resolvers, NewSourceResolver(src))
	}
	return &ChainResolver{resolvers: resolvers}
}

// Resolve resolves a template reference using the chain of resolvers.
func (c *ChainResolver) Resolve(ref template.TemplateRef) (*template.ResolvedTemplate, error) {
	if len(c.resolvers) == 0 {
		return nil, &template.TemplateNotFoundError{Name: ref.Name}
	}

	var errs []error
	for _, r := range c.resolvers {
		resolved, err := r.Resolve(ref)
		if err == nil {
			return resolved, nil
		}
		errs = append(errs, err)
	}

	return nil, errors.Join(errs...)
}
