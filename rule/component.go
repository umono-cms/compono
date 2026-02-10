package rule

import (
	"regexp"

	"github.com/umono-cms/compono/selector"
)

// Local components definition wrapper
type localCompDefWrapper struct{}

func newLocalCompDefWrapper() Rule {
	return &localCompDefWrapper{}
}

func (_ *localCompDefWrapper) Name() string {
	return "local-comp-def-wrapper"
}

func (_ *localCompDefWrapper) Selectors() []selector.Selector {
	return []selector.Selector{
		selector.NewSinceFirstMatchInner(`\n*~\s+[A-Z0-9]+(?:_[A-Z0-9]+)*`),
	}
}

func (_ *localCompDefWrapper) Rules() []Rule {
	return []Rule{
		newLocalCompDef(),
	}
}

// Local component definition
type localCompDef struct{}

func newLocalCompDef() Rule {
	return &localCompDef{}
}

func (_ *localCompDef) Name() string {
	return "local-comp-def"
}

func (_ *localCompDef) Selectors() []selector.Selector {
	seli, _ := selector.NewStartEndLeftInner(`(?:\n|\A)~\s+[A-Z0-9]+(?:_[A-Z0-9]+)*`, `\n~\s+[A-Z0-9]+(?:_[A-Z0-9]+)*|\z`)
	return []selector.Selector{
		seli,
	}
}

func (_ *localCompDef) Rules() []Rule {
	return []Rule{
		newLocalCompDefHead(),
		newLocalCompDefContent(),
	}
}

// Local component definition head
type localCompDefHead struct{}

func newLocalCompDefHead() Rule {
	return &localCompDefHead{}
}

func (_ *localCompDefHead) Name() string {
	return "local-comp-def-head"
}

func (_ *localCompDefHead) Selectors() []selector.Selector {
	se, _ := selector.NewStartEnd(`(?:\n|\A)~\s+`, `\s*\n`)
	return []selector.Selector{
		se,
	}
}

func (_ *localCompDefHead) Rules() []Rule {
	return []Rule{
		newLocalCompName(),
		newCompParams(),
	}
}

// Local component name
type localCompName struct{}

func newLocalCompName() Rule {
	return &localCompName{}
}

func (_ *localCompName) Name() string {
	return "local-comp-name"
}

func (_ *localCompName) Selectors() []selector.Selector {
	return []selector.Selector{
		selector.NewStartEndInner(`(?:\n|\A)~\s+`, ` +|\n|\z`),
	}
}

func (_ *localCompName) Rules() []Rule {
	return []Rule{}
}

// Component parameters
type compParams struct{}

func newCompParams() Rule {
	return &compParams{}
}

func (_ *compParams) Name() string {
	return "comp-params"
}

func (_ *compParams) Selectors() []selector.Selector {
	se, _ := selector.NewStartEnd(`.`, `.`)
	p, _ := selector.NewPattern(`([a-z][a-z0-9-]*)(?:[\s\n\r]*=[\s\n\r]*(".*?"|\d+(?:\.\d+)?|true|false|[A-Z0-9]+(?:_[A-Z0-9]+)*))?`)
	return []selector.Selector{
		selector.NewBounds(se, p),
	}
}

func (_ *compParams) Rules() []Rule {
	return []Rule{
		newCompParam(),
	}
}

// Component parameter
type compParam struct{}

func newCompParam() Rule {
	return &compParam{}
}

func (_ *compParam) Name() string {
	return "comp-param"
}

func (_ *compParam) Selectors() []selector.Selector {
	p, _ := selector.NewPattern(`([a-z][a-z0-9-]*)(?:[\s\n\r]*=[\s\n\r]*(".*?"|\d+(?:\.\d+)?|true|false|[A-Z0-9]+(?:_[A-Z0-9]+)*))?`)
	return []selector.Selector{
		p,
	}
}

func (_ *compParam) Rules() []Rule {
	return []Rule{
		newCompParamName(),
		newCompParamType(),
	}
}

// Component parameter name
type compParamName struct{}

func newCompParamName() Rule {
	return &compParamName{}
}

func (_ *compParamName) Name() string {
	return "comp-param-name"
}

func (_ *compParamName) Selectors() []selector.Selector {
	seli, _ := selector.NewStartEndLeftInner(`([a-z][a-z0-9-]*)\s*`, `=`)
	return []selector.Selector{
		seli,
		selector.NewAll(),
	}
}

func (_ *compParamName) Rules() []Rule {
	return []Rule{}
}

// Component parameter type
type compParamType struct{}

func newCompParamType() Rule {
	return &compParamType{}
}

func (_ *compParamType) Name() string {
	return "comp-param-type"
}

func (_ *compParamType) Selectors() []selector.Selector {
	p, _ := selector.NewPattern(`[\s\n\r]*(".*?"|\d+(?:\.\d+)?|true|false|[A-Z0-9]+(?:_[A-Z0-9]+)*)`)
	return []selector.Selector{
		p,
	}
}

func (_ *compParamType) Rules() []Rule {
	return []Rule{
		newCompStringParam(),
		newCompNumberParam(),
		newCompBoolParam(),
		newCompCompParam(),
	}
}

// Component's string parameter
type compStringParam struct{}

func newCompStringParam() Rule {
	return &compStringParam{}
}

func (_ *compStringParam) Name() string {
	return "comp-string-param"
}

func (_ *compStringParam) Selectors() []selector.Selector {
	return []selector.Selector{
		selector.NewStartEndInner(`[\s\n\r]*"`, `"[\s\n\r]*`),
	}
}

func (_ *compStringParam) Rules() []Rule {
	return []Rule{
		newCompParamDefaValue(),
	}
}

// Component's number parameter
type compNumberParam struct{}

func newCompNumberParam() Rule {
	return &compNumberParam{}
}

func (_ *compNumberParam) Name() string {
	return "comp-number-param"
}

func (_ *compNumberParam) Selectors() []selector.Selector {
	p, _ := selector.NewPattern(`\d+(?:\.\d+)?`)
	return []selector.Selector{
		p,
	}
}

func (_ *compNumberParam) Rules() []Rule {
	return []Rule{
		newCompParamDefaValue(),
	}
}

// Component's bool parameter
type compBoolParam struct{}

func newCompBoolParam() Rule {
	return &compBoolParam{}
}

func (_ *compBoolParam) Name() string {
	return "comp-bool-param"
}

func (_ *compBoolParam) Selectors() []selector.Selector {
	p, _ := selector.NewPattern(`true|false`)
	return []selector.Selector{
		p,
	}
}

func (_ *compBoolParam) Rules() []Rule {
	return []Rule{
		newCompParamDefaValue(),
	}
}

// Component's component parameter (SCREAMING_CASE default)
type compCompParam struct{}

func newCompCompParam() Rule {
	return &compCompParam{}
}

func (_ *compCompParam) Name() string {
	return "comp-comp-param"
}

func (_ *compCompParam) Selectors() []selector.Selector {
	p, _ := selector.NewPattern(`[A-Z0-9]+(?:_[A-Z0-9]+)*`)
	return []selector.Selector{
		p,
	}
}

func (_ *compCompParam) Rules() []Rule {
	return []Rule{
		newCompParamDefaValue(),
	}
}

// Component parameter default value
type compParamDefaValue struct{}

func newCompParamDefaValue() Rule {
	return &compParamDefaValue{}
}

func (_ *compParamDefaValue) Name() string {
	return "comp-param-defa-value"
}

func (_ *compParamDefaValue) Selectors() []selector.Selector {
	return []selector.Selector{
		selector.NewAll(),
	}
}

func (_ *compParamDefaValue) Rules() []Rule {
	return []Rule{}
}

// Local component definition content
type localCompDefContent struct{}

func newLocalCompDefContent() Rule {
	return &localCompDefContent{}
}

func (_ *localCompDefContent) Name() string {
	return "local-comp-def-content"
}

func (_ *localCompDefContent) Selectors() []selector.Selector {
	seli, _ := selector.NewStartEndLeftInner(`^`, `\n~\s+[A-Z0-9]+(?:_[A-Z0-9]+)*|\z`)
	return []selector.Selector{
		seli,
	}
}

func (_ *localCompDefContent) Rules() []Rule {
	return []Rule{
		newCodeBlock(),
		newH6(),
		newH5(),
		newH4(),
		newH3(),
		newH2(),
		newH1(),
		newBlockCompCall(),
		newBlockParamCompCall(),
		newP(),
	}
}

// Parameter reference
type paramRef struct{}

func newParamRef() Rule {
	return &paramRef{}
}

func (_ *paramRef) Name() string {
	return "param-ref"
}

func (_ *paramRef) Selectors() []selector.Selector {
	se, _ := selector.NewStartEnd(`\{\{\s*[a-z][a-z0-9-]*`, `\s*\}\}`)
	return []selector.Selector{
		se,
	}
}

func (_ *paramRef) Rules() []Rule {
	return []Rule{
		newParamRefName(),
	}
}

// Parameter reference's name
type paramRefName struct{}

func newParamRefName() Rule {
	return &paramRefName{}
}

func (_ *paramRefName) Name() string {
	return "param-ref-name"
}

func (_ *paramRefName) Selectors() []selector.Selector {
	sei := selector.NewStartEndInner(`\{\{\s*`, `\s*\}\}`)
	return []selector.Selector{
		sei,
	}
}

func (_ *paramRefName) Rules() []Rule {
	return []Rule{}
}

// Block component call
type blockCompCall struct {
	*compCall
}

func newBlockCompCall() Rule {
	cc := newCompCall()
	return &blockCompCall{
		compCall: cc.(*compCall),
	}
}

func (_ *blockCompCall) Name() string {
	return "block-comp-call"
}

func (_ *blockCompCall) Selectors() []selector.Selector {
	se, _ := selector.NewStartEnd(`\{\{\s*[A-Z0-9]+(?:_[A-Z0-9]+)*`, `\s*\}\}`)
	return []selector.Selector{
		selector.NewFilter(se, func(source []byte, index [][2]int) [][2]int {
			if len(index) == 0 {
				return [][2]int{}
			}

			filtered := [][2]int{}

			for _, ind := range index {
				start, end := ind[0], ind[1]

				leftOK := true
				for i := start - 1; i >= 0 && source[i] != '\n'; i-- {
					if source[i] != ' ' && source[i] != '\t' {
						leftOK = false
						break
					}
				}

				rightOK := true
				for i := end; i < len(source) && source[i] != '\n'; i++ {
					if source[i] != ' ' && source[i] != '\t' {
						rightOK = false
						break
					}
				}

				insideOK := true
				re := regexp.MustCompile(`}}`)
				closingInd := re.FindAllStringIndex(string(source[ind[0]:ind[1]]), -1)
				if len(closingInd) > 1 {
					insideOK = false
				}

				if leftOK && rightOK && insideOK {
					filtered = append(filtered, ind)
				}
			}

			return filtered
		}),
	}
}

// Inline component call
type inlineCompCall struct {
	*compCall
}

func newInlineCompCall() Rule {
	cc := newCompCall()
	return &inlineCompCall{
		compCall: cc.(*compCall),
	}
}

func (_ *inlineCompCall) Name() string {
	return "inline-comp-call"
}

// Component call
type compCall struct{}

func newCompCall() Rule {
	return &compCall{}
}

func (_ *compCall) Name() string {
	return "comp-call"
}

func (_ *compCall) Selectors() []selector.Selector {
	seSelector, _ := selector.NewStartEnd(`\{\{\s*[A-Z0-9]+(?:_[A-Z0-9]+)*`, `\s*\}\}`)
	return []selector.Selector{
		seSelector,
	}
}

func (_ *compCall) Rules() []Rule {
	return []Rule{
		newCompCallName(),
		newCompCallArgs(),
	}
}

// Component call name
type compCallName struct{}

func newCompCallName() Rule {
	return &compCallName{}
}

func (_ *compCallName) Name() string {
	return "comp-call-name"
}

func (_ *compCallName) Selectors() []selector.Selector {
	p, _ := selector.NewPattern(`\s*[A-Z0-9]+(?:_[A-Z0-9]+)*\s*`)
	return []selector.Selector{
		selector.NewFilter(p, func(source []byte, index [][2]int) [][2]int {
			if len(index) > 0 {
				return [][2]int{index[0]}
			}
			return [][2]int{}
		}),
	}
}

func (_ *compCallName) Rules() []Rule {
	return []Rule{}
}

// Commponent call arguments
type compCallArgs struct{}

func newCompCallArgs() Rule {
	return &compCallArgs{}
}

func (_ *compCallArgs) Name() string {
	return "comp-call-args"
}

func (_ *compCallArgs) Selectors() []selector.Selector {
	p, _ := selector.NewPattern(`([a-z][a-z0-9-]*)[\s\n\r]*=[\s\n\r]*(".*?"|\d+(?:\.\d+)?|true|false|\$[a-z][a-z0-9-]*|[A-Z0-9]+(?:_[A-Z0-9]+)*)`)
	return []selector.Selector{
		selector.NewFilter(p, func(source []byte, index [][2]int) [][2]int {
			if len(index) == 0 {
				return [][2]int{}
			}

			start := index[0][0]
			end := index[0][1]

			for _, i := range index[1:] {
				if i[0] < start {
					start = i[0]
				}
				if i[1] > end {
					end = i[1]
				}
			}

			return [][2]int{{start, end}}
		}),
	}
}

func (_ *compCallArgs) Rules() []Rule {
	return []Rule{
		newCompCallArg(),
	}
}

// Component call argument
type compCallArg struct{}

func newCompCallArg() Rule {
	return &compCallArg{}
}

func (_ *compCallArg) Name() string {
	return "comp-call-arg"
}

func (_ *compCallArg) Selectors() []selector.Selector {
	p, _ := selector.NewPattern(`([a-z][a-z0-9-]*)[\s\n\r]*=[\s\n\r]*(".*?"|\d+(?:\.\d+)?|true|false|\$[a-z][a-z0-9-]*|[A-Z0-9]+(?:_[A-Z0-9]+)*)`)
	return []selector.Selector{
		p,
	}
}

func (_ *compCallArg) Rules() []Rule {
	return []Rule{
		newCompCallArgName(),
		newCompCallArgType(),
	}
}

// Component call argument name
type compCallArgName struct{}

func newCompCallArgName() Rule {
	return &compCallArgName{}
}

func (_ *compCallArgName) Name() string {
	return "comp-call-arg-name"
}

func (_ *compCallArgName) Selectors() []selector.Selector {
	seli, _ := selector.NewStartEndLeftInner(`([a-z][a-z0-9-]*)\s*`, `=`)
	return []selector.Selector{
		seli,
	}
}

func (_ *compCallArgName) Rules() []Rule {
	return []Rule{}
}

// Component call argument type
type compCallArgType struct{}

func newCompCallArgType() Rule {
	return &compCallArgType{}
}

func (_ *compCallArgType) Name() string {
	return "comp-call-arg-type"
}

func (_ *compCallArgType) Selectors() []selector.Selector {
	p, _ := selector.NewPattern(`[\s\n\r]*(".*?"|\d+(?:\.\d+)?|true|false|\$[a-z][a-z0-9-]*|[A-Z0-9]+(?:_[A-Z0-9]+)*)`)
	return []selector.Selector{
		p,
	}
}

func (_ *compCallArgType) Rules() []Rule {
	return []Rule{
		newCompCallStringArg(),
		newCompCallNumberArg(),
		newCompCallBoolArg(),
		newCompCallParamArg(),
		newCompCallCompArg(),
	}
}

// Component call's string argument
type compCallStringArg struct{}

func newCompCallStringArg() Rule {
	return &compCallStringArg{}
}

func (_ *compCallStringArg) Name() string {
	return "comp-call-string-arg"
}

func (_ *compCallStringArg) Selectors() []selector.Selector {
	return []selector.Selector{
		selector.NewStartEndInner(`[\s\n\r]*"`, `"[\s\n\r]*`),
	}
}

func (_ *compCallStringArg) Rules() []Rule {
	return []Rule{
		newCompCallArgValue(),
	}
}

// Component call's number argument
type compCallNumberArg struct{}

func newCompCallNumberArg() Rule {
	return &compCallNumberArg{}
}

func (_ *compCallNumberArg) Name() string {
	return "comp-call-number-arg"
}

func (_ *compCallNumberArg) Selectors() []selector.Selector {
	p, _ := selector.NewPattern(`\d+(?:\.\d+)?`)
	return []selector.Selector{
		p,
	}
}

func (_ *compCallNumberArg) Rules() []Rule {
	return []Rule{
		newCompCallArgValue(),
	}
}

// Component call's bool argument
type compCallBoolArg struct{}

func newCompCallBoolArg() Rule {
	return &compCallBoolArg{}
}

func (_ *compCallBoolArg) Name() string {
	return "comp-call-bool-arg"
}

func (_ *compCallBoolArg) Selectors() []selector.Selector {
	p, _ := selector.NewPattern(`true|false`)
	return []selector.Selector{
		p,
	}
}

func (_ *compCallBoolArg) Rules() []Rule {
	return []Rule{
		newCompCallArgValue(),
	}
}

// Component call's param argument
type compCallParamArg struct{}

func newCompCallParamArg() Rule {
	return &compCallParamArg{}
}

func (_ *compCallParamArg) Name() string {
	return "comp-call-param-arg"
}

func (_ *compCallParamArg) Selectors() []selector.Selector {
	return []selector.Selector{
		selector.NewStartEndInner(`[\s\n\r]*\$`, `\z`),
	}
}

func (_ *compCallParamArg) Rules() []Rule {
	return []Rule{
		newCompCallArgValue(),
	}
}

type compCallArgValue struct{}

func newCompCallArgValue() Rule {
	return &compCallArgValue{}
}

func (_ *compCallArgValue) Name() string {
	return "comp-call-arg-value"
}

func (_ *compCallArgValue) Selectors() []selector.Selector {
	return []selector.Selector{
		selector.NewAll(),
	}
}

func (_ *compCallArgValue) Rules() []Rule {
	return []Rule{}
}

// Component call's component argument (SCREAMING_CASE)
type compCallCompArg struct{}

func newCompCallCompArg() Rule {
	return &compCallCompArg{}
}

func (_ *compCallCompArg) Name() string {
	return "comp-call-comp-arg"
}

func (_ *compCallCompArg) Selectors() []selector.Selector {
	p, _ := selector.NewPattern(`[A-Z0-9]+(?:_[A-Z0-9]+)*`)
	return []selector.Selector{
		p,
	}
}

func (_ *compCallCompArg) Rules() []Rule {
	return []Rule{
		newCompCallArgValue(),
	}
}

// Param component call (base)
type paramCompCall struct{}

func newParamCompCall() Rule {
	return &paramCompCall{}
}

func (_ *paramCompCall) Name() string {
	return "param-comp-call"
}

func (_ *paramCompCall) Selectors() []selector.Selector {
	se, _ := selector.NewStartEnd(`\{\{\s*\$[a-z][a-z0-9-]*`, `\s*\}\}`)
	return []selector.Selector{
		se,
	}
}

func (_ *paramCompCall) Rules() []Rule {
	return []Rule{
		newParamCompCallName(),
		newCompCallArgs(),
	}
}

// Block param component call
type blockParamCompCall struct {
	*paramCompCall
}

func newBlockParamCompCall() Rule {
	pcc := newParamCompCall()
	return &blockParamCompCall{
		paramCompCall: pcc.(*paramCompCall),
	}
}

func (_ *blockParamCompCall) Name() string {
	return "block-param-comp-call"
}

func (_ *blockParamCompCall) Selectors() []selector.Selector {
	se, _ := selector.NewStartEnd(`\{\{\s*\$[a-z][a-z0-9-]*`, `\s*\}\}`)
	return []selector.Selector{
		selector.NewFilter(se, func(source []byte, index [][2]int) [][2]int {
			if len(index) == 0 {
				return [][2]int{}
			}

			filtered := [][2]int{}

			for _, ind := range index {
				start, end := ind[0], ind[1]

				leftOK := true
				for i := start - 1; i >= 0 && source[i] != '\n'; i-- {
					if source[i] != ' ' && source[i] != '\t' {
						leftOK = false
						break
					}
				}

				rightOK := true
				for i := end; i < len(source) && source[i] != '\n'; i++ {
					if source[i] != ' ' && source[i] != '\t' {
						rightOK = false
						break
					}
				}

				insideOK := true
				re := regexp.MustCompile(`}}`)
				closingInd := re.FindAllStringIndex(string(source[ind[0]:ind[1]]), -1)
				if len(closingInd) > 1 {
					insideOK = false
				}

				if leftOK && rightOK && insideOK {
					filtered = append(filtered, ind)
				}
			}

			return filtered
		}),
	}
}

// Inline param component call
type inlineParamCompCall struct {
	*paramCompCall
}

func newInlineParamCompCall() Rule {
	pcc := newParamCompCall()
	return &inlineParamCompCall{
		paramCompCall: pcc.(*paramCompCall),
	}
}

func (_ *inlineParamCompCall) Name() string {
	return "inline-param-comp-call"
}

// Param component call name
type paramCompCallName struct{}

func newParamCompCallName() Rule {
	return &paramCompCallName{}
}

func (_ *paramCompCallName) Name() string {
	return "param-comp-call-name"
}

func (_ *paramCompCallName) Selectors() []selector.Selector {
	return []selector.Selector{
		selector.NewStartEndInner(`\{\{\s*\$`, `\s+|\s*\}\}`),
	}
}

func (_ *paramCompCallName) Rules() []Rule {
	return []Rule{}
}

// Global components definition wrapper
type globalCompDefWrapper struct{}

func NewGlobalCompDefWrapper() Rule {
	return &globalCompDefWrapper{}
}

func (_ *globalCompDefWrapper) Name() string {
	return "global-comp-def-wrapper"
}

func (_ *globalCompDefWrapper) Selectors() []selector.Selector {
	return []selector.Selector{}
}

func (_ *globalCompDefWrapper) Rules() []Rule {
	return []Rule{}
}

// Global component definition
type globalCompDef struct{}

func NewGlobalCompDef() Rule {
	return &globalCompDef{}
}

func (_ *globalCompDef) Name() string {
	return "global-comp-def"
}

func (_ *globalCompDef) Selectors() []selector.Selector {
	return []selector.Selector{
		selector.NewAll(),
	}
}

func (_ *globalCompDef) Rules() []Rule {
	return []Rule{
		NewGlobalCompDefHead(),
		newGlobalCompDefContent(),
		newLocalCompDefWrapper(),
	}
}

// Global component name
type globalCompName struct{}

func NewGlobalCompName() Rule {
	return &globalCompName{}
}

func (_ *globalCompName) Name() string {
	return "global-comp-name"
}

func (_ *globalCompName) Selectors() []selector.Selector {
	return []selector.Selector{}
}

func (_ *globalCompName) Rules() []Rule {
	return []Rule{}
}

// Global component definition head
type globalCompDefHead struct{}

func NewGlobalCompDefHead() Rule {
	return &globalCompDefHead{}
}

func (_ *globalCompDefHead) Name() string {
	return "global-comp-def-head"
}

func (_ *globalCompDefHead) Selectors() []selector.Selector {
	p, _ := selector.NewStartEnd(`^([a-z][a-z0-9-]*)[ \t\r\n]*=[ \t\r\n]*(".*?"|\d+(?:\.\d+)?|true|false)`, `\n|\z`)
	return []selector.Selector{
		p,
	}
}

func (_ *globalCompDefHead) Rules() []Rule {
	return []Rule{
		newCompParams(),
	}
}

// Global component definition content
type globalCompDefContent struct{}

func newGlobalCompDefContent() Rule {
	return &globalCompDefContent{}
}

func (_ *globalCompDefContent) Name() string {
	return "global-comp-def-content"
}

func (_ *globalCompDefContent) Selectors() []selector.Selector {
	seli, _ := selector.NewStartEndLeftInner(`^`, `\n~\s+[A-Z0-9]+(?:_[A-Z0-9]+)*|\z`)
	return []selector.Selector{
		seli,
	}
}

func (_ *globalCompDefContent) Rules() []Rule {
	return []Rule{
		newCodeBlock(),
		newH6(),
		newH5(),
		newH4(),
		newH3(),
		newH2(),
		newH1(),
		newBlockCompCall(),
		newBlockParamCompCall(),
		newP(),
	}
}
