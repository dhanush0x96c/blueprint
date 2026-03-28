package resolver

import (
	"errors"

	"github.com/dhanush0x96c/blueprint/internal/template"
)

// ChainResolver is a resolver that chains multiple resolvers together.
type ChainResolver struct {
	resolvers []Resolver
}

// NewChainResolver creates a new chain resolver.
func NewChainResolver(resolvers ...Resolver) *ChainResolver {
	return &ChainResolver{resolvers: resolvers}
}

// Resolve resolves a template reference using the chain of resolvers.
func (c *ChainResolver) Resolve(ref TemplateRef) (*ResolvedTemplate, error) {
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
