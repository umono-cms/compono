package compono

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnalyzeDirectives(t *testing.T) {
	source := `{{ LINK text = "a" url = "/" current = "__auto" }}
{{ LINK text = "b" url = "/b" current = " __auto " }}
{{ LINK text = "c" url = "/c" current = "page" }}
{{ LINK text = "__d" url = "/__d" current = "" }}
{{ LINK text = "e" url = "/e" current = context(site/current) }}
{{ LIST items = [] item = LINK(current = "__page") }}
{{ MENU }}

~ MENU c = LINK(current = "__auto")
{{ LINK text = "m" url = "/m" current = "__auto" }}
{{ LINK text = "p" url = "/p" current = c }}
{{ c }}
`
	analysis, err := Analyze([]byte(source))
	require.NoError(t, err)

	assert.Equal(t, []Directive{
		{Builtin: "LINK", Param: "current", Value: "__auto", In: ""},
		{Builtin: "LINK", Param: "current", Value: "__auto", In: ""},
		{Builtin: "LINK", Param: "current", Value: "__page", In: ""},
		{Builtin: "LINK", Param: "current", Value: "__auto", In: "MENU"},
		{Builtin: "LINK", Param: "current", Value: "__auto", In: "MENU"},
	}, analysis.Directives)
}

func TestAnalyzeDirectivesInGlobal(t *testing.T) {
	source := `~ c = LINK(current = "__auto")
{{ LINK text = "g" url = "/" current = "__auto" }}
{{ c }}
`
	analysis, err := Analyze([]byte(source))
	require.NoError(t, err)

	assert.Equal(t, []Directive{
		{Builtin: "LINK", Param: "current", Value: "__auto", In: ""},
		{Builtin: "LINK", Param: "current", Value: "__auto", In: ""},
	}, analysis.Directives)
}

func TestAnalyzeDirectivesNotForLocalWithBuiltinName(t *testing.T) {
	source := `{{ LINK current = "__auto" }}

~ LINK current = ""
{{ current }}
`
	analysis, err := Analyze([]byte(source))
	require.NoError(t, err)
	assert.Nil(t, analysis.Directives)
}

func TestAnalyzeDirectivesNilWithoutDirectives(t *testing.T) {
	analysis, err := Analyze([]byte(`{{ LINK text = "a" url = "/" current = "page" }} __auto`))
	require.NoError(t, err)
	assert.Nil(t, analysis.Directives)
}
