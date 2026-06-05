package html

import (
	"strings"

	"github.com/umono-cms/compono/ast"
	"github.com/umono-cms/compono/renderer/hook"
)

type compCall struct {
	baseRenderable
	renderer *renderer
}

func newCompCall(rend *renderer) renderableNode {
	return &compCall{
		renderer: rend,
	}
}

func (cc *compCall) New() renderableNode {
	return newCompCall(cc.renderer)
}

func (_ *compCall) Condition(invoker renderableNode, node ast.Node) bool {
	return ast.IsRuleNameOneOf(node, []string{"block-comp-call", "inline-comp-call"})
}

func (cc *compCall) Render() string {
	inlineCompCall := false
	if ast.IsRuleName(cc.Node(), "inline-comp-call") {
		inlineCompCall = true
	}

	compCallName := ast.FindNodeByRuleName(cc.Node().Children(), "comp-call-name")
	globalCompDefAnc := ast.FindNode(ast.GetAncestors(cc.Node()), func(anc ast.Node) bool {
		return ast.IsRuleName(anc, "global-comp-def")
	})

	localCompDefSrc := cc.renderer.root
	if globalCompDefAnc != nil {
		localCompDefSrc = globalCompDefAnc
	}

	localCompDef := cc.renderer.findLocalCompDef(localCompDefSrc, string(compCallName.Raw()))
	if localCompDef != nil {
		localCompDefContent := ast.FindNodeByRuleName(localCompDef.Children(), "local-comp-def-content")
		if localCompDefContent == nil {
			return ""
		}
		if inlineCompCall {
			return cc.renderInlineCompCall(strings.TrimSpace(string(compCallName.Raw())), localCompDefContent)
		}
		return cc.renderer.renderChildren(cc, localCompDefContent.Children())
	}

	globalCompDef := cc.renderer.findGlobalCompDef(string(compCallName.Raw()))
	if globalCompDef != nil {
		globalCompDefContent := ast.FindNodeByRuleName(globalCompDef.Children(), "global-comp-def-content")
		if globalCompDefContent == nil {
			return ""
		}
		if inlineCompCall {
			return cc.renderInlineCompCall(strings.TrimSpace(string(compCallName.Raw())), globalCompDefContent)
		}
		return cc.renderer.renderChildren(cc, globalCompDefContent.Children())
	}

	builtinComp := cc.renderer.findBuiltinComp(string(compCallName.Raw()))
	if builtinComp != nil {
		output := builtinComp.Render(cc, cc.Node())
		builtinParams := cc.renderer.extractBuiltinParams(cc, cc.Node())
		return cc.renderer.applyHooks(output, hook.KindBuiltin, strings.TrimSpace(string(compCallName.Raw())), builtinParams)
	}

	return ""
}

func (cc *compCall) renderInlineCompCall(compName string, compDefContent ast.Node) string {
	childCount := len(compDefContent.Children())
	if childCount == 0 {
		return ""
	}
	p := ast.FindNodeByRuleName(compDefContent.Children(), "p")
	pContent := ast.FindNodeByRuleName(p.Children(), "p-content")
	return cc.renderer.renderChildren(cc, pContent.Children())
}
