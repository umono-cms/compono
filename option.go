package compono

import (
	"fmt"

	"github.com/umono-cms/compono/ast"
)

type ConvertOption interface {
	applyConvert(*compono, *convertConfig) error
}

type convertOptionFunc func(*compono, *convertConfig) error

func (f convertOptionFunc) applyConvert(c *compono, cfg *convertConfig) error {
	return f(c, cfg)
}

type globalComponentNode struct {
	name          string
	node          ast.Node
	source        []byte
	isolated      bool
	subComponents []*globalComponentNode
}

type convertConfig struct {
	globalComponents   []*globalComponentNode
	contextValues      map[string]any
	isolatedScope      bool
	attributeHook      AttributeHookFunc
	attributeHookSet   bool
	urlReporter        URLReporterFunc
	urlReporterSet     bool
	currentResolver    CurrentResolverFunc
	currentResolverSet bool
}

type globalComponentOption struct {
	name   string
	source []byte
	opts   []ConvertOption
}

// WithGlobalComponent returns a ConvertOption that injects a global component
// for a single conversion. The name must be SCREAMING_SNAKE_CASE. Sub-options
// may include nested WithGlobalComponent calls (which become sub-components of
// this global) and WithIsolatedScope. WithContext, WithAttributeHook,
// WithURLReporter and WithCurrentResolver are not permitted inside a global
// component at any depth.
func WithGlobalComponent(name string, source []byte, opts ...ConvertOption) ConvertOption {
	return &globalComponentOption{
		name:   name,
		source: source,
		opts:   opts,
	}
}

func (o *globalComponentOption) applyConvert(c *compono, cfg *convertConfig) error {
	globalNode, err := c.newGlobalComponentNode(o.name, o.source)
	if err != nil {
		return err
	}

	subCfg := &convertConfig{}
	for _, opt := range o.opts {
		if opt == nil {
			continue
		}
		if isConversionOption(opt) {
			return NewComponoError(ErrConversionOptionInGlobal, fmt.Sprintf("conversion option not allowed inside global component %q: WithContext, WithAttributeHook, WithURLReporter and WithCurrentResolver are forbidden in global scope", o.name))
		}
		if err := opt.applyConvert(c, subCfg); err != nil {
			return err
		}
	}

	if err := checkDuplicateSubComponents(subCfg.globalComponents); err != nil {
		return err
	}

	node := &globalComponentNode{
		name:          o.name,
		node:          globalNode,
		source:        o.source,
		isolated:      subCfg.isolatedScope,
		subComponents: subCfg.globalComponents,
	}

	cfg.globalComponents = append(cfg.globalComponents, node)
	return nil
}

func isConversionOption(opt ConvertOption) bool {
	switch opt.(type) {
	case *contextOption, *attributeHookOption, *urlReporterOption, *currentResolverOption:
		return true
	default:
		return false
	}
}

func checkDuplicateSubComponents(components []*globalComponentNode) error {
	seen := make(map[string]bool)
	for _, comp := range components {
		if seen[comp.name] {
			return NewComponoError(ErrDuplicateSubComponent, fmt.Sprintf("duplicate sub-component %q in the same scope", comp.name))
		}
		seen[comp.name] = true
	}
	return nil
}

type contextOption struct {
	values map[string]any
}

// WithContext returns a ConvertOption that adds context values for template
// rendering. This option is not permitted inside WithGlobalComponent.
func WithContext(values map[string]any) ConvertOption {
	return &contextOption{values: values}
}

func (o *contextOption) applyConvert(_ *compono, cfg *convertConfig) error {
	if len(o.values) == 0 {
		return nil
	}

	if cfg.contextValues == nil {
		cfg.contextValues = map[string]any{}
	}

	for key, value := range o.values {
		cfg.contextValues[key] = value
	}

	return nil
}

type isolatedScopeOption struct{}

// WithIsolatedScope returns a ConvertOption that marks a global component's
// scope as isolated. Calls in the global's body and its sub components' bodies
// resolve only to locals, its sub components and built-ins. Given to Convert
// directly, it returns an error.
func WithIsolatedScope() ConvertOption {
	return &isolatedScopeOption{}
}

func (o *isolatedScopeOption) applyConvert(_ *compono, cfg *convertConfig) error {
	cfg.isolatedScope = true
	return nil
}
