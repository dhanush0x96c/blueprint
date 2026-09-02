// Package resolver provides mechanisms to discover and resolve templates across multiple sources.
package resolver

import (
	"errors"

	"github.com/dhanush0x96c/blueprint/internal/template"
)

// Resolver resolves a template reference.
type Resolver interface {
	Resolve(ref template.Ref) (*template.ResolvedTemplate, error)
}

// ChainResolver is a resolver that chains multiple resolvers together.
type ChainResolver struct {
	resolvers []Resolver
}

// NewChainResolver creates a new chain resolver from the provided sources.
func NewChainResolver(sources ...Source) *ChainResolver {
	resolvers := make([]Resolver, 0, len(sources))
	for _, src := range sources {
		resolvers = append(resolvers, NewSourceResolver(src))
	}
	return &ChainResolver{resolvers: resolvers}
}

// Resolve resolves a template reference using the chain of resolvers.
func (c *ChainResolver) Resolve(ref template.Ref) (*template.ResolvedTemplate, error) {
	if len(c.resolvers) == 0 {
		return nil, &template.NotFoundError{Name: ref.Name}
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
