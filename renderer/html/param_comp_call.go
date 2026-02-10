package html

import (
	"strings"

	"github.com/umono-cms/compono/ast"
)

type paramCompCall struct {
	baseRenderable
	renderer *renderer
}

func newParamCompCall(rend *renderer) renderableNode {
	return &paramCompCall{
		renderer: rend,
	}
}

func (pcc *paramCompCall) New() renderableNode {
	return newParamCompCall(pcc.renderer)
}

func (_ *paramCompCall) Condition(invoker renderableNode, node ast.Node) bool {
	return ast.IsRuleNameOneOf(node, []string{"block-param-comp-call", "inline-param-comp-call"})
}

func (pcc *paramCompCall) Render() string {
	inlineCall := ast.IsRuleName(pcc.Node(), "inline-param-comp-call")

	paramName := pcc.getParamName()
	compName := pcc.resolveCompName(paramName)
	if compName == "" {
		return ""
	}

	globalCompDefAnc := ast.FindNode(ast.GetAncestors(pcc.Node()), func(anc ast.Node) bool {
		return ast.IsRuleName(anc, "global-comp-def")
	})

	localCompDefSrc := pcc.renderer.root
	if globalCompDefAnc != nil {
		localCompDefSrc = globalCompDefAnc
	}

	localCompDef := pcc.renderer.findLocalCompDef(localCompDefSrc, compName)
	if localCompDef != nil {
		localCompDefContent := ast.FindNodeByRuleName(localCompDef.Children(), "local-comp-def-content")
		if localCompDefContent == nil {
			return ""
		}
		if inlineCall {
			return pcc.renderInlineParamCompCall(localCompDefContent)
		}
		return pcc.renderer.renderChildren(pcc, localCompDefContent.Children())
	}

	globalCompDef := pcc.renderer.findGlobalCompDef(compName)
	if globalCompDef != nil {
		globalCompDefContent := ast.FindNodeByRuleName(globalCompDef.Children(), "global-comp-def-content")
		if globalCompDefContent == nil {
			return ""
		}
		if inlineCall {
			return pcc.renderInlineParamCompCall(globalCompDefContent)
		}
		return pcc.renderer.renderChildren(pcc, globalCompDefContent.Children())
	}

	return ""
}

func (pcc *paramCompCall) getParamName() string {
	nameNode := ast.FindNodeByRuleName(pcc.Node().Children(), "param-comp-call-name")
	if nameNode == nil {
		return ""
	}
	return strings.TrimSpace(string(nameNode.Raw()))
}

func (pcc *paramCompCall) resolveCompName(paramName string) string {
	invokerAncestors := getAncestorsByInvoker(pcc)

	for _, anc := range invokerAncestors {
		if !ast.IsRuleNameOneOf(anc, []string{"block-comp-call", "inline-comp-call", "block-param-comp-call", "inline-param-comp-call"}) {
			continue
		}

		compCallArgs := ast.FindNodeByRuleName(anc.Children(), "comp-call-args")
		if compCallArgs != nil {
			compCallArg := ast.FindNode(compCallArgs.Children(), func(cca ast.Node) bool {
				argName := ast.FindNodeByRuleName(cca.Children(), "comp-call-arg-name")
				return argName != nil && strings.TrimSpace(string(argName.Raw())) == paramName
			})
			if compCallArg != nil {
				return resolveCompCallArgValueRaw(compCallArg, invokerAncestors, anc, pcc.renderer)
			}
		}

		compDef := findCompDefFromCompCall(anc, pcc.renderer)
		if compDef != nil {
			val := getCompParamDefault(compDef, paramName)
			if val != "" {
				return val
			}
		}
	}

	return ""
}

func resolveCompCallArgValueRaw(compCallArg ast.Node, invokerAncestors []ast.Node, currentCompCall ast.Node, r *renderer) string {
	compCallArgType := ast.FindNodeByRuleName(compCallArg.Children(), "comp-call-arg-type")
	argTypeNode := ast.FindNode(compCallArgType.Children(), func(node ast.Node) bool {
		return ast.IsRuleNameOneOf(node, []string{"comp-call-string-arg", "comp-call-number-arg", "comp-call-bool-arg", "comp-call-param-arg", "comp-call-comp-arg"})
	})
	if argTypeNode == nil {
		return ""
	}

	argValue := ast.FindNodeByRuleName(argTypeNode.Children(), "comp-call-arg-value")
	if argValue == nil {
		return ""
	}

	if ast.IsRuleName(argTypeNode, "comp-call-param-arg") {
		referencedParamName := strings.TrimSpace(string(argValue.Raw()))
		var remainingAncestors []ast.Node
		for i, anc := range invokerAncestors {
			if anc == currentCompCall {
				remainingAncestors = invokerAncestors[i+1:]
				break
			}
		}
		return resolveParamFromAncestorsRaw(referencedParamName, remainingAncestors, r)
	}

	return strings.TrimSpace(string(argValue.Raw()))
}

func resolveParamFromAncestorsRaw(paramName string, invokerAncestors []ast.Node, r *renderer) string {
	for _, anc := range invokerAncestors {
		if !ast.IsRuleNameOneOf(anc, []string{"block-comp-call", "inline-comp-call", "block-param-comp-call", "inline-param-comp-call"}) {
			continue
		}

		compCallArgs := ast.FindNodeByRuleName(anc.Children(), "comp-call-args")
		if compCallArgs != nil {
			compCallArg := ast.FindNode(compCallArgs.Children(), func(cca ast.Node) bool {
				argName := ast.FindNodeByRuleName(cca.Children(), "comp-call-arg-name")
				return strings.TrimSpace(string(argName.Raw())) == paramName
			})

			if compCallArg != nil {
				return resolveCompCallArgValueRaw(compCallArg, invokerAncestors, anc, r)
			}
		}

		if r != nil {
			compDef := findCompDefFromCompCall(anc, r)
			if compDef != nil {
				val := getCompParamDefault(compDef, paramName)
				if val != "" {
					return val
				}
			}
		}
	}
	return ""
}

func getCompParamDefault(compDef ast.Node, paramName string) string {
	var compDefHead ast.Node
	compDefHead = ast.FindNodeByRuleName(compDef.Children(), "local-comp-def-head")
	if compDefHead == nil {
		compDefHead = ast.FindNodeByRuleName(compDef.Children(), "global-comp-def-head")
	}
	if compDefHead == nil {
		return ""
	}

	compParams := ast.FindNodeByRuleName(compDefHead.Children(), "comp-params")
	if compParams == nil {
		return ""
	}

	compParam := ast.FindNode(compParams.Children(), func(cp ast.Node) bool {
		cpName := ast.FindNodeByRuleName(cp.Children(), "comp-param-name")
		return cpName != nil && strings.TrimSpace(string(cpName.Raw())) == paramName
	})
	if compParam == nil {
		return ""
	}

	compParamType := ast.FindNodeByRuleName(compParam.Children(), "comp-param-type")
	if compParamType == nil {
		return ""
	}

	typeNode := ast.FindNode(compParamType.Children(), func(node ast.Node) bool {
		return ast.IsRuleNameOneOf(node, []string{"comp-string-param", "comp-number-param", "comp-bool-param", "comp-comp-param"})
	})
	if typeNode == nil {
		return ""
	}

	defaValue := ast.FindNodeByRuleName(typeNode.Children(), "comp-param-defa-value")
	if defaValue == nil {
		return ""
	}

	return strings.TrimSpace(string(defaValue.Raw()))
}

func (pcc *paramCompCall) renderInlineParamCompCall(compDefContent ast.Node) string {
	childCount := len(compDefContent.Children())
	if childCount == 0 {
		return ""
	}
	p := ast.FindNodeByRuleName(compDefContent.Children(), "p")
	pContent := ast.FindNodeByRuleName(p.Children(), "p-content")
	return pcc.renderer.renderChildren(pcc, pContent.Children())
}
