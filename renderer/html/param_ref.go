package html

import (
	"html"
	"strings"

	"github.com/umono-cms/compono/ast"
)

type baseParamRef struct {
	baseRenderable
	renderer *renderer
}

func (bpr *baseParamRef) paramRefName() string {
	paramRefName := ast.FindNodeByRuleName(bpr.Node().Children(), "param-ref-name")
	return strings.TrimSpace(string(paramRefName.Raw()))
}

type paramRefInLocalCompDef struct {
	baseParamRef
}

func newParamRefInLocalCompDef(rend *renderer) renderableNode {
	return &paramRefInLocalCompDef{
		baseParamRef: baseParamRef{
			renderer: rend,
		},
	}
}

func (p *paramRefInLocalCompDef) New() renderableNode {
	return newParamRefInLocalCompDef(p.renderer)
}

func (_ *paramRefInLocalCompDef) Condition(invoker renderableNode, node ast.Node) bool {
	if !ast.IsRuleName(node, "param-ref") {
		return false
	}
	localCompDef := ast.FindNodeByRuleName(ast.GetAncestors(node), "local-comp-def")
	if localCompDef == nil {
		return false
	}
	return true
}

func (p *paramRefInLocalCompDef) Render() string {
	paramRefName := p.paramRefName()

	localCompDef := ast.FindNodeByRuleName(ast.GetAncestors(p.Node()), "local-comp-def")
	localCompDefHead := ast.FindNodeByRuleName(localCompDef.Children(), "local-comp-def-head")
	compParams := ast.FindNodeByRuleName(localCompDefHead.Children(), "comp-params")

	var compParam ast.Node
	if compParams != nil {
		compParam = ast.FindNode(compParams.Children(), func(cp ast.Node) bool {
			compParamName := ast.FindNodeByRuleName(cp.Children(), "comp-param-name")
			if strings.TrimSpace(string(compParamName.Raw())) == paramRefName {
				return true
			}
			return false
		})
	}

	if compParam != nil {
		compCall := ast.FindNode(getAncestorsByInvoker(p), func(node ast.Node) bool {
			return ast.IsRuleNameOneOf(node, []string{"block-comp-call", "inline-comp-call", "block-param-comp-call", "inline-param-comp-call"})
		})

		compCallArgs := ast.FindNodeByRuleName(compCall.Children(), "comp-call-args")
		if compCallArgs != nil {
			compCallArg := ast.FindNode(compCallArgs.Children(), func(cca ast.Node) bool {
				argName := ast.FindNodeByRuleName(cca.Children(), "comp-call-arg-name")
				if strings.TrimSpace(string(argName.Raw())) == paramRefName {
					return true
				}
				return false
			})
			if compCallArg != nil {
				return resolveCompCallArgValue(compCallArg, getAncestorsByInvoker(p), compCall, p.renderer)
			}
		}

		compParamType := ast.FindNodeByRuleName(compParam.Children(), "comp-param-type")
		if compParamType == nil {
			return ""
		}
		compParamDefaValue := ast.FindNodeByRuleName(ast.FindNode(compParamType.Children(), func(node ast.Node) bool {
			return ast.IsRuleNameOneOf(node, []string{"comp-string-param", "comp-number-param", "comp-bool-param", "comp-comp-param"})
		}).Children(), "comp-param-defa-value")

		if compParamDefaValue == nil {
			return ""
		}

		return html.EscapeString(strings.TrimSpace(string(compParamDefaValue.Raw())))
	}

	globalCompDef := ast.FindNodeByRuleName(ast.GetAncestors(p.Node()), "global-comp-def")
	if globalCompDef == nil {
		return ""
	}

	globalCompDefHead := ast.FindNodeByRuleName(globalCompDef.Children(), "global-comp-def-head")
	if globalCompDefHead == nil {
		return ""
	}

	globalCompParams := ast.FindNodeByRuleName(globalCompDefHead.Children(), "comp-params")
	if globalCompParams == nil {
		return ""
	}

	globalCompParam := ast.FindNode(globalCompParams.Children(), func(cp ast.Node) bool {
		compParamName := ast.FindNodeByRuleName(cp.Children(), "comp-param-name")
		if strings.TrimSpace(string(compParamName.Raw())) == paramRefName {
			return true
		}
		return false
	})

	if globalCompParam == nil {
		return ""
	}

	for _, anc := range getAncestorsByInvoker(p) {
		if !ast.IsRuleNameOneOf(anc, []string{"block-comp-call", "inline-comp-call", "block-param-comp-call", "inline-param-comp-call"}) {
			continue
		}
		compCallArgs := ast.FindNodeByRuleName(anc.Children(), "comp-call-args")
		if compCallArgs == nil {
			continue
		}
		compCallArg := ast.FindNode(compCallArgs.Children(), func(cca ast.Node) bool {
			argName := ast.FindNodeByRuleName(cca.Children(), "comp-call-arg-name")
			if strings.TrimSpace(string(argName.Raw())) == paramRefName {
				return true
			}
			return false
		})
		if compCallArg != nil {
			return resolveCompCallArgValue(compCallArg, getAncestorsByInvoker(p), anc, p.renderer)
		}
	}

	compParamDefaValue := ast.FindNodeByRuleName(ast.FindNode(ast.FindNodeByRuleName(globalCompParam.Children(), "comp-param-type").Children(), func(node ast.Node) bool {
		return ast.IsRuleNameOneOf(node, []string{"comp-string-param", "comp-number-param", "comp-bool-param", "comp-comp-param"})
	}).Children(), "comp-param-defa-value")

	if compParamDefaValue == nil {
		return ""
	}

	return html.EscapeString(strings.TrimSpace(string(compParamDefaValue.Raw())))
}

type paramRefInGlobalCompDef struct {
	baseParamRef
}

func newParamRefInGlobalCompDef(rend *renderer) renderableNode {
	return &paramRefInGlobalCompDef{
		baseParamRef: baseParamRef{
			renderer: rend,
		},
	}
}

func (p *paramRefInGlobalCompDef) New() renderableNode {
	return newParamRefInGlobalCompDef(p.renderer)
}

func (_ *paramRefInGlobalCompDef) Condition(invoker renderableNode, node ast.Node) bool {
	if !ast.IsRuleName(node, "param-ref") {
		return false
	}
	localCompDef := ast.FindNodeByRuleName(ast.GetAncestors(node), "local-comp-def")
	if localCompDef != nil {
		return false
	}
	globalCompDef := ast.FindNodeByRuleName(ast.GetAncestors(node), "global-comp-def")
	if globalCompDef == nil {
		return false
	}
	return true
}

func (p *paramRefInGlobalCompDef) Render() string {
	paramRefName := p.paramRefName()

	globalCompDef := ast.FindNodeByRuleName(ast.GetAncestors(p.Node()), "global-comp-def")
	globalCompDefHead := ast.FindNodeByRuleName(globalCompDef.Children(), "global-comp-def-head")

	compParams := ast.FindNodeByRuleName(globalCompDefHead.Children(), "comp-params")

	compParam := ast.FindNode(compParams.Children(), func(cp ast.Node) bool {
		compParamName := ast.FindNodeByRuleName(cp.Children(), "comp-param-name")
		if strings.TrimSpace(string(compParamName.Raw())) == paramRefName {
			return true
		}
		return false
	})

	compCall := ast.FindNode(getAncestorsByInvoker(p), func(node ast.Node) bool {
		return ast.IsRuleNameOneOf(node, []string{"block-comp-call", "inline-comp-call", "block-param-comp-call", "inline-param-comp-call"})
	})

	compCallArgs := ast.FindNodeByRuleName(compCall.Children(), "comp-call-args")
	if compCallArgs != nil {
		compCallArg := ast.FindNode(compCallArgs.Children(), func(cca ast.Node) bool {
			argName := ast.FindNodeByRuleName(cca.Children(), "comp-call-arg-name")
			if strings.TrimSpace(string(argName.Raw())) == paramRefName {
				return true
			}
			return false
		})
		if compCallArg != nil {
			return resolveCompCallArgValue(compCallArg, getAncestorsByInvoker(p), compCall, p.renderer)
		}
	}

	compParamDefaValue := ast.FindNodeByRuleName(ast.FindNode(ast.FindNodeByRuleName(compParam.Children(), "comp-param-type").Children(), func(node ast.Node) bool {
		return ast.IsRuleNameOneOf(node, []string{"comp-string-param", "comp-number-param", "comp-bool-param", "comp-comp-param"})
	}).Children(), "comp-param-defa-value")

	if compParamDefaValue == nil {
		return ""
	}

	return html.EscapeString(strings.TrimSpace(string(compParamDefaValue.Raw())))
}

func resolveCompCallArgValue(compCallArg ast.Node, invokerAncestors []ast.Node, currentCompCall ast.Node, r ...*renderer) string {
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
		var rend *renderer
		if len(r) > 0 {
			rend = r[0]
		}
		return resolveParamFromAncestors(referencedParamName, remainingAncestors, rend)
	}

	return html.EscapeString(strings.TrimSpace(string(argValue.Raw())))
}

func resolveParamFromAncestors(paramName string, invokerAncestors []ast.Node, r *renderer) string {
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
				return resolveCompCallArgValue(compCallArg, invokerAncestors, anc, r)
			}
		}

		if r != nil {
			val := resolveParamDefaultFromCompCall(anc, paramName, r)
			if val != "" {
				return val
			}
		}
	}
	return ""
}

func resolveParamDefaultFromCompCall(compCallNode ast.Node, paramName string, r *renderer) string {
	compDef := findCompDefFromCompCall(compCallNode, r)
	if compDef == nil {
		return ""
	}

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

	return html.EscapeString(strings.TrimSpace(string(defaValue.Raw())))
}

func findCompDefFromCompCall(compCallNode ast.Node, r *renderer) ast.Node {
	compCallNameNode := ast.FindNodeByRuleName(compCallNode.Children(), "comp-call-name")
	if compCallNameNode == nil {
		return nil
	}
	compName := strings.TrimSpace(string(compCallNameNode.Raw()))

	globalCompDefAnc := ast.FindNode(ast.GetAncestors(compCallNode), func(anc ast.Node) bool {
		return ast.IsRuleName(anc, "global-comp-def")
	})

	localCompDefSrc := r.root
	if globalCompDefAnc != nil {
		localCompDefSrc = globalCompDefAnc
	}

	localCompDef := r.findLocalCompDef(localCompDefSrc, compName)
	if localCompDef != nil {
		return localCompDef
	}

	return r.findGlobalCompDef(compName)
}
