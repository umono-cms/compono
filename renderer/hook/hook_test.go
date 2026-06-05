package hook

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRendererHookContextFields(t *testing.T) {
	ctx := RendererHookContext{
		Kind:   KindMarkdown,
		Name:   "h1",
		Params: Params{"text": NewString("Hello")},
		Output: "<h1>Hello</h1>",
	}

	assert.Equal(t, KindMarkdown, ctx.Kind)
	assert.Equal(t, "h1", ctx.Name)
	value, ok := ctx.Params.String("text")
	assert.True(t, ok)
	assert.Equal(t, "Hello", value)
	assert.Equal(t, "<h1>Hello</h1>", ctx.Output)
}

func TestParamValueHelpers(t *testing.T) {
	params := Params{
		"title": NewString("Hello"),
		"items": NewArray([]ParamValue{
			NewRecord(map[string]ParamValue{
				"label": NewString("Home"),
				"url":   NewString("/"),
			}),
		}),
	}

	title, ok := params.String("title")
	assert.True(t, ok)
	assert.Equal(t, "Hello", title)

	items, ok := params.Array("items")
	assert.True(t, ok)
	item, ok := items.Record(0)
	assert.True(t, ok)
	label, ok := item.String("label")
	assert.True(t, ok)
	assert.Equal(t, "Home", label)

	_, ok = params.String("missing")
	assert.False(t, ok)
	_, ok = params.Get("items").String()
	assert.False(t, ok)
}

func TestRendererHookFunc(t *testing.T) {
	hookFn := RendererHookFunc(func(ctx RendererHookContext) string {
		return "<modified>" + ctx.Output + "</modified>"
	})

	ctx := RendererHookContext{
		Kind:   KindMarkdown,
		Name:   "p",
		Params: Params{},
		Output: "<p>Hello</p>",
	}

	result := hookFn(ctx)
	assert.Equal(t, "<modified><p>Hello</p></modified>", result)
}

func TestKindConstants(t *testing.T) {
	assert.Equal(t, Kind("markdown"), KindMarkdown)
	assert.Equal(t, Kind("builtin"), KindBuiltin)
	assert.NotEqual(t, KindMarkdown, KindBuiltin)
}
