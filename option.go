package compono

import (
	"github.com/umono-cms/compono/ast"
	"github.com/umono-cms/compono/renderer/hook"
)

type ConvertOption interface {
	applyConvert(*compono, *convertConfig) error
}

type convertOptionFunc func(*compono, *convertConfig) error

func (f convertOptionFunc) applyConvert(c *compono, cfg *convertConfig) error {
	return f(c, cfg)
}

type convertConfig struct {
	globalComponents []ast.Node
	contextValues    map[string]any
	rendererHooks    []hook.RendererHookFunc
}

func WithGlobalComponent(name string, source []byte) ConvertOption {
	return convertOptionFunc(func(c *compono, cfg *convertConfig) error {
		globalComponent, err := c.newGlobalComponentNode(name, source)
		if err != nil {
			return err
		}

		cfg.globalComponents = append(cfg.globalComponents, globalComponent)
		return nil
	})
}

func WithContext(values map[string]any) ConvertOption {
	return convertOptionFunc(func(_ *compono, cfg *convertConfig) error {
		if len(values) == 0 {
			return nil
		}

		if cfg.contextValues == nil {
			cfg.contextValues = map[string]any{}
		}

		for key, value := range values {
			cfg.contextValues[key] = value
		}

		return nil
	})
}

func WithRendererHook(fn hook.RendererHookFunc) ConvertOption {
	return convertOptionFunc(func(_ *compono, cfg *convertConfig) error {
		if fn == nil {
			return nil
		}
		cfg.rendererHooks = append(cfg.rendererHooks, fn)
		return nil
	})
}
