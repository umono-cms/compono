package errwrap

import (
	"regexp"
	"strings"

	"github.com/umono-cms/compono/ast"
	"github.com/umono-cms/compono/rule"
	"github.com/umono-cms/compono/util"
)

var screamingCaseRe = regexp.MustCompile(`^[A-Z0-9]+(?:_[A-Z0-9]+)*$`)

type ErrorWrapper interface {
	Wrap(ast.Node)
}

func DefaultErrorWrapper() ErrorWrapper {
	return &errorWrapper{}
}

type errorWrapper struct {
	root ast.Node
}

func (ew *errorWrapper) Wrap(root ast.Node) {
	ew.root = root
	ew.wrapInvalidParamCompCallRef(root)
	ew.wrapInfiniteCompCall(root)
	ew.wrapInvalidParamRef(root)
	ew.wrapInvalidCompCall(root)
}

func (ew *errorWrapper) wrapInfiniteCompCall(root ast.Node) {
	ew.detectInfiniteCompCall(root, []string{}, nil)
}

func (ew *errorWrapper) detectInfiniteCompCall(node ast.Node, callStack []string, currentCallNode ast.Node) {
	ruleName := node.Rule().Name()

	if ruleName == "block-comp-call" || ruleName == "inline-comp-call" {

		block := false
		if ruleName == "block-comp-call" {
			block = true
		}

		compCallName := ew.getCompCallName(node)
		if compCallName == "" {
			return
		}

		if ew.isInCallStack(compCallName, callStack) {
			ew.wrapWithErr(node, "Infinite component call", "The call to component **"+compCallName+"** creates an infinite loop and was skipped.", block)
			return
		}

		compDef := ew.findCompDef(node, compCallName)
		if compDef == nil {
			if ew.isBuiltInComp(compCallName) {
				return
			}
			ew.wrapWithErr(node, "Unknown component", "The component **"+compCallName+"** is not defined or not registered.", block)
			return
		}

		newStack := append(callStack, compCallName)
		compDefContent := ew.getCompDefContent(compDef)
		if compDefContent != nil {
			ew.detectInfiniteCompCall(compDefContent, newStack, node)
		}
	}

	if currentCallNode != nil && (ruleName == "block-param-comp-call" || ruleName == "inline-param-comp-call") {
		block := ruleName == "block-param-comp-call"

		compName := ew.resolveParamCompCallName(node, currentCallNode)
		if compName == "" {
			return
		}

		if ew.isInCallStack(compName, callStack) {
			ew.wrapWithErr(node, "Infinite component call", "The call to component **"+compName+"** creates an infinite loop and was skipped.", block)
			return
		}

		compDef := ew.findCompDef(node, compName)
		if compDef == nil {
			if ew.isBuiltInComp(compName) {
				return
			}
			ew.wrapWithErr(node, "Unknown component", "The component **"+compName+"** is not defined or not registered.", block)
			return
		}

		newStack := append(callStack, compName)
		compDefContent := ew.getCompDefContent(compDef)
		if compDefContent != nil {
			ew.detectInfiniteCompCall(compDefContent, newStack, currentCallNode)
		}
	}

	for _, child := range node.Children() {
		ew.detectInfiniteCompCall(child, callStack, currentCallNode)
	}
}

func (ew *errorWrapper) wrapInvalidParamRef(root ast.Node) {
	paramRefs := ast.FilterNodesInTree(root, func(node ast.Node) bool {
		return ast.IsRuleName(node, "param-ref")
	})
	for _, pr := range paramRefs {
		compDefContent := ast.FindNode(ast.GetAncestors(pr), func(anc ast.Node) bool {
			return ast.IsRuleNameOneOf(anc, []string{"local-comp-def-content", "global-comp-def-content"})
		})

		if compDefContent == nil {
			ew.wrapWithErr(pr, "Invalid parameter usage", "Parameters cannot be used in the root context.", false)
			continue
		}

		if ast.IsRuleName(compDefContent, "local-comp-def-content") {
			parRefNa := strings.TrimSpace(string(ast.FindNodeByRuleName(pr.Children(), "param-ref-name").Raw()))

			localCompDef := ast.FindNodeByRuleName(ast.GetAncestors(pr), "local-comp-def")
			localCompDefHead := ast.FindNodeByRuleName(localCompDef.Children(), "local-comp-def-head")
			compParams := ast.FindNodeByRuleName(localCompDefHead.Children(), "comp-params")

			title := "Unknown parameter"
			msg := "The parameter **" + parRefNa + "** is not defined for this component."

			found := false
			if compParams != nil {
				compParam := ast.FindNode(compParams.Children(), func(cp ast.Node) bool {
					compParamName := ast.FindNodeByRuleName(cp.Children(), "comp-param-name")
					if strings.TrimSpace(string(compParamName.Raw())) == parRefNa {
						return true
					}
					return false
				})
				if compParam != nil {
					found = true
				}
			}

			if !found {
				globalCompDef := ast.FindNodeByRuleName(ast.GetAncestors(pr), "global-comp-def")
				if globalCompDef != nil {
					globalCompDefHead := ast.FindNodeByRuleName(globalCompDef.Children(), "global-comp-def-head")
					if globalCompDefHead != nil {
						globalCompParams := ast.FindNodeByRuleName(globalCompDefHead.Children(), "comp-params")
						if globalCompParams != nil {
							globalCompParam := ast.FindNode(globalCompParams.Children(), func(cp ast.Node) bool {
								compParamName := ast.FindNodeByRuleName(cp.Children(), "comp-param-name")
								if strings.TrimSpace(string(compParamName.Raw())) == parRefNa {
									return true
								}
								return false
							})
							if globalCompParam != nil {
								found = true
							}
						}
					}
				}
			}

			if !found {
				ew.wrapWithErr(pr, title, msg, false)
				continue
			}
		}

		if ast.IsRuleName(compDefContent, "global-comp-def-content") {
			parRefNa := strings.TrimSpace(string(ast.FindNodeByRuleName(pr.Children(), "param-ref-name").Raw()))

			globalCompDef := ast.FindNodeByRuleName(ast.GetAncestors(pr), "global-comp-def")
			globalCompDefHead := ast.FindNodeByRuleName(globalCompDef.Children(), "global-comp-def-head")

			title := "Unknown parameter"
			msg := "The parameter **" + parRefNa + "** is not defined for this component."

			if globalCompDefHead == nil {
				ew.wrapWithErr(pr, title, msg, false)
				continue
			}

			compParams := ast.FindNodeByRuleName(globalCompDefHead.Children(), "comp-params")
			if compParams == nil {
				ew.wrapWithErr(pr, title, msg, false)
				continue
			}

			compParam := ast.FindNode(compParams.Children(), func(cp ast.Node) bool {
				compParamName := ast.FindNodeByRuleName(cp.Children(), "comp-param-name")
				if strings.TrimSpace(string(compParamName.Raw())) == parRefNa {
					return true
				}
				return false
			})

			if compParam == nil {
				ew.wrapWithErr(pr, title, msg, false)
				continue
			}
		}
	}
}

func (ew *errorWrapper) wrapInvalidParamCompCallRef(root ast.Node) {
	paramCompCalls := ast.FilterNodesInTree(root, func(node ast.Node) bool {
		return ast.IsRuleNameOneOf(node, []string{"block-param-comp-call", "inline-param-comp-call"})
	})

	for _, pcc := range paramCompCalls {
		block := ast.IsRuleName(pcc, "block-param-comp-call")

		paramNameNode := ast.FindNodeByRuleName(pcc.Children(), "param-comp-call-name")
		if paramNameNode == nil {
			continue
		}
		paramName := strings.TrimSpace(string(paramNameNode.Raw()))

		compDef := ast.FindNode(ast.GetAncestors(pcc), func(anc ast.Node) bool {
			return ast.IsRuleNameOneOf(anc, []string{"local-comp-def", "global-comp-def"})
		})
		if compDef == nil {
			continue
		}

		var compDefHead ast.Node
		compDefHead = ast.FindNodeByRuleName(compDef.Children(), "local-comp-def-head")
		if compDefHead == nil {
			compDefHead = ast.FindNodeByRuleName(compDef.Children(), "global-comp-def-head")
		}
		if compDefHead == nil {
			ew.wrapParamCompCallError(pcc, "Unknown parameter", "The parameter **"+paramName+"** is not defined for this component.", block)
			continue
		}

		compParams := ast.FindNodeByRuleName(compDefHead.Children(), "comp-params")
		if compParams == nil {
			ew.wrapParamCompCallError(pcc, "Unknown parameter", "The parameter **"+paramName+"** is not defined for this component.", block)
			continue
		}

		compParam := ast.FindNode(compParams.Children(), func(cp ast.Node) bool {
			cpName := ast.FindNodeByRuleName(cp.Children(), "comp-param-name")
			return cpName != nil && strings.TrimSpace(string(cpName.Raw())) == paramName
		})

		if compParam == nil {
			ew.wrapParamCompCallError(pcc, "Unknown parameter", "The parameter **"+paramName+"** is not defined for this component.", block)
			continue
		}

		compParamType := ast.FindNodeByRuleName(compParam.Children(), "comp-param-type")
		if compParamType == nil {
			// bare param (no type) → treated as component type
			continue
		}

		// Check raw content of comp-param-type: if it doesn't match SCREAMING_CASE,
		// the param is not a component type. We use raw content check because
		// Pattern selector ignores already-claimed regions, causing false child matches.
		compParamTypeRaw := strings.TrimSpace(string(compParamType.Raw()))
		if !screamingCaseRe.MatchString(compParamTypeRaw) {
			ew.wrapParamCompCallError(pcc, "Not component parameter", "The parameter **"+paramName+"** is not component parameter", block)
			continue
		}
	}
}

func (ew *errorWrapper) wrapParamCompCallError(self ast.Node, title, msg string, isBlock bool) {
	if isBlock {
		// Block param-comp-call errors need p > p-content > inline-error wrapping
		errNode := ew.createError("inline-error", self, title, msg)

		pRule := rule.NewDynamic("p")
		pContentRule := rule.NewDynamic("p-content")

		pContentNode := ast.DefaultEmptyNode()
		pContentNode.SetRule(pContentRule)

		inlineErrChild := ast.DefaultEmptyNode()
		inlineErrChild.SetRule(errNode.Rule())
		inlineErrChild.SetChildren(errNode.Children())
		inlineErrChild.SetRaw(errNode.Raw())
		inlineErrChild.SetParent(pContentNode)

		pContentNode.SetChildren([]ast.Node{inlineErrChild})

		self.SetRule(pRule)
		self.SetChildren([]ast.Node{pContentNode})
		self.SetRaw([]byte{})
		pContentNode.SetParent(self)
	} else {
		ew.wrapWithErr(self, title, msg, false)
	}
}

func (ew *errorWrapper) wrapInvalidCompCall(root ast.Node) {
	inlineCompCalls := ast.FilterNodesInTree(root, func(node ast.Node) bool {
		return ast.IsRuleNameOneOf(node, []string{"inline-comp-call", "inline-param-comp-call"})
	})

	for _, icc := range inlineCompCalls {
		var compCallName string
		if ast.IsRuleName(icc, "inline-comp-call") {
			compCallName = ew.getCompCallName(icc)
		} else {
			// For inline-param-comp-calls, resolve from comp-def default only.
			// If the default comp doesn't exist, skip — it may be overridden at runtime.
			compCallName = ew.resolveParamCompCallName(icc, nil)
		}
		if compCallName == "" {
			continue
		}
		compDef := ew.findCompDef(icc, compCallName)
		if compDef == nil {
			continue
		}
		compDefContent := ew.getCompDefContent(compDef)
		if compDefContent == nil {
			continue
		}
		childrenCount := len(compDefContent.Children())
		if childrenCount == 0 {
			continue
		}

		wrap := false
		if childrenCount > 1 {
			wrap = true
		}

		var p ast.Node
		if !wrap {
			p = ast.FindNodeByRuleName(compDefContent.Children(), "p")
			if p == nil {
				wrap = true
			}
		}

		if !wrap {
			pc := ast.FindNodeByRuleName(p.Children(), "p-content")
			sb := ast.FindNodeByRuleName(pc.Children(), "soft-break")
			if sb != nil {
				wrap = true
			}
		}

		if wrap {
			ew.wrapWithErr(icc, "Invalid component usage", "The component **"+compCallName+"** is a block component and cannot be used inline.", false)
		}

	}
}

func (ew *errorWrapper) resolveParamCompCallName(pccNode ast.Node, currentCallNode ast.Node) string {
	paramNameNode := ast.FindNodeByRuleName(pccNode.Children(), "param-comp-call-name")
	if paramNameNode == nil {
		return ""
	}
	paramName := strings.TrimSpace(string(paramNameNode.Raw()))

	if currentCallNode != nil {
		compCallArgs := ast.FindNodeByRuleName(currentCallNode.Children(), "comp-call-args")
		if compCallArgs != nil {
			compCallArg := ast.FindNode(compCallArgs.Children(), func(cca ast.Node) bool {
				argName := ast.FindNodeByRuleName(cca.Children(), "comp-call-arg-name")
				return argName != nil && strings.TrimSpace(string(argName.Raw())) == paramName
			})
			if compCallArg != nil {
				compCallArgType := ast.FindNodeByRuleName(compCallArg.Children(), "comp-call-arg-type")
				if compCallArgType != nil {
					compArg := ast.FindNodeByRuleName(compCallArgType.Children(), "comp-call-comp-arg")
					if compArg != nil {
						argValue := ast.FindNodeByRuleName(compArg.Children(), "comp-call-arg-value")
						if argValue != nil {
							return strings.TrimSpace(string(argValue.Raw()))
						}
					}
				}
			}
		}
	}

	compDef := ast.FindNode(ast.GetAncestors(pccNode), func(anc ast.Node) bool {
		return ast.IsRuleNameOneOf(anc, []string{"local-comp-def", "global-comp-def"})
	})
	if compDef == nil {
		return ""
	}

	return ew.getCompParamDefault(compDef, paramName)
}

func (ew *errorWrapper) getCompParamDefault(compDef ast.Node, paramName string) string {
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

	compCompParam := ast.FindNodeByRuleName(compParamType.Children(), "comp-comp-param")
	if compCompParam == nil {
		return ""
	}

	defaValue := ast.FindNodeByRuleName(compCompParam.Children(), "comp-param-defa-value")
	if defaValue == nil {
		return ""
	}

	return strings.TrimSpace(string(defaValue.Raw()))
}

func (ew *errorWrapper) getCompCallName(node ast.Node) string {
	compCallNameNode := ast.FindNodeByRuleName(node.Children(), "comp-call-name")
	if compCallNameNode != nil {
		return strings.TrimSpace(string(compCallNameNode.Raw()))
	}
	return ""
}

func (ew *errorWrapper) isBuiltInComp(compCallName string) bool {
	if util.InSliceString(compCallName, []string{"LINK"}) {
		return true
	}
	return false
}

func (ew *errorWrapper) isInCallStack(name string, callStack []string) bool {
	for _, n := range callStack {
		if n == name {
			return true
		}
	}
	return false
}

func (ew *errorWrapper) findCompDef(compCallNode ast.Node, name string) ast.Node {
	globalCompDefAnc := ast.FindNode(ast.GetAncestors(compCallNode), func(anc ast.Node) bool {
		return ast.IsRuleName(anc, "global-comp-def")
	})

	localCompDefSrc := ew.root
	if globalCompDefAnc != nil {
		localCompDefSrc = globalCompDefAnc
	}

	localCompDef := ast.FindLocalCompDef(localCompDefSrc, name)
	if localCompDef != nil {
		return localCompDef
	}

	globalCompDef := ast.FindGlobalCompDef(ew.root, name)
	if globalCompDef != nil {
		return globalCompDef
	}

	return nil
}

func (ew *errorWrapper) getCompDefContent(compDef ast.Node) ast.Node {
	for _, child := range compDef.Children() {
		if child.Rule() == nil {
			continue
		}
		ruleName := child.Rule().Name()
		if ruleName == "local-comp-def-content" || ruleName == "global-comp-def-content" {
			return child
		}
	}
	return nil
}

func (ew *errorWrapper) wrapWithErr(self ast.Node, title, msg string, block bool) {
	var errNode ast.Node
	if block {
		errNode = ew.createBlockError(self, title, msg)
	} else {
		errNode = ew.createInlineError(self, title, msg)
	}

	self.SetRule(errNode.Rule())
	self.SetChildren(errNode.Children())
	self.SetRaw(errNode.Raw())
}

func (ew *errorWrapper) createBlockError(node ast.Node, title, msg string) ast.Node {
	return ew.createError("block-error", node, title, msg)
}

func (ew *errorWrapper) createInlineError(node ast.Node, title, msg string) ast.Node {
	return ew.createError("inline-error", node, title, msg)
}

func (ew *errorWrapper) createError(errRuleName string, node ast.Node, title, msg string) ast.Node {
	err := rule.NewDynamic(errRuleName)
	errTitle := rule.NewDynamic("error-title")
	errMsg := rule.NewDynamic("error-message")
	self := rule.NewDynamic("self")

	errNode := ast.DefaultEmptyNode()
	errNode.SetRule(err)

	errTitleNode := ast.DefaultEmptyNode()
	errTitleNode.SetRule(errTitle)
	errTitleNode.SetParent(errNode)
	errTitleNode.SetRaw([]byte(title))

	errMsgNode := ast.DefaultEmptyNode()
	errMsgNode.SetRule(errMsg)
	errMsgNode.SetParent(errNode)
	errMsgNode.SetRaw([]byte(msg))

	selfNode := ast.DefaultEmptyNode()
	selfNode.SetRule(self)
	selfNode.SetParent(errNode)
	selfNode.SetChildren(node.Children())

	errNode.SetChildren([]ast.Node{
		errTitleNode,
		errMsgNode,
		selfNode,
	})

	return errNode
}
