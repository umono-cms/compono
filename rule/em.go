package rule

import "github.com/umono-cms/compono/selector"

type em struct{}

func newEm() Rule {
	return &em{}
}

func (_ *em) Name() string {
	return "em"
}

func (_ *em) Selectors() []selector.Selector {
	seSelector, _ := selector.NewStartEnd(`\*[^\s\*]`, `[^\s\*]\*`)
	return []selector.Selector{
		seSelector,
	}
}

func (_ *em) Rules() []Rule {
	return []Rule{
		newEmContent(),
	}
}

type emContent struct{}

func newEmContent() Rule {
	return &emContent{}
}

func (_ *emContent) Name() string {
	return "em-content"
}

func (_ *emContent) Selectors() []selector.Selector {
	return []selector.Selector{
		selector.NewStartEndInner(`\*`, `\*`),
	}
}

func (_ *emContent) Rules() []Rule {
	return []Rule{
		newInlineCompCall(),
		newInlineParamCompCall(),
		newParamRef(),
		newPlain(),
	}
}
