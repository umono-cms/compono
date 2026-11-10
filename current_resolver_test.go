package compono

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const invalidCurrentValuePrefix = "The value **"

const invalidCurrentValueSuffix = "** is not allowed for the parameter **current**. Allowed values: **page**, **step**, **location**, **date**, **time**, **true**, **__auto**."

const resolverValuePrefix = "The current resolver returned **"

const resolverValueMiddle = "** for the URL **"

const resolverValueSuffix = "**. Allowed values: **page**, **step**, **location**, **date**, **time**, **true**."

// convertWithResolver converts source and returns the output, the diagnostics
// and the urls the resolver was asked about, in order.
func convertWithResolver(t *testing.T, source string, resolve func(url string) string, opts ...ConvertOption) (string, []Diagnostic, []string) {
	t.Helper()
	var urls []string
	opts = append(opts, WithCurrentResolver(func(url string) string {
		urls = append(urls, url)
		return resolve(url)
	}))
	var buf bytes.Buffer
	diags, err := New().Convert([]byte(source), &buf, opts...)
	require.NoError(t, err)
	return buf.String(), diags, urls
}

func convertPlain(t *testing.T, source string, opts ...ConvertOption) (string, []Diagnostic) {
	t.Helper()
	var buf bytes.Buffer
	diags, err := New().Convert([]byte(source), &buf, opts...)
	require.NoError(t, err)
	return buf.String(), diags
}

func blogIsPage(url string) string {
	if url == "/blog" {
		return "page"
	}
	return ""
}

func TestCurrentAutoWithoutResolverIsEmpty(t *testing.T) {
	out, diags := convertPlain(t, `{{ LINK text = "a" url = "/a" current = "__auto" }}`)
	assert.Empty(t, diags)
	assert.Equal(t, `<a href="/a">a</a>`, out)
}

func TestCurrentAutoMatchesCurrentPageAndEmpty(t *testing.T) {
	auto := `Menü: {{ LINK text = "Home" url = "/" current = "__auto" }} {{ LINK text = "Blog" url = "/blog" current = "__auto" }}`
	written := `Menü: {{ LINK text = "Home" url = "/" current = "" }} {{ LINK text = "Blog" url = "/blog" current = "page" }}`

	out, diags, urls := convertWithResolver(t, auto, blogIsPage)
	want, _ := convertPlain(t, written)

	assert.Empty(t, diags)
	assert.Equal(t, want, out)
	assert.Equal(t, `<p>Menü: <a href="/">Home</a> <a href="/blog" aria-current="page">Blog</a></p>`, out)
	assert.Equal(t, []string{"/", "/blog"}, urls)
}

func TestCurrentAutoResolverGetsTheHrefValue(t *testing.T) {
	_, _, urls := convertWithResolver(t, `{{ LINK text = "a" url = "  /x?a=1&b=2  " current = "__auto" }}`, blogIsPage)
	assert.Equal(t, []string{"/x?a=1&b=2"}, urls)
}

func TestCurrentAutoResolverValueIsTrimmed(t *testing.T) {
	out, diags, _ := convertWithResolver(t, `{{ LINK text = "a" url = "/a" current = " __auto " }}`, func(string) string { return " true " })
	assert.Empty(t, diags)
	assert.Equal(t, `<a href="/a" aria-current="true">a</a>`, out)
}

func TestCurrentAutoInvalidResolverValueDropsLink(t *testing.T) {
	for _, value := range []string{"Page", "false", "__auto", "<b>"} {
		t.Run(value, func(t *testing.T) {
			source := `Önce {{ LINK text = "a" url = "/a" current = "__auto" }} sonra`
			out, diags, _ := convertWithResolver(t, source, func(string) string { return value })
			assert.Equal(t, `<p>Önce  sonra</p>`, out)
			require.Len(t, diags, 1)
			assert.Equal(t, CodeInvalidCurrentValue, diags[0].Code)
			assert.Equal(t, resolverValuePrefix+value+resolverValueMiddle+"/a"+resolverValueSuffix, diags[0].Message)
			assert.Equal(t, span(t, source, `{{ LINK text = "a" url = "/a" current = "__auto" }}`, 0), diags[0].Range)
		})
	}
}

func TestCurrentAutoInvalidResolverValueDropsBlockLink(t *testing.T) {
	source := "{{ LINK text = \"a\" url = \"/a\" current = \"__auto\" }}\n\nsonra"
	out, diags, _ := convertWithResolver(t, source, func(string) string { return "Page" })
	assert.Equal(t, `<p>sonra</p>`, out)
	require.Len(t, diags, 1)
	assert.Equal(t, CodeInvalidCurrentValue, diags[0].Code)
	assert.Equal(t, resolverValuePrefix+"Page"+resolverValueMiddle+"/a"+resolverValueSuffix, diags[0].Message)
}

func TestCurrentAutoInvalidResolverValueMessageHasHrefURL(t *testing.T) {
	source := `{{ LINK text = "a" url = "  /x?a=1&b=2  " current = "__auto" }}`
	_, diags, urls := convertWithResolver(t, source, func(string) string { return " Page " })
	require.Len(t, diags, 1)
	assert.Equal(t, CodeInvalidCurrentValue, diags[0].Code)
	assert.Equal(t, resolverValuePrefix+"Page"+resolverValueMiddle+"/x?a=1&b=2"+resolverValueSuffix, diags[0].Message)
	assert.Equal(t, []string{"/x?a=1&b=2"}, urls)
}

func TestCurrentAutoInvalidResolverValueSkipsAttributeHook(t *testing.T) {
	called := 0
	hook := WithAttributeHook(func(string, []Frame) map[string]string {
		called++
		return nil
	})
	_, diags, _ := convertWithResolver(t, `{{ LINK text = "a" url = "/a" current = "__auto" }}`, func(string) string { return "Page" }, hook)
	require.Len(t, diags, 1)
	assert.Equal(t, 0, called)
}

func TestCurrentAutoResolverNotCalledForOtherLinks(t *testing.T) {
	source := `{{ LINK text = "a" url = "/a" current = "page" }} {{ LINK text = "b" url = "/b" current = "" }} {{ LINK text = "c" url = "/c" }} [d](/d)`
	_, diags, urls := convertWithResolver(t, source, blogIsPage)
	assert.Empty(t, diags)
	assert.Empty(t, urls)
}

func TestCurrentAutoResolverNotCalledForDroppedLink(t *testing.T) {
	_, diags, urls := convertWithResolver(t, `{{ LINK text = "" url = "/blog" current = "__auto" }}`, blogIsPage)
	require.Len(t, diags, 1)
	assert.Equal(t, CodeEmptyText, diags[0].Code)
	assert.Empty(t, urls)
}

func TestCurrentUnknownDirective(t *testing.T) {
	for _, value := range []string{"__page", "__", "__Auto", "_auto"} {
		t.Run(value, func(t *testing.T) {
			_, diags, urls := convertWithResolver(t, `{{ LINK text = "a" url = "/a" current = "`+value+`" }}`, blogIsPage)
			require.Len(t, diags, 1)
			assert.Equal(t, CodeInvalidCurrentValue, diags[0].Code)
			assert.Equal(t, invalidCurrentValuePrefix+value+invalidCurrentValueSuffix, diags[0].Message)
			assert.Empty(t, urls)
		})
	}
}

func TestCurrentAutoFromParameterRecordAndContext(t *testing.T) {
	source := "{{ ITEM item = { text: \"Blog\", url: \"/blog\", current: \"__auto\" } }}\n\n" +
		"{{ LINK text = context(m).text url = context(m).url current = context(m).current }}\n\n" +
		"{{ LIST items = context(menu) item = ITEM }}\n\n" +
		"~ ITEM item = {}\n" +
		"{{ LINK text = item.text url = item.url current = item.current }}\n"
	menu := []any{
		map[string]any{"text": "Home", "url": "/", "current": "__auto"},
		map[string]any{"text": "Blog", "url": "/blog", "current": "__auto"},
	}
	ctx := WithContext(map[string]any{
		"m":    map[string]any{"text": "Blog", "url": "/blog", "current": "__auto"},
		"menu": menu,
	})

	out, diags, urls := convertWithResolver(t, source, blogIsPage, ctx)
	assert.Empty(t, diags)
	assert.Equal(t, []string{"/blog", "/blog", "/", "/blog"}, urls)
	assert.Contains(t, out, `<a href="/">Home</a>`)
	assert.Equal(t, 3, bytes.Count([]byte(out), []byte(`<a href="/blog" aria-current="page">Blog</a>`)))
}

func TestCurrentAutoInGlobal(t *testing.T) {
	out, diags, urls := convertWithResolver(t, `{{ MENU }}`, blogIsPage,
		WithGlobalComponent("MENU", []byte(`{{ LINK text = "Blog" url = "/blog" current = "__auto" }}`)),
	)
	assert.Empty(t, diags)
	assert.Equal(t, []string{"/blog"}, urls)
	assert.Equal(t, `<a href="/blog" aria-current="page">Blog</a>`, out)
}

func TestCurrentResolverSetTwice(t *testing.T) {
	fn := func(string) string { return "" }
	_, err := New().Convert([]byte("x"), io.Discard, WithCurrentResolver(fn), WithCurrentResolver(fn))

	require.Error(t, err)
	var compErr *ComponoError
	require.ErrorAs(t, err, &compErr)
	assert.Equal(t, ErrCurrentResolverAlreadySet, compErr.Code)
}

func TestCurrentResolverNil(t *testing.T) {
	out, diags := convertPlain(t, `{{ LINK text = "a" url = "/a" current = "__auto" }}`, WithCurrentResolver(nil))
	assert.Empty(t, diags)
	assert.Equal(t, `<a href="/a">a</a>`, out)
}

func TestCurrentResolverNotAllowedInGlobal(t *testing.T) {
	fn := func(string) string { return "" }
	_, err := New().Convert([]byte("{{ G }}"), io.Discard,
		WithGlobalComponent("G", []byte("x"), WithCurrentResolver(fn)),
	)

	require.Error(t, err)
	var compErr *ComponoError
	require.ErrorAs(t, err, &compErr)
	assert.Equal(t, ErrConversionOptionInGlobal, compErr.Code)
}
