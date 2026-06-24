package builtin

import (
	"regexp"
	"strings"

	"github.com/umono-cms/compono/ast"
)

var webGridAreaPattern = regexp.MustCompile(`^[a-z]+(?:-[a-z0-9]+)*$`)

func BuiltinComponents() []Definition {
	return []Definition{
		{
			Name: "LINK",
			Params: []Param{
				{
					Name:         "text",
					Schema:       String(),
					DefaultValue: "",
				},
				{
					Name:         "url",
					Schema:       String(),
					DefaultValue: "",
				},
				{
					Name:         "new-tab",
					Schema:       Bool(),
					DefaultValue: false,
				},
			},
			InlineRenderable: true,
		},
		{
			Name: "IMAGE",
			Params: []Param{
				{
					Name:         "media",
					Schema:       imageMediaSchema(),
					DefaultValue: map[string]any{},
					IsRequired:   true,
				},
				{
					Name:         "alt",
					Schema:       String(),
					DefaultValue: "",
				},
			},
			InlineRenderable: true,
		},
		// Deprecated: WEB_GRID is deprecated and will be removed in v1.
		// Compono will now only carry semantic content. Built-in components
		// that tell the presentation layer what to do will no longer be added to Compono.
		{
			Name: "WEB_GRID",
			Params: []Param{
				{
					Name:         "items",
					Schema:       ArrayOf(webGridItemSchema()).Min(1),
					DefaultValue: []any{},
					IsRequired:   true,
					Diagnostic:   webGridItemsDiagnostic,
				},
				{
					Name:         "grid-template-columns",
					Schema:       ArrayOf(String()).Min(1),
					DefaultValue: []any{},
					IsRequired:   true,
					Diagnostic:   webGridTemplateColumnsDiagnostic,
				},
				{
					Name:         "grid-template-rows",
					Schema:       ArrayOf(String()).Min(1),
					DefaultValue: []any{},
					IsRequired:   true,
					Diagnostic:   webGridTemplateRowsDiagnostic,
				},
				{
					Name:         "grid-template-areas",
					Schema:       ArrayOf(ArrayOf(String()).Min(1)).Min(1),
					DefaultValue: []any{},
					IsRequired:   true,
					Diagnostic:   webGridTemplateAreasDiagnostic,
				},
				{
					Name:         "sm-grid-template-columns",
					Schema:       ArrayOf(String()).Min(1),
					DefaultValue: []any{},
					Diagnostic:   webGridTemplateColumnsDiagnostic,
				},
				{
					Name:         "md-grid-template-columns",
					Schema:       ArrayOf(String()).Min(1),
					DefaultValue: []any{},
					Diagnostic:   webGridTemplateColumnsDiagnostic,
				},
				{
					Name:         "lg-grid-template-columns",
					Schema:       ArrayOf(String()).Min(1),
					DefaultValue: []any{},
					Diagnostic:   webGridTemplateColumnsDiagnostic,
				},
				{
					Name:         "xl-grid-template-columns",
					Schema:       ArrayOf(String()).Min(1),
					DefaultValue: []any{},
					Diagnostic:   webGridTemplateColumnsDiagnostic,
				},
				{
					Name:         "xxl-grid-template-columns",
					Schema:       ArrayOf(String()).Min(1),
					DefaultValue: []any{},
					Diagnostic:   webGridTemplateColumnsDiagnostic,
				},
				{
					Name:         "sm-grid-template-rows",
					Schema:       ArrayOf(String()).Min(1),
					DefaultValue: []any{},
					Diagnostic:   webGridTemplateRowsDiagnostic,
				},
				{
					Name:         "md-grid-template-rows",
					Schema:       ArrayOf(String()).Min(1),
					DefaultValue: []any{},
					Diagnostic:   webGridTemplateRowsDiagnostic,
				},
				{
					Name:         "lg-grid-template-rows",
					Schema:       ArrayOf(String()).Min(1),
					DefaultValue: []any{},
					Diagnostic:   webGridTemplateRowsDiagnostic,
				},
				{
					Name:         "xl-grid-template-rows",
					Schema:       ArrayOf(String()).Min(1),
					DefaultValue: []any{},
					Diagnostic:   webGridTemplateRowsDiagnostic,
				},
				{
					Name:         "xxl-grid-template-rows",
					Schema:       ArrayOf(String()).Min(1),
					DefaultValue: []any{},
					Diagnostic:   webGridTemplateRowsDiagnostic,
				},
				{
					Name:         "sm-grid-template-areas",
					Schema:       ArrayOf(ArrayOf(String()).Min(1)).Min(1),
					DefaultValue: []any{},
					Diagnostic:   webGridTemplateAreasDiagnostic,
				},
				{
					Name:         "md-grid-template-areas",
					Schema:       ArrayOf(ArrayOf(String()).Min(1)).Min(1),
					DefaultValue: []any{},
					Diagnostic:   webGridTemplateAreasDiagnostic,
				},
				{
					Name:         "lg-grid-template-areas",
					Schema:       ArrayOf(ArrayOf(String()).Min(1)).Min(1),
					DefaultValue: []any{},
					Diagnostic:   webGridTemplateAreasDiagnostic,
				},
				{
					Name:         "xl-grid-template-areas",
					Schema:       ArrayOf(ArrayOf(String()).Min(1)).Min(1),
					DefaultValue: []any{},
					Diagnostic:   webGridTemplateAreasDiagnostic,
				},
				{
					Name:         "xxl-grid-template-areas",
					Schema:       ArrayOf(ArrayOf(String()).Min(1)).Min(1),
					DefaultValue: []any{},
					Diagnostic:   webGridTemplateAreasDiagnostic,
				},
			},
		},
		{
			Name: "NAVIGATION",
			Params: []Param{
				{
					Name:         "items",
					Schema:       ArrayOf(navigationItemSchema()).Min(1),
					DefaultValue: []any{},
					IsRequired:   true,
				},
			},
		},
	}
}

func webGridItemsDiagnostic(name string, value ast.ResolvedValue) (ValidationDiagnostic, bool) {
	if value.Type != "array" || len(value.Items) > 0 {
		return ValidationDiagnostic{}, false
	}
	return ValidationDiagnostic{
		Title:   "Empty items",
		Message: "The parameter **" + name + "** cannot be an empty array.",
	}, true
}

func webGridTemplateColumnsDiagnostic(name string, value ast.ResolvedValue) (ValidationDiagnostic, bool) {
	if value.Type != "array" || len(value.Items) > 0 {
		return ValidationDiagnostic{}, false
	}
	return ValidationDiagnostic{
		Title:   "Empty grid template columns",
		Message: "The parameter **" + name + "** cannot be an empty array.",
	}, true
}

func webGridTemplateRowsDiagnostic(name string, value ast.ResolvedValue) (ValidationDiagnostic, bool) {
	if value.Type != "array" || len(value.Items) > 0 {
		return ValidationDiagnostic{}, false
	}
	return ValidationDiagnostic{
		Title:   "Empty grid template rows",
		Message: "The parameter **" + name + "** cannot be an empty array.",
	}, true
}

func webGridTemplateAreasDiagnostic(name string, value ast.ResolvedValue) (ValidationDiagnostic, bool) {
	if value.Type != "array" {
		return ValidationDiagnostic{}, false
	}
	if len(value.Items) == 0 {
		return ValidationDiagnostic{
			Title:   "Empty grid template area",
			Message: "The parameter **" + name + "** cannot be empty.",
		}, true
	}
	for _, row := range value.Items {
		if row.Type == "array" && len(row.Items) == 0 {
			return ValidationDiagnostic{
				Title:   "Empty grid template area",
				Message: "The parameter **" + name + "** cannot be empty.",
			}, true
		}
	}
	return ValidationDiagnostic{}, false
}

func imageMediaSchema() ValueSchema {
	return Record(
		Field("url", String()).Required(),
		Field("width", Integer()).Required(),
		Field("height", Integer()).Required(),
		Field("mime-type", String()).Required(),
		Field("variants", ArrayOf(imageVariantSchema())),
	).DisallowUnknownKeys()
}

func imageVariantSchema() ValueSchema {
	return Record(
		Field("url", String()).Required(),
		Field("width", Integer()).Required(),
		Field("height", Integer()).Required(),
		Field("mime-type", String()).Required(),
	).DisallowUnknownKeys()
}

func webGridItemSchema() ValueSchema {
	return Record(
		Field("component", Component()).Required(),
		Field("grid-area", String().Matching(func(value ast.ResolvedValue) bool {
			raw := strings.TrimSpace(value.Raw)
			if raw == "" {
				return false
			}
			if !webGridAreaPattern.MatchString(raw) {
				return false
			}
			for _, ch := range raw {
				if ch < '0' || ch > '9' {
					return true
				}
			}
			return false
		})).Required(),
	).DisallowUnknownKeys()
}

func navigationItemSchema() ValueSchema {
	return Record(
		Field("label", String()).Required(),
		Field("target", String()).Required(),
		Field("children", ArrayOf(Lazy(navigationItemSchema))),
	).DisallowUnknownKeys()
}
