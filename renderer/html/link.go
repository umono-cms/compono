package html

import (
	"html"
	"strings"

	"github.com/umono-cms/compono/ast"
	"github.com/umono-cms/compono/renderer/hook"
)

type linkElement struct {
	baseRenderable
	renderer *renderer
}

func newLinkElement(rend *renderer) renderableNode {
	return &linkElement{
		renderer: rend,
	}
}

func (l *linkElement) New() renderableNode {
	return newLinkElement(l.renderer)
}

func (_ *linkElement) Condition(invoker renderableNode, node ast.Node) bool {
	return ast.IsRuleName(node, "link")
}

func (l *linkElement) Render() string {
	linkText := ast.FindNodeByRuleName(l.Node().Children(), "link-text")
	linkURL := ast.FindNodeByRuleName(l.Node().Children(), "link-url")

	text := ""
	url := ""

	if linkText != nil {
		text = l.renderer.renderChildren(l, linkText.Children())
	}

	if linkURL != nil {
		url = html.EscapeString(strings.TrimSpace(string(linkURL.Raw())))
	}

	output := `<compono-link><a href="` + url + `">` + text + `</a></compono-link>`

	params := hook.Params{}
	if linkText != nil {
		raw := strings.TrimSpace(string(linkText.Raw()))
		params["text"] = hook.NewString(raw)
	}
	if linkURL != nil {
		raw := strings.TrimSpace(string(linkURL.Raw()))
		params["url"] = hook.NewString(raw)
	}

	return l.renderer.applyHooks(output, hook.KindMarkdown, "link", params)
}

type linkTextElement struct {
	baseRenderable
	renderer *renderer
}

func newLinkTextElement(rend *renderer) renderableNode {
	return &linkTextElement{
		renderer: rend,
	}
}

func (lt *linkTextElement) New() renderableNode {
	return newLinkTextElement(lt.renderer)
}

func (_ *linkTextElement) Condition(invoker renderableNode, node ast.Node) bool {
	return ast.IsRuleName(node, "link-text")
}

func (lt *linkTextElement) Render() string {
	return html.EscapeString(strings.TrimSpace(string(lt.Node().Raw())))
}

type linkURLElement struct {
	baseRenderable
	renderer *renderer
}

func newLinkURLElement(rend *renderer) renderableNode {
	return &linkURLElement{
		renderer: rend,
	}
}

func (lu *linkURLElement) New() renderableNode {
	return newLinkURLElement(lu.renderer)
}

func (_ *linkURLElement) Condition(invoker renderableNode, node ast.Node) bool {
	return ast.IsRuleName(node, "link-url")
}

func (lu *linkURLElement) Render() string {
	return html.EscapeString(strings.TrimSpace(string(lu.Node().Raw())))
}
