package renderer

import (
	"io"

	"github.com/umono-cms/compono/ast"
	"github.com/umono-cms/compono/logger"
	"github.com/umono-cms/compono/renderer/hook"
	"github.com/umono-cms/compono/renderer/html"
)

type Renderer interface {
	Render(writer io.Writer, root ast.Node) error
}

type HookSetter interface {
	SetRendererHooks([]hook.RendererHookFunc)
}

func DefaultRenderer(log logger.Logger) Renderer {
	return html.NewRenderer(log)
}
