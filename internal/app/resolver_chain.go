package app

type ChainResolver struct {
	resolvers []Resolver
}

func NewChainResolver(resolvers ...Resolver) *ChainResolver {
	return &ChainResolver{resolvers: resolvers}
}

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
