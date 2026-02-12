package app

// ChainResolver is a resolver that chains multiple resolvers together.
type ChainResolver struct {
	resolvers []Resolver
}

// NewChainResolver creates a new chain resolver.
func NewChainResolver(resolvers ...Resolver) *ChainResolver {
	return &ChainResolver{resolvers: resolvers}
}

// Resolve resolves a template reference using the chain of resolvers.
func (c *ChainResolver) Resolve(ctx *Context, ref TemplateRef) (*ResolvedTemplate, error) {
	var lastErr error

	for _, r := range c.resolvers {
		path, err := r.Resolve(ctx, ref)
		if err == nil {
			return path, nil
		}
		lastErr = err
	}

	if lastErr != nil {
		return nil, lastErr
	}

	return nil, ErrTemplateNotFound
}
