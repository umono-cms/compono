package compono

import (
	"strings"

	"github.com/umono-cms/compono/ast"
	"github.com/umono-cms/compono/util"
)

// Directive is a directive given as a literal to a parameter of a built-in:
// a value of the parameter that starts with "__".
type Directive struct {
	Builtin string // "LINK"
	Param   string // "current"
	Value   string // "__auto", trimmed
	In      string // the local the call is in, "" outside locals
}

// directiveParams are the parameters of the built-ins whose values that start
// with "__" are directives.
var directiveParams = map[string][]string{"LINK": {"current"}}

// analysisDirectives returns the directives given as literals in the
// references refs, whose nodes are the keys of refNodes, nil if there are none.
func analysisDirectives(refs []Ref, refNodes map[ast.Node]int) []Directive {
	nodes := make([]ast.Node, len(refs))
	for node, i := range refNodes {
		nodes[i] = node
	}

	var directives []Directive
	for i, ref := range refs {
		params, ok := directiveParams[ref.Name]
		if !ok || ref.Local {
			continue
		}
		argsRule := "comp-bound-args"
		if ref.Kind == "call" {
			argsRule = "comp-call-args"
		}
		args := ast.FindNodeByRuleName(nodes[i].Children(), argsRule)
		if args == nil {
			continue
		}
		for _, arg := range args.Children() {
			if !ast.IsRuleName(arg, "comp-call-arg") {
				continue
			}
			name := ast.GetArgNameFromCompCallArg(arg)
			if !util.InSliceString(name, params) {
				continue
			}
			value, ok := literalStringArg(arg)
			if !ok || !strings.HasPrefix(value, "__") {
				continue
			}
			directives = append(directives, Directive{Builtin: ref.Name, Param: name, Value: value, In: ref.In})
		}
	}
	return directives
}

// literalStringArg returns the trimmed value of arg when it is a string
// literal.
func literalStringArg(arg ast.Node) (string, bool) {
	typ := ast.FindNodeByRuleName(arg.Children(), "comp-call-arg-type")
	if typ == nil {
		return "", false
	}
	str := ast.FindNodeByRuleName(typ.Children(), "comp-call-string-arg")
	if str == nil {
		return "", false
	}
	raw := ""
	if value := ast.FindNodeByRuleName(str.Children(), "comp-call-arg-value"); value != nil {
		raw = strings.ReplaceAll(string(value.Raw()), quote+quote, quote)
	}
	return util.TrimWhitespace(raw), true
}
