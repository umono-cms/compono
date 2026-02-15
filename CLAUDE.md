# CLAUDE.md - Compono Project Guidelines

## Project Overview

Compono is a **platform-agnostic**, component-based domain-specific language (DSL) that extends Markdown syntax with reusable components. Originally developed for [Umono CMS](https://github.com/umono-cms/umono), it can be used in any Go project that needs a flexible templating solution.

**Repository:** `github.com/umono-cms/compono`  
**Language:** Go 1.23+  
**License:** MIT

### Platform-Agnostic Design

Compono's syntax is designed to be **renderer-independent**. Currently, the only renderer is the HTML renderer (`renderer/html/`), but the architecture supports future renderers (e.g., PDF, plain text, etc.).

**Important:** When extending Compono syntax, avoid adding platform-specific features. The syntax should remain abstract and let renderers handle platform-specific output.

## Quick Reference

### Common Commands

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run specific package tests
go test ./parser/...
go test ./renderer/...
go test ./validator/...

# Run tests with coverage
go test -cover ./...

# Build/verify compilation
go build ./...

# Format code
gofmt -w .

# Lint (if golangci-lint is installed)
golangci-lint run
```

### Golden Tests

The project uses golden file testing in `testdata/`:
- Input files: `testdata/input/*.comp`
- Expected output: `testdata/output/*.golden`

Golden files are manually created. When adding a new feature, run tests to verify.

## Architecture

```
compono/
├── compono.go          # Main API (Compono interface, New(), Convert())
├── ast/                # Abstract Syntax Tree node definitions
├── parser/             # Source → AST conversion
├── renderer/           # Renderer interface + implementations
│   └── html/           # Default HTML renderer (currently the only one)
├── rule/               # Grammar rules (components, markdown, etc.)
├── selector/           # Text pattern matching for parser
├── validator/          # AST validation
├── errwrap/            # Error wrapping for user-friendly messages
├── logger/             # Debug logging
├── util/               # Utilities (string helpers, etc.)
└── testdata/           # Golden test files
```

### Core Flow

```
Source (.comp) → Parser → AST → Validator → ErrorWrapper → Renderer → Output
                                                              ↓
                                                    (HTML, PDF, etc.)
```

### Key Interfaces

```go
// Main entry point
type Compono interface {
    Convert(source []byte, writer io.Writer) error
    RegisterGlobalComponent(name string, source []byte) error
    UnregisterGlobalComponent(name string) error
    ConvertGlobalComponent(name string, source []byte, writer io.Writer) error
}

// Internal components
type Parser interface {
    Parse(source []byte, node ast.Node) ast.Node
}

type Renderer interface {
    Render(writer io.Writer, root ast.Node) error
}

type Validator interface {
    Validate(ast.Node) error
}
```

## Syntax Reference

### Component Naming
- Component names: `SCREAMING_SNAKE_CASE` (e.g., `USER_CARD`, `NAV_MENU`)
- Parameter names: `kebab-case` (e.g., `user-name`, `is-active`)

### Local Component Definition

Local components are defined within the same source file where they're used:

```
{{ GREETING name="World" }}

~ GREETING name="Guest"
# Hello, {{ name }}!
```

The `~` prefix marks a local component definition. Parameters with default values are declared after the component name.

### Global Component Definition

Global components are registered programmatically and can be used across multiple conversions:

```go
c := compono.New()

// Simple global component
c.RegisterGlobalComponent("FOOTER", []byte(`© 2025 My Company`))

// Global component with parameters
c.RegisterGlobalComponent("CARD", []byte(`title="" content=""
## {{ title }}
{{ content }}
`))

// Use in any conversion
c.Convert([]byte(`{{ FOOTER }}`), &buf)
```

### Component Call
```
{{ COMPONENT_NAME param1="value" param2=42 }}
```

### Override Hierarchy

When a component is called, Compono resolves it in this order:

```
1. Local Component    (defined in same file with ~)
2. Global Component   (registered via RegisterGlobalComponent)
3. Built-in Component (provided by the renderer, e.g., LINK)
```

The first match wins. This means:
- A local `~ LINK` definition will override the built-in `LINK` component
- A global component can be overridden per-file by defining a local one
- Built-ins serve as fallbacks when no custom definition exists

### Parameter Passing with `$`

The `$` prefix enables two powerful features:

#### 1. Passing a Component as a Parameter

A parameter can hold a component name, and `{{ $param }}` will render that component:

```
{{ DEFAULT_TEMPLATE comp=HOME }}
{{ DEFAULT_TEMPLATE comp=ABOUT }}

~ DEFAULT_TEMPLATE comp=EMPTY
{{ HEADER }}
{{ $comp }}
{{ FOOTER }}

~ HOME
I am home page

~ ABOUT
I am about page
```

Here `comp=HOME` passes the component name, and `{{ $comp }}` renders the `HOME` component.

#### 2. Forwarding Parameters to Child Components

Use `$param` to pass a parent's parameter value to a child component:

```
{{ PROFILE_WRAPPER }}
{{ PROFILE_WRAPPER profile-component=PROFILE_TYPE_3 name="John Doe" }}

~ PROFILE_WRAPPER profile-component=PROFILE_TYPE_1 name="Jane Doe" email="example@example.com"
{{ $profile-component name=$name email=$email }}

~ PROFILE_TYPE_1 name="" email=""
**{{ name }}** ({{ email }})

~ PROFILE_TYPE_2 name="" email=""
{{ email }} - {{ name }}

~ PROFILE_TYPE_3 name="" email=""
# {{ name }}
## {{ email }}
```

In this example:
- `$profile-component` renders whichever component was passed (default: `PROFILE_TYPE_1`)
- `name=$name` and `email=$email` forward the parent's parameters to the child
- This enables flexible template composition and parameter delegation

## Code Patterns

### Adding a New Rule

1. Create rule in `rule/` package implementing the `Rule` interface:
```go
type myRule struct{}

func newMyRule() Rule { return &myRule{} }

func (_ *myRule) Name() string { return "my-rule" }

func (_ *myRule) Selectors() []selector.Selector {
    // Return selectors that match this rule
}

func (_ *myRule) Rules() []Rule {
    // Return child rules
}
```

2. Register in parent rule's `Rules()` method

### Adding a Renderable Node

1. Create in `renderer/html/` implementing `renderableNode`:
```go
type myRenderer struct {
    baseRenderable
    renderer *renderer
}

func newMyRenderer(rend *renderer) renderableNode {
    return &myRenderer{renderer: rend}
}

func (m *myRenderer) New() renderableNode { return newMyRenderer(m.renderer) }

func (_ *myRenderer) Condition(invoker renderableNode, node ast.Node) bool {
    return ast.IsRuleName(node, "my-rule")
}

func (m *myRenderer) Render() string {
    // Return HTML output
}
```

2. Register in `renderer.go` `NewRenderer()` function

### Adding a Built-in Component

1. Create in `renderer/html/` implementing `builtinComponent`:
```go
type myBuiltin struct {
    renderer *renderer
}

func newMyBuiltin(r *renderer) builtinComponent {
    return &myBuiltin{renderer: r}
}

func (_ *myBuiltin) Name() string { return "MY_BUILTIN" }

func (m *myBuiltin) Render(invoker renderableNode, node ast.Node) string {
    // Extract params and return HTML
}
```

2. Register in `renderer.go` `builtinCompMap`

## Testing Guidelines

### Test File Structure
- Unit tests: `*_test.go` in same package
- Golden tests: `testdata/input/*.comp` → `testdata/output/*.golden`

### Writing Tests
```go
func TestMyFeature(t *testing.T) {
    c := compono.New()
    
    source := []byte(`{{ MY_COMPONENT }}`)
    var buf bytes.Buffer
    
    err := c.Convert(source, &buf)
    require.NoError(t, err)
    assert.Equal(t, "<expected>output</expected>", buf.String())
}
```

### Adding Golden Tests
1. Create `testdata/input/XXXX_description.comp`
2. Run tests to generate golden file
3. Verify output in `testdata/output/XXXX_description.golden`

## AST Utilities

Common AST helper functions in `ast/` package:
```go
ast.FindNodeByRuleName(children, "rule-name")  // Find first child by rule
ast.FindNode(nodes, predicateFn)               // Find node matching predicate
ast.FilterNodes(nodes, predicateFn)            // Filter nodes
ast.IsRuleName(node, "rule-name")              // Check rule name
ast.IsRuleNameOneOf(node, []string{...})       // Check multiple rules
ast.GetAncestors(node)                         // Get parent chain
```

## Error Handling

Error codes defined in `compono.go`:
- `ErrInvalidGlobalName` - Invalid component name format
- `ErrGlobalAlreadyRegistered` - Duplicate global component
- `ErrGlobalNotExist` - Unregistering non-existent component
- `ErrInvalidAST` - Validation failure
- `ErrRender` - Rendering error

User-facing errors are wrapped via `errwrap/` package and rendered as custom HTML elements:
- `<compono-error-block>` - Block-level errors
- `<compono-error-inline>` - Inline errors

## Important Notes

- Components must be defined AFTER their usage in source (bottom of file)
- Empty source input returns `nil` error with no output
- The parser uses regex-based selectors (not a traditional lexer/parser)
- Infinite loop detection is built into component resolution

## Dependencies

- `github.com/stretchr/testify` - Testing assertions only
- No runtime dependencies

## Debugging

Enable debug logging:
```go
c := compono.New()
c.Logger().SetLevel(logger.Parser | logger.Detail)
```

Log categories: `Parser`, `Detail`, `Renderer`
