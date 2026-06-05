package html

import (
	"io"
	"strings"

	"github.com/umono-cms/compono/ast"
	"github.com/umono-cms/compono/logger"
	"github.com/umono-cms/compono/renderer/hook"
)

type renderer struct {
	logger          logger.Logger
	renderableNodes []renderableNode
	root            ast.Node
	builtinCompMap  map[string]builtinComponent
	hooks           []hook.RendererHookFunc
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
		newNonVoidElement(r),
		newNonVoidElementContent(r),
		newContextRef(r),
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
		newImage(r),
		newWebGrid(r),
		newNavigation(r),
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
	return ast.FindLocalCompDef(srcNode, name)
}

func (r *renderer) findGlobalCompDef(name string) ast.Node {
	return ast.FindGlobalCompDef(r.root, name)
}

func (r *renderer) findBuiltinComp(name string) builtinComponent {
	if r.findBuiltinCompDef(name) == nil {
		return nil
	}

	bc, ok := r.builtinCompMap[strings.TrimSpace(name)]
	if !ok {
		return nil
	}
	return bc.New()
}

func (r *renderer) findBuiltinCompDef(name string) ast.Node {
	return ast.FindBuiltinCompDef(r.root, name)
}

func (r *renderer) SetRendererHooks(hooks []hook.RendererHookFunc) {
	r.hooks = hooks
}

func (r *renderer) applyHooks(output string, kind hook.Kind, name string, params hook.Params) string {
	if len(r.hooks) == 0 {
		return output
	}
	ctx := hook.RendererHookContext{
		Kind:   kind,
		Name:   name,
		Params: params,
		Output: output,
	}
	for _, h := range r.hooks {
		ctx.Output = h(ctx)
	}
	return ctx.Output
}

func (r *renderer) extractBuiltinParams(invoker renderableNode, compCall ast.Node) hook.Params {
	params := hook.Params{}

	compCallName := ast.FindNodeByRuleName(compCall.Children(), "comp-call-name")
	if compCallName == nil {
		return params
	}
	name := strings.TrimSpace(string(compCallName.Raw()))

	compDef := ast.FindBuiltinCompDef(r.root, name)
	if compDef == nil {
		return params
	}

	for _, compParam := range ast.GetCompParamsFromCompDef(compDef) {
		paramName := ast.GetParamNameFromCompParam(compParam)
		if paramName == "" {
			continue
		}
		resolved := ast.ResolveParamDefaultFromCompCall(r.root, compCall, paramName)
		if !resolved.IsZero() && resolved.MissingContextKey == "" {
			params[paramName] = resolvedValueToHookParam(resolved)
		}
	}

	for _, arg := range ast.GetCompCallArgsFromCompCall(compCall) {
		argName := ast.GetArgNameFromCompCallArg(arg)
		if argName == "" {
			continue
		}
		resolved := ast.ResolveCompCallArgValue(r.root, arg, getAncestorsByInvoker(invoker), compCall)
		if resolved.IsZero() || resolved.MissingContextKey != "" {
			continue
		}
		params[argName] = resolvedValueToHookParam(resolved)
	}

	return params
}

func resolvedValueToHookParam(value ast.ResolvedValue) hook.ParamValue {
	switch value.Type {
	case "array":
		items := make([]hook.ParamValue, 0, len(value.Items))
		for _, item := range value.Items {
			items = append(items, resolvedValueToHookParam(item))
		}
		return hook.NewArray(items)
	case "record":
		fields := make(map[string]hook.ParamValue, len(value.Fields))
		for key, field := range value.Fields {
			fields[key] = resolvedValueToHookParam(field)
		}
		return hook.NewRecord(fields)
	default:
		return hook.NewString(strings.TrimSpace(value.Raw))
	}
}
