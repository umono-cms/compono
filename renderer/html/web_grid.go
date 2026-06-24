package html

import (
	"encoding/json"
	"html"
	"strings"

	"github.com/umono-cms/compono/ast"
	"github.com/umono-cms/compono/rule"
)

// Deprecated: WEB_GRID is deprecated and will be removed in v1.
// Compono will now only carry semantic content. Built-in components
// that tell the presentation layer what to do will no longer be added to Compono.
type webGrid struct {
	renderer *renderer
}

func newWebGrid(rend *renderer) builtinComponent {
	return &webGrid{
		renderer: rend,
	}
}

func (wg *webGrid) New() builtinComponent {
	return newWebGrid(wg.renderer)
}

func (_ *webGrid) Name() string {
	return "WEB_GRID"
}

func (wg *webGrid) Render(invoker renderableNode, node ast.Node) string {
	attrs := []string{
		`data-grid-template-columns="` + html.EscapeString(wg.joinScalarArray(invoker, node, "grid-template-columns")) + `"`,
		`data-grid-template-rows="` + html.EscapeString(wg.joinScalarArray(invoker, node, "grid-template-rows")) + `"`,
		`data-grid-template-areas='` + wg.mustJSON(wg.areaMatrix(invoker, node, "grid-template-areas")) + `'`,
	}

	for _, breakpoint := range []string{"sm", "md", "lg", "xl", "xxl"} {
		columnsName := breakpoint + "-grid-template-columns"
		rowsName := breakpoint + "-grid-template-rows"
		areasName := breakpoint + "-grid-template-areas"
		if !wg.hasExplicitArg(node, columnsName) {
			continue
		}

		attrs = append(attrs,
			`data-`+columnsName+`="`+html.EscapeString(wg.joinScalarArray(invoker, node, columnsName))+`"`,
			`data-`+rowsName+`="`+html.EscapeString(wg.joinScalarArray(invoker, node, rowsName))+`"`,
			`data-`+areasName+`='`+wg.mustJSON(wg.areaMatrix(invoker, node, areasName))+`'`,
		)
	}

	itemsValue := wg.resolveArg(invoker, node, "items")
	renderedItems := make([]string, 0, len(itemsValue.Items))
	for _, item := range itemsValue.Items {
		if item.Type != "record" {
			continue
		}

		area, ok := item.Fields["grid-area"]
		if !ok || area.Type != "string" {
			continue
		}

		component, ok := item.Fields["component"]
		if !ok || component.Type != "comp" {
			continue
		}

		renderedItems = append(renderedItems, `<compono-web-grid-item data-grid-area="`+html.EscapeString(area.Raw)+`">`+wg.renderComponent(invoker, node, component.Raw, component.Scope)+`</compono-web-grid-item>`)
	}

	return `<compono-web-grid ` + strings.Join(attrs, " ") + `>` + strings.Join(renderedItems, "") + `</compono-web-grid>`
}

func (wg *webGrid) renderComponent(invoker renderableNode, parent ast.Node, name string, scope ast.Node) string {
	renderCtx := newPassthroughRenderable(parent, invoker)

	localCompDefSrc := scope
	if localCompDefSrc == nil {
		localCompDefSrc = localCompSourceFromNode(parent, wg.renderer.root)
	}

	localCompDef := wg.renderer.findLocalCompDef(localCompDefSrc, name)
	if localCompDef == nil {
		currentGlobalCompDef := ast.FindNode(ast.GetAncestors(parent), func(anc ast.Node) bool {
			return ast.IsRuleName(anc, "global-comp-def")
		})
		if currentGlobalCompDef != nil && currentGlobalCompDef != localCompDefSrc {
			localCompDef = wg.renderer.findLocalCompDef(currentGlobalCompDef, name)
		}
	}
	if localCompDef != nil {
		localCompDefContent := ast.FindNodeByRuleName(localCompDef.Children(), "local-comp-def-content")
		if localCompDefContent == nil {
			return ""
		}
		return wg.renderer.renderChildren(renderCtx, localCompDefContent.Children())
	}

	globalCompDef := wg.renderer.findGlobalCompDef(name)
	if globalCompDef != nil {
		globalCompDefContent := ast.FindNodeByRuleName(globalCompDef.Children(), "global-comp-def-content")
		if globalCompDefContent == nil {
			return ""
		}
		return wg.renderer.renderChildren(renderCtx, globalCompDefContent.Children())
	}

	compCall := ast.DefaultEmptyNode()
	compCall.SetRule(rule.NewDynamic("block-comp-call"))
	compCall.SetParent(parent)

	compCallName := ast.DefaultEmptyNode()
	compCallName.SetRule(rule.NewDynamic("comp-call-name"))
	compCallName.SetParent(compCall)
	compCallName.SetRaw([]byte(name))

	compCall.SetChildren([]ast.Node{compCallName})
	re := wg.renderer.findRenderable(renderCtx, compCall)
	if re == nil {
		return ""
	}
	return renderNode(re, renderCtx, compCall)
}

func (wg *webGrid) resolveArg(invoker renderableNode, compCall ast.Node, name string) ast.ResolvedValue {
	arg := ast.GetCompCallArgByParamName(ast.GetCompCallArgsFromCompCall(compCall), name)
	if arg != nil {
		invokerAncestors := append([]ast.Node{compCall}, webGridInvokerAncestors(invoker)...)
		return ast.ResolveCompCallArgValue(wg.renderer.root, arg, invokerAncestors, compCall)
	}
	return ast.ResolveParamDefaultFromCompCall(wg.renderer.root, compCall, name)
}

func (wg *webGrid) joinScalarArray(invoker renderableNode, compCall ast.Node, name string) string {
	value := wg.resolveArg(invoker, compCall, name)
	parts := make([]string, 0, len(value.Items))
	for _, item := range value.Items {
		parts = append(parts, item.Raw)
	}
	return strings.Join(parts, " ")
}

func (wg *webGrid) areaMatrix(invoker renderableNode, compCall ast.Node, name string) [][]string {
	value := wg.resolveArg(invoker, compCall, name)
	rows := make([][]string, 0, len(value.Items))
	for _, row := range value.Items {
		cols := make([]string, 0, len(row.Items))
		for _, col := range row.Items {
			cols = append(cols, col.Raw)
		}
		rows = append(rows, cols)
	}
	return rows
}

func (wg *webGrid) mustJSON(value any) string {
	raw, _ := json.Marshal(value)
	return string(raw)
}

func (wg *webGrid) hasExplicitArg(compCall ast.Node, name string) bool {
	return ast.GetCompCallArgByParamName(ast.GetCompCallArgsFromCompCall(compCall), name) != nil
}

func webGridInvokerAncestors(invoker renderableNode) []ast.Node {
	if invoker == nil {
		return nil
	}
	return append([]ast.Node{invoker.Node()}, webGridInvokerAncestors(invoker.Invoker())...)
}
