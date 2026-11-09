package compono

// CurrentResolverFunc receives the url of a LINK whose current is "__auto" and
// returns the value of current: an allowed value such as "page" or "true", or
// "".
type CurrentResolverFunc func(url string) string

type currentResolverOption struct {
	fn CurrentResolverFunc
}

// WithCurrentResolver returns a ConvertOption that registers a function
// invoked, in render order, for each LINK whose current is "__auto". It may be
// used at most once per conversion and is not permitted inside
// WithGlobalComponent. A nil function is the same as not setting the resolver.
func WithCurrentResolver(fn CurrentResolverFunc) ConvertOption {
	return &currentResolverOption{fn: fn}
}

func (o *currentResolverOption) applyConvert(_ *compono, cfg *convertConfig) error {
	if cfg.currentResolverSet {
		return NewComponoError(ErrCurrentResolverAlreadySet, "current resolver is already set: WithCurrentResolver can be used at most once per conversion")
	}
	cfg.currentResolverSet = true
	cfg.currentResolver = o.fn
	return nil
}
