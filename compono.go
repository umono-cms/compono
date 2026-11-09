package compono

import (
	"errors"
	"fmt"
	"io"

	"github.com/umono-cms/compono/ast"
	"github.com/umono-cms/compono/builtin"
	"github.com/umono-cms/compono/errwrap"
	"github.com/umono-cms/compono/internal/attrhook"
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
	_                              // 2: ErrGlobalAlreadyRegistered, removed in v0.7
	_                              // 3: ErrGlobalNotExist, removed in v0.7
	ErrInvalidAST
	ErrRender
	ErrUnsupportedType
	ErrUnsupportedKeyNotation
	_ // 8: ErrErrorStylesheetAlreadySet, removed in v0.7
	ErrConversionOptionInGlobal
	ErrIsolatedScopeInConvert
	ErrDuplicateSubComponent
	ErrAttributeHookAlreadySet
	ErrInvalidAttributeName
	ErrAttributeConflict
	ErrDuplicateGlobalComponent
	ErrURLReporterAlreadySet
	ErrCurrentResolverAlreadySet
)

type Compono interface {
	Convert(source []byte, writer io.Writer, opts ...ConvertOption) ([]Diagnostic, error)
	Parser() parser.Parser
	Renderer() renderer.Renderer
	Validator() validator.Validator
	ErrorWrapper() errwrap.ErrorWrapper
	Logger() logger.Logger
}

func New() Compono {
	log := logger.NewLogger()

	p := parser.DefaultParser(log)
	r := renderer.DefaultRenderer(log)
	v := validator.DefaultValidator()
	ew := errwrap.DefaultErrorWrapper()

	c := &compono{
		parser:       p,
		renderer:     r,
		validator:    v,
		errorWrapper: ew,
		logger:       log,
	}

	return c
}

type compono struct {
	parser       parser.Parser
	renderer     renderer.Renderer
	validator    validator.Validator
	errorWrapper errwrap.ErrorWrapper
	logger       logger.Logger
}

func (c *compono) Convert(source []byte, writer io.Writer, opts ...ConvertOption) ([]Diagnostic, error) {
	if len(source) == 0 {
		return nil, nil
	}

	cfg, err := c.newConvertConfig(opts...)
	if err != nil {
		return nil, err
	}

	root := c.parser.Parse(source, ast.DefaultRootNode())

	contextWrapper, err := buildContextWrapper(root, cfg.contextValues)
	if err != nil {
		return nil, err
	}
	root.SetChildren(append(root.Children(), contextWrapper))

	globalWrapper := c.newGlobalWrapper(cfg.globalComponents)
	globalWrapper.SetParent(root)
	root.SetChildren(append(root.Children(), globalWrapper))

	builtinWrapper := ast.DefaultEmptyNode()
	builtinWrapper.SetRule(rule.NewDynamic("builtin-comp-wrapper"))
	builtinWrapper.SetParent(root)
	builtinWrapper.SetChildren(builtin.BuildASTNodes(builtinWrapper))
	root.SetChildren(append(root.Children(), builtinWrapper))

	err = c.validator.Validate(root)
	if err != nil {
		return nil, NewComponoError(ErrInvalidAST, err.Error())
	}

	c.errorWrapper.Wrap(root)

	rendered, err := c.renderer.Render(writer, root, renderer.Options{
		AttributeHook:   cfg.attributeHook,
		URLReporter:     cfg.urlReporter,
		CurrentResolver: cfg.currentResolver,
	})
	if err != nil {
		switch {
		case errors.Is(err, attrhook.ErrInvalidAttributeName):
			return nil, NewComponoError(ErrInvalidAttributeName, err.Error())
		case errors.Is(err, attrhook.ErrAttributeConflict):
			return nil, NewComponoError(ErrAttributeConflict, err.Error())
		default:
			return nil, NewComponoError(ErrRender, err.Error())
		}
	}
	return newDiagnostics(root, sourcesOf(root, source, cfg.globalComponents), rendered), nil
}

func (c *compono) Parser() parser.Parser {
	return c.parser
}

func (c *compono) Renderer() renderer.Renderer {
	return c.renderer
}

func (c *compono) Validator() validator.Validator {
	return c.validator
}

func (c *compono) ErrorWrapper() errwrap.ErrorWrapper {
	return c.errorWrapper
}

func (c *compono) Logger() logger.Logger {
	return c.logger
}

func (c *compono) newConvertConfig(opts ...ConvertOption) (*convertConfig, error) {
	cfg := &convertConfig{}
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		if _, ok := opt.(*isolatedScopeOption); ok {
			return nil, NewComponoError(ErrIsolatedScopeInConvert, "WithIsolatedScope is not allowed at the top level of Convert")
		}
		if err := opt.applyConvert(c, cfg); err != nil {
			return nil, err
		}
	}
	seen := make(map[string]bool)
	for _, comp := range cfg.globalComponents {
		if seen[comp.name] {
			return nil, NewComponoError(ErrDuplicateGlobalComponent, fmt.Sprintf("duplicate global component %q in the root scope", comp.name))
		}
		seen[comp.name] = true
	}
	return cfg, nil
}

func sourcesOf(root ast.Node, source []byte, globals []*globalComponentNode) map[ast.Node][]byte {
	sources := map[ast.Node][]byte{root: source}
	addGlobalSources(sources, globals)
	return sources
}

func addGlobalSources(sources map[ast.Node][]byte, globals []*globalComponentNode) {
	for _, g := range globals {
		sources[g.node] = g.source
		addGlobalSources(sources, g.subComponents)
	}
}

func (c *compono) newGlobalWrapper(injected []*globalComponentNode) ast.Node {
	gw := ast.DefaultEmptyNode()
	gw.SetRule(rule.NewGlobalCompDefWrapper())

	var children []ast.Node
	for _, node := range injected {
		c.buildGlobalCompDefNode(node)
		children = append(children, node.node)
	}
	gw.SetChildren(children)

	for _, child := range gw.Children() {
		child.SetParent(gw)
	}

	return gw
}

func (c *compono) buildGlobalCompDefNode(node *globalComponentNode) {
	if node.isolated {
		isolatedMarker := ast.DefaultEmptyNode()
		isolatedMarker.SetRule(rule.NewDynamic("isolated-scope"))
		isolatedMarker.SetParent(node.node)
		node.node.SetChildren(append(node.node.Children(), isolatedMarker))
	}

	if len(node.subComponents) > 0 {
		subWrapper := ast.DefaultEmptyNode()
		subWrapper.SetRule(rule.NewGlobalCompDefWrapper())
		subWrapper.SetParent(node.node)

		var subChildren []ast.Node
		for _, sub := range node.subComponents {
			c.buildGlobalCompDefNode(sub)
			subChildren = append(subChildren, sub.node)
		}
		subWrapper.SetChildren(subChildren)

		for _, child := range subWrapper.Children() {
			child.SetParent(subWrapper)
		}

		node.node.SetChildren(append(node.node.Children(), subWrapper))
	}
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

type ComponoError struct {
	Code    ErrorCode
	Message string
}

func (e *ComponoError) Error() string { return e.Message }

func NewComponoError(code ErrorCode, msg string) *ComponoError {
	return &ComponoError{Code: code, Message: msg}
}
