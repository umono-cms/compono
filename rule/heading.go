package rule

import "github.com/umono-cms/compono/selector"

func headingSelectors(prefix string) []selector.Selector {
	se, _ := selector.NewStartEnd(prefix+` (\t| )*`, `\n|\z`)
	return []selector.Selector{
		selector.NewFilter(se, func(source []byte, index [][2]int) [][2]int {
			if len(index) == 0 {
				return [][2]int{}
			}

			filtered := [][2]int{}

		outer:
			for _, ind := range index {
				start := ind[0]

				for i := start - 1; i >= 0 && source[i] != '\n'; i-- {
					if source[i] != ' ' && source[i] != '\t' {
						continue outer
					}
				}

				filtered = append(filtered, ind)
			}

			return filtered
		}),
	}
}

type h1 struct{}

func newH1() Rule {
	return &h1{}
}

func (_ *h1) Name() string {
	return "h1"
}

func (_ *h1) Selectors() []selector.Selector {
	return headingSelectors("#")
}

func (_ *h1) Rules() []Rule {
	return []Rule{
		newH1Content(),
	}
}

type h1Content struct{}

func newH1Content() Rule {
	return &h1Content{}
}

func (_ *h1Content) Name() string {
	return "h1-content"
}

func (_ *h1Content) Selectors() []selector.Selector {
	return []selector.Selector{
		selector.NewStartEndInner(`#\s+`, `\n|\z`),
	}
}

func (_ *h1Content) Rules() []Rule {
	return []Rule{
		newLink(),
		newEm(),
		newStrong(),
		newInlineCode(),
		newInlineCompCall(),
		newInlineParamCompCall(),
		newParamRef(),
		newPlain(),
	}
}

type h2 struct{}

func newH2() Rule {
	return &h2{}
}

func (_ *h2) Name() string {
	return "h2"
}

func (_ *h2) Selectors() []selector.Selector {
	return headingSelectors("##")
}

func (_ *h2) Rules() []Rule {
	return []Rule{
		newH2Content(),
	}
}

type h2Content struct{}

func newH2Content() Rule {
	return &h2Content{}
}

func (_ *h2Content) Name() string {
	return "h2-content"
}

func (_ *h2Content) Selectors() []selector.Selector {
	return []selector.Selector{
		selector.NewStartEndInner(`##\s+`, `\n|\z`),
	}
}

func (_ *h2Content) Rules() []Rule {
	return []Rule{
		newLink(),
		newStrong(),
		newEm(),
		newInlineCode(),
		newInlineCompCall(),
		newInlineParamCompCall(),
		newParamRef(),
		newPlain(),
	}
}

type h3 struct{}

func newH3() Rule {
	return &h3{}
}

func (_ *h3) Name() string {
	return "h3"
}

func (_ *h3) Selectors() []selector.Selector {
	return headingSelectors("###")
}

func (_ *h3) Rules() []Rule {
	return []Rule{
		newH3Content(),
	}
}

type h3Content struct{}

func newH3Content() Rule {
	return &h3Content{}
}

func (_ *h3Content) Name() string {
	return "h3-content"
}

func (_ *h3Content) Selectors() []selector.Selector {
	return []selector.Selector{
		selector.NewStartEndInner(`###\s+`, `\n|\z`),
	}
}

func (_ *h3Content) Rules() []Rule {
	return []Rule{
		newLink(),
		newStrong(),
		newEm(),
		newInlineCode(),
		newInlineCompCall(),
		newInlineParamCompCall(),
		newParamRef(),
		newPlain(),
	}
}

type h4 struct{}

func newH4() Rule {
	return &h4{}
}

func (_ *h4) Name() string {
	return "h4"
}

func (_ *h4) Selectors() []selector.Selector {
	return headingSelectors("####")
}

func (_ *h4) Rules() []Rule {
	return []Rule{
		newH4Content(),
	}
}

type h4Content struct{}

func newH4Content() Rule {
	return &h4Content{}
}

func (_ *h4Content) Name() string {
	return "h4-content"
}

func (_ *h4Content) Selectors() []selector.Selector {
	return []selector.Selector{
		selector.NewStartEndInner(`####\s+`, `\n|\z`),
	}
}

func (_ *h4Content) Rules() []Rule {
	return []Rule{
		newLink(),
		newStrong(),
		newEm(),
		newInlineCode(),
		newInlineCompCall(),
		newInlineParamCompCall(),
		newParamRef(),
		newPlain(),
	}
}

type h5 struct{}

func newH5() Rule {
	return &h5{}
}

func (_ *h5) Name() string {
	return "h5"
}

func (_ *h5) Selectors() []selector.Selector {
	return headingSelectors("#####")
}

func (_ *h5) Rules() []Rule {
	return []Rule{
		newH5Content(),
	}
}

type h5Content struct{}

func newH5Content() Rule {
	return &h5Content{}
}

func (_ *h5Content) Name() string {
	return "h5-content"
}

func (_ *h5Content) Selectors() []selector.Selector {
	return []selector.Selector{
		selector.NewStartEndInner(`#####\s+`, `\n|\z`),
	}
}

func (_ *h5Content) Rules() []Rule {
	return []Rule{
		newLink(),
		newStrong(),
		newEm(),
		newInlineCode(),
		newInlineCompCall(),
		newInlineParamCompCall(),
		newParamRef(),
		newPlain(),
	}
}

type h6 struct{}

func newH6() Rule {
	return &h6{}
}

func (_ *h6) Name() string {
	return "h6"
}

func (_ *h6) Selectors() []selector.Selector {
	return headingSelectors("######")
}

func (_ *h6) Rules() []Rule {
	return []Rule{
		newH6Content(),
	}
}

type h6Content struct{}

func newH6Content() Rule {
	return &h6Content{}
}

func (_ *h6Content) Name() string {
	return "h6-content"
}

func (_ *h6Content) Selectors() []selector.Selector {
	return []selector.Selector{
		selector.NewStartEndInner(`######\s+`, `\n|\z`),
	}
}

func (_ *h6Content) Rules() []Rule {
	return []Rule{
		newLink(),
		newStrong(),
		newEm(),
		newInlineCode(),
		newInlineCompCall(),
		newInlineParamCompCall(),
		newParamRef(),
		newPlain(),
	}
}
