package compono

import (
	"slices"
	"strings"

	"github.com/umono-cms/compono/ast"
	"github.com/umono-cms/compono/renderer"
)

// DiagnosticCode is the fixed name of the kind of a diagnostic. It is the
// kebab-case form of the error title. Several situations may share a code; the
// message tells them apart. A code is never renamed.
type DiagnosticCode string

const (
	CodeArrayIndexOutOfRange       DiagnosticCode = "array-index-out-of-range"
	CodeDisallowedURLScheme        DiagnosticCode = "disallowed-url-scheme"
	CodeDuplicateArgument          DiagnosticCode = "duplicate-argument"
	CodeDuplicateVariant           DiagnosticCode = "duplicate-variant"
	CodeEmptyTable                 DiagnosticCode = "empty-table"
	CodeEmptyTableHeader           DiagnosticCode = "empty-table-header"
	CodeEmptyText                  DiagnosticCode = "empty-text"
	CodeEmptyURL                   DiagnosticCode = "empty-url"
	CodeInconsistentAspectRatio    DiagnosticCode = "inconsistent-aspect-ratio"
	CodeInfiniteComponentCall      DiagnosticCode = "infinite-component-call"
	CodeInvalidBuiltinArguments    DiagnosticCode = "invalid-built-in-arguments"
	CodeInvalidCodeBlockLanguage   DiagnosticCode = "invalid-code-block-language"
	CodeInvalidComponentUsage      DiagnosticCode = "invalid-component-usage"
	CodeInvalidContextAccess       DiagnosticCode = "invalid-context-access"
	CodeInvalidContextUsage        DiagnosticCode = "invalid-context-usage"
	CodeInvalidCurrentValue        DiagnosticCode = "invalid-current-value"
	CodeInvalidDimension           DiagnosticCode = "invalid-dimension"
	CodeInvalidExpression          DiagnosticCode = "invalid-expression"
	CodeInvalidItemComponent       DiagnosticCode = "invalid-item-component"
	CodeInvalidListItem            DiagnosticCode = "invalid-list-item"
	CodeInvalidParameterAccess     DiagnosticCode = "invalid-parameter-access"
	CodeInvalidParameterDefinition DiagnosticCode = "invalid-parameter-definition"
	CodeInvalidParameterUsage      DiagnosticCode = "invalid-parameter-usage"
	CodeInvalidTableDelimiter      DiagnosticCode = "invalid-table-delimiter"
	CodeInvalidTableRow            DiagnosticCode = "invalid-table-row"
	CodeInvalidValue               DiagnosticCode = "invalid-value"
	CodeInvalidVariantURL          DiagnosticCode = "invalid-variant-url"
	CodeMissingArgument            DiagnosticCode = "missing-argument"
	CodeNotComponentParameter      DiagnosticCode = "not-component-parameter"
	CodeTableCellCountMismatch     DiagnosticCode = "table-cell-count-mismatch"
	CodeUnclosedExpression         DiagnosticCode = "unclosed-expression"
	CodeUnknownComponent           DiagnosticCode = "unknown-component"
	CodeUnknownKey                 DiagnosticCode = "unknown-key"
	CodeUnknownParameter           DiagnosticCode = "unknown-parameter"
	CodeUnknownRecordKey           DiagnosticCode = "unknown-record-key"
	CodeUnsupportedMimeType        DiagnosticCode = "unsupported-mime-type"
	CodeWrongArgumentType          DiagnosticCode = "wrong-argument-type"
)

// Position is a point in a source: a 0-based byte offset, a 1-based line and a
// 1-based column counted in runes.
type Position = ast.Position

// Range is a [Start, End) range in a source.
type Range struct {
	Start Position
	End   Position
}

// Diagnostic is an error in a source. The unit the error belongs to is dropped
// from the output; the rest of the output is written and valid.
type Diagnostic struct {
	// Code is the kind of the error. Programs look at the code, not at the
	// message.
	Code DiagnosticCode
	// Message is an English description. Values are emphasized with **. It is
	// not HTML.
	Message string
	// Source is the source the dropped unit is written in. It is empty for the
	// converted source. For a global it is its scope path ending with its own
	// name: [LAYOUT CARD] for the sub component CARD of LAYOUT.
	Source []string
	// Range is the range of the dropped unit in Source.
	Range Range
	// Calls are the component calls around the dropped unit, outermost first.
	// It is empty when the unit is in the converted source itself.
	Calls []Call
}

// Call is a component call around a dropped unit.
type Call struct {
	Name string
	Kind FrameKind
	// Source is the source the call is written in, as in Diagnostic.
	Source []string
	// Range is the range of the call in Source.
	Range Range
}

func newDiagnostics(root ast.Node, sources map[ast.Node][]byte, rendered []renderer.Diagnostic) []Diagnostic {
	if len(rendered) == 0 {
		return nil
	}
	diags := make([]Diagnostic, 0, len(rendered))
	for _, r := range rendered {
		var calls []Call
		for _, c := range r.Calls {
			calls = append(calls, Call{
				Name:   c.Name,
				Kind:   c.Kind,
				Source: sourcePath(c.Node),
				Range:  rangeOf(root, sources, c.Node),
			})
		}
		diags = append(diags, Diagnostic{
			Code:    DiagnosticCode(strings.ReplaceAll(strings.ToLower(r.Title), " ", "-")),
			Message: r.Message,
			Source:  sourcePath(r.Node),
			Range:   rangeOf(root, sources, r.Node),
			Calls:   calls,
		})
	}
	return diags
}

func sourcePath(node ast.Node) []string {
	var names []string
	for cur := node; cur != nil; cur = cur.Parent() {
		if ast.IsRuleName(cur, "global-comp-def") {
			names = append(names, globalCompName(cur))
		}
	}
	if len(names) == 0 {
		return nil
	}
	slices.Reverse(names)
	return names
}

func globalCompName(node ast.Node) string {
	nameNode := ast.FindNodeByRuleName(node.Children(), "global-comp-name")
	if nameNode == nil {
		return ""
	}
	return strings.TrimSpace(string(nameNode.Raw()))
}

func sourceDef(root, node ast.Node) ast.Node {
	for cur := node; cur != nil; cur = cur.Parent() {
		if ast.IsRuleName(cur, "global-comp-def") {
			return cur
		}
	}
	return root
}

func rangeOf(root ast.Node, sources map[ast.Node][]byte, node ast.Node) Range {
	source := sources[sourceDef(root, node)]
	rng := node.Range()
	return Range{
		Start: positionIn(source, rng.Start),
		End:   positionIn(source, rng.End),
	}
}

func positionIn(source []byte, offset int) Position {
	if offset < 0 {
		offset = 0
	}
	if offset > len(source) {
		offset = len(source)
	}
	return ast.PositionAt(source, offset)
}
