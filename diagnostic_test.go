package compono

import (
	goast "go/ast"
	"go/parser"
	"go/token"
	"io"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var diagnosticCodeTitles = []struct {
	constant DiagnosticCode
	title    string
}{
	{CodeArrayIndexOutOfRange, "Array index out of range"},
	{CodeDisallowedURLScheme, "Disallowed URL scheme"},
	{CodeDuplicateArgument, "Duplicate argument"},
	{CodeDuplicateVariant, "Duplicate variant"},
	{CodeEmptyTable, "Empty table"},
	{CodeEmptyTableHeader, "Empty table header"},
	{CodeEmptyText, "Empty text"},
	{CodeEmptyURL, "Empty URL"},
	{CodeInconsistentAspectRatio, "Inconsistent aspect ratio"},
	{CodeInfiniteComponentCall, "Infinite component call"},
	{CodeInvalidBuiltinArguments, "Invalid built-in arguments"},
	{CodeInvalidCodeBlockLanguage, "Invalid code block language"},
	{CodeInvalidComponentUsage, "Invalid component usage"},
	{CodeInvalidContextAccess, "Invalid context access"},
	{CodeInvalidContextUsage, "Invalid context usage"},
	{CodeInvalidCurrentValue, "Invalid current value"},
	{CodeInvalidDimension, "Invalid dimension"},
	{CodeInvalidExpression, "Invalid expression"},
	{CodeInvalidItemComponent, "Invalid item component"},
	{CodeInvalidListItem, "Invalid list item"},
	{CodeInvalidParameterAccess, "Invalid parameter access"},
	{CodeInvalidParameterDefinition, "Invalid parameter definition"},
	{CodeInvalidParameterUsage, "Invalid parameter usage"},
	{CodeInvalidTableDelimiter, "Invalid table delimiter"},
	{CodeInvalidTableRow, "Invalid table row"},
	{CodeInvalidValue, "Invalid value"},
	{CodeInvalidVariantURL, "Invalid variant URL"},
	{CodeMissingArgument, "Missing argument"},
	{CodeNotComponentParameter, "Not component parameter"},
	{CodeTableCellCountMismatch, "Table cell count mismatch"},
	{CodeUnclosedExpression, "Unclosed expression"},
	{CodeUnknownComponent, "Unknown component"},
	{CodeUnknownKey, "Unknown key"},
	{CodeUnknownParameter, "Unknown parameter"},
	{CodeUnknownRecordKey, "Unknown record key"},
	{CodeUnsupportedMimeType, "Unsupported mime-type"},
	{CodeWrongArgumentType, "Wrong argument type"},
}

func TestDiagnosticCodes(t *testing.T) {
	for _, tt := range diagnosticCodeTitles {
		t.Run(tt.title, func(t *testing.T) {
			expected := strings.ReplaceAll(strings.ToLower(tt.title), " ", "-")
			assert.Equal(t, expected, string(tt.constant))
		})
	}
}

func TestDiagnosticCodesAreUnique(t *testing.T) {
	seen := make(map[string]string, len(diagnosticCodeTitles))

	for _, tt := range diagnosticCodeTitles {
		previous, exists := seen[string(tt.constant)]
		assert.Falsef(t, exists, "duplicate code %q for %q and %q", tt.constant, previous, tt.title)
		seen[string(tt.constant)] = tt.title
	}
}

// TestDiagnosticCodeTitlesCoverAllConstants checks that diagnosticCodeTitles
// lists every DiagnosticCode constant of diagnostic.go.
func TestDiagnosticCodeTitlesCoverAllConstants(t *testing.T) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "diagnostic.go", nil, 0)
	require.NoError(t, err)

	var constants []string
	for _, decl := range file.Decls {
		genDecl, ok := decl.(*goast.GenDecl)
		if !ok || genDecl.Tok != token.CONST {
			continue
		}
		for _, spec := range genDecl.Specs {
			valueSpec, ok := spec.(*goast.ValueSpec)
			if !ok {
				continue
			}
			ident, ok := valueSpec.Type.(*goast.Ident)
			if !ok || ident.Name != "DiagnosticCode" {
				continue
			}
			for _, value := range valueSpec.Values {
				basicLit, ok := value.(*goast.BasicLit)
				if !ok {
					continue
				}
				unquoted, err := strconv.Unquote(basicLit.Value)
				require.NoError(t, err)
				constants = append(constants, unquoted)
			}
		}
	}

	titles := make([]string, len(diagnosticCodeTitles))
	for i, tt := range diagnosticCodeTitles {
		titles[i] = string(tt.constant)
	}

	assert.ElementsMatch(t, constants, titles)
}

func TestGeneralDiagnosticMessages(t *testing.T) {
	cases := []struct {
		name    string
		source  string
		context map[string]any
		code    DiagnosticCode
		message string
	}{
		{
			name:    "unknown component",
			source:  "{{ FOO }}",
			code:    CodeUnknownComponent,
			message: "The component **FOO** is not defined.",
		},
		{
			name:    "unknown components",
			source:  "{{ X a = A b = B }}\n\n~ X a = NO_MATTER b = NO_MATTER\n{{ a }}{{ b }}",
			code:    CodeUnknownComponent,
			message: "The components **A**, **B** are not defined.",
		},
		{
			name:    "not component parameter",
			source:  "{{ X }}\n\n~ X title = \"\"\n{{ title x = 1 }}",
			code:    CodeNotComponentParameter,
			message: "The parameter **title** is not a component parameter and cannot be called.",
		},
		{
			name:    "parameter in root",
			source:  "{{ title }}",
			code:    CodeInvalidParameterUsage,
			message: "The parameter **title** cannot be used outside a component.",
		},
		{
			name:    "context index out of range",
			source:  "{{ context(site/items)[5] }}",
			context: map[string]any{"site/items": []any{"a"}},
			code:    CodeArrayIndexOutOfRange,
			message: "The index used for context value **site/items** is out of range.",
		},
		{
			name:    "context array rendered directly",
			source:  "{{ context(site/items) }}",
			context: map[string]any{"site/items": []any{"a"}},
			code:    CodeInvalidContextUsage,
			message: "The context value **site/items** is an array and cannot be rendered directly.",
		},
		{
			name:    "context record rendered directly",
			source:  "{{ context(site/info) }}",
			context: map[string]any{"site/info": map[string]any{"k": "v"}},
			code:    CodeInvalidContextUsage,
			message: "The context value **site/info** is a record and cannot be rendered directly.",
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			var (
				diags []Diagnostic
				err   error
			)
			if tt.context != nil {
				diags, err = New().Convert([]byte(tt.source), io.Discard, WithContext(tt.context))
			} else {
				diags, err = New().Convert([]byte(tt.source), io.Discard)
			}
			require.NoError(t, err)
			require.Len(t, diags, 1)
			assert.Equal(t, tt.code, diags[0].Code)
			assert.Equal(t, tt.message, diags[0].Message)
		})
	}
}
