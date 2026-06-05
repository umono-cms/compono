package compono

import (
	"fmt"
	"io"

	"github.com/umono-cms/compono/ast"
	"github.com/umono-cms/compono/builtin"
	"github.com/umono-cms/compono/errwrap"
	"github.com/umono-cms/compono/logger"
	"github.com/umono-cms/compono/parser"
	"github.com/umono-cms/compono/renderer"
	"github.com/umono-cms/compono/rule"
	"github.com/umono-cms/compono/util"
	"github.com/umono-cms/compono/validator"
)

type ErrorCode int

const (
	ErrInvalidGlobalName ErrorCode = iota + 1
	ErrGlobalAlreadyRegistered
	ErrGlobalNotExist
	ErrInvalidAST
	ErrRender
	ErrUnsupportedType
	ErrUnsupportedKeyNotation
)

type Compono interface {
	Convert(source []byte, writer io.Writer, opts ...ConvertOption) error
	RegisterGlobalComponent(string, []byte) error
	UnregisterGlobalComponent(string) error
	Parser() parser.Parser
	SetParser(parser.Parser)
	Renderer() renderer.Renderer
	SetRenderer(renderer.Renderer)
	Validator() validator.Validator
	SetValidator(validator.Validator)
	ErrorWrapper() errwrap.ErrorWrapper
	SetErrorWrapper(errwrap.ErrorWrapper)
	Logger() logger.Logger
	SetLogger(logger.Logger)
}

func New() Compono {
	log := logger.NewLogger()

	p := parser.DefaultParser(log)
	r := renderer.DefaultRenderer(log)
	v := validator.DefaultValidator()
	ew := errwrap.DefaultErrorWrapper()

	gw := ast.DefaultEmptyNode()
	gw.SetRule(rule.NewGlobalCompDefWrapper())

	bw := ast.DefaultEmptyNode()
	bw.SetRule(rule.NewDynamic("builtin-comp-wrapper"))

	c := &compono{
		parser:         p,
		renderer:       r,
		validator:      v,
		errorWrapper:   ew,
		logger:         log,
		globalWrapper:  gw,
		builtinWrapper: bw,
	}

	c.fillBuiltins()

	return c
}

type compono struct {
	parser         parser.Parser
	renderer       renderer.Renderer
	validator      validator.Validator
	errorWrapper   errwrap.ErrorWrapper
	logger         logger.Logger
	globalWrapper  ast.Node
	builtinWrapper ast.Node
}

func (c *compono) Convert(source []byte, writer io.Writer, opts ...ConvertOption) error {
	if len(source) == 0 {
		return nil
	}

	cfg, err := c.newConvertConfig(opts...)
	if err != nil {
		return err
	}

	root := c.parser.Parse(source, ast.DefaultRootNode())

	contextWrapper, err := buildContextWrapper(root, cfg.contextValues)
	if err != nil {
		return err
	}
	root.SetChildren(append(root.Children(), contextWrapper))

	globalWrapper := c.newGlobalWrapper(cfg.globalComponents)
	globalWrapper.SetParent(root)
	root.SetChildren(append(root.Children(), globalWrapper))

	c.builtinWrapper.SetParent(root)
	root.SetChildren(append(root.Children(), c.builtinWrapper))

	err = c.validator.Validate(root)
	if err != nil {
		return NewComponoError(ErrInvalidAST, err.Error())
	}

	c.errorWrapper.Wrap(root)

	if hs, ok := c.renderer.(renderer.HookSetter); ok {
		hs.SetRendererHooks(cfg.rendererHooks)
	}

	err = c.renderer.Render(writer, root)
	if err != nil {
		return NewComponoError(ErrRender, err.Error())
	}
	return nil
}

func (c *compono) RegisterGlobalComponent(name string, source []byte) error {
	if registered := c.getGlobalCompDefByName(name); registered != nil {
		return NewComponoError(ErrGlobalAlreadyRegistered, fmt.Sprintf("cannot register global component %q: already registered", name))
	}

	parsed, err := c.newGlobalComponentNode(name, source)
	if err != nil {
		return err
	}

	c.globalWrapper.SetChildren(append([]ast.Node{parsed}, c.globalWrapper.Children()...))

	return nil
}

func (c *compono) UnregisterGlobalComponent(name string) error {
	if registered := c.getGlobalCompDefByName(name); registered == nil {
		return NewComponoError(ErrGlobalNotExist, fmt.Sprintf("cannot unregister global component %q: does not exist", name))
	}

	globalComps := ast.FilterNodes(c.globalWrapper.Children(), func(gc ast.Node) bool {
		globalCompName := ast.FindNodeByRuleName(gc.Children(), "global-comp-name")
		if string(globalCompName.Raw()) == name {
			return false
		}
		return true
	})

	c.globalWrapper.SetChildren(globalComps)
	return nil
}

func (c *compono) Parser() parser.Parser {
	return c.parser
}

func (c *compono) SetParser(parser parser.Parser) {
	c.parser = parser
}

func (c *compono) Renderer() renderer.Renderer {
	return c.renderer
}

func (c *compono) SetRenderer(renderer renderer.Renderer) {
	c.renderer = renderer
}

func (c *compono) Validator() validator.Validator {
	return c.validator
}

func (c *compono) SetValidator(vldtr validator.Validator) {
	c.validator = vldtr
}

func (c *compono) ErrorWrapper() errwrap.ErrorWrapper {
	return c.errorWrapper
}

func (c *compono) SetErrorWrapper(ew errwrap.ErrorWrapper) {
	c.errorWrapper = ew
}

func (c *compono) Logger() logger.Logger {
	return c.logger
}

func (c *compono) SetLogger(logger logger.Logger) {
	c.logger = logger
}

func (c *compono) getGlobalCompDefByName(name string) ast.Node {
	for _, gcd := range c.globalWrapper.Children() {
		if gcd.Rule().Name() != "global-comp-def" {
			continue
		}
		for _, child := range gcd.Children() {
			if child.Rule().Name() == "global-comp-name" && name == string(child.Raw()) {
				return gcd
			}
		}
	}
	return nil
}

func (c *compono) cloneGlobalComponents() []ast.Node {
	children := c.globalWrapper.Children()
	if len(children) == 0 {
		return nil
	}

	cloned := make([]ast.Node, len(children))
	for i, child := range children {
		cloned[i] = c.cloneNode(child)
	}
	return cloned
}

func (c *compono) cloneNode(node ast.Node) ast.Node {
	if node == nil {
		return nil
	}

	clone := ast.DefaultEmptyNode()
	clone.SetRule(node.Rule())
	clone.SetRaw(node.Raw())

	children := node.Children()
	if len(children) > 0 {
		clonedChildren := make([]ast.Node, len(children))
		for i, child := range children {
			clonedChild := c.cloneNode(child)
			clonedChild.SetParent(clone)
			clonedChildren[i] = clonedChild
		}
		clone.SetChildren(clonedChildren)
	}

	return clone
}

func (c *compono) newConvertConfig(opts ...ConvertOption) (*convertConfig, error) {
	cfg := &convertConfig{}
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		if err := opt.applyConvert(c, cfg); err != nil {
			return nil, err
		}
	}
	return cfg, nil
}

func (c *compono) newGlobalWrapper(injected []ast.Node) ast.Node {
	gw := ast.DefaultEmptyNode()
	gw.SetRule(rule.NewGlobalCompDefWrapper())

	children := append([]ast.Node{}, injected...)
	children = append(children, c.cloneGlobalComponents()...)
	gw.SetChildren(children)

	for _, child := range gw.Children() {
		child.SetParent(gw)
	}

	return gw
}

func (c *compono) newGlobalComponentNode(name string, source []byte) (ast.Node, error) {
	if !util.IsScreamingSnakeCase(name) {
		return nil, NewComponoError(ErrInvalidGlobalName, fmt.Sprintf("invalid global component name %q: must be SCREAMING_SNAKE_CASE (digits allowed)", name))
	}

	node := ast.DefaultEmptyNode()
	node.SetRule(rule.NewGlobalCompDef())

	parsed := c.parser.Parse(source, node)

	globalCompName := ast.DefaultEmptyNode()
	globalCompName.SetRule(rule.NewGlobalCompName())
	globalCompName.SetParent(parsed)
	globalCompName.SetRaw([]byte(name))

	parsed.SetChildren(append([]ast.Node{globalCompName}, parsed.Children()...))
	return parsed, nil
}

func (c *compono) fillBuiltins() {
	c.builtinWrapper.SetChildren(builtin.BuildASTNodes(c.builtinWrapper))
}

type ComponoError struct {
	Code    ErrorCode
	Message string
}

func (e *ComponoError) Error() string { return e.Message }

func NewComponoError(code ErrorCode, msg string) *ComponoError {
	return &ComponoError{Code: code, Message: msg}
}
