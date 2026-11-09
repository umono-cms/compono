package html

import (
	"html"

	"github.com/umono-cms/compono/ast"
	"github.com/umono-cms/compono/errwrap"
	"github.com/umono-cms/compono/util"
)

type builtinComponent interface {
	New() builtinComponent
	Name() string
	Render(invoker renderableNode, node ast.Node) string
}

// currentAuto is the reserved LINK "current" value that is resolved when the
// link is rendered.
const currentAuto = "__auto"

type link struct {
	renderer *renderer
}

func newLink(rend *renderer) builtinComponent {
	return &link{
		renderer: rend,
	}
}

func (l *link) New() builtinComponent {
	return newLink(l.renderer)
}

func (_ *link) Name() string {
	return "LINK"
}

func (l *link) Render(invoker renderableNode, node ast.Node) string {
	newTabStr := ""
	newTab, ok := getBoolArgValue(l.renderer, invoker, node, "new-tab")
	if ok && newTab {
		newTabStr = ` target="_blank" rel="noopener noreferrer"`
	}
	ariaCurrentStr := ""
	current, ok := getRawArgValue(l.renderer, invoker, node, "current")
	if ok && current == currentAuto {
		current = ""
		if l.renderer.currentResolver != nil {
			url, _ := getRawArgValue(l.renderer, invoker, node, "url")
			current = util.TrimWhitespace(l.renderer.currentResolver(util.NormalizeURL(url)))
			if current != "" && !errwrap.LinkCurrentHTMLValue(current) {
				title, message := errwrap.InvalidCurrentValueError(current)
				return l.renderer.recordDiagnostic(node, title, message, isBlockBuiltinCall(l.renderer, node))
			}
		}
	}
	if ok && current != "" {
		ariaCurrentStr = ` aria-current="` + html.EscapeString(current) + `"`
	}
	attrs := l.renderer.callAttributeHook(l.Name(), linkConflictAttributes)
	return "<a href=\"" + util.NormalizeURL(getArgValueWithDefa(l.renderer, invoker, node, "url", "url")) + "\"" + newTabStr + ariaCurrentStr + formatHookAttributes(attrs) + ">" + getArgValueWithDefa(l.renderer, invoker, node, "text", "") + "</a>"
}

var linkConflictAttributes = map[string]bool{
	"href":         true,
	"target":       true,
	"rel":          true,
	"aria-current": true,
}

func getArgValueWithDefa(r *renderer, invoker renderableNode, compCall ast.Node, name string, defa string) string {
	value, ok := getArgValue(r, invoker, compCall, name)
	if !ok {
		return defa
	}
	return value
}

func getArgValue(r *renderer, invoker renderableNode, compCall ast.Node, name string) (string, bool) {
	value, ok := getRawArgValue(r, invoker, compCall, name)
	if !ok {
		return "", false
	}

	return html.EscapeString(value), true
}

// getRawArgValue resolves the argument value of a builtin call without
// escaping it.
func getRawArgValue(r *renderer, invoker renderableNode, compCall ast.Node, name string) (string, bool) {
	resolved, ok := resolveBuiltinArg(r, invoker, compCall, name)
	if !ok || resolved.IsZero() || resolved.Type == "array" || resolved.Type == "record" || resolved.MissingContextKey != "" {
		return "", false
	}

	return util.TrimWhitespace(resolved.Raw), true
}

// isBlockBuiltinCall reports whether node, a call or a parameter reference
// that renders a built-in, is written as a block.
func isBlockBuiltinCall(r *renderer, node ast.Node) bool {
	switch {
	case ast.IsRuleName(node, "block-comp-call"):
		return true
	case ast.IsRuleName(node, "inline-comp-call"):
		return false
	}
	return !(r.inlineBodies > 0 || isInlineCompParamRef(node))
}

func getBoolArgValue(r *renderer, invoker renderableNode, callNode ast.Node, name string) (bool, bool) {
	value, ok := getArgValue(r, invoker, callNode, name)
	if !ok {
		return false, false
	}
	if value == "false" {
		return false, true
	}
	if value == "true" {
		return true, true
	}
	return false, false
}
