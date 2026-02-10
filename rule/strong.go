package rule

import "github.com/umono-cms/compono/selector"

type strong struct{}

func newStrong() Rule {
	return &strong{}
}

func (_ *strong) Name() string {
	return "strong"
}

func (_ *strong) Selectors() []selector.Selector {
	seSelector, _ := selector.NewStartEnd(`\*\*[^\s]`, `[^\s]\*\*`)
	return []selector.Selector{
		seSelector,
	}
}

func (_ *strong) Rules() []Rule {
	return []Rule{
		newStrongContent(),
	}
}

type strongContent struct{}

func newStrongContent() Rule {
	return &strongContent{}
}

func (sc *strongContent) Name() string {
	return "strong-content"
}

func (sc *strongContent) Selectors() []selector.Selector {
	return []selector.Selector{
		selector.NewStartEndInner(`\*\*`, `\*\*`),
	}
}

func (_ *strongContent) Rules() []Rule {
	return []Rule{
		newInlineCompCall(),
		newInlineParamCompCall(),
		newParamRef(),
		newPlain(),
	}
}
