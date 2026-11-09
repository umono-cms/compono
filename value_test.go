package compono

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// dq is a double quote, so that sources with doubled quotes stay readable.
const dq = "\""

func TestParseValue(t *testing.T) {
	menu := "\n[\n" +
		"  { text: " + dq + "Home" + dq + ", url: " + dq + "/" + dq + ", current: " + dq + "__auto" + dq + ", children: [] },\n" +
		"  { text: " + dq + "Blog" + dq + ", url: " + dq + "/blog" + dq + ", current: " + dq + "__auto" + dq + ", children: [] }\n" +
		"]\n"

	tests := []struct {
		name string
		src  string
		want any
	}{
		{"string", dq + "Home" + dq, "Home"},
		{"empty string", dq + dq, ""},
		{"doubled quote", dq + "a" + dq + dq + "b" + dq, "a" + dq + "b"},
		{"surrounding whitespace", " \t\n" + dq + " a " + dq + "\r\n ", " a "},
		{"integer", "42", 42},
		{"negative integer", "-3", -3},
		{"true", "true", true},
		{"false", "false", false},
		{"array", "[1, " + dq + "x" + dq + ", false]", []any{1, "x", false}},
		{"empty array", "[]", []any{}},
		{"nested array", "[[1, 2], [3]]", []any{[]any{1, 2}, []any{3}}},
		{"record", "{ text: " + dq + "Home" + dq + ", url: " + dq + "/" + dq + " }", map[string]any{"text": "Home", "url": "/"}},
		{"empty record", "{}", map[string]any{}},
		{"nested record", "{ seo: { title: " + dq + "T" + dq + ", index: true } }", map[string]any{"seo": map[string]any{"title": "T", "index": true}}},
		{"menu", menu, []any{
			map[string]any{"text": "Home", "url": "/", "current": "__auto", "children": []any{}},
			map[string]any{"text": "Blog", "url": "/blog", "current": "__auto", "children": []any{}},
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, diags := ParseValue([]byte(tt.src))
			assert.Empty(t, diags)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParseValueInvalidValue(t *testing.T) {
	const message = "The value must be a string, integer, bool, array or record."

	tests := []struct {
		name string
		src  string
		text string // the text the diagnostic covers
	}{
		{"component", "FOO", "FOO"},
		{"no matter", "NO_MATTER", "NO_MATTER"},
		{"component with bound arguments", "FOO(a = 1)", "FOO(a = 1)"},
		{"parameter reference", "title", "title"},
		{"context", "context(site/name)", "context(site/name)"},
		{"component in array", "[1, FOO]", "FOO"},
		{"context in record", "{ a: context(site/x) }", "context(site/x)"},
		{"parameter in nested record", "{ a: { b: c } }", "c"},
		{"after whitespace", "\n  FOO", "FOO"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, diags := ParseValue([]byte(tt.src))
			assert.Nil(t, got)
			require.Len(t, diags, 1)
			assert.Equal(t, CodeInvalidValue, diags[0].Code)
			assert.Equal(t, message, diags[0].Message)
			assert.Equal(t, span(t, tt.src, tt.text, 0), diags[0].Range)
		})
	}
}

func TestParseValueEmpty(t *testing.T) {
	for _, src := range []string{"", "   ", "\n\t\n"} {
		got, diags := ParseValue([]byte(src))
		assert.Nil(t, got)
		require.Len(t, diags, 1)
		assert.Equal(t, CodeInvalidValue, diags[0].Code)
	}
}

func TestParseValueGrammarError(t *testing.T) {
	for _, src := range []string{"[1, ", "{ a: }", "1 2", dq + "a" + dq + " b = 1", "1 }} x {{ Y"} {
		t.Run(src, func(t *testing.T) {
			got, diags := ParseValue([]byte(src))
			assert.Nil(t, got)
			require.NotEmpty(t, diags)
			for _, d := range diags {
				assert.LessOrEqual(t, d.Range.End.Offset, len(src))
				assert.GreaterOrEqual(t, d.Range.Start.Offset, 0)
			}
		})
	}
}

// TestParseValueEquivalence pins that giving the result of ParseValue to
// WithContext renders the same output and diagnostics as writing the same
// source as an argument.
func TestParseValueEquivalence(t *testing.T) {
	tests := []struct {
		name  string
		value string
		defa  string
		body  string
	}{
		{"string", dq + "Merhaba <b>" + dq + dq + dq, dq + dq, "{{ v }}"},
		{"integer", "7", "0", "{{ v }}"},
		{"bool", "true", "false", "{{ v }}"},
		{"array", "[" + dq + "a" + dq + ", " + dq + "b" + dq + "]", "[]", "{{ v[1] }}"},
		{"array out of range", "[" + dq + "a" + dq + "]", "[]", "{{ v[3] }}"},
		{"record", "{ text: " + dq + "Home" + dq + ", url: " + dq + "/" + dq + " }", "{}", "{{ LINK text = v.text url = v.url }}"},
		{"record unknown key", "{ text: " + dq + "Home" + dq + " }", "{}", "{{ v.url }}"},
		{"array of records", "[{ text: " + dq + "A" + dq + " }, { text: " + dq + "B" + dq + " }]", "[]", "{{ v[1].text }}"},
		{"direct record", "{ text: " + dq + "Home" + dq + " }", "{}", "{{ v }}"},
	}

	c := New()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, diags := ParseValue([]byte(tt.value))
			require.Empty(t, diags)

			def := "\n\n~ A v = " + tt.defa + "\n" + tt.body + "\n"

			var literal bytes.Buffer
			literalDiags, err := c.Convert([]byte("{{ A v = "+tt.value+" }}"+def), &literal)
			require.NoError(t, err)

			var fromContext bytes.Buffer
			contextDiags, err := c.Convert([]byte("{{ A v = context(k) }}"+def), &fromContext, WithContext(map[string]any{"k": value}))
			require.NoError(t, err)

			assert.Equal(t, literal.String(), fromContext.String())
			require.Len(t, contextDiags, len(literalDiags))
			for i := range literalDiags {
				assert.Equal(t, literalDiags[i].Code, contextDiags[i].Code)
			}
		})
	}
}
