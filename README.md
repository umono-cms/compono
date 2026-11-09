# Compono

Compono is a Markdown-based domain-specific language (DSL) for semantic HTML. It extends a restricted Markdown syntax with reusable components and built-in components that produce semantic HTML elements.

Compono is developed for [Umono CMS](https://github.com/umono-cms/umono), and its design follows Umono's needs. It is a Go library and can be used in other Go projects, but HTML is its only output target.

## Installation

```bash
go get github.com/umono-cms/compono
```

## Quick Start

```go
package main

import (
    "bytes"
    "fmt"
    "github.com/umono-cms/compono"
)

func main() {
    c := compono.New()

    source := []byte(`{{ SAY_HELLO name="World" }}

~ SAY_HELLO name="Guest"
# Hello, {{ name }}!
`)

    var buf bytes.Buffer
    if _, err := c.Convert(source, &buf); err != nil {
        panic(err)
    }

    fmt.Println(buf.String())
    // Output: <h1>Hello, World!</h1>
}
```

## Syntax

### Markdown Support

Compono supports common Markdown elements:

```
# Heading 1
## Heading 2
### Heading 3

This is a paragraph with **bold** and *italic* text.

`inline code`

[Link text](https://example.com)
```

The Markdown surface is closed: the elements described in this section are all there is, and every other syntax, such as `_text_`, `~~text~~`, `![alt](url)`, `# Heading {#id}`, `~~~` fences or raw HTML, stays plain text. What Markdown does not cover is done with built-in components.

`*text*` renders `<em>` and `**text**` renders `<strong>`. A delimiter is a run of `*`: a run of one opens and closes an em, a run of two a strong, and a run of three or more is text. An opener is not followed by a space and a closer is not preceded by one. An em opener pairs with the next run of `*` and a strong opener with the next run of two; if that run cannot close, the opener is text. Both ends are on the same line. em and strong do not nest: `**a *b* c**` is a strong whose text is `a *b* c`. `_` and `__` are text.

`` `code` `` renders `<code>` with its content as raw, escaped text. A single backtick pairs with the next backtick run on the same line; if that run is longer than one, the backtick is text, so ` `` ` and ` ``code`` ` are text. Code spans and `{{ }}` units are recognized first, whichever opens first: `` `{{ name }}` `` is a code span, and in ``{{ LINK text = "`a` [b](/c)" url = "/" }}`` the argument is plain text. Links come next, then strong, then em, and none of them can cut through a code span or a `{{ }}` unit.

`[text](address)` renders `<a href="address">`. The text runs from `[` to the first `]` that the `[`s inside it do not balance, so `[[1]](#ref)` has the text `[1]` and `[[a](/b)](/c)` has the plain text `[a](/b)`: links do not nest. The address ends at the first `)`. A link is a single line, and `]` and `(` are written together. The text is trimmed and may hold strong, em, code spans and parameter references. A component call in the text or the address is an `invalid-component-usage` error, an empty text is `empty-text` and an empty address is `empty-url`; each drops the link. The address must pass the [URL scheme allowlist](#url-scheme-allowlist).

Code blocks are also supported:

~~~
```go
fmt.Println("Hello")
```
~~~

```html
<pre><code data-lang="go">fmt.Println(&#34;Hello&#34;)
</code></pre>
```

The language is written as `data-lang` on `<code>`. A code block without a language renders `<pre><code>` with no attribute. A language contains only lowercase ASCII letters, digits and `+ # . _ -`, starts with a letter or digit and is at most 32 bytes (`c++`, `c#`, `objective-c`). Otherwise the code block renders nothing and returns an `invalid-code-block-language` diagnostic. An empty code block, or one with only whitespace, renders nothing.

The opening line is a fence of three or more backticks, followed without a space by the optional language. It is closed by the first following line that is exactly a fence of the same length; every line in between is content, kept as written. So a code that holds a ```` ``` ```` line is written with a ```` ```` ```` fence. The indentation of the closing line is not part of the content. An opening line that is never closed is text. `~~~` fences and indented code blocks do not exist.

A line that is exactly `---` (spaces and tabs around it are ignored) is a thematic break and renders `<hr>`. It interrupts a paragraph. There are no setext headings: `---` or `===` under a line never turns it into a heading. `***`, `___` and `----` are text.

A line that starts with `> ` is a quote line, and a line that is only `>` is an empty quote line. Leading spaces and tabs are ignored. Consecutive quote lines render a `<blockquote>` whose paragraphs are always wrapped in `<p>`:

```
> First paragraph,
> second line.
>
> Second paragraph.
```

```html
<blockquote><p>First paragraph,<br>second line.</p><p>Second paragraph.</p></blockquote>
```

A quote interrupts a paragraph and ends at the first line that does not start with `>`. There is no lazy continuation and no nesting. A quote holds only paragraphs: `> > text`, `> # text`, `> - text` and `>text` are text. A block component call inside a quote is an invalid component usage. A blank line between two quotes renders two blockquotes.

A line that starts with `- ` followed by content is a list item. Leading spaces and tabs are ignored. Consecutive item lines render a `<ul>` with one `<li>` per line, and the item content is not wrapped in `<p>`:

```
- Apple
- Pear
  - Quince
```

```html
<ul><li>Apple</li><li>Pear</li><li>Quince</li></ul>
```

A list interrupts a paragraph and ends at the first line that does not start with `- `. There is no lazy continuation and no nesting: indentation and `- - text` make sibling items. Only `-` starts an item, so `+ text`, `* text`, `1. text`, `-text` and a lone `-` are text. An item is a single line of inline content, so a block component call inside it is an invalid component usage. An item that resolves empty renders no `<li>`. Ordered and nested lists are built with `LIST`.

A line that starts with `|` (after leading spaces and tabs) is a table row, and consecutive table rows form a table. The first row is the header, the second is the delimiter row and the rest are body rows:

```
| Plan     | Price |
| -------- | ----- |
| Personal | 0 ₺   |
| Team     | 490 ₺ |
```

```html
<table><thead><tr><th scope="col">Plan</th><th scope="col">Price</th></tr></thead><tbody><tr><td>Personal</td><td>0 ₺</td></tr><tr><td>Team</td><td>490 ₺</td></tr></tbody></table>
```

- A row starts and ends with `|`. The `|`s in between separate the cells. A `|` inside a code span or a `{{ }}` unit does not separate cells, and there is no other way to write a `|` in a cell.
- Each delimiter cell is only `-`, at least three of them. There is no alignment.
- A cell is a single line of inline content. It is trimmed and written even when it is empty. A block component call inside it is an invalid component usage that drops only the call.
- A table interrupts a paragraph and ends at the first line that does not start with `|`. Comment lines inside a table do not split it. Two tables are separated by a line that is not a table row, such as a blank line.
- There is no caption, row header, header-less table or multi-line cell.

A table that breaks this structure does not fall back to text. It renders nothing and returns one diagnostic, for the first broken row from the top: `invalid-table-row` (a row that does not start and end with `|`), `invalid-table-delimiter` (a missing or wrong delimiter row), `table-cell-count-mismatch` (a row whose cell count differs from the header's), `empty-table-header` (a header cell with an empty source) or `empty-table` (no body row). A header cell whose source is not empty but resolves empty is not an error.

### Empty Content

Compono writes no empty element. A heading, paragraph, list item, list, blockquote, `*em*` or `**strong**` whose content is empty after trimming spaces and tabs renders nothing, and so does a code block that is empty or only whitespace. The rule works from inside out: `## **{{ subtitle }}**` renders nothing when `subtitle` is empty. A paragraph line that resolves empty is dropped with its line break:

```
A
{{ empty }}
B
```

```html
<p>A<br>B</p>
```

Empty content is not an error, and a unit dropped by an error still returns its diagnostic. Table cells are the only exception: `<th>` and `<td>` are written even when empty, so the other cells of the row keep their columns.

### Comments

A line that starts with `//` is a comment line. Leading spaces and tabs are ignored. Comment lines are removed from the source, with their line breaks, before anything else is read, and they write nothing to the output:

```
# About
// TODO: confirm the founding year
Umono was founded in 2024.
```

Output:

```html
<h1>About</h1><p>Umono was founded in 2024.</p>
```

- A comment line is not an empty line, so it does not split a paragraph. `First line`, `// note`, `Second line` give `<p>First line<br>Second line</p>`.
- Comment lines are not lines of a component body. A component whose body is a comment line and one paragraph line is still inline.
- The content of a comment is not read. A call on a comment line is not called and produces no error, so `// {{ CARD }}` disables a line.
- There are no inline, multi-line or block comments. `Text // note`, `/* note */` and `<!-- note -->` are plain text.
- `//` does not start a comment inside a code block, inside a `{{ }}` unit that spans several lines, or after a heading's `# `.

Comments work in the converted source, in local components and in global components.

### Invalid and Unclosed Units

Every `{{` in inline content starts a unit. A unit that is not a component call, a parameter reference or a `context(key)`, such as `{{ title | upper }}`, `{{ Title }}` or `{{ X (a = 1) }}`, is an `invalid-expression` error and drops as a whole. The names inside it are not checked. A valid unit ends where its grammar ends, so a `}}` inside a string does not close it. An invalid unit ends at the first `}}` and cannot run past its paragraph, heading, list item, quote line or table row. On its own lines it splits the paragraph like a block call.

A `{{` with no `}}` after it, or with another `{{` before the first `}}`, is an `unclosed-expression` error and only the `{{` drops: `Hello {{ name } world` renders `<p>Hello  name } world</p>`, and in `{{ name } {{ FOOTER }}` the second unit is still rendered. A `{{` is never plain text in inline content; to write one, put it in a code span or a code block.

### Components

Components are the core feature of Compono. They allow you to create reusable content blocks.

#### Defining a Local Component

Local components are defined in the same scope where they're used:

```
{{ GREETING }}

~ GREETING
Welcome to our website!
```

The `~ COMPONENT_NAME` syntax marks the beginning of a local component definition. Everything after it becomes the component's content. A component definition ends when another component definition starts or at EOF.

#### Components with Parameters

Components can accept parameters with default values:

```
{{ USER_CARD name="Anonymous" role="Guest" }}

~ USER_CARD name="" role=""
## {{ name }}
*{{ role }}*
```

#### Block vs Inline Components

Components containing multiple paragraphs or block elements are **block components**:

```
{{ ARTICLE }}

~ ARTICLE
# Title
First paragraph.

Second paragraph.
```

Components with single-line content can be used **inline**:

```
Welcome, {{ USERNAME }}!

~ USERNAME
John
```

A component whose body is a single component call takes the class of the called component, so a wrapper is transparent. A wrapper of an inline component is inline, and a wrapper of `LINK` or `IMAGE` can be called both inline and on its own line:

```
See {{ DOCS_LINK }} for details.

~ DOCS_LINK
{{ LINK text = "Docs" url = "/docs" }}
```

This outputs `<p>See <a href="/docs">Docs</a> for details.</p>`. The rule is recursive. Comment lines do not count.

A body that is only a reference to a component parameter (`{{ content }}`, also `{{ item.content }}` when the default is `item = { content: NO_MATTER }`) is block whatever its value is. The type of a field or an index is read from the same place in the default value. When the default does not have it (`item = {}`, `list = []`), the type is unknown and a body that is only `{{ item.content }}` is a paragraph line, so the component is inline. If such a body is rendered inline and the value is a component, an inline component is placed inline and a block component is an `invalid-component-usage` error that drops the reference. Called on its own line, the body renders the value as a block. A body that is a call to an unknown component or a call that loops back to itself is block too. Calling a block component inline is an `invalid-component-usage` error and the call is dropped.

### Global Components

Global components are given to `Convert` with `WithGlobalComponent` and exist only in that conversion:

```go
c := compono.New()

c.Convert([]byte(`
# Page Title
Content here...
{{ FOOTER }}
`), &buf, compono.WithGlobalComponent("FOOTER", []byte(`© 2026 My Company`)))
```

A `Compono` value does not store global components. Pass them again to every conversion that uses them.

Global components can also have parameters. They are defined on the **parameter line** at the start of the source. It starts with `~` and is the first line of a local component definition without the name, because the name is given by `WithGlobalComponent`:

```go
c.Convert(source, &buf, compono.WithGlobalComponent("BLOG_PAGE", []byte(`~ title="" content=""
## {{ title }}
{{ content }}`)))
```

- Comment lines and blank lines may come before the parameter line.
- If the first line does not start with `~`, the global component has no parameters and that line is content. A line like `title = ""` renders as text.
- If the first line starts with `~ NAME`, it begins a local component definition and the global component has no parameters.
- The parameter line follows the same rules as the parameters of a local component definition. A malformed line and a `~` line without parameters are errors. See [Parameter Definition Errors](#parameter-definition-errors).

### Global Component Scopes

All global components share one namespace, so two unrelated groups of components can clash: a name may be defined twice, or a call inside one group may resolve to a component of the other. Scopes solve this. A `WithGlobalComponent` can take other `WithGlobalComponent`s as options. These are its **sub components**, and it is their **owner**:

```go
c.Convert(pageSource, w,
	compono.WithGlobalComponent("CARD", cardSource),
	compono.WithGlobalComponent("LAYOUT", layoutSource,
		compono.WithIsolatedScope(),
		compono.WithGlobalComponent("CARD", otherCardSource),
		compono.WithGlobalComponent("TABLE", tableSource),
	),
)
```

There are two `CARD`s here and both work. A `CARD` call in the page resolves to the first one, and a `CARD` call inside `LAYOUT` resolves to the second one. `LAYOUT` cannot reach the page's `CARD`.

**Visibility:** A sub component can only be called from its owner's body and from its siblings' bodies. It is not visible from the converted source or from other global components. It neither overrides a same-named global component outside nor is overridden by one.

**Resolution:** A call inside a global component's body looks in this order:

1. The local components of the source the call is written in.
2. The global component's sub components.
3. The scope the global component is defined in, then outward through the scopes of its owners, and finally the global components given to `Convert` directly.
4. Built-in components.

A global component without sub components resolves exactly as before: `Local > Global > Built-in`. Nesting depth is unlimited and the same rules apply at every level.

**Isolation:** `WithIsolatedScope()` stops step 3. Calls in the body of that global component and in the bodies of its sub components resolve only to locals, its sub components and built-ins. Giving it more than once has the same effect as giving it once.

**Syntactic resolution:** A call resolves in the scope of the source it is written in, not where it is rendered. If the page passes `{{ LAYOUT content = BODY }}`, the calls inside `BODY` resolve in the page's scope even though they are rendered inside `LAYOUT`. `LAYOUT`'s sub components are not visible from `BODY`.

**Built-in names:** Sub components override built-in components, and Compono does not check sub component names against the built-in list. Adding a new built-in in a later version does not change the output of a sub component with that name.

**Errors:** `Convert` returns an error when:

- a conversion option (`WithContext`, `WithAttributeHook`, `WithURLReporter`) is given to a global component at any depth, even with an empty value (`ErrConversionOptionInGlobal`),
- `WithIsolatedScope` is given directly to `Convert` (`ErrIsolatedScopeInConvert`),
- two sub components of the same owner share a name (`ErrDuplicateSubComponent`),
- two global components given to `Convert` directly share a name (`ErrDuplicateGlobalComponent`).

## Built-in Components

### LINK

Creates an anchor element with optional target blank:

```
{{ LINK text="Visit us" url="https://example.com" new-tab=true }}
```

Output:
```html
<a href="https://example.com" target="_blank" rel="noopener noreferrer">Visit us</a>
```

| Parameter | Type   | Required | Default | Description |
| --------- | ------ | -------- | ------- | ----------- |
| `text`    | String | Yes      |         | The visible text. It is trimmed and cannot be empty (`empty-text`). |
| `url`     | String | Yes      |         | The `href`. It cannot be empty (`empty-url`) and must pass the URL scheme allowlist. |
| `new-tab` | Bool   | No       | `false` | Writes `target="_blank" rel="noopener noreferrer"`. |
| `current` | String | No       | `""`    | Writes `aria-current` when it is not empty. One of `page`, `step`, `location`, `date`, `time` or `true`, otherwise `invalid-current-value`. |

```
{{ LINK text = "Home" url = "/" current = "page" }}
```

Output:
```html
<a href="/" aria-current="page">Home</a>
```

A missing `text` or `url` is an `invalid-built-in-arguments` error. The attributes are written in the order `href`, `target`, `rel`, `aria-current`.

#### URL Scheme Allowlist

The `url` of `LINK` and the URL of a markdown link must use `http`, `https`, `mailto` or `tel`, or be relative (`/docs`, `docs`, `#part`, `?page=2`, `//cdn.example.com`). Any other scheme, such as `javascript:` or `data:`, is a `disallowed-url-scheme` error and the link is dropped.

The scheme is found the way browsers find it: leading and trailing control characters and spaces are trimmed, ASCII tabs and newlines inside the URL are removed, and the scheme is compared case-insensitively, so `java	script:` and `JavaScript:` are rejected. The `href` is the URL after this normalization. A relative path whose first segment contains `:` is read as a scheme. Write it as `./docs:intro`.

### IMAGE

Creates a semantic web image output from a `media` record and optional responsive variants.

Basic usage:

```
{{ IMAGE media = {
  url: "https://cdn.example.com/my-photo.jpg",
  width: 1600,
  height: 900,
  mime-type: "image/jpeg",
  variants: [
    {
      url: "https://cdn.example.com/my-photo-640.avif",
      width: 640,
      height: 360,
      mime-type: "image/avif"
    },
    {
      url: "https://cdn.example.com/my-photo-1280.avif",
      width: 1280,
      height: 720,
      mime-type: "image/avif"
    },
    {
      url: "https://cdn.example.com/my-photo-1600.avif",
      width: 1600,
      height: 900,
      mime-type: "image/avif"
    },
    {
      url: "https://cdn.example.com/my-photo-640.webp",
      width: 640,
      height: 360,
      mime-type: "image/webp"
    },
    {
      url: "https://cdn.example.com/my-photo-1280.webp",
      width: 1280,
      height: 720,
      mime-type: "image/webp"
    },
    {
      url: "https://cdn.example.com/my-photo-1600.webp",
      width: 1600,
      height: 900,
      mime-type: "image/webp"
    }
  ]
} alt = "My photo" }}
```

Output:
```html
<picture><source type="image/avif" srcset="https://cdn.example.com/my-photo-640.avif 640w, https://cdn.example.com/my-photo-1280.avif 1280w, https://cdn.example.com/my-photo-1600.avif 1600w"><source type="image/webp" srcset="https://cdn.example.com/my-photo-640.webp 640w, https://cdn.example.com/my-photo-1280.webp 1280w, https://cdn.example.com/my-photo-1600.webp 1600w"><img src="https://cdn.example.com/my-photo.jpg" alt="My photo" width="1600" height="900"></picture>
```

Without variants, `IMAGE` still renders a `picture` element that holds only the `img`:

```
{{ IMAGE media = {
  url: "https://cdn.example.com/avatar.png",
  width: 512,
  height: 512,
  mime-type: "image/png"
} alt = "Profile avatar" }}
```

Output:
```html
<picture><img src="https://cdn.example.com/avatar.png" alt="Profile avatar" width="512" height="512"></picture>
```

`IMAGE` can be used inline or as a block component depending on where it is called:

```
Gallery cover: {{ IMAGE media = {
  url: "https://cdn.example.com/gallery-cover.jpg",
  width: 800,
  height: 450,
  mime-type: "image/jpeg"
} alt = "Gallery cover" }} is ready.
```

#### IMAGE Parameters

- `media` is required and must be a record with:
  - `url`
  - `width`
  - `height`
  - `mime-type`
  - optional `variants`
- `alt` is a required string. It may be empty: pass `alt = ""` for decorative images. An `IMAGE` without `alt` is an `invalid-built-in-arguments` error.

Each item in `variants` must be a record with:

- `url`
- `width`
- `height`
- `mime-type`

Supported mime types:

- `image/jpeg`
- `image/png`
- `image/webp`
- `image/gif`
- `image/avif`

#### IMAGE Behavior

- `media` is always the fallback image source.
- HTML output is always a `picture` element, with a `source` for every variant group and the fallback `img` last.
- every `url` must pass the [URL scheme allowlist](#url-scheme-allowlist) and is written after its normalization. An empty `url` is an `empty-url` error.
- a variant `url` must be usable in `srcset`: it cannot contain whitespace or start or end with a comma (`invalid-variant-url`).
- variants are grouped by first-seen `mime-type`, preserving the original group order.
- within each mime type group, `srcset` entries are sorted by ascending width.
- an empty `variants` array is valid and renders only the fallback `img`.
- all widths and heights must be greater than `0`.
- all variants must preserve the same aspect ratio as the main `media`.
- duplicate `mime-type` + `width` pairs are invalid.

When validation fails, the IMAGE call renders nothing instead of producing invalid markup, and `Convert` returns a diagnostic. Common IMAGE-specific diagnostic codes include:

- `invalid-built-in-arguments`
- `empty-url`
- `disallowed-url-scheme`
- `invalid-variant-url`
- `unsupported-mime-type`
- `invalid-dimension`
- `duplicate-variant`
- `inconsistent-aspect-ratio`

### LIST

`LIST` renders its item component once for every record of `items` and wraps each result in an `<li>`, inside a `<ul>`, or an `<ol>` when `ordered = true`.

```
{{ LIST items = [{ text: "Apple" }, { text: "Pear" }] item = FRUIT }}

~ FRUIT item = {}
{{ item.text }}
```

Output:
```html
<ul><li>Apple</li><li>Pear</li></ul>
```

| Parameter | Type      | Required | Default | Description |
| --------- | --------- | -------- | ------- | ----------- |
| `items`   | Array     | Yes      |         | The records to iterate. Every element must be a record. |
| `item`    | Component | Yes      |         | The **item component**, rendered for every element. |
| `ordered` | Bool      | No       | `false` | `true` renders an `<ol>`, `false` a `<ul>`. |

- The item component must define a record parameter named `item`. `LIST` gives it every element as `item`, so `item` may be defined as required: `~ CARD item! = {}`. A built-in or a component without such a parameter is an `invalid-item-component` error.
- Other parameters of the item component are given with [argument binding](#argument-binding): `item = CARD(more = more)`. Binding `item` itself is a `duplicate-argument` error. A required parameter that is not bound is a `missing-argument` error, also when `items` is empty.
- An inline item component is written in the `<li>` without a `<p>`. A component that delegates to an inline component or to `LINK` is inline too. A block item component is written like a block call.
- An element that is not a record is an `invalid-list-item` error and drops only that element. An element that renders nothing writes no `<li>`, and the numbers of an ordered list follow the written elements. A `LIST` without any `<li>`, for example with empty `items`, renders nothing.
- While it renders its elements, `LIST` is a `builtin` frame of the attribute hook chain and of the diagnostics' call chains.

Nested lists are built by an item component that calls `LIST` again with its own data and itself as the item component:

```
{{ LIST items = menu item = NAV_ITEM }}

~ NAV_ITEM item = {}
{{ LINK text = item.text url = item.url }}
{{ LIST items = item.children item = NAV_ITEM }}
```

Entering the item component again with the same signature while it is being rendered is normally an `infinite-component-call` error ([What Is Dropped](#what-is-dropped)). `LIST` allows it when the inner `LIST`'s `items` are equal to a part of the current `item` of the outer `LIST` reached through at least one key or index access, also when passed through parameters. The parts are compared by value, not by where they are written. Every such step renders a smaller part of the data, so the recursion ends where the data ends. `{{ LIST items = menu item = NAV_ITEM(menu = menu) }}` inside `NAV_ITEM` does not shrink the data and is still an infinite call.

### HEADER, MAIN and FOOTER

These built-ins wrap the component given by their required `content` parameter in a `<header>`, `<main>` or `<footer>` element. They are block components and take no other parameter. Together they form the landmarks of a page skeleton:

```
{{ HEADER content = SITE_HEADER(site-name = site-name) }}
{{ MAIN content = content }}
{{ FOOTER content = SITE_FOOTER }}

~ SITE_HEADER site-name = ""
{{ LINK text = site-name url = "/" }}

~ SITE_FOOTER
© 2026 Umono
```

Output (with `site-name = "Umono"` and a `content` whose body is `# Hello`):
```html
<header><a href="/">Umono</a></header><main><h1>Hello</h1></main><footer><p>© 2026 Umono</p></footer>
```

- `content` is rendered like a block component call on its own line. Arguments are given to it with [argument binding](#argument-binding).
- If `content` renders nothing, the built-in renders nothing. No empty element is written.
- An error inside `content` drops only its own unit. The built-in is still rendered.
- `MAIN` cannot be placed inside another built-in, and a conversion renders only its first `MAIN` call. `HEADER` and `FOOTER` cannot be placed inside a `HEADER` or a `FOOTER`. These rules look at where the call is, not at what it renders, and break with `invalid-component-usage`.
- While these built-ins render their content, they are `builtin` frames of the [attribute hook](#attribute-hook) chain and of the call chain of [diagnostics](#diagnostics).

### ARTICLE and ASIDE

`ARTICLE` wraps its required `content` in an `<article>`: a self-contained piece such as a post or a comment. `ASIDE` wraps it in an `<aside>`: content that is related to its surroundings but can be removed, such as related posts. `ASIDE` takes an optional `label`, written as `aria-label` when its trimmed value is not empty.

```
{{ ARTICLE content = POST(title = "Hello") }}

~ POST title = ""
{{ HEADER content = POST_HEADER(title = title) }}
Body.
{{ ASIDE content = RELATED label = "Related posts" }}

~ POST_HEADER title = ""
# {{ title }}

~ RELATED
First post.
```

Both behave like `HEADER`, `MAIN` and `FOOTER`: they are block components, render nothing when their content renders nothing, and can be nested. They have no placement rule of their own. A `MAIN` cannot be inside them, which is `MAIN`'s own rule.

### NAV

`NAV` wraps its required `content` in a `<nav>`, a navigation region. Its optional `label` is the accessible name of the region and is written as `aria-label` when its trimmed value is not empty. Give every `NAV` of a page a different `label`.

```
{{ NAV label = "Footer" content = FOOTER_LINKS }}

~ FOOTER_LINKS
{{ LINK text = "Privacy" url = "/privacy" }} · {{ LINK text = "Contact" url = "/contact" }}
```

Output:
```html
<nav aria-label="Footer"><p><a href="/privacy">Privacy</a> · <a href="/contact">Contact</a></p></nav>
```

`NAV` behaves like the other content built-ins: it is a block component, renders nothing when its content renders nothing and can be nested.

### DETAILS

`DETAILS` renders a section that the visitor opens and closes: a `<details>` whose `<summary>` is the required `text` and whose body is the required `content`. `open = true` makes the section open when the page loads.

```
{{ DETAILS text = "How do I install a theme?" content = ANSWER open = true }}

~ ANSWER
Use the CLI.
```

Output:
```html
<details open><summary>How do I install a theme?</summary><p>Use the CLI.</p></details>
```

- `text` is trimmed and escaped. An empty `text` is an `empty-text` error.
- If `content` renders nothing, `DETAILS` renders nothing, even with a `text`.
- `DETAILS` can be given as the content of another built-in, for example a menu that opens on small screens: `{{ NAV label = "Main" content = DETAILS(text = "Menu" content = MAIN_MENU) }}`.

### Removed Built-in Components

#### WEB_GRID

`WEB_GRID` was removed in v0.7. It described a grid layout through parameters such as columns, rows, areas and breakpoints. That is presentation, not meaning. Compono only carries semantic content, and layout belongs to the stylesheet of whoever renders the output. Keeping it would also have frozen its output DOM as part of a stable contract.

A `WEB_GRID` call is now an `unknown-component` error and renders nothing. Local or global components named `WEB_GRID` are not affected and work as regular components.

#### NAVIGATION

`NAVIGATION` was removed in v0.7. It bundled a whole menu concept (`<nav>`, `<ul>`, `<li>` and `<a>`) into one built-in. It always produced an unordered list, so ordered navigations like breadcrumbs could not be expressed, and it could not carry attributes such as `aria-current` or `aria-label`. Its item records also used `label`/`target` instead of Compono's `text`/`url` naming.

A `NAVIGATION` call is now an `unknown-component` error and renders nothing. A menu is now built with [NAV](#nav) and `LINK`. Local or global components named `NAVIGATION` are not affected and work as regular components.

## Parameters

Components can accept parameters. Each parameter must have a **default value** defined in the component definition.

If a parameter value is not provided during the call, the **default value is used**. A parameter can also be marked as [required](#required-parameters).

```
{{ SAY_HELLO name="Jane" }}

~ SAY_HELLO name="John"
# Hello, {{ name }}!
```

### Supported Types

Supported parameter types:

- **String** → `name = "John"`. A `"` inside a string is written twice: `title = "Umono's ""theme"" system"`. This is the only escape. `""` is the empty string, `""""` is a string of one `"`, and `\` is a plain character (`"C:\Users"`).
- **Integer** → `age = 25`
- **Bool** → `active = true`
- **Component** → `comp = COMP`
- **Array** → `items = ["Jane", 22, true, COMP]`
- **Record** → `config = { lang: "tr", for-admin: true }`

### Whitespace

Whitespace in Compono is only the space (U+0020) and the tab (U+0009). Other space characters, such as U+00A0 and U+3000, are content.

- Values are never trimmed. A parameter value, a `context(...)` value, an array element and a record value are placed as they are: `Before[{{ t }}]after` with `t = "  a  "` gives `Before[  a  ]after`.
- The content of a heading or a paragraph is trimmed at its edges after the values are placed. `# {{ t }}` with `t = " a "` gives the same output as `# a`.
- The string arguments of built-in components, such as the `text` of `LINK` and the `alt` of `IMAGE`, are trimmed the same way.
- Trimming happens before the units dropped by errors are removed. A dropped unit counts as content, so the whitespace around it is kept: `Before {{ UNKNOWN }}` gives `<p>Before </p>`. A parameter with an empty value is not dropped and is trimmed: `Before {{ t }}` with `t = ""` gives `<p>Before</p>`. The output with the error and the output after the error is fixed differ only by the unit itself. [Empty content](#empty-content) is checked after the dropped units are removed, so `{{ UNKNOWN }} {{ UNKNOWN }}` renders nothing.
- A value of only U+00A0 is not empty. It renders in a heading and in the `text` of `LINK`.

### Required Parameters

A parameter whose name is followed by `!` is required. A call that does not give it is dropped and returns a `missing-argument` diagnostic:

```
{{ COVER media = context(media-by-alias/cover) }}

~ COVER media! = {} alt! = ""
{{ IMAGE media = media alt = alt }}
```

This drops the `COVER` call with the message `The parameter **alt** of component **COVER** is required.` When several are missing, the message is `The parameters **a**, **b** of component **[component]** are required.`

- The `!` is glued to the name: `alt! = ""`.
- The default value of a required parameter only declares its type. It is never used or resolved.
- An empty value satisfies the requirement. `alt = ""` is a given argument. The component decides what an empty value means.
- A required argument can be given in the call or bound with [Argument Binding](#argument-binding). Bound and given arguments are checked together.
- The call that misses the argument is dropped. If the called component comes from a component parameter, the call through the parameter (`{{ content }}`) is checked every time it is rendered and dropped where it is written.
- `!` is for local and global components. Required parameters of built-in components are checked by their own errors.

### Parameter Definition Errors

A parameter definition must be `name = default` or `name! = default`. A definition without a default value (`~ X a`) or with any other shape (`~ X a ! = ""`, `~ X !a = ""`) is an `invalid-parameter-definition` error with the message `The parameter definition **[text]** of component **[component]** is invalid.` Every call of the component is dropped and returns this diagnostic.

The same applies to the [parameter line](#global-components) of a global component. `~ title ! = ""` and a `~` line without parameters are this error too.

A definition line, the first line of a local component definition or the parameter line of a global component, is a single line. A non-blank line that starts with a space or a tab right after it is also an `invalid-parameter-definition` error, so the first line of a component body cannot be indented.

---

### Passing Parameters to Other Components

A parameter can be passed directly to another component call.

```
{{ USER age=31 }}

~ USER age=18
{{ ANOTHER_COMP another-integer-param=age }}

~ ANOTHER_COMP another-integer-param=0
Integer: *{{ another-integer-param }}*
```

Here:

- `USER` receives `age`
- it forwards that value to `ANOTHER_COMP`

---

### Passing Components as Parameters

Components themselves can also be passed as parameters.

```
{{ USER name="Yunus Emre" age=31 age-wrapper=AGE_WRAPPER_2 }}

~ USER name="John" age=25 age-wrapper=AGE_WRAPPER_1
# Welcome **{{ name }}**!
{{ age-wrapper age=age }}

~ AGE_WRAPPER_1 age=0
Your age: *{{ age }}*

~ AGE_WRAPPER_2 age=0
*{{ age }}*
```

Here:

- `age-wrapper` receives a **component**
- that component is executed inside `USER`

---

### Argument Binding

When a component is passed as a value, it is not known who renders it and with which arguments. Arguments can be bound to a component value in parentheses:

```
{{ SITE_NAV menu-title = "Main Menu" }}

~ SITE_NAV menu-title = ""
{{ WRAPPER content = MENU(title = menu-title) }}

~ WRAPPER content = NO_MATTER
{{ content }}

~ MENU title = ""
## {{ title }}
```

Output:

```html
<h2>Main Menu</h2>
```

- The parentheses hold an argument list, just like a component call. Arguments are separated by spaces and can span multiple lines: `CARD(title = "Hello" tags = ["a", "b"])`.
- Values can be anything a call argument can be: literals, parameter references (including accessors such as `item.children`), `context(key)`, arrays, records and components.
- The name and the parenthesis must be adjacent and the parentheses cannot be empty. `CARD (title = "Hello")` and `CARD()` are not argument binding. Write only the name when no argument is bound.
- A bound value is resolved where it is written. In `MENU(title = menu-title)`, `menu-title` is the parameter of `SITE_NAV`, no matter where `MENU` is rendered.
- A bound component can be used anywhere a component value can: as a call argument, a default value, an array item or a record value. Bindings can be nested: `CARD(footer = FOOTER(text = "Bye"))`.
- Built-in components can be bound too: `content = LINK(text = text url = url)`.
- When a bound component is rendered, the bound arguments and the arguments given by the caller are used together. Parameters that are neither bound nor given use their default values.

```
{{ PAGE }}

~ PAGE card = CARD(title = "Bound title")
{{ card body = "Given by the caller" }}

~ CARD title = "" body = ""
## {{ title }}
{{ body }}
```

Output:

```html
<h2>Bound title</h2><p>Given by the caller</p>
```

#### Argument Binding Errors

- Binding a parameter that the component does not define is an `unknown-parameter` error.
- Binding a value of the wrong type is a `wrong-argument-type` error.
- Required arguments of a bound component are validated with the bound and the given arguments together. This applies to built-in components and to [required parameters](#required-parameters) of local and global components.
- Giving an argument to a parameter that is already bound is a `duplicate-argument` error with the message `The parameter **[name]** of component **[component]** is already bound.` The call that gives the argument is dropped. There is no precedence rule between a bound and a given argument.

```
{{ WRAPPER content = CARD(title = "Bound") }}

~ WRAPPER content = NO_MATTER
{{ content title = "Given" }}

~ CARD title = ""
# {{ title }}
```

An error of a bound value drops the call the binding is written in. For a value bound in a default value, it drops the call that uses the default value. A missing or duplicate argument is found where the component value is called (`{{ content }}`) and drops that call.

---

### Parameter Visibility

A component only sees the parameters it defines. There is no parameter inheritance:

- a local component of a global component does not see the parameters of that global component
- a component does not see the parameters of the component calling it

Every value a component needs is passed to it as an argument. Referencing a parameter that is not defined by the component is an `unknown-parameter` error and drops the `{{ }}` unit.

```
{{ OUTER title = "Hello" }}

~ OUTER title = ""
{{ INNER }}

~ INNER
# {{ title }}
// Unknown parameter: INNER does not define title
```

Pass the value explicitly instead:

```
~ OUTER title = ""
{{ INNER title = title }}

~ INNER title = ""
# {{ title }}
```

The same applies to global components:

```go
c.Convert(source, &buf, compono.WithGlobalComponent("PROFILE_PAGE", []byte(`
~ name="Guest"

{{ PROFILE_CARD name = name }}

~ PROFILE_CARD name = ""
## {{ name }}
Welcome to the profile page.
`)))
```

Usage:

```
{{ PROFILE_PAGE name="Yunus" }}
```

Output:

```
<h2>Yunus</h2>
<p>Welcome to the profile page.</p>
```

---

### Array Parameters
```
{{ WRAPPER names = ["John", "Jane"] }}

~ WRAPPER names = []
{{ SAY_HELLO name = names[0] }}
{{ SAY_HELLO name = names[1] }}

~ SAY_HELLO name = ""
# Hello **{{ name }}**!
```

Arrays do not have to be homogeneous.
```
~ COMP mix = ["Jane", 22, true, SAY_HELLO]
We can reach an element via index.
{{ mix[2] }}
// true
```

Arrays can be nested.
```
{{ TABLE data = [
  [1,2],
  [3,4],
]}}

~ TABLE data = []
{{ data[0][0] }} - {{ data[0][1] }}
{{ data[1][0] }} - {{ data[1][1] }}
```

---

### Record Parameters
Pass data as key - value

```
{{ COMP record = { title: "Hello", content: "Here Content" } }}

~ COMP record = {}
# {{ record.title }}
{{ record.content }}
```
Records can be nested
```
{{ COMP nested = {record: {key-1: "string", key-2: 123}, empty-record: {} } }}

~ COMP nested = {}
{{ nested.record.key-1 }} - {{ nested.record.key-2 }}
```

---

## Context

`context(key)` is a built-in reference mechanism for injecting immutable values at convert time.

Use it with `compono.WithContext`:

```go
type CurrentUser struct {
    FirstName string `compono:"first-name"`
    LastName  string `compono:"last-name"`
}

_, err := c.Convert(source, &buf, compono.WithContext(map[string]any{
    "app/version":   "1.2.0",
    "feature/live":  true,
    "stats/numbers": []int{10, 20, 30},
    "current-user": CurrentUser{
        FirstName: "Yunus",
        LastName:  "Emre",
    },
}))
```

Direct usage:

```
Version: {{ context(app/version) }}
```

It can also be used as:

- a component argument
- a default parameter value
- an array item
- a record value
- built-in component arguments

Example:

```
{{ LINK text=context(link/text) url=context(link/url) new-tab=context(link/new-tab) }}

~ GREETING name=context(current-user/first-name)
Hello **{{ name }}**!
```

If the resolved value is a record or array, you can keep using normal access syntax:

```
# {{ context(current-user).first-name }}
## {{ context(stats/numbers)[1] }}
```

### Context Keys

Keys are static and unquoted. They are made of `segment`s joined by `/`.

- segments may contain lowercase Latin letters, numbers, and `-`
- `/` cannot appear at the beginning or end
- repeated separators are invalid
- whitespace is allowed around the key inside `context(...)`

An empty `context()` or a key that does not follow this format is not a context reference. It is not plain text either: the `{{ }}` unit it is in, or the call whose argument it is, is an `invalid-expression` error and drops ([Invalid and Unclosed Units](#invalid-and-unclosed-units)).

### Supported Go Types

`WithContext` supports:

- `string`
- `bool`
- `int`, `int8`, `int16`, `int32`, `int64`
- `[]T` and `[N]T`
- `map[string]T`
- structs

Notes:

- `nil` is not supported
- map keys must be `string`
- struct fields must be exported
- `compono` struct tags must be valid `kebab-case`
- if a struct field has no `compono` tag, its name is converted to `kebab-case`
- unsupported types such as floats, pointers, functions, and channels return a fatal conversion error
- a map or slice that contains itself is a fatal conversion error (`ErrUnsupportedType`). The same map or slice may appear more than once side by side.
- fatal context injection errors are returned as `ErrUnsupportedType` or `ErrUnsupportedKeyNotation`

### Missing Keys and Errors

If a referenced key is not injected, `Convert` returns a diagnostic with:

- Code: `unknown-key`
- Message: `The key **[key]** is not injected.`

What renders nothing depends on how `context(...)` is used:

- direct usage drops the `{{ context(...) }}` unit
- using it in an argument of a component call drops that call, also when the value reaches the argument through a parameter
- default values are resolved lazily, so no error is produced unless that parameter is actually used; then the `{{ }}` unit that uses it drops

## Attribute Hook

An attribute hook lets an application add HTML attributes to the output of built-in component calls, for example a theme's `class`. It is registered per conversion with `compono.WithAttributeHook`. The hook only returns attributes. Where they are written is defined by each built-in.

```go
_, err := c.Convert(pageSource, w,
	compono.WithGlobalComponent("LAYOUT", layoutSource,
		compono.WithIsolatedScope(),
		compono.WithGlobalComponent("MAIN_MENU", menuSource),
	),
	compono.WithAttributeHook(func(builtin string, chain []compono.Frame) map[string]string {
		if builtin == "LINK" && len(chain) > 0 && chain[len(chain)-1].Name == "MAIN_MENU" {
			return map[string]string{"class": "main-menu-link"}
		}
		return nil
	}),
)
```

For a `LINK` call in the body of `MAIN_MENU`, the hook receives `"LINK"` and this chain:

```go
[]compono.Frame{
	{Name: "LAYOUT", Kind: compono.FrameGlobal, ScopePath: []string{}},
	{Name: "MAIN_MENU", Kind: compono.FrameGlobal, ScopePath: []string{"LAYOUT"}},
}
```

### When the Hook Is Called

- Only for real built-in calls. A call that resolves to a local or global component is not a built-in call, even if its name is a built-in name. Markdown elements never call the hook.
- Once per call, before the built-in's output is written, in render order, on the goroutine running `Convert`.
- Never for a call that an error drops.

### The Chain

The chain is the call stack at the moment the built-in is rendered, ordered from the outermost frame to the innermost.

- Each frame has a `Name` and a `Kind`: `compono.FrameGlobal`, `compono.FrameLocal` or `compono.FrameBuiltin`.
- `ScopePath` is set only for global frames: the names of the global component's owners, outermost first (see [Global Component Scopes](#global-component-scopes)). It is an empty slice for a global in the root scope and `nil` for other kinds.
- The converted source is not a frame. A built-in written directly in the source gets an empty chain.
- The chain does not contain the rendered built-in itself. Its name is the `builtin` argument.
- The chain is dynamic. Content passed through a component parameter belongs to the frame it is rendered in, although its calls are resolved where they are written.
- The slices are only valid during the call. Copy them to keep them.

### Writing Attributes

- A `nil` or empty map writes nothing.
- Values are escaped and quoted. Attributes are written after Compono's own attributes, sorted by name, so the output is deterministic.
- Each attribute goes to the root element of the built-in, unless the built-in names a special point for it:
  - `LINK`: everything goes to `<a>`.
  - `IMAGE`: `sizes` goes to every element that carries `srcset` (each `<source>`), and the other attributes go to the root element. An `IMAGE` without variants has no `<source>`, so `sizes` is not written. This is not an error.
- Built-ins never write `class` or `sizes` themselves, so the hook's value is the only value. There is no merging.

### Hook Errors

`Convert` returns an error when:

- `WithAttributeHook` is given more than once in a conversion, even with `nil` (`ErrAttributeHookAlreadySet`). A `nil` hook given once is the same as no hook.
- it is given to a global component (`ErrConversionOptionInGlobal`),
- the hook returns a name that does not match `[a-z][a-z0-9-]*` (`ErrInvalidAttributeName`),
- the hook returns an attribute that Compono writes on that element itself, such as `href`, `target` or `rel` on `LINK`'s `<a>` (`ErrAttributeConflict`).

Nothing is written to the writer in these cases. The hook cannot return an error. Validating the values is the job of the application that gives the hook.

## URL Reporter

`compono.WithURLReporter` registers a function that receives every URL the browser loads as a resource from the output. In v0.7 these are the URLs of `IMAGE`: the `src` of `<img>` and every candidate of the `srcset` of `<source>`. Links (`href`) are not reported. An application can derive, for example, the `img-src` origins of its Content-Security-Policy from them.

```go
var urls []compono.URLReport
_, err := c.Convert(pageSource, w,
	compono.WithURLReporter(func(r compono.URLReport) {
		urls = append(urls, r)
	}),
)
```

- `URLReport` has `Element` (`img` or `source`), `Attribute` (`src` or `srcset`) and `URL`. A `srcset` candidate is reported without its width descriptor.
- `URL` is the value written to the output after the [URL scheme allowlist](#url-scheme-allowlist) normalization and before HTML escaping. Compono does not extract origins or remove duplicates.
- URLs are reported in output order, synchronously during `Convert`. A URL written twice is reported twice.
- The URLs of a dropped call are not reported.
- `WithURLReporter` may be given once per conversion (`ErrURLReporterAlreadySet`) and not to a global component (`ErrConversionOptionInGlobal`). A `nil` reporter is the same as no reporter. If `Convert` returns an error, the reported URLs should not be used.

## Error Handling

There are two kinds of errors:

- **Fatal errors:** unexpected failures and invalid options. `Convert` returns them as `error` and writes no output.
- **Diagnostics:** syntax and logic errors in the source. An error drops the part of the output it belongs to: that part renders nothing. The rest of the output is written and valid, and `Convert` returns a `compono.Diagnostic` for the dropped part.

Compono never writes an error into the output. Who sees an error, and how, is up to the application.

```go
diagnostics, err := c.Convert(source, &buf)
if err != nil {
    return err
}
for _, d := range diagnostics {
    fmt.Printf("%v %d:%d %s: %s\n", d.Source, d.Range.Start.Line, d.Range.Start.Column, d.Code, d.Message)
}
```

### Diagnostics

A diagnostic has these fields:

- `Code`: the kind of the error, the kebab-case form of its title, such as `unknown-component`. Codes are exported as constants (`compono.CodeUnknownComponent`) and are never renamed. Several situations may share a code; the message tells them apart. Programs should look at the code, not the message.
- `Message`: an English description. Values are emphasized with `**`. It is not HTML; escaping it is the job of whoever shows it.
- `Source`: the source the dropped part is written in. It is empty for the converted source. For a global component it is its scope path ending with its own name: `[LAYOUT CARD]` for the sub component `CARD` of `LAYOUT`.
- `Range`: the `[Start, End)` range of the dropped part in `Source`. Each end has a 0-based byte `Offset`, a 1-based `Line` and a 1-based `Column` counted in runes. Positions are relative to the source as given: comment lines, a global's parameter line and leading blank lines are counted.
- `Calls`: the component calls around the dropped part, outermost first, each with its `Name`, `Kind` (`builtin`, `global` or `local`), `Source` and `Range`. It is empty when the dropped part is in the converted source itself.

### What Is Dropped

An error drops the smallest part it belongs to:

- **A component call** (built-in, global or local) for its own errors: an unknown component, argument errors, built-in specific errors, an invalid parameter definition of the called component and an infinite call. An error of an argument's value also drops the call the argument belongs to.
- **A `{{ }}` unit** (a parameter reference or `context(key)`) for its own errors, and a unit that follows no grammar. An unclosed `{{` drops only itself.
- **A markdown link** for an error in its address and for a component call in its text or address, defined or not. The call is not rendered, so its own errors are not reported: `[{{ UNDEFINED }}](/z)` gives only `invalid-component-usage`. An error of a `{{ }}` unit in the link text drops only that unit: `[Docs {{ x }}](/docs)` renders `<a href="/docs">Docs </a>` when `x` cannot be used.

The paragraph, heading or component body around a dropped part is still rendered. For example, `Hello {{ FOO }} world` with an undefined `FOO` renders `<p>Hello  world</p>`.

An error that depends on the value of a parameter is found where the value is used, separately for every render. A `{{ }}` unit with an index out of range, an unknown record key, an array or record used directly, or a block component used inline drops only itself, in the component where it is written. A call whose argument comes from a parameter, such as an `IMAGE` whose `media` is passed down or a call that forwards a value of the wrong type, is checked the same way and drops where it is written. The same unit can drop in one call and render in another:

```
{{ ITEM list = [1] }}

{{ ITEM list = [] }}

~ ITEM list = []
Item: {{ list[0] }}
```

renders `<p>Item: 1</p><p>Item: </p>` and returns one `array-index-out-of-range` diagnostic whose `Calls` holds the second `ITEM` call.

A call that enters a component already being rendered with the same signature never ends. The signature is the called definition and the definitions its Component parameters resolve to, each in the scope the value is written in: a local `CARD` of the converted source and a local `CARD` of a global are different values. Such a call is an `infinite-component-call` and drops where it is written, so in a cycle only the call that closes it drops. A component that calls itself with other component values is not a loop.

Bound arguments and String, Integer, Bool, Array and Record parameters are not part of the signature. A value can grow on every entry, as `c` does in `{{ A c = B(x = c) }}`, and the chain would never see the same signature twice. Since the definitions are finite, so are the signatures, and every chain that never ends is stopped. The cost is that some chains that would end are dropped too.

### Rules

- `Convert` returns one diagnostic for each dropped part. A global component called 10 times with an error inside returns 10 diagnostics, each with its own `Calls`; grouping them is up to the application.
- A part with several errors returns one diagnostic.
- An error that drops nothing returns no diagnostic, such as one inside a component that is never called.
- Diagnostics are returned in the order of the dropped parts in the output. The same source and options always return the same diagnostics in the same order.
- A diagnostic is not a failure: the output is written and valid. `error` is only for fatal errors, and then no diagnostics are returned.

### General Errors

These errors belong to the language itself, not to a built-in or a Markdown element. The errors of built-ins and Markdown elements are described with them.

| Code | When | Message |
| --- | --- | --- |
| `unknown-component` | A called name, or a name given as a value, resolves to no local, global or built-in component. This includes the default value of a Component parameter when it is used | `The component **[name]** is not defined.` |
| `unknown-parameter` | A reference to a parameter the component does not define, or an argument in a call or a binding | `The parameter **[name]** is not defined for this component.` |
| `wrong-argument-type` | The type of an argument does not match the type of the parameter. For built-ins the type comes from their parameter table | `The parameter **[name]** has the wrong type.` |
| `missing-argument` | A required parameter gets no argument | `The parameter **[name]** of component **[component]** is required.` |
| `duplicate-argument` | The caller gives an argument to a parameter that is already bound | `The parameter **[name]** of component **[component]** is already bound.` |
| `invalid-parameter-definition` | A parameter definition does not follow the format | `The parameter definition **[text]** of component **[component]** is invalid.` |
| `invalid-built-in-arguments` | The arguments of a built-in do not match its schema: a required argument is missing or the shape of a record argument does not match | `The parameter **[name]** does not match the schema of the built-in component **[component]**.` |
| `invalid-component-usage` | A block component is used inline, directly or through a parameter reference | `The component **[name]** is a block component and cannot be used inline.` |
| `infinite-component-call` | A call enters a component that is already being rendered with the same signature | `The call to component **[name]** creates an infinite loop and was skipped.` |
| `not-component-parameter` | A parameter called with arguments (`{{ title x = 1 }}`) is not a Component parameter | `The parameter **[name]** is not a component parameter and cannot be called.` |
| `invalid-parameter-usage` | A parameter reference in the root of the converted source, which is not a component body and defines no parameters | `The parameter **[name]** cannot be used outside a component.` |
| `invalid-parameter-usage` | An array or record parameter rendered without access (`{{ items }}`) | `The parameter **[name]** is an array and cannot be rendered directly.` or `… is a record …` |
| `invalid-parameter-access` | An index on a parameter that is not an array, or a key on one that is not a record | `The parameter **[name]** is not an array and cannot be indexed.` or `The parameter **[name]** is not a record and cannot be accessed by key.` |
| `array-index-out-of-range` | The index is past the end of the array | `The index used for parameter **[name]** is out of range.` or `The index used for context value **[key]** is out of range.` |
| `unknown-record-key` | The record has no such key | `The key **[key]** is not defined in this record.` |
| `unknown-key` | The key of `context(key)` is not injected | `The key **[key]** is not injected.` |
| `invalid-context-usage` | An array or record context value rendered without access | `The context value **[key]** is an array and cannot be rendered directly.` or `… is a record …` |
| `invalid-context-access` | An index on a context value that is not an array, or a key on one that is not a record | `The context value **[key]** is not an array and cannot be indexed.` or `The context value **[key]** is not a record and cannot be accessed by key.` |
| `invalid-expression` | A `{{ }}` unit follows no grammar ([Invalid and Unclosed Units](#invalid-and-unclosed-units)) | `The content between **{{** and **}}** must be a component call, a parameter reference or a context reference.` |
| `unclosed-expression` | A `{{` has no matching `}}` | `This **{{** has no matching **}}**.` |
| `invalid-value` | The source given to [`ParseValue`](#value-parsing) is not a string, integer, bool, array or record | `The value must be a string, integer, bool, array or record.` |

- When the same error occurs for several names in one part, one diagnostic is returned and the message is plural, such as `The components **A**, **B** are not defined.`
- An error of a value that comes from a parameter has an `invalid-parameter-*` code, and one that comes from `context(key)` has an `invalid-context-*` code.
- `[name]`, `[component]`, `[key]` and `[text]` are as written in the source. Messages are not part of the API; codes are.

## API Reference

### Core Methods

```go
// Create a new Compono instance
c := compono.New()

// Convert source to HTML and get the diagnostics of the conversion
diagnostics, err := c.Convert(source []byte, writer io.Writer, opts ...compono.ConvertOption)

// Give a global component for a single conversion
_, err := c.Convert(source, writer, compono.WithGlobalComponent(name, globalSource))

// Give a global component its own sub components and isolate its scope
_, err := c.Convert(source, writer, compono.WithGlobalComponent(name, globalSource,
    compono.WithIsolatedScope(),
    compono.WithGlobalComponent(subName, subSource),
))

// Inject convert-time context values
_, err := c.Convert(source, writer, compono.WithContext(map[string]any{
    "app/version": "1.2.0",
}))

// Add attributes to built-in component calls (once per conversion)
_, err := c.Convert(source, writer, compono.WithAttributeHook(func(builtin string, chain []compono.Frame) map[string]string {
    return nil
}))

// Receive the image URLs written to the output (once per conversion)
_, err := c.Convert(source, writer, compono.WithURLReporter(func(r compono.URLReport) {}))
```

### Concurrency

A `Compono` value is stateless. `Convert` depends only on the source and the options, and it changes nothing in the `Compono` value. Everything that belongs to a conversion (global components, context, attribute hook) is given to `Convert` as an option and applies only to that conversion. One `Compono` value can be shared and its `Convert` called from several goroutines at once, with different options. The configuration does not change after `New`.

### Analysis

`Analyze` parses a source without rendering it and returns facts about the source itself. It takes no global components and no context, and its `error` is only for fatal errors. It is a package-level function and can be called from several goroutines at once.

```go
analysis, err := compono.Analyze(source)

compono.Builtins()         // "LINK", "IMAGE", "HEADER", …
compono.MarkdownElements() // "blockquote", "br", "code-block", …
```

A source that starts with a parameter line (`~ title = ""`) is analyzed as a global component, any other source as a converted source. An `Analysis` holds:

- `Params`: the parameter line of a global source, `nil` for a converted source. A definition that does not follow the format is left out and reported in `Diagnostics`.
- `Locals`: the local definitions, each with its name, parameters and range.
- `Refs`: the references to components. `Kind` is `call` for a `{{ X … }}` call, `value` for a component given as a value (an argument, a bound argument, an array element or a record value; `X(c = Y(a = 1))` gives two) and `default` for a component written in a parameter definition. `In` is the local the reference is in, `Local` tells whether the name resolves to a local of the source, `Args` are the names of the arguments given or bound and `Block` tells whether a call is written as a block. Whether a name that resolves to no local is a global or a built-in is up to the caller.
- `Elements`: the Markdown elements used, with the names of `MarkdownElements`. A heading of any level is `heading` and a line break in a paragraph is `br`.
- `Root`: the units of the source outside its local definitions. A block call is a `call` unit with the index of its reference in `Refs`. A paragraph is split into its inline calls, its inline Markdown elements (`markdown`) and its `text`; any other block is `markdown`, and a `{{ }}` unit that follows no grammar is `text`. Comment lines, the parameter line, line breaks and whitespace are not units.
- `Diagnostics`: the errors that do not depend on values, in the whole source, also in locals that are never called. An invalid parameter definition is reported at the definition, and also at every call to it. Since no globals and no context are given, `unknown-component` and `unknown-key` are not reported. `Source` and `Calls` are empty.

Ranges are in the format of diagnostic ranges, relative to the source, with leading and trailing whitespace left out. Every list is in source order, and the same source always gives the same `Analysis`. `Builtins` and `MarkdownElements` return a new slice on every call.

### Value Parsing

`ParseValue` parses the source of a value, written as the value of an argument, and returns it as a Go value that `WithContext` accepts. It lets an application keep values that its users write in Compono's syntax and give them as context.

```go
value, diagnostics := compono.ParseValue([]byte(`[ { text: "Home", url: "/" } ]`))
// value: []any{map[string]any{"text": "Home", "url": "/"}}

_, err := c.Convert(page, w, compono.WithContext(map[string]any{"site/menu": value}))
```

- The syntax is the syntax of an argument value: a string (with the `""` escape), an integer, a bool, an array or a record ([Supported Types](#supported-types)). Leading and trailing spaces, tabs and line breaks are ignored.
- Only plain values. A component (a name or `NO_MATTER`), a parameter reference and `context(...)` are `invalid-value` errors, at the top and nested.
- The Go types are `string`, `int`, `bool`, `[]any` and `map[string]any`.
- Giving the result to `WithContext` renders the same output and the same diagnostics as writing the source as an argument.
- Errors are diagnostics. A grammar error has the code it has in an argument value (`invalid-expression` and others), and a source that is not a value is `invalid-value`. Ranges are relative to the given source. When there are diagnostics the value is `nil`.
- It is a package-level function and can be called from several goroutines at once.

## Component Naming Convention

Component names must be in `SCREAMING_SNAKE_CASE`:

- ✓ `HEADER`
- ✓ `USER_PROFILE`
- ✓ `NAV_MENU_ITEM`
- ✗ `header`
- ✗ `userProfile`

## Parameter Naming Convention

Parameter names must be in `kebab-case`:

- ✓ `name`
- ✓ `user-name`
- ✓ `is-active`
- ✗ `userName`
- ✗ `user_name`

## Component Override Behavior

When multiple components share the same name, Compono follows a clear override hierarchy:

```
Local Component > Global Component > Built-in Component
```

**Local always wins:**

```
{{ LINK }}

~ LINK
I override the built-in LINK component!
```

This outputs `<p>I override the built-in LINK component!</p>` instead of an anchor tag.

**Global overrides built-in:**

```go
c.Convert(source, &buf, compono.WithGlobalComponent("LINK", []byte(`Custom link behavior`)))
```

Now all `{{ LINK }}` calls in that conversion will use your global definition instead of the built-in one.

This allows you to customize or extend built-in components without modifying the library.

Inside a global component with sub components, the sub components come between the local components and the outer global components. See [Global Component Scopes](#global-component-scopes).

## License

MIT License - see [LICENSE](LICENSE) for details.
