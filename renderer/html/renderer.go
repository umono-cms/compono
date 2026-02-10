package html

import (
	"io"
	"strings"

	"github.com/umono-cms/compono/ast"
	"github.com/umono-cms/compono/logger"
)

type renderer struct {
	logger          logger.Logger
	renderableNodes []renderableNode
	root            ast.Node
	builtinCompMap  map[string]builtinComponent
}

func NewRenderer(log logger.Logger) *renderer {

	r := &renderer{
		logger: log,
	}

	r.renderableNodes = []renderableNode{
		newErr(r),
		newRoot(r),
		newRootContent(r),
		newCompCall(r),
		newParamCompCall(r),
		newNonVoidElement(r),
		newNonVoidElementContent(r),
		newParamRefInLocalCompDef(r),
		newParamRefInGlobalCompDef(r),
		newPlain(r),
		newCodeBlock(r),
		newCodeBlockContent(r),
		newInlineCode(r),
		newInlineCodeContent(r),
		newRaw(r),
		newLinkElement(r),
		newLinkTextElement(r),
		newLinkURLElement(r),
		newBr(r),
	}

	r.builtinCompMap = make(map[string]builtinComponent)

	builtinComps := []builtinComponent{
		newLink(r),
	}

	for _, bc := range builtinComps {
		r.builtinCompMap[bc.Name()] = bc
	}

	return r
}

func (r *renderer) Render(writer io.Writer, root ast.Node) error {

	r.root = root

	_, err := writer.Write([]byte(r.render(root)))
	if err != nil {
		return err
	}

	return nil
}

func (r *renderer) render(node ast.Node) string {
	rn := r.findRenderable(nil, node)
	if rn != nil {
		return renderNode(rn, nil, node)
	}
	return ""
}

func (r *renderer) renderChildren(invoker renderableNode, children []ast.Node) string {
	result := ""
	for _, child := range children {
		re := r.findRenderable(invoker, child)
		if re != nil {
			result += renderNode(re, invoker, child)
		}
	}
	return result
}

func (r *renderer) findRenderable(invoker renderableNode, node ast.Node) renderableNode {
	for _, rn := range r.renderableNodes {
		if cond := rn.Condition(invoker, node); cond {
			return rn.New()
		}
	}
	return nil
}

func (r *renderer) findLocalCompDef(srcNode ast.Node, name string) ast.Node {
	localCompDefWrapper := ast.FindNodeByRuleName(srcNode.Children(), "local-comp-def-wrapper")
	if localCompDefWrapper == nil {
		return nil
	}

	return ast.FindNode(localCompDefWrapper.Children(), func(child ast.Node) bool {
		if !ast.IsRuleName(child, "local-comp-def") {
			return false
		}

		localCompDefHead := ast.FindNodeByRuleName(child.Children(), "local-comp-def-head")
		if localCompDefHead == nil {
			return false
		}

		localCompName := ast.FindNodeByRuleName(localCompDefHead.Children(), "local-comp-name")
		if localCompName == nil {
			return false
		}

		if strings.TrimSpace(string(localCompName.Raw())) != strings.TrimSpace(name) {
			return false
		}

		return true
	})
}

func (r *renderer) findGlobalCompDef(name string) ast.Node {
	globalCompDefWrapper := ast.FindNodeByRuleName(r.root.Children(), "global-comp-def-wrapper")
	if globalCompDefWrapper == nil {
		return nil
	}

	return ast.FindNode(globalCompDefWrapper.Children(), func(child ast.Node) bool {
		if !ast.IsRuleName(child, "global-comp-def") {
			return false
		}

		globalCompName := ast.FindNodeByRuleName(child.Children(), "global-comp-name")
		if globalCompName == nil {
			return false
		}

		if strings.TrimSpace(string(globalCompName.Raw())) != strings.TrimSpace(name) {
			return false
		}

		return true
	})
}

func (r *renderer) findBuiltinComp(name string) builtinComponent {
	bc, ok := r.builtinCompMap[strings.TrimSpace(name)]
	if !ok {
		return nil
	}
	return bc.New()
}
