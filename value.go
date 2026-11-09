package compono

import (
	"strconv"
	"strings"

	"github.com/umono-cms/compono/ast"
)

// invalidValueMessage is the message of the diagnostic that reports a value
// that is not a string, an integer, a bool, an array or a record.
const invalidValueMessage = "The value must be a string, integer, bool, array or record."

// valuePrefix opens the source a value is parsed in. The value is written as
// the only argument of the call it names.
const valuePrefix = "{{ PARSE_VALUE v = "

// valueSuffix closes the source a value is parsed in.
const valueSuffix = " }}"

// quote is the character that is doubled to escape it inside a string.
const quote = "\""

// ParseValue parses src as the value of an argument: a string, an integer,
// a bool, an array or a record. Leading and trailing spaces, tabs and line
// breaks are ignored. A component, a parameter reference or a context
// reference, at the top or nested, is an invalid-value diagnostic. Grammar
// errors are returned with the codes they have in an argument value. Ranges
// are relative to src. When there are diagnostics the value is nil.
func ParseValue(src []byte) (any, []Diagnostic) {
	lead, trimmed := trimValue(src)
	if len(trimmed) == 0 {
		return nil, []Diagnostic{invalidValueDiagnostic(src, 0, len(src))}
	}

	wrapped := valuePrefix + string(trimmed) + valueSuffix
	analysis, err := Analyze([]byte(wrapped))
	if err != nil {
		return nil, []Diagnostic{invalidValueDiagnostic(src, lead, lead+len(trimmed))}
	}
	if len(analysis.Diagnostics) > 0 {
		diags := make([]Diagnostic, 0, len(analysis.Diagnostics))
		for _, d := range analysis.Diagnostics {
			diags = append(diags, Diagnostic{
				Code:    d.Code,
				Message: d.Message,
				Range: Range{
					Start: positionIn(src, valueOffset(d.Range.Start.Offset, lead, trimmed)),
					End:   positionIn(src, valueOffset(d.Range.End.Offset, lead, trimmed)),
				},
			})
		}
		return nil, diags
	}
	if !isValueCall(analysis) {
		return nil, []Diagnostic{invalidValueDiagnostic(src, lead, lead+len(trimmed))}
	}

	value, rng, ok := convertValueNode(valueNodeOf([]byte(wrapped)), nil)
	if !ok {
		return nil, []Diagnostic{invalidValueDiagnostic(src,
			valueOffset(rng.Start, lead, trimmed),
			valueOffset(rng.End, lead, trimmed),
		)}
	}

	return value, nil
}

// trimValue trims the spaces, tabs and line breaks at both ends of src. It
// returns the offset the trimmed value starts at and the trimmed value.
func trimValue(src []byte) (int, []byte) {
	start := 0
	for start < len(src) && isValueSpace(src[start]) {
		start++
	}
	end := len(src)
	for end > start && isValueSpace(src[end-1]) {
		end--
	}
	return start, src[start:end]
}

// isValueSpace reports whether the byte is a space, a tab or a line break.
func isValueSpace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\r' || b == '\n'
}

// valueOffset maps an offset in the source a value is parsed in back to the
// source the value is given in.
func valueOffset(offset, lead int, trimmed []byte) int {
	mapped := offset - len(valuePrefix)
	if mapped < 0 {
		mapped = 0
	}
	if mapped > len(trimmed) {
		mapped = len(trimmed)
	}
	return lead + mapped
}

// invalidValueDiagnostic returns the diagnostic that reports a value that is
// not a string, an integer, a bool, an array or a record. The range is the
// given range of the source.
func invalidValueDiagnostic(src []byte, start, end int) Diagnostic {
	return Diagnostic{
		Code:    CodeInvalidValue,
		Message: invalidValueMessage,
		Range:   Range{Start: positionIn(src, start), End: positionIn(src, end)},
	}
}

// isValueCall reports whether the analysis is of a source that holds the
// single call the value is parsed in.
func isValueCall(analysis Analysis) bool {
	if len(analysis.Root) != 1 {
		return false
	}
	unit := analysis.Root[0]
	if unit.Kind != "call" || unit.Ref < 0 || unit.Ref >= len(analysis.Refs) {
		return false
	}
	ref := analysis.Refs[unit.Ref]
	if ref.Name != "PARSE_VALUE" {
		return false
	}
	return len(ref.Args) == 1 && ref.Args[0] == "v"
}

// valueNodeOf returns the node of the type of the value in the source a
// value is parsed in, nil if there is none.
func valueNodeOf(wrapped []byte) ast.Node {
	top, _ := analysisTree(wrapped)
	calls := ast.FilterNodesInTree(top, func(node ast.Node) bool {
		return ast.IsRuleNameOneOf(node, []string{"block-comp-call", "inline-comp-call"})
	})
	if len(calls) == 0 {
		return nil
	}
	args := ast.FindNodeByRuleName(calls[0].Children(), "comp-call-args")
	if args == nil {
		return nil
	}
	arg := ast.FindNodeByRuleName(args.Children(), "comp-call-arg")
	if arg == nil {
		return nil
	}
	return ast.FindNodeByRuleName(arg.Children(), "comp-call-arg-type")
}

// convertValueNode converts the node of a value to a Go value. wrapper is
// the node of the type the value is written in, whose range is the range of
// the diagnostic of an invalid value. The range of the invalid value is
// returned with false when the value is not valid.
func convertValueNode(node ast.Node, wrapper ast.Node) (any, ast.Range, bool) {
	if node == nil {
		return nil, invalidValueRange(node, wrapper), false
	}

	switch node.Rule().Name() {
	case "comp-call-arg-type", "comp-call-array-arg-value-type", "comp-call-record-arg-value-type":
		child := typedValueNode(node)
		if child == nil {
			return nil, node.Range(), false
		}
		return convertValueNode(child, node)
	case "comp-call-string-arg":
		value := ast.FindNodeByRuleName(node.Children(), "comp-call-arg-value")
		if value == nil {
			return "", ast.Range{}, true
		}
		return strings.ReplaceAll(string(value.Raw()), quote+quote, quote), ast.Range{}, true
	case "comp-call-integer-arg":
		number, err := strconv.Atoi(argValueText(node))
		if err != nil {
			return nil, invalidValueRange(node, wrapper), false
		}
		return number, ast.Range{}, true
	case "comp-call-bool-arg":
		switch argValueText(node) {
		case "true":
			return true, ast.Range{}, true
		case "false":
			return false, ast.Range{}, true
		}
		return nil, invalidValueRange(node, wrapper), false
	case "comp-call-array-arg":
		return convertArrayValue(node)
	case "comp-call-record-arg":
		return convertRecordValue(node)
	default:
		return nil, invalidValueRange(node, wrapper), false
	}
}

// typedValueNode returns the node of the type of the value in the given node,
// nil if there is none.
func typedValueNode(node ast.Node) ast.Node {
	return ast.FindNode(node.Children(), func(child ast.Node) bool {
		return ast.IsRuleNameOneOf(child, []string{
			"comp-call-string-arg",
			"comp-call-integer-arg",
			"comp-call-bool-arg",
			"comp-call-array-arg",
			"comp-call-record-arg",
			"comp-call-comp-arg",
			"comp-call-param-arg",
			"comp-call-context-arg",
		})
	})
}

// argValueText returns the trimmed text of the value of the given node.
func argValueText(node ast.Node) string {
	value := ast.FindNodeByRuleName(node.Children(), "comp-call-arg-value")
	if value == nil {
		return ""
	}
	return strings.TrimSpace(string(value.Raw()))
}

// invalidValueRange returns the range of the diagnostic of an invalid value:
// the range of the type the value is written in, the range of the node itself
// when there is no such type.
func invalidValueRange(node ast.Node, wrapper ast.Node) ast.Range {
	if wrapper != nil {
		return wrapper.Range()
	}
	if node == nil {
		return ast.Range{}
	}
	return node.Range()
}

// convertArrayValue converts the node of an array to a Go value.
func convertArrayValue(node ast.Node) (any, ast.Range, bool) {
	items := []any{}
	values := ast.FindNodeByRuleName(node.Children(), "comp-call-array-arg-values")
	if values == nil {
		return items, ast.Range{}, true
	}
	for _, valueNode := range values.Children() {
		if !ast.IsRuleName(valueNode, "comp-call-array-arg-value") {
			continue
		}
		valueType := ast.FindNodeByRuleName(valueNode.Children(), "comp-call-array-arg-value-type")
		if valueType == nil {
			continue
		}
		item, rng, ok := convertValueNode(valueType, node)
		if !ok {
			return nil, rng, false
		}
		items = append(items, item)
	}
	return items, ast.Range{}, true
}

// convertRecordValue converts the node of a record to a Go value.
func convertRecordValue(node ast.Node) (any, ast.Range, bool) {
	fields := map[string]any{}
	values := ast.FindNodeByRuleName(node.Children(), "comp-call-record-arg-values")
	if values == nil {
		return fields, ast.Range{}, true
	}
	for _, valueNode := range values.Children() {
		if !ast.IsRuleName(valueNode, "comp-call-record-arg-value") {
			continue
		}
		key := ast.FindNodeByRuleName(valueNode.Children(), "comp-call-record-arg-key")
		valueType := ast.FindNodeByRuleName(valueNode.Children(), "comp-call-record-arg-value-type")
		if key == nil || valueType == nil {
			continue
		}
		value, rng, ok := convertValueNode(valueType, node)
		if !ok {
			return nil, rng, false
		}
		fields[strings.TrimSpace(string(key.Raw()))] = value
	}
	return fields, ast.Range{}, true
}
