# Compono

Compono is a **platform-agnostic**, component-based domain-specific language (DSL) that extends Markdown syntax with reusable components.

Originally developed for [Umono CMS](https://github.com/umono-cms/umono), Compono can be used in any Go project that needs a flexible templating solution.

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
    if err := c.Convert(source, &buf); err != nil {
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

Code blocks are also supported:

~~~
```go
fmt.Println("Hello")
```
~~~

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

### Global Components

Global components can be registered once and used across multiple conversions:

```go
c := compono.New()

// Register a global component
c.RegisterGlobalComponent("FOOTER", []byte(`© 2026 My Company`))

// Use it in any conversion
c.Convert([]byte(`
# Page Title
Content here...
{{ FOOTER }}
`), &buf)
```

Global components can also have parameters:

```go
c.RegisterGlobalComponent("BLOG_PAGE", []byte(`title="" content=""
## {{ title }}
{{ content }}`))
```

## Built-in Components

### LINK

Creates an anchor element with optional target blank:

```
{{ LINK text="Visit us" url="https://example.com" new-tab=true }}
```

Output:
```html
<compono-link><a href="https://example.com" target="_blank" rel="noopener noreferrer">Visit us</a></compono-link>
```

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
<compono-image><picture><source type="image/avif" srcset="https://cdn.example.com/my-photo-640.avif 640w, https://cdn.example.com/my-photo-1280.avif 1280w, https://cdn.example.com/my-photo-1600.avif 1600w"><source type="image/webp" srcset="https://cdn.example.com/my-photo-640.webp 640w, https://cdn.example.com/my-photo-1280.webp 1280w, https://cdn.example.com/my-photo-1600.webp 1600w"><img src="https://cdn.example.com/my-photo.jpg" alt="My photo" width="1600" height="900"></picture></compono-image>
```

Without variants, `IMAGE` renders a plain `img` element inside the `compono-image` custom element:

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
<compono-image><img src="https://cdn.example.com/avatar.png" alt="Profile avatar" width="512" height="512"></compono-image>
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
- `alt` is a string. Pass `alt=""` for decorative images.

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
- HTML output is always wrapped with the `compono-image` custom element.
- variants are grouped by first-seen `mime-type`, preserving the original group order.
- within each mime type group, `srcset` entries are sorted by ascending width.
- an empty `variants` array is valid and renders only the fallback `img` inside `compono-image`.
- all widths and heights must be greater than `0`.
- all variants must preserve the same aspect ratio as the main `media`.
- duplicate `mime-type` + `width` pairs are invalid.

When validation fails, Compono renders an error placeholder instead of silently producing invalid markup. Common IMAGE-specific errors include:

- `Invalid built-in arguments`
- `Unsupported mime-type`
- `Invalid dimension`
- `Duplicate variant`
- `Inconsistent aspect ratio`

### WEB_GRID

Creates a web grid wrapper from component items and grid template definitions:

```
{{ WEB_GRID
  items = [
    { component: HEADER, grid-area: "header" },
    { component: CONTENT, grid-area: "content" },
    { component: FOOTER, grid-area: "footer" }
  ]
  grid-template-columns = ["1fr"]
  grid-template-rows = ["min-content", "1fr", "min-content"]
  grid-template-areas = [
    ["header"],
    ["content"],
    ["footer"]
  ]
}}

~ HEADER
# Header

~ CONTENT
Main content.

~ FOOTER
Footer
```

Output:
```html
<compono-web-grid data-grid-template-columns="1fr" data-grid-template-rows="min-content 1fr min-content" data-grid-template-areas='[["header"],["content"],["footer"]]'><compono-web-grid-item data-grid-area="header"><h1>Header</h1></compono-web-grid-item><compono-web-grid-item data-grid-area="content"><p>Main content.</p></compono-web-grid-item><compono-web-grid-item data-grid-area="footer"><p>Footer</p></compono-web-grid-item></compono-web-grid>
```

`WEB_GRID` also supports responsive breakpoint variants for the grid template parameters:
`sm-grid-template-columns`, `md-grid-template-columns`, `lg-grid-template-columns`, `xl-grid-template-columns`, `xxl-grid-template-columns`, and the corresponding `*-grid-template-rows` / `*-grid-template-areas` parameters.

### NAVIGATION

Creates a platform navigation tree from an `items` array. In the HTML renderer, it outputs a `<compono-navigation>` wrapper containing a semantic `nav` list. Compono does not define the custom element; that belongs to the application or runtime using the generated output.

Basic usage:

```
{{ NAVIGATION
  items = [
    { label: "Home", target: "/" },
    { label: "About", target: "/about" },
    { label: "Contact", target: "/contact" }
  ]
}}
```

Output:
```html
<compono-navigation><nav><ul><li><a href="/">Home</a></li><li><a href="/about">About</a></li><li><a href="/contact">Contact</a></li></ul></nav></compono-navigation>
```

Items can contain nested `children` arrays:

```
{{ NAVIGATION
  items = [
    { label: "Home", target: "/" },
    {
      label: "Docs",
      target: "/docs",
      children: [
        { label: "Guide", target: "/docs/guide" },
        {
          label: "API",
          target: "/docs/api",
          children: [
            { label: "Components", target: "/docs/api/components" }
          ]
        }
      ]
    }
  ]
}}
```

Output:
```html
<compono-navigation><nav><ul><li><a href="/">Home</a></li><li><a href="/docs">Docs</a><ul><li><a href="/docs/guide">Guide</a></li><li><a href="/docs/api">API</a><ul><li><a href="/docs/api/components">Components</a></li></ul></li></ul></li></ul></nav></compono-navigation>
```

#### NAVIGATION Parameters

- `items` is an array of navigation item records.
- each item must include:
  - `label`: string rendered as the visible link text.
  - `target`: string rendered as the link `href`.
  - optional `children`: array of navigation item records.

`NAVIGATION` is a block built-in component. It can receive `items` from a parameter or `context(key)`, and label/target values are HTML escaped by the HTML renderer.

## Parameters

Components can accept parameters. Each parameter must have a **default value** defined in the component definition.

If a parameter value is not provided during the call, the **default value is used**.

```
{{ SAY_HELLO name="Jane" }}

~ SAY_HELLO name="John"
# Hello, {{ name }}!
```

### Supported Types

Supported parameter types:

- **String** → `name = "John"`
- **Integer** → `age = 25`
- **Bool** → `active = true`
- **Component** → `comp = COMP`
- **Array** → `items = ["Jane", 22, true, COMP]`
- **Record** → `config = { lang: "tr", for-admin: true }`

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

### Global Parameter Visibility in Local Components

When a **global component** defines parameters, those parameters are **visible to local components inside it**.

```
c.RegisterGlobalComponent("PROFILE_PAGE", []byte(`
name="Guest"

{{ PROFILE_CARD }}

~ PROFILE_CARD
## {{ name }}
Welcome to the profile page.
`))
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

The local component `PROFILE_CARD` can directly access the global parameter `name`.

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

err := c.Convert(source, &buf, compono.WithContext(map[string]any{
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

If `context()` is empty or the key format is invalid, it is treated as plain text instead of a context reference.

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
- fatal context injection errors are returned as `ErrUnsupportedType` or `ErrUnsupportedKeyNotation`

### Missing Keys and Errors

If a referenced key is not injected, Compono renders an error placeholder with:

- Title: `Unknown key`
- Message: `The key **[key]** is not injected.`

Error placement depends on how `context(...)` is used:

- direct usage always renders an inline error
- using it in a block component call renders a block error at the call site
- using it in an inline component call renders an inline error at the call site
- default values are resolved lazily, so no error is produced unless that parameter is actually used

## Error Handling

Compono provides error feedback by rendering placeholders where errors occur.
Fatal errors during conversion stop the process and no output is produced.

## API Reference

### Core Methods

```go
// Create a new Compono instance
c := compono.New()

// Convert source to HTML
err := c.Convert(source []byte, writer io.Writer, opts ...compono.ConvertOption)

// Register a global component
err := c.RegisterGlobalComponent(name string, source []byte)

// Unregister a global component
err := c.UnregisterGlobalComponent(name string)

// Inject a global component for a single conversion
err := c.Convert(source, writer, compono.WithGlobalComponent(name, globalSource))

// Inject convert-time context values
err := c.Convert(source, writer, compono.WithContext(map[string]any{
    "app/version": "1.2.0",
}))
```

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
c.RegisterGlobalComponent("LINK", []byte(`Custom link behavior`))
```

Now all `{{ LINK }}` calls will use your global definition instead of the built-in one.

This allows you to customize or extend built-in components without modifying the library.

## License

MIT License - see [LICENSE](LICENSE) for details.
