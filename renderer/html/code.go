package html

import (
	"html"
	"strings"

	"github.com/umono-cms/compono/ast"
	"github.com/umono-cms/compono/renderer/hook"
)

type codeBlock struct {
	baseRenderable
	renderer *renderer
}

func newCodeBlock(rend *renderer) renderableNode {
	return &codeBlock{
		renderer: rend,
	}
}

func (cb *codeBlock) New() renderableNode {
	return newCodeBlock(cb.renderer)
}

func (_ *codeBlock) Condition(invoker renderableNode, node ast.Node) bool {
	return ast.IsRuleName(node, "code-block")
}

func (cb *codeBlock) Render() string {
	langClass := "language-plaintext"
	params := hook.Params{}
	cbl := ast.FindNodeByRuleName(cb.Node().Children(), "code-block-lang")
	if cbl != nil {
		lang := html.EscapeString(strings.TrimSpace(string(cbl.Raw())))
		if lang != "" {
			langClass = "language-" + lang
			params["lang"] = hook.NewString(strings.TrimSpace(string(cbl.Raw())))
		}
	}
	content := cb.renderer.renderChildren(cb, cb.Node().Children())
	params["content"] = hook.NewString(html.UnescapeString(content))
	output := `<pre><code class="` + langClass + `">` + content + `</code></pre>`
	return cb.renderer.applyHooks(output, hook.KindMarkdown, "code-block", params)
}

type codeBlockContent struct {
	baseRenderable
	renderer *renderer
}

func newCodeBlockContent(rend *renderer) renderableNode {
	return &codeBlockContent{
		renderer: rend,
	}
}

func (cbc *codeBlockContent) New() renderableNode {
	return newCodeBlockContent(cbc.renderer)
}

func (_ *codeBlockContent) Condition(invoker renderableNode, node ast.Node) bool {
	return ast.IsRuleName(node, "code-block-content")
}

func (cbc *codeBlockContent) Render() string {
	return cbc.renderer.renderChildren(cbc, cbc.Node().Children())
}

type inlineCode struct {
	baseRenderable
	renderer *renderer
}

func newInlineCode(rend *renderer) renderableNode {
	return &inlineCode{
		renderer: rend,
	}
}

func (ic *inlineCode) New() renderableNode {
	return newInlineCode(ic.renderer)
}

func (_ *inlineCode) Condition(_ renderableNode, node ast.Node) bool {
	return ast.IsRuleName(node, "inline-code")
}

func (ic *inlineCode) Render() string {
	content := ic.renderer.renderChildren(ic, ic.Node().Children())
	output := `<code style="white-space: pre">` + content + "</code>"
	return ic.renderer.applyHooks(output, hook.KindMarkdown, "inline-code", hook.Params{"content": hook.NewString(html.UnescapeString(content))})
}

type inlineCodeContent struct {
	baseRenderable
	renderer *renderer
}

func newInlineCodeContent(rend *renderer) renderableNode {
	return &inlineCodeContent{
		renderer: rend,
	}
}

func (icc *inlineCodeContent) New() renderableNode {
	return newInlineCodeContent(icc.renderer)
}

func (_ *inlineCodeContent) Condition(_ renderableNode, node ast.Node) bool {
	return ast.IsRuleName(node, "inline-code-content")
}

func (icc *inlineCodeContent) Render() string {
	return icc.renderer.renderChildren(icc, icc.Node().Children())
}
