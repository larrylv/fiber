// ⚡️ Fiber is an Express inspired web framework written in Go with ☕️
// 🤖 Github Repository: https://github.com/gofiber/fiber
// 📌 API Documentation: https://docs.gofiber.io
// ⚠️ This path parser was inspired by ucarion/urlpath (MIT License).
// 💖 Maintained and modified for Fiber by @renewerner87

package fiber

import (
	"strings"

	utils "github.com/gofiber/utils"
)

// routeSegment holds the segment metadata
type routeSegment struct {
	Param      string
	Const      string
	IsParam    bool
	IsOptional bool
	IsLast     bool
	EndChar    byte
}

// RouteNode is a radix tree node that holds metadata for a section of a
// registered route. Route sections are separated by `/`. `App` should hold a
// pointer of the root RouteNode, and use it to find the handler given a url.
type RouteNode struct {
	pathPretty     string                // Create a stripped path in-case sensitive / trailing slashes
	Path           string                // Original registered route path
	MethodHandlers [][]Handler           // key is the http method
	ChildrenNodes  map[string]*RouteNode // key is the full section
	Segments       []routeSegment        // Segments stored the route parts separated by hyphen `-` and colon `:`. TODO(larrylv): add regexp support
	Params         []string              // Case sensitive param keys
}

var routeSegmentDelimiter = ".-"

func (app *App) findHandlers(path string, methodInt int) []Handler {
	pathSections := strings.SplitAfter(path, "/")
	currentNode := app.rootRouteNode
	if path == "/" && currentNode.pathPretty == "" { // it's the root node
		handlers := currentNode.MethodHandlers[methodInt]
		return handlers
	}

	// If the path has a trailing slash, the last part will be an empty string,
	// and we should just get rid of it.
	if len(pathSections) > 1 && pathSections[len(pathSections)-1] == "" {
		pathSections = pathSections[:len(pathSections)-1]
	}

	return currentNode.findHandlers(
		pathSections,
		path,
		methodInt,
		app.Settings.CaseSensitive,
		app.Settings.StrictRouting,
	)
}

func (app *App) buildRouteNode(method, path string, handlers ...Handler) {
	pathSections := strings.SplitAfter(path, "/")
	// The first section should always be `/` since callers should preappend a
	// `/` to the path if the path doesn't start with `/`. This check makes sure
	// we tolerate if callers make a mistake, and does nothing but return.
	if len(pathSections) == 0 || pathSections[0] != "/" {
		return
	}

	buildChildRouteNode(
		app.rootRouteNode,
		pathSections[1:],
		method,
		app.Settings.CaseSensitive,
		app.Settings.StrictRouting,
		handlers...,
	)
}

// buildChildRouteNode tries to build the child node for the passed parent node.
// If the child node is nil, it adds the handlers to parent node. Otherwise, it
// saves the child node on the `ChildrennNodes` field on parent node.
func buildChildRouteNode(parentNode *RouteNode, pathSections []string, method string, isCaseSensitive bool, isStrictRouting bool, handlers ...Handler) *RouteNode {
	// When there is only an empty string in the `pathSections`, that means we
	// reach the end of the url path, and the last parent node is the leaf node.
	if len(pathSections) == 0 || (len(pathSections) == 1 && pathSections[0] == "") {
		addHandlersToNode(parentNode, method, handlers...)
		return nil
	}

	sectionRaw := pathSections[0]
	sectionPretty := sectionRaw
	// Case sensitive routing, all to lowercase
	if !isCaseSensitive {
		sectionPretty = utils.ToLower(sectionPretty)
	}
	// We should remove trailing slashes when the current section is not `/`, and
	// either of the conditions below is true:
	// 1. this is not the last path section
	// 2. this is the last path section, and strict routing is disabled.
	if len(sectionPretty) > 1 {
		if len(pathSections) > 1 || !isStrictRouting {
			sectionPretty = utils.TrimRight(sectionPretty, '/')
		}
	}

	currentRouteNode, ok := parentNode.ChildrenNodes[sectionPretty]
	// If there isn't a RouteNode for the current section, we create one
	if !ok {
		currentRouteNode = &RouteNode{
			pathPretty:    sectionPretty,
			Path:          sectionRaw,
			ChildrenNodes: make(map[string]*RouteNode),
		}
		currentRouteNode.build()
		parentNode.ChildrenNodes[sectionPretty] = currentRouteNode
	}

	buildChildRouteNode(
		currentRouteNode,
		pathSections[1:],
		method,
		isCaseSensitive,
		isStrictRouting,
		handlers...,
	)

	return currentRouteNode
}

func addHandlersToNode(node *RouteNode, method string, handlers ...Handler) {
	if node.MethodHandlers == nil {
		node.MethodHandlers = make([][]Handler, len(intMethod))
	}

	mIndex := methodInt(method)
	node.MethodHandlers[mIndex] = append(
		node.MethodHandlers[mIndex],
		handlers...,
	)
}

func (node *RouteNode) findHandlers(pathSections []string, path string, methodInt int, isCaseSensitive, isStrictRouting bool) []Handler {
	if len(pathSections) == 0 {
		return nil
	}

	currentSection := pathSections[0]
	isLastSection := len(pathSections) == 1
	if node.match(currentSection, isLastSection, isCaseSensitive, isStrictRouting) {
		if isLastSection {
			if node.MethodHandlers == nil {
				return nil
			}
			return node.MethodHandlers[methodInt]
		}

		for _, childNode := range node.ChildrenNodes {
			handlers := childNode.findHandlers(pathSections[1:], path, methodInt, isCaseSensitive, isStrictRouting)
			if handlers != nil {
				return handlers
			}
		}
	} else {
		return nil
	}

	return nil
}

func (node *RouteNode) match(s string, isLastSection, isCaseSensitive, isStrictRouting bool) bool {
	if !isCaseSensitive {
		s = utils.ToLower(s)
	}

	if isLastSection {
		if !isStrictRouting {
			return node.pathPretty == s || node.pathPretty+"/" == s
		} else {
			return node.pathPretty == s
		}
	}

	// pathPretty of non-last section doesn't have trailing slash
	return node.pathPretty+"/" == s
}

func (node *RouteNode) build() {
	pattern := node.pathPretty
	part, delimiterPos := "", findNextRouteSegmentDelimiter(pattern)

	// If the initial `delimiterPos` is `-1`, we could avoid storing it in the
	// `Segments` since we could just use `node.pathPretty` directly
	for len(pattern) > 0 && delimiterPos != -1 {
		if delimiterPos != -1 {
			part = pattern[:delimiterPos]
		} else {
			part = pattern
		}

		partLen, lastSeg := len(part), len(node.Segments)-1
		if partLen == 0 { // skip empty parts
			if len(pattern) > 0 {
				// remove first char
				pattern = pattern[1:]
			}
			continue
		}
		// is parameter ?
		if part[0] == '*' || part[0] == ':' {
			node.Segments = append(node.Segments, routeSegment{
				Param:      utils.GetTrimmedParam(part),
				IsParam:    true,
				IsOptional: part == wildcardParam || part[partLen-1] == '?',
			})
			lastSeg = len(node.Segments) - 1
			node.Params = append(node.Params, node.Segments[lastSeg].Param)
			// combine const segments
		} else if lastSeg >= 0 && !node.Segments[lastSeg].IsParam {
			node.Segments[lastSeg].Const += string(node.Segments[lastSeg].EndChar) + part
			// create new const segment
		} else {
			node.Segments = append(node.Segments, routeSegment{
				Const: part,
			})
			lastSeg = len(node.Segments) - 1
		}

		if delimiterPos != -1 && len(pattern) >= delimiterPos+1 {
			node.Segments[lastSeg].EndChar = pattern[delimiterPos]
			pattern = pattern[delimiterPos+1:]
		}

		delimiterPos = findNextRouteSegmentDelimiter(pattern)
	}
	if len(node.Segments) > 0 {
		node.Segments[len(node.Segments)-1].IsLast = true
	}
}

// findNextRouteSegmentDelimiter searches in the route for the next end position for a segment
func findNextRouteSegmentDelimiter(search string) int {
	return strings.IndexAny(search, routeSegmentDelimiter)
}
