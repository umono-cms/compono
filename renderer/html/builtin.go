package html

import (
	"html"
	"strings"

	"github.com/umono-cms/compono/ast"
)

type builtinComponent interface {
	New() builtinComponent
	Name() string
	Render(invoker renderableNode, node ast.Node) string
}

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
	newTab, ok := getBoolArgValue(invoker, node, "new-tab")
	if ok && newTab {
		newTabStr = ` target="_blank" rel="noopener noreferrer"`
	}
	return "<compono-link><a href=\"" + getArgValueWithDefa(l.renderer, invoker, node, "url", "url") + "\"" + newTabStr + ">" + getArgValueWithDefa(l.renderer, invoker, node, "text", "") + "</a></compono-link>"
}

func getArgValueWithDefa(r *renderer, invoker renderableNode, compCall ast.Node, name string, defa string) string {
	value, ok := getArgValue(r, invoker, compCall, name)
	if !ok {
		return defa
	}
	return value
}

func getArgValue(r *renderer, invoker renderableNode, compCall ast.Node, name string) (string, bool) {
	compCallArg := ast.GetCompCallArgByParamName(ast.GetCompCallArgsFromCompCall(compCall), name)
	if compCallArg == nil {
		return "", false
	}

	resolved := ast.ResolveCompCallArgValue(r.root, compCallArg, getAncestorsByInvoker(invoker), compCall)
	if resolved.IsZero() || resolved.Type == "array" || resolved.Type == "record" || resolved.MissingContextKey != "" {
		return "", false
	}

	return html.EscapeString(strings.TrimSpace(resolved.Raw)), true
}

func getBoolArgValue(invoker renderableNode, callNode ast.Node, name string) (bool, bool) {
	rnd, _ := invoker.(*compCall)
	if rnd == nil {
		return false, false
	}

	value, ok := getArgValue(rnd.renderer, invoker, callNode, name)
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
