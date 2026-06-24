// Deprecated: WEB_GRID is deprecated and will be removed in v1.
// Compono will now only carry semantic content. Built-in components
// that tell the presentation layer what to do will no longer be added to Compono.
package errwrap

import (
	"regexp"
	"strings"

	"github.com/umono-cms/compono/ast"
)

var (
	webGridBreakpoints   = []string{"sm", "md", "lg", "xl", "xxl"}
	webGridReservedAreas = []string{"span", "auto", "inherit", "initial", "unset", "revert", "revert-layer"}
	webGridSizePattern   = regexp.MustCompile(`^[1-9][0-9]*fr$`)
	webGridMinmaxPattern = regexp.MustCompile(`^minmax\((min-content|max-content|[1-9][0-9]*fr),(min-content|max-content|[1-9][0-9]*fr)\)$`)
)

type webGridTemplate struct {
	breakpoint string
	explicit   bool
	columns    ast.ResolvedValue
	rows       ast.ResolvedValue
	areas      ast.ResolvedValue
}

type webGridError struct {
	title   string
	message string
}

func invalidWebGrid() conditionAnalyzer {
	return conditionAnalyzer{
		conditions: []func(*wrapContext, ast.Node) bool{
			isRuleNameOneOf("block-comp-call", "inline-comp-call"),
			not(isInsideCompDef()),
			isWebGridWithSpecificError(),
		},
		title: func(ctx *wrapContext, node ast.Node) string {
			return getWebGridError(ctx, node).title
		},
		message: func(ctx *wrapContext, node ast.Node) string {
			return getWebGridError(ctx, node).message
		},
		block: blockFromRuleName,
	}
}

func unknownWebGridItemComponent() conditionAnalyzer {
	return conditionAnalyzer{
		conditions: []func(*wrapContext, ast.Node) bool{
			isRuleNameOneOf("block-comp-call", "inline-comp-call"),
			not(isInsideCompDef()),
			hasUnknownWebGridItemComponent(),
		},
		title: func(ctx *wrapContext, node ast.Node) string {
			return getUnknownWebGridItemComponentError(ctx, node).title
		},
		message: func(ctx *wrapContext, node ast.Node) string {
			return getUnknownWebGridItemComponentError(ctx, node).message
		},
		block: blockFromRuleName,
	}
}

func isInsideCompDef() func(*wrapContext, ast.Node) bool {
	return func(_ *wrapContext, node ast.Node) bool {
		return findEnclosingCompDef(node) != nil
	}
}

func isWebGridWithSpecificError() func(*wrapContext, ast.Node) bool {
	return func(ctx *wrapContext, node ast.Node) bool {
		return getWebGridError(ctx, node).title != ""
	}
}

func hasUnknownWebGridItemComponent() func(*wrapContext, ast.Node) bool {
	return func(ctx *wrapContext, node ast.Node) bool {
		return getUnknownWebGridItemComponentError(ctx, node).title != ""
	}
}

func getWebGridError(ctx *wrapContext, node ast.Node) webGridError {
	if !ast.IsRuleNameOneOf(node, []string{"block-comp-call", "inline-comp-call"}) {
		return webGridError{}
	}

	return walkWebGridCallTree(ctx, node, func(target ast.Node) webGridError {
		return getWebGridErrorForCompCalls(ctx, node, target)
	})
}

func getUnknownWebGridItemComponentError(ctx *wrapContext, node ast.Node) webGridError {
	if !ast.IsRuleNameOneOf(node, []string{"block-comp-call", "inline-comp-call"}) {
		return webGridError{}
	}

	return walkWebGridCallTree(ctx, node, func(target ast.Node) webGridError {
		items := resolveWebGridArg(ctx, node, target, "items")
		return getWebGridUnknownComponentError(ctx, target, items)
	})
}

func walkWebGridCallTree(ctx *wrapContext, ownerCompCall ast.Node, visit func(ast.Node) webGridError) webGridError {
	seen := map[ast.Node]bool{}

	var walk func(ast.Node) webGridError
	walk = func(current ast.Node) webGridError {
		if seen[current] {
			return webGridError{}
		}
		seen[current] = true

		if getCompCallNameStr(current) == "WEB_GRID" {
			if err := visit(current); err.title != "" {
				return err
			}
		}

		compName := getCompCallNameStr(current)
		if compName == "" {
			return webGridError{}
		}

		compDef := findCompDef(ctx.root, current, compName)
		if compDef == nil {
			return webGridError{}
		}

		compDefContent := getCompDefContent(compDef)
		if compDefContent == nil {
			return webGridError{}
		}

		for _, nested := range ast.FilterNodesInTree(compDefContent, func(child ast.Node) bool {
			return ast.IsRuleNameOneOf(child, []string{"block-comp-call", "inline-comp-call"})
		}) {
			if err := walk(nested); err.title != "" {
				return err
			}
		}

		return webGridError{}
	}

	_ = ownerCompCall
	return walk(ownerCompCall)
}

func getWebGridErrorForCompCalls(ctx *wrapContext, ownerCompCall ast.Node, targetCompCall ast.Node) webGridError {
	if getCompCallNameStr(targetCompCall) != "WEB_GRID" {
		return webGridError{}
	}

	targetCompDef := findCompDef(ctx.root, targetCompCall, "WEB_GRID")
	if targetCompDef == nil || !ast.IsRuleName(targetCompDef, "builtin-comp") {
		return webGridError{}
	}

	items := resolveWebGridArg(ctx, ownerCompCall, targetCompCall, "items")
	if err := getWebGridItemsRecordError(items); err.title != "" {
		return err
	}
	if err := getWebGridUniqueAreasError(items); err.title != "" {
		return err
	}
	if err := getWebGridReservedAreaError(items); err.title != "" {
		return err
	}
	if ast.FindNode(ast.GetAncestors(targetCompCall), func(anc ast.Node) bool {
		return ast.IsRuleName(anc, "global-comp-def")
	}) != nil {
		if err := getWebGridUnknownComponentError(ctx, targetCompCall, items); err.title != "" {
			return err
		}
	}
	if err := getWebGridItemImageError(ctx, ownerCompCall, targetCompCall, items); err.title != "" {
		return err
	}
	if err := getWebGridBreakpointError(targetCompCall); err.title != "" {
		return err
	}

	if diagnostic := firstBuiltinSchemaMismatchDiagnostic(getBuiltinSchemaMismatchesForCompCall(ctx, ownerCompCall, targetCompCall)); diagnostic.Title != "" {
		return webGridError{}
	}

	templates := resolveWebGridTemplates(ctx, ownerCompCall, targetCompCall)

	if err := getWebGridUnsupportedSizeError(templates); err.title != "" {
		return err
	}
	if err := getWebGridUnknownAreaError(items, templates); err.title != "" {
		return err
	}
	if err := getWebGridUnmatchedRowsError(templates); err.title != "" {
		return err
	}
	if err := getWebGridUnmatchedColumnsError(templates); err.title != "" {
		return err
	}
	if err := getWebGridInvalidShapeError(templates); err.title != "" {
		return err
	}
	if err := getWebGridUnusedAreaError(items, templates); err.title != "" {
		return err
	}
	return webGridError{}
}

func getWebGridItemImageError(ctx *wrapContext, ownerCompCall ast.Node, targetCompCall ast.Node, items ast.ResolvedValue) webGridError {
	if items.Type != "array" {
		return webGridError{}
	}

	for _, item := range items.Items {
		if item.Type != "record" {
			continue
		}

		component, ok := item.Fields["component"]
		if !ok || component.Type != "comp" || component.Raw == "" {
			continue
		}

		err := imageErrorForComponentTarget(ctx, targetCompCall, imageComponentTarget{
			name:  component.Raw,
			scope: component.Scope,
		}, ast.GetAncestors(ownerCompCall), map[string]bool{})
		if err.title == "" {
			continue
		}

		return webGridError{
			title:   err.title,
			message: err.message,
		}
	}

	return webGridError{}
}

func resolveWebGridArg(ctx *wrapContext, ownerCompCall ast.Node, targetCompCall ast.Node, name string) ast.ResolvedValue {
	arg := ast.GetCompCallArgByParamName(ast.GetCompCallArgsFromCompCall(targetCompCall), name)
	if arg != nil {
		invokerAncestors := append([]ast.Node{targetCompCall, ownerCompCall}, ast.GetAncestors(ownerCompCall)...)
		return ast.ResolveCompCallArgValue(ctx.root, arg, invokerAncestors, targetCompCall)
	}
	return ast.ResolveParamDefaultFromCompCall(ctx.root, targetCompCall, name)
}

func resolveWebGridTemplates(ctx *wrapContext, ownerCompCall ast.Node, targetCompCall ast.Node) []webGridTemplate {
	templates := []webGridTemplate{
		{
			explicit: true,
			columns:  resolveWebGridArg(ctx, ownerCompCall, targetCompCall, "grid-template-columns"),
			rows:     resolveWebGridArg(ctx, ownerCompCall, targetCompCall, "grid-template-rows"),
			areas:    resolveWebGridArg(ctx, ownerCompCall, targetCompCall, "grid-template-areas"),
		},
	}

	for _, breakpoint := range webGridBreakpoints {
		hasColumns := webGridHasExplicitArg(targetCompCall, breakpoint+"-grid-template-columns")
		hasRows := webGridHasExplicitArg(targetCompCall, breakpoint+"-grid-template-rows")
		hasAreas := webGridHasExplicitArg(targetCompCall, breakpoint+"-grid-template-areas")
		templates = append(templates, webGridTemplate{
			breakpoint: breakpoint,
			explicit:   hasColumns || hasRows || hasAreas,
			columns:    resolveWebGridArg(ctx, ownerCompCall, targetCompCall, breakpoint+"-grid-template-columns"),
			rows:       resolveWebGridArg(ctx, ownerCompCall, targetCompCall, breakpoint+"-grid-template-rows"),
			areas:      resolveWebGridArg(ctx, ownerCompCall, targetCompCall, breakpoint+"-grid-template-areas"),
		})
	}

	return templates
}

func getWebGridItemsRecordError(items ast.ResolvedValue) webGridError {
	if items.Type != "array" {
		return webGridError{}
	}

	for _, item := range items.Items {
		if item.Type != "record" {
			continue
		}

		for key := range item.Fields {
			if key != "component" && key != "grid-area" {
				return webGridError{
					title:   "Unknown key",
					message: "The key **" + key + "** is not defined for this record.",
				}
			}
		}

		if value, ok := item.Fields["component"]; ok && value.Type != "comp" {
			return webGridError{
				title:   "Wrong value type",
				message: "The value of **component** has the wrong type.",
			}
		}
		if value, ok := item.Fields["grid-area"]; ok && value.Type != "string" {
			return webGridError{
				title:   "Wrong value type",
				message: "The value of **grid-area** has the wrong type.",
			}
		}
	}

	return webGridError{}
}

func getWebGridUniqueAreasError(items ast.ResolvedValue) webGridError {
	seen := map[string]bool{}
	for _, area := range getWebGridItemAreas(items) {
		if seen[area] {
			return webGridError{
				title:   "Grid areas must be unique",
				message: "The grid area **" + area + "** is used more than once in **items**.",
			}
		}
		seen[area] = true
	}
	return webGridError{}
}

func getWebGridReservedAreaError(items ast.ResolvedValue) webGridError {
	for _, area := range getWebGridItemAreas(items) {
		for _, reserved := range webGridReservedAreas {
			if area != reserved {
				continue
			}
			return webGridError{
				title:   "Reserved CSS keyword",
				message: "The grid area **" + area + "** uses a reserved CSS keyword and cannot be used.",
			}
		}
	}
	return webGridError{}
}

func getWebGridUnknownComponentError(ctx *wrapContext, targetCompCall ast.Node, items ast.ResolvedValue) webGridError {
	if items.Type != "array" {
		return webGridError{}
	}

	for _, item := range items.Items {
		if item.Type != "record" {
			continue
		}

		component, ok := item.Fields["component"]
		if !ok || component.Type != "comp" || component.Raw == "" {
			continue
		}

		if findWebGridItemComponentDef(ctx.root, targetCompCall, component.Raw, component.Scope) != nil {
			continue
		}

		return webGridError{
			title:   "Unknown component",
			message: "The component **" + component.Raw + "** is not defined or not registered.",
		}
	}

	return webGridError{}
}

func findWebGridItemComponentDef(root ast.Node, targetCompCall ast.Node, name string, scope ast.Node) ast.Node {
	if name == "" {
		return nil
	}

	localCompDefSrc := scope
	if localCompDefSrc == nil {
		localCompDefSrc = ast.GetLocalCompSourceFromNode(targetCompCall, root)
	}

	localCompDef := ast.FindLocalCompDef(localCompDefSrc, name)
	if localCompDef != nil {
		return localCompDef
	}

	currentGlobalCompDef := ast.FindNode(ast.GetAncestors(targetCompCall), func(anc ast.Node) bool {
		return ast.IsRuleName(anc, "global-comp-def")
	})
	if currentGlobalCompDef != nil && currentGlobalCompDef != localCompDefSrc {
		localCompDef = ast.FindLocalCompDef(currentGlobalCompDef, name)
		if localCompDef != nil {
			return localCompDef
		}
	}

	globalCompDef := ast.FindGlobalCompDef(root, name)
	if globalCompDef != nil {
		return globalCompDef
	}

	return ast.FindBuiltinCompDef(root, name)
}

func getWebGridBreakpointError(node ast.Node) webGridError {
	for _, breakpoint := range webGridBreakpoints {
		hasColumns := webGridHasExplicitArg(node, breakpoint+"-grid-template-columns")
		hasRows := webGridHasExplicitArg(node, breakpoint+"-grid-template-rows")
		hasAreas := webGridHasExplicitArg(node, breakpoint+"-grid-template-areas")
		if !hasColumns && !hasRows && !hasAreas {
			continue
		}

		if hasColumns && hasRows && hasAreas {
			continue
		}

		return webGridError{
			title:   "Missing breakpoint grid template parameters",
			message: "The breakpoint **" + breakpoint + "** must define **grid-template-columns**, **grid-template-rows**, and **grid-template-areas** together.",
		}
	}

	return webGridError{}
}

func getWebGridUnsupportedSizeError(templates []webGridTemplate) webGridError {
	for _, template := range templates {
		if !template.explicit {
			continue
		}
		for _, value := range template.columns.Items {
			if !isSupportedWebGridSize(value.Raw) {
				return webGridError{
					title:   "Unsupported size unit",
					message: "The value **" + value.Raw + "** uses an unsupported size unit.",
				}
			}
		}
		for _, value := range template.rows.Items {
			if !isSupportedWebGridSize(value.Raw) {
				return webGridError{
					title:   "Unsupported size unit",
					message: "The value **" + value.Raw + "** uses an unsupported size unit.",
				}
			}
		}
	}

	return webGridError{}
}

func getWebGridUnknownAreaError(items ast.ResolvedValue, templates []webGridTemplate) webGridError {
	known := map[string]bool{}
	for _, area := range getWebGridItemAreas(items) {
		known[area] = true
	}

	for _, template := range templates {
		if !template.explicit {
			continue
		}
		for _, row := range template.areas.Items {
			for _, area := range row.Items {
				if area.Raw == "." || known[area.Raw] {
					continue
				}
				return webGridError{
					title:   "Unknown grid area",
					message: "The grid area **" + area.Raw + "** is used in **" + webGridTemplateParamName(template.breakpoint, "grid-template-areas") + "** but is not defined in **items**.",
				}
			}
		}
	}

	return webGridError{}
}

func getWebGridUnmatchedRowsError(templates []webGridTemplate) webGridError {
	for _, template := range templates {
		if !template.explicit {
			continue
		}
		if len(template.areas.Items) != len(template.rows.Items) {
			return webGridError{
				title:   "Unmatched rows",
				message: "The number of rows in **" + webGridTemplateParamName(template.breakpoint, "grid-template-areas") + "** does not match **" + webGridTemplateParamName(template.breakpoint, "grid-template-rows") + "**.",
			}
		}
	}

	return webGridError{}
}

func getWebGridUnmatchedColumnsError(templates []webGridTemplate) webGridError {
	for _, template := range templates {
		if !template.explicit {
			continue
		}
		for _, row := range template.areas.Items {
			if len(row.Items) != len(template.columns.Items) {
				return webGridError{
					title:   "Unmatched columns",
					message: "The number of columns in **" + webGridTemplateParamName(template.breakpoint, "grid-template-areas") + "** does not match **" + webGridTemplateParamName(template.breakpoint, "grid-template-columns") + "**.",
				}
			}
		}
	}

	return webGridError{}
}

func getWebGridInvalidShapeError(templates []webGridTemplate) webGridError {
	for _, template := range templates {
		if !template.explicit {
			continue
		}
		areas := map[string][][2]int{}
		for rowIdx, row := range template.areas.Items {
			for colIdx, area := range row.Items {
				if area.Raw == "." {
					continue
				}
				areas[area.Raw] = append(areas[area.Raw], [2]int{rowIdx, colIdx})
			}
		}

		for area, points := range areas {
			minRow, maxRow := points[0][0], points[0][0]
			minCol, maxCol := points[0][1], points[0][1]
			pointSet := map[[2]int]bool{}
			for _, point := range points {
				pointSet[point] = true
				if point[0] < minRow {
					minRow = point[0]
				}
				if point[0] > maxRow {
					maxRow = point[0]
				}
				if point[1] < minCol {
					minCol = point[1]
				}
				if point[1] > maxCol {
					maxCol = point[1]
				}
			}

			rectSize := (maxRow - minRow + 1) * (maxCol - minCol + 1)
			if rectSize != len(points) {
				if hasSeparatedWebGridIslands(pointSet) {
					return webGridError{
						title:   "Multiple grid area shapes",
						message: "The grid area **" + area + "** in **" + webGridTemplateParamName(template.breakpoint, "grid-template-areas") + "** creates multiple separate shapes.",
					}
				}
				return webGridError{
					title:   "Invalid grid area shape",
					message: "The grid area **" + area + "** in **" + webGridTemplateParamName(template.breakpoint, "grid-template-areas") + "** must form a rectangle.",
				}
			}
		}
	}

	return webGridError{}
}

func getWebGridUnusedAreaError(items ast.ResolvedValue, templates []webGridTemplate) webGridError {
	used := map[string]bool{}
	for _, template := range templates {
		if !template.explicit {
			continue
		}
		for _, row := range template.areas.Items {
			for _, area := range row.Items {
				if area.Raw != "." {
					used[area.Raw] = true
				}
			}
		}
	}

	for _, area := range getWebGridItemAreas(items) {
		if used[area] {
			continue
		}
		return webGridError{
			title:   "Unused grid area",
			message: "The grid area **" + area + "** is defined in **items** but is not used in any grid template areas.",
		}
	}

	return webGridError{}
}

func getWebGridItemAreas(items ast.ResolvedValue) []string {
	areas := []string{}
	if items.Type != "array" {
		return areas
	}

	for _, item := range items.Items {
		if item.Type != "record" {
			continue
		}
		area, ok := item.Fields["grid-area"]
		if !ok || area.Type != "string" || area.Raw == "" {
			continue
		}
		areas = append(areas, area.Raw)
	}
	return areas
}

func hasSeparatedWebGridIslands(points map[[2]int]bool) bool {
	if len(points) == 0 {
		return false
	}

	queue := [][2]int{}
	visited := map[[2]int]bool{}
	for point := range points {
		queue = append(queue, point)
		visited[point] = true
		break
	}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		neighbors := [][2]int{
			{current[0] - 1, current[1]},
			{current[0] + 1, current[1]},
			{current[0], current[1] - 1},
			{current[0], current[1] + 1},
		}

		for _, neighbor := range neighbors {
			if !points[neighbor] || visited[neighbor] {
				continue
			}
			visited[neighbor] = true
			queue = append(queue, neighbor)
		}
	}

	return len(visited) != len(points)
}

func isSupportedWebGridSize(value string) bool {
	value = strings.TrimSpace(value)
	if value == "min-content" || value == "max-content" {
		return true
	}
	if webGridSizePattern.MatchString(value) {
		return true
	}
	return webGridMinmaxPattern.MatchString(value)
}

func webGridHasExplicitArg(compCall ast.Node, name string) bool {
	return ast.GetCompCallArgByParamName(ast.GetCompCallArgsFromCompCall(compCall), name) != nil
}

func webGridTemplateParamName(breakpoint string, suffix string) string {
	if breakpoint == "" {
		return suffix
	}
	return breakpoint + "-" + suffix
}
