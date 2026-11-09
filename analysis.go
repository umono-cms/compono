package compono

import (
	"sort"
	"strings"

	"github.com/umono-cms/compono/ast"
	"github.com/umono-cms/compono/builtin"
	"github.com/umono-cms/compono/errwrap"
)

// Analysis holds the facts of a source that do not depend on rendering,
// on values or on global components. Every list is in source order.
type Analysis struct {
	// Params is the parameter line of a global source, nil if there is none.
	Params []Param
	Locals []Local
	// Refs are the references to components, those inside locals included.
	Refs []Ref
	// Elements are the Markdown elements used in the source.
	Elements []Element
	// Root holds the units of the root, local definitions excluded.
	Root []Unit
	// Directives are the directives given as literals to built-in
	// parameters, in source order, those in locals included.
	Directives []Directive
	// Diagnostics are the errors that do not depend on values, those in
	// uncalled locals included. Source and Calls are empty.
	Diagnostics []Diagnostic
}

// ParamType is the type of a parameter: "string", "integer", "bool",
// "component", "array", "record" or "context".
type ParamType string

// Param is a parameter definition.
type Param struct {
	Name     string
	Type     ParamType
	Required bool
	Range    Range
}

// Local is a local component definition.
type Local struct {
	Name   string
	Params []Param
	Range  Range
}

// RefKind is the kind of a reference: "call" for a {{ X … }} call, "value"
// for a component given as a value (an argument, a bound argument, an array
// element or a record value) and "default" for a component written in a
// parameter definition.
type RefKind string

// Ref is a reference to a component.
type Ref struct {
	Name string
	Kind RefKind
	// In is the name of the local the reference is in, "" outside locals.
	In string
	// Local reports whether the name resolves to a local of the source.
	Local bool
	// Args are the names of the arguments given or bound, nil if none.
	Args []string
	// Block reports whether a call is written as a block. Only for "call".
	Block bool
	Range Range
}

// Element is a use of a Markdown element.
type Element struct {
	// Name is one of MarkdownElements.
	Name  string
	In    string
	Range Range
}

// UnitKind is the kind of a root unit: "call", "markdown" or "text".
type UnitKind string

// Unit is a unit of the root.
type Unit struct {
	Kind UnitKind
	// Ref is the index of the call in Refs when Kind is "call", -1 otherwise.
	Ref   int
	Range Range
}

// Builtins returns the names of the built-in components. It returns a new
// slice on every call.
func Builtins() []string {
	defs := builtin.BuiltinComponents()
	names := make([]string, 0, len(defs))
	for _, def := range defs {
		names = append(names, def.Name)
	}
	return names
}

// MarkdownElements returns the names of the Markdown elements, the names of
// their spec directories. It returns a new slice on every call.
func MarkdownElements() []string {
	return []string{"blockquote", "br", "code-block", "em", "heading", "hr", "inline-code", "link", "list", "p", "strong", "table"}
}

// Analyze parses source without rendering it and returns its facts. Global
// components are not given: a name that resolves to no local is left to the
// caller, so unknown-component is not reported, and neither is unknown-key,
// since no context is given. The error is only for fatal errors.
func Analyze(source []byte) (Analysis, error) {
	if len(source) == 0 {
		return Analysis{}, nil
	}

	top, root := analysisTree(source)
	refs, refNodes := analysisRefs(source, top)
	analysis := Analysis{
		Locals:   analysisLocals(source, top),
		Refs:     refs,
		Elements: analysisElements(source, top),
		Root:     analysisRoot(source, top, refNodes),
	}
	analysis.Directives = analysisDirectives(refs, refNodes)
	if head := ast.FindNodeByRuleName(top.Children(), "global-comp-def-head"); head != nil {
		analysis.Params = analysisParams(source, ast.FindNodeByRuleName(head.Children(), "comp-params"))
		if analysis.Params == nil {
			analysis.Params = []Param{}
		}
	}

	attachBuiltins(root)
	errwrap.DefaultErrorWrapper().Wrap(root)
	diags := append(analysisDefinitionDiagnostics(source, top), analysisDiagnostics(source, top)...)
	sort.SliceStable(diags, func(i, j int) bool {
		return diags[i].Range.Start.Offset < diags[j].Range.Start.Offset
	})
	if len(diags) > 0 {
		analysis.Diagnostics = diags
	}

	return analysis, nil
}

// analysisDiagnostics returns the errors the error wrapper found under top,
// in source order, without those that depend on globals or context.
func analysisDiagnostics(source []byte, top ast.Node) []Diagnostic {
	var diags []Diagnostic
	var walk func(node ast.Node)
	walk = func(node ast.Node) {
		if ast.IsRuleNameOneOf(node, []string{"block-error", "inline-error"}) {
			title := ast.FindNodeByRuleName(node.Children(), "error-title")
			message := ast.FindNodeByRuleName(node.Children(), "error-message")
			if title == nil || message == nil {
				return
			}
			code := DiagnosticCode(strings.ReplaceAll(strings.ToLower(strings.TrimSpace(string(title.Raw()))), " ", "-"))
			if code == CodeUnknownComponent || code == CodeUnknownKey {
				return
			}
			rng := node.Range()
			diags = append(diags, Diagnostic{
				Code:    code,
				Message: strings.TrimSpace(string(message.Raw())),
				Range:   Range{Start: positionIn(source, rng.Start), End: positionIn(source, rng.End)},
			})
			return
		}
		for _, child := range node.Children() {
			walk(child)
		}
	}
	walk(top)
	return diags
}
