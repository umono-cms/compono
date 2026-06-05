package compono

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/umono-cms/compono/logger"
	"github.com/umono-cms/compono/renderer/hook"
)

type componoTestSuite struct {
	suite.Suite
}

type invalidContextKeyNotation struct {
	Title string `compono:"invalid_key"`
}

func (s *componoTestSuite) TestGolden() {
	inputFiles, err := filepath.Glob("testdata/input/*.comp")
	require.Nil(s.T(), err)
	require.NotEmpty(s.T(), inputFiles, "no .comp files found")

	for _, inputPath := range inputFiles {
		name := filepath.Base(inputPath)
		input, err := os.ReadFile(inputPath)
		require.Nil(s.T(), err)

		globalFiles, err := filepath.Glob("testdata/input/global/" + strings.TrimSuffix(name, ".comp") + "/*.comp")
		require.Nil(s.T(), err)

		comp := New()
		comp.Logger().SetLogLevel(logger.All)

		for _, gPath := range globalFiles {
			globalCompName := filepath.Base(gPath)
			globalInput, err := os.ReadFile(gPath)
			require.Nil(s.T(), err)

			err = comp.RegisterGlobalComponent(strings.TrimSuffix(globalCompName, ".comp"), []byte(strings.TrimSpace(string(globalInput))))
			assert.Nil(s.T(), err)
		}

		opts := []ConvertOption{}
		contextPath := filepath.Join("testdata/input/context", strings.TrimSuffix(name, ".comp")+".json")
		if _, err := os.Stat(contextPath); err == nil {
			contextValues, err := readContextFixture(contextPath)
			require.Nil(s.T(), err)
			opts = append(opts, WithContext(contextValues))
		}

		var buf bytes.Buffer
		err = comp.Convert([]byte(strings.TrimSpace(string(input))), &buf, opts...)
		assert.Nil(s.T(), err)

		goldenPath := filepath.Join(
			"testdata/output",
			strings.TrimSuffix(name, ".comp")+".golden",
		)

		golden, err := os.ReadFile(goldenPath)
		require.Nil(s.T(), err, "golden file missing")

		assert.Equal(s.T(), strings.TrimSpace(string(golden)), buf.String(), "from %s", inputPath)
	}
}

func (s *componoTestSuite) TestGoldenForWithGlobalComponent() {
	inputFiles, err := filepath.Glob("testdata/global_input/*.comp")
	require.Nil(s.T(), err)
	require.NotEmpty(s.T(), inputFiles, "no .comp files found")

	for _, inputPath := range inputFiles {
		name := filepath.Base(inputPath)
		input, err := os.ReadFile(inputPath)
		require.Nil(s.T(), err)

		globalFiles, err := filepath.Glob("testdata/global_input/global/" + strings.TrimSuffix(name, ".comp") + "/*.comp")
		require.Nil(s.T(), err)

		comp := New()
		comp.Logger().SetLogLevel(logger.All)

		for _, gPath := range globalFiles {
			globalCompName := filepath.Base(gPath)
			globalInput, err := os.ReadFile(gPath)
			require.Nil(s.T(), err)

			err = comp.RegisterGlobalComponent(strings.TrimSuffix(globalCompName, ".comp"), []byte(strings.TrimSpace(string(globalInput))))
			assert.Nil(s.T(), err)
		}

		var buf bytes.Buffer
		err = comp.Convert(
			[]byte(`{{ `+strings.TrimSuffix(name, ".comp")+` }}`),
			&buf,
			WithGlobalComponent(strings.TrimSuffix(name, ".comp"), []byte(strings.TrimSpace(string(input)))),
		)
		assert.Nil(s.T(), err)

		goldenPath := filepath.Join(
			"testdata/global_output",
			strings.TrimSuffix(name, ".comp")+".golden",
		)

		golden, err := os.ReadFile(goldenPath)
		require.Nil(s.T(), err, "golden file missing")

		assert.Equal(s.T(), strings.TrimSpace(string(golden)), buf.String(), "from %s", inputPath)
	}
}

func (s *componoTestSuite) TestUnregisterGlobalComponent() {
	compono := New().(*compono)
	err := compono.RegisterGlobalComponent("SAY_HELLO", []byte("# Hello"))
	require.Nil(s.T(), err)
	err = compono.UnregisterGlobalComponent("SAY_HELLO")
	require.Nil(s.T(), err)
	assert.Equal(s.T(), 0, len(compono.globalWrapper.Children()))
}

func (s *componoTestSuite) TestConvertWithContextErrUnsupportedType() {
	compono := New()

	var buf bytes.Buffer
	err := compono.Convert([]byte("context"), &buf, WithContext(map[string]any{
		"value": func() {},
	}))

	require.Error(s.T(), err)

	var compErr *ComponoError
	require.ErrorAs(s.T(), err, &compErr)
	assert.Equal(s.T(), ErrUnsupportedType, compErr.Code)
	assert.Contains(s.T(), compErr.Message, "unsupported context value type")
}

func (s *componoTestSuite) TestConvertWithContextErrUnsupportedKeyNotation() {
	compono := New()

	var buf bytes.Buffer
	err := compono.Convert([]byte("context"), &buf, WithContext(map[string]any{
		"value": invalidContextKeyNotation{Title: "Hello"},
	}))

	require.Error(s.T(), err)

	var compErr *ComponoError
	require.ErrorAs(s.T(), err, &compErr)
	assert.Equal(s.T(), ErrUnsupportedKeyNotation, compErr.Code)
	assert.Contains(s.T(), compErr.Message, `invalid compono struct tag "invalid_key"`)
}

func (s *componoTestSuite) TestNavigationRequiresItems() {
	compono := New()

	var buf bytes.Buffer
	err := compono.Convert([]byte(`{{ NAVIGATION }}`), &buf)
	require.NoError(s.T(), err)

	assert.Equal(
		s.T(),
		`<compono-error-block><div slot="title">Invalid built-in arguments</div><div slot="description">The parameter <strong>items</strong> does not match the schema of the built-in component <strong>NAVIGATION</strong>.</div></compono-error-block>`,
		buf.String(),
	)
}

func (s *componoTestSuite) TestNavigationRejectsEmptyItems() {
	compono := New()

	var buf bytes.Buffer
	err := compono.Convert([]byte(`{{ NAVIGATION items = [] }}`), &buf)
	require.NoError(s.T(), err)

	assert.Equal(
		s.T(),
		`<compono-error-block><div slot="title">Invalid built-in arguments</div><div slot="description">The parameter <strong>items</strong> does not match the schema of the built-in component <strong>NAVIGATION</strong>.</div></compono-error-block>`,
		buf.String(),
	)
}

func (s *componoTestSuite) TestRendererHookForMarkdownElement() {
	c := New()

	var calls []hook.RendererHookContext
	hookFn := func(ctx hook.RendererHookContext) string {
		calls = append(calls, ctx)
		return ctx.Output
	}

	var buf bytes.Buffer
	err := c.Convert([]byte("# Hello"), &buf, WithRendererHook(hookFn))
	require.NoError(s.T(), err)

	found := false
	for _, call := range calls {
		if call.Kind == hook.KindMarkdown && call.Name == "h1" {
			found = true
			assert.Equal(s.T(), "<h1>Hello</h1>", call.Output)
			assert.Equal(s.T(), "Hello", requireHookString(s.T(), call.Params, "content"))
		}
	}
	assert.True(s.T(), found, "expected hook to be called for h1 markdown element")
}

func (s *componoTestSuite) TestRendererHookForParagraph() {
	c := New()

	var calls []hook.RendererHookContext
	hookFn := func(ctx hook.RendererHookContext) string {
		calls = append(calls, ctx)
		return ctx.Output
	}

	var buf bytes.Buffer
	err := c.Convert([]byte("Some text"), &buf, WithRendererHook(hookFn))
	require.NoError(s.T(), err)

	found := false
	for _, call := range calls {
		if call.Kind == hook.KindMarkdown && call.Name == "p" {
			found = true
			assert.Equal(s.T(), "<p>Some text</p>", call.Output)
		}
	}
	assert.True(s.T(), found, "expected hook to be called for p markdown element")
}

func (s *componoTestSuite) TestRendererHookForCodeBlock() {
	c := New()

	var calls []hook.RendererHookContext
	hookFn := func(ctx hook.RendererHookContext) string {
		calls = append(calls, ctx)
		return ctx.Output
	}

	source := "```go\nfmt.Println(\"hi\")\n```"
	var buf bytes.Buffer
	err := c.Convert([]byte(source), &buf, WithRendererHook(hookFn))
	require.NoError(s.T(), err)

	found := false
	for _, call := range calls {
		if call.Kind == hook.KindMarkdown && call.Name == "code-block" {
			found = true
			assert.Equal(s.T(), "go", requireHookString(s.T(), call.Params, "lang"))
			assert.Contains(s.T(), call.Output, "<pre><code")
		}
	}
	assert.True(s.T(), found, "expected hook to be called for code-block markdown element")
}

func (s *componoTestSuite) TestRendererHookForInlineCode() {
	c := New()

	var calls []hook.RendererHookContext
	hookFn := func(ctx hook.RendererHookContext) string {
		calls = append(calls, ctx)
		return ctx.Output
	}

	var buf bytes.Buffer
	err := c.Convert([]byte("Use `fmt.Println`"), &buf, WithRendererHook(hookFn))
	require.NoError(s.T(), err)

	found := false
	for _, call := range calls {
		if call.Kind == hook.KindMarkdown && call.Name == "inline-code" {
			found = true
			assert.Contains(s.T(), call.Output, "<code")
		}
	}
	assert.True(s.T(), found, "expected hook to be called for inline-code markdown element")
}

func (s *componoTestSuite) TestRendererHookForMarkdownLink() {
	c := New()

	var calls []hook.RendererHookContext
	hookFn := func(ctx hook.RendererHookContext) string {
		calls = append(calls, ctx)
		return ctx.Output
	}

	var buf bytes.Buffer
	err := c.Convert([]byte("[click](https://example.com)"), &buf, WithRendererHook(hookFn))
	require.NoError(s.T(), err)

	found := false
	for _, call := range calls {
		if call.Kind == hook.KindMarkdown && call.Name == "link" {
			found = true
			assert.Equal(s.T(), "click", requireHookString(s.T(), call.Params, "text"))
			assert.Equal(s.T(), "https://example.com", requireHookString(s.T(), call.Params, "url"))
			assert.Contains(s.T(), call.Output, "<compono-link>")
		}
	}
	assert.True(s.T(), found, "expected hook to be called for link markdown element")
}

func (s *componoTestSuite) TestRendererHookForBuiltinLink() {
	c := New()

	var calls []hook.RendererHookContext
	hookFn := func(ctx hook.RendererHookContext) string {
		calls = append(calls, ctx)
		return ctx.Output
	}

	var buf bytes.Buffer
	err := c.Convert([]byte(`{{ LINK text="Visit" url="https://example.com" new-tab=true }}`), &buf, WithRendererHook(hookFn))
	require.NoError(s.T(), err)

	found := false
	for _, call := range calls {
		if call.Kind == hook.KindBuiltin && call.Name == "LINK" {
			found = true
			assert.Equal(s.T(), "Visit", requireHookString(s.T(), call.Params, "text"))
			assert.Equal(s.T(), "https://example.com", requireHookString(s.T(), call.Params, "url"))
			assert.Equal(s.T(), "true", requireHookString(s.T(), call.Params, "new-tab"))
			assert.Contains(s.T(), call.Output, "<compono-link>")
		}
	}
	assert.True(s.T(), found, "expected hook to be called for LINK builtin component")
}

func (s *componoTestSuite) TestRendererHookModifiesOutput() {
	c := New()

	hookFn := func(ctx hook.RendererHookContext) string {
		if ctx.Kind == hook.KindMarkdown && ctx.Name == "h1" {
			return strings.Replace(ctx.Output, "<h1>", "<h1 class=\"title\">", 1)
		}
		return ctx.Output
	}

	var buf bytes.Buffer
	err := c.Convert([]byte("# Hello"), &buf, WithRendererHook(hookFn))
	require.NoError(s.T(), err)
	assert.Equal(s.T(), `<h1 class="title">Hello</h1>`, buf.String())
}

func (s *componoTestSuite) TestRendererHookParamsAreRaw() {
	c := New()

	var calls []hook.RendererHookContext
	hookFn := func(ctx hook.RendererHookContext) string {
		calls = append(calls, ctx)
		return ctx.Output
	}

	var buf bytes.Buffer
	err := c.Convert([]byte(`{{ LINK text="<script>alert('xss')</script>" url="https://example.com?a=1&b=2" }}`), &buf, WithRendererHook(hookFn))
	require.NoError(s.T(), err)

	found := false
	for _, call := range calls {
		if call.Kind == hook.KindBuiltin && call.Name == "LINK" {
			found = true
			text := requireHookString(s.T(), call.Params, "text")
			url := requireHookString(s.T(), call.Params, "url")
			assert.Contains(s.T(), text, "<script>")
			assert.Contains(s.T(), url, "&")
			assert.NotContains(s.T(), text, "&lt;")
			assert.NotContains(s.T(), url, "&amp;")
		}
	}
	assert.True(s.T(), found, "expected hook to be called with raw params")
}

func (s *componoTestSuite) TestRendererHookParamsForMarkdownElementAreRaw() {
	c := New()

	var calls []hook.RendererHookContext
	hookFn := func(ctx hook.RendererHookContext) string {
		calls = append(calls, ctx)
		return ctx.Output
	}

	var buf bytes.Buffer
	err := c.Convert([]byte("# <script>alert('xss')</script>"), &buf, WithRendererHook(hookFn))
	require.NoError(s.T(), err)

	found := false
	for _, call := range calls {
		if call.Kind == hook.KindMarkdown && call.Name == "h1" {
			found = true
			content := requireHookString(s.T(), call.Params, "content")
			assert.Contains(s.T(), content, "<script>")
			assert.NotContains(s.T(), content, "&lt;")
		}
	}
	assert.True(s.T(), found, "expected hook to be called for h1 markdown element")
}

func (s *componoTestSuite) TestRendererHookParamsForBuiltinLinkAreRaw() {
	c := New()

	var calls []hook.RendererHookContext
	hookFn := func(ctx hook.RendererHookContext) string {
		calls = append(calls, ctx)
		return ctx.Output
	}

	var buf bytes.Buffer
	err := c.Convert([]byte(`{{ LINK text="<b>bold</b>" url="https://example.com?a=1&b=2" }}`), &buf, WithRendererHook(hookFn))
	require.NoError(s.T(), err)

	found := false
	for _, call := range calls {
		if call.Kind == hook.KindBuiltin && call.Name == "LINK" {
			found = true
			assert.Equal(s.T(), "<b>bold</b>", requireHookString(s.T(), call.Params, "text"))
			assert.Equal(s.T(), "https://example.com?a=1&b=2", requireHookString(s.T(), call.Params, "url"))
		}
	}
	assert.True(s.T(), found, "expected hook to be called for LINK builtin component")
}

func (s *componoTestSuite) TestRendererHookParamsForAllMarkdownElementsAreRaw() {
	c := New()

	var calls []hook.RendererHookContext
	hookFn := func(ctx hook.RendererHookContext) string {
		calls = append(calls, ctx)
		return ctx.Output
	}

	source := "# H1 & <tag>\n## H2 & <tag>\n### H3 & <tag>\n#### H4 & <tag>\n##### H5 & <tag>\n###### H6 & <tag>\n\npara & <tag>\n\n**bold & <tag>**\n\n*em & <tag>*\n\n```html\n<div>&</div>\n```\n\n`inline & <tag>`\n\n[link & <tag>](https://example.com?a=1&b=2)"

	var buf bytes.Buffer
	err := c.Convert([]byte(source), &buf, WithRendererHook(hookFn))
	require.NoError(s.T(), err)

	expectedCalls := map[string]bool{
		"h1": false, "h2": false, "h3": false, "h4": false, "h5": false, "h6": false,
		"p": false, "strong": false, "em": false,
		"code-block": false, "inline-code": false, "link": false,
	}

	for _, call := range calls {
		if call.Kind != hook.KindMarkdown {
			continue
		}
		if _, ok := expectedCalls[call.Name]; !ok {
			continue
		}
		expectedCalls[call.Name] = true

		switch call.Name {
		case "h1", "h2", "h3", "h4", "h5", "h6":
			content := requireHookString(s.T(), call.Params, "content")
			assert.Contains(s.T(), content, "&")
			assert.Contains(s.T(), content, "<tag>")
			assert.NotContains(s.T(), content, "&amp;")
			assert.NotContains(s.T(), content, "&lt;")
		case "p":
			content := requireHookString(s.T(), call.Params, "content")
			assert.Contains(s.T(), content, "&")
			assert.Contains(s.T(), content, "<tag>")
			assert.NotContains(s.T(), content, "&amp;")
		case "strong":
			content := requireHookString(s.T(), call.Params, "content")
			assert.Contains(s.T(), content, "&")
			assert.NotContains(s.T(), content, "&amp;")
		case "em":
			content := requireHookString(s.T(), call.Params, "content")
			assert.Contains(s.T(), content, "&")
			assert.NotContains(s.T(), content, "&amp;")
		case "code-block":
			content := requireHookString(s.T(), call.Params, "content")
			assert.Equal(s.T(), "html", requireHookString(s.T(), call.Params, "lang"))
			assert.Contains(s.T(), content, "<div>")
			assert.Contains(s.T(), content, "&")
			assert.NotContains(s.T(), content, "&lt;")
		case "inline-code":
			content := requireHookString(s.T(), call.Params, "content")
			assert.Contains(s.T(), content, "&")
			assert.Contains(s.T(), content, "<tag>")
			assert.NotContains(s.T(), content, "&amp;")
		case "link":
			assert.Equal(s.T(), "link & <tag>", requireHookString(s.T(), call.Params, "text"))
			assert.Equal(s.T(), "https://example.com?a=1&b=2", requireHookString(s.T(), call.Params, "url"))
		}
	}

	for name, found := range expectedCalls {
		assert.True(s.T(), found, "expected hook call for markdown element: %s", name)
	}
}

func (s *componoTestSuite) TestRendererHookParamsForAllBuiltinComponents() {
	c := New()

	var calls []hook.RendererHookContext
	hookFn := func(ctx hook.RendererHookContext) string {
		calls = append(calls, ctx)
		return ctx.Output
	}

	source := `{{ LINK text="<b>bold</b>" url="https://example.com?a=1&b=2" new-tab=true }}`

	var buf bytes.Buffer
	err := c.Convert([]byte(source), &buf, WithRendererHook(hookFn))
	require.NoError(s.T(), err)

	foundLink := false
	for _, call := range calls {
		if call.Kind == hook.KindBuiltin && call.Name == "LINK" {
			foundLink = true
			assert.Equal(s.T(), "<b>bold</b>", requireHookString(s.T(), call.Params, "text"))
			assert.Equal(s.T(), "https://example.com?a=1&b=2", requireHookString(s.T(), call.Params, "url"))
			assert.Equal(s.T(), "true", requireHookString(s.T(), call.Params, "new-tab"))
		}
	}
	assert.True(s.T(), foundLink, "expected hook call for LINK builtin")

	imageSource := `{{ IMAGE media = {
  url: "https://cdn.example.com/photo.jpg",
  width: 800,
  height: 600,
  mime-type: "image/jpeg"
} alt = "A <photo> & memory" }}`

	buf.Reset()
	err = c.Convert([]byte(imageSource), &buf, WithRendererHook(hookFn))
	require.NoError(s.T(), err)

	foundImage := false
	for _, call := range calls {
		if call.Kind == hook.KindBuiltin && call.Name == "IMAGE" {
			foundImage = true
			media := requireHookRecord(s.T(), call.Params, "media")
			assert.Equal(s.T(), "A <photo> & memory", requireHookString(s.T(), call.Params, "alt"))
			assert.Equal(s.T(), "https://cdn.example.com/photo.jpg", requireHookRecordString(s.T(), media, "url"))
			assert.Equal(s.T(), "800", requireHookRecordString(s.T(), media, "width"))
			assert.Equal(s.T(), "600", requireHookRecordString(s.T(), media, "height"))
			assert.Equal(s.T(), "image/jpeg", requireHookRecordString(s.T(), media, "mime-type"))
		}
	}
	assert.True(s.T(), foundImage, "expected hook call for IMAGE builtin")

	webGridSource := `{{ WEB_GRID
  items = [
    { component: MY_COMP, grid-area: "header" }
  ]
  grid-template-columns = ["1fr"]
  grid-template-rows = ["min-content"]
  grid-template-areas = [
    ["header"]
  ]
}}

~ MY_COMP
# Header`

	buf.Reset()
	err = c.Convert([]byte(webGridSource), &buf, WithRendererHook(hookFn))
	require.NoError(s.T(), err)

	foundWebGrid := false
	for _, call := range calls {
		if call.Kind == hook.KindBuiltin && call.Name == "WEB_GRID" {
			foundWebGrid = true
			items := requireHookArray(s.T(), call.Params, "items")
			item, ok := items.Record(0)
			require.True(s.T(), ok)
			assert.Equal(s.T(), "MY_COMP", requireHookRecordString(s.T(), item, "component"))
			assert.Equal(s.T(), "header", requireHookRecordString(s.T(), item, "grid-area"))

			columns := requireHookArray(s.T(), call.Params, "grid-template-columns")
			rows := requireHookArray(s.T(), call.Params, "grid-template-rows")
			areas := requireHookArray(s.T(), call.Params, "grid-template-areas")
			areaRow, ok := areas.Array(0)
			require.True(s.T(), ok)
			assert.Equal(s.T(), "1fr", requireHookArrayString(s.T(), columns, 0))
			assert.Equal(s.T(), "min-content", requireHookArrayString(s.T(), rows, 0))
			assert.Equal(s.T(), "header", requireHookArrayString(s.T(), areaRow, 0))
		}
	}
	assert.True(s.T(), foundWebGrid, "expected hook call for WEB_GRID builtin")

	navSource := `{{ NAVIGATION items = [
  { label: "Home", target: "/" },
  { label: "About", target: "/about" }
] }}`

	buf.Reset()
	err = c.Convert([]byte(navSource), &buf, WithRendererHook(hookFn))
	require.NoError(s.T(), err)

	foundNav := false
	for _, call := range calls {
		if call.Kind == hook.KindBuiltin && call.Name == "NAVIGATION" {
			foundNav = true
			items := requireHookArray(s.T(), call.Params, "items")
			first, ok := items.Record(0)
			require.True(s.T(), ok)
			second, ok := items.Record(1)
			require.True(s.T(), ok)
			assert.Equal(s.T(), "Home", requireHookRecordString(s.T(), first, "label"))
			assert.Equal(s.T(), "/", requireHookRecordString(s.T(), first, "target"))
			assert.Equal(s.T(), "About", requireHookRecordString(s.T(), second, "label"))
			assert.Equal(s.T(), "/about", requireHookRecordString(s.T(), second, "target"))
		}
	}
	assert.True(s.T(), foundNav, "expected hook call for NAVIGATION builtin")
}

func (s *componoTestSuite) TestMultipleRendererHooks() {
	c := New()

	hook1 := func(ctx hook.RendererHookContext) string {
		if ctx.Kind == hook.KindMarkdown && ctx.Name == "h1" {
			return ctx.Output + "<!-- hook1 -->"
		}
		return ctx.Output
	}

	hook2 := func(ctx hook.RendererHookContext) string {
		if ctx.Kind == hook.KindMarkdown && ctx.Name == "h1" {
			return ctx.Output + "<!-- hook2 -->"
		}
		return ctx.Output
	}

	var buf bytes.Buffer
	err := c.Convert([]byte("# Hello"), &buf, WithRendererHook(hook1), WithRendererHook(hook2))
	require.NoError(s.T(), err)
	assert.Equal(s.T(), `<h1>Hello</h1><!-- hook1 --><!-- hook2 -->`, buf.String())
}

func (s *componoTestSuite) TestNilRendererHookIsIgnored() {
	c := New()

	var buf bytes.Buffer
	err := c.Convert([]byte("# Hello"), &buf, WithRendererHook(nil))
	require.NoError(s.T(), err)
	assert.Equal(s.T(), `<h1>Hello</h1>`, buf.String())
}

func (s *componoTestSuite) TestRendererHookNotCalledWithoutHook() {
	c := New()

	var buf bytes.Buffer
	err := c.Convert([]byte("# Hello"), &buf)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), `<h1>Hello</h1>`, buf.String())
}

func (s *componoTestSuite) TestRendererHookForEmAndStrong() {
	c := New()

	var calls []hook.RendererHookContext
	hookFn := func(ctx hook.RendererHookContext) string {
		calls = append(calls, ctx)
		return ctx.Output
	}

	var buf bytes.Buffer
	err := c.Convert([]byte("**bold** and *italic*"), &buf, WithRendererHook(hookFn))
	require.NoError(s.T(), err)

	foundStrong := false
	foundEm := false
	for _, call := range calls {
		if call.Kind == hook.KindMarkdown && call.Name == "strong" {
			foundStrong = true
		}
		if call.Kind == hook.KindMarkdown && call.Name == "em" {
			foundEm = true
		}
	}
	assert.True(s.T(), foundStrong, "expected hook to be called for strong markdown element")
	assert.True(s.T(), foundEm, "expected hook to be called for em markdown element")
}

func TestComponoTestSuite(t *testing.T) {
	suite.Run(t, new(componoTestSuite))
}

func requireHookString(t *testing.T, params hook.Params, name string) string {
	t.Helper()

	value, ok := params.String(name)
	require.True(t, ok, "expected hook param %q to be a string", name)
	return value
}

func requireHookArray(t *testing.T, params hook.Params, name string) hook.Array {
	t.Helper()

	value, ok := params.Array(name)
	require.True(t, ok, "expected hook param %q to be an array", name)
	return value
}

func requireHookRecord(t *testing.T, params hook.Params, name string) hook.Record {
	t.Helper()

	value, ok := params.Record(name)
	require.True(t, ok, "expected hook param %q to be a record", name)
	return value
}

func requireHookArrayString(t *testing.T, array hook.Array, index int) string {
	t.Helper()

	value, ok := array.String(index)
	require.True(t, ok, "expected hook array item %d to be a string", index)
	return value
}

func requireHookRecordString(t *testing.T, record hook.Record, name string) string {
	t.Helper()

	value, ok := record.String(name)
	require.True(t, ok, "expected hook record field %q to be a string", name)
	return value
}

func readContextFixture(path string) (map[string]any, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	decoder := json.NewDecoder(f)
	decoder.UseNumber()

	values := map[string]any{}
	if err := decoder.Decode(&values); err != nil {
		return nil, err
	}

	return normalizeJSONValue(values).(map[string]any), nil
}

func normalizeJSONValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		result := make(map[string]any, len(typed))
		for key, item := range typed {
			result[key] = normalizeJSONValue(item)
		}
		return result
	case []any:
		result := make([]any, 0, len(typed))
		for _, item := range typed {
			result = append(result, normalizeJSONValue(item))
		}
		return result
	case json.Number:
		if i, err := strconv.ParseInt(string(typed), 10, 64); err == nil {
			return i
		}
		return typed
	default:
		return value
	}
}
