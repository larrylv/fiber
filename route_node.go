// ⚡️ Fiber is an Express inspired web framework written in Go with ☕️
// 🤖 Github Repository: https://github.com/gofiber/fiber
// 📌 API Documentation: https://docs.gofiber.io
// ⚠️ This path parser was inspired by ucarion/urlpath (MIT License).
// 💖 Maintained and modified for Fiber by @renewerner87

package fiber

import (
	"fmt"
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
	Segments       []routeSegment        // Segments are the route path separated by hyphen `-` and colon `:`. TODO(larrylv): add regexp support
	Params         []string              // Case sensitive param keys
}

var routeSegmentDelimiter = ".-"

func (app *App) findHandlers(path string, methodInt int) []Handler {
	currentNode := app.rootRouteNode
	if path == "/" {
		handlers := currentNode.MethodHandlers[methodInt]
		return handlers
	}

	pathSections := strings.SplitAfter(path, "/")
	// If the path has a trailing slash, the last section will be an empty
	// string, and we get rid of it since it's useless for routing.
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
	// `/` to the path if the path doesn't start with `/`.
	if len(pathSections) == 0 || pathSections[0] != "/" {
		panic(fmt.Sprintf("buildRouteNode: invalid path %s, should start with a /\n", path))
	}
	// If the path has a trailing slash, the last section will be an empty
	// string, and we get rid of it since it's useless for routing.
	if len(pathSections) > 1 && pathSections[len(pathSections)-1] == "" {
		pathSections = pathSections[:len(pathSections)-1]
	}
	methodInt := methodInt(method)
	if methodInt == -1 {
		panic(fmt.Sprintf("buildRouteNode: invalid http method %s\n", method))
	}

	buildChildRouteNode(
		app.rootRouteNode,
		pathSections[1:],
		methodInt,
		app.Settings.CaseSensitive,
		app.Settings.StrictRouting,
		handlers...,
	)
}

// buildChildRouteNode tries to build the child node and its children nodes given
// a parentNode and all path sections. If path sections are empty, that means the
// passed parentNode is the leaf node that should holder the handlers.
func buildChildRouteNode(parentNode *RouteNode, pathSections []string, methodInt int, isCaseSensitive bool, isStrictRouting bool, handlers ...Handler) *RouteNode {
	if len(pathSections) == 0 {
		addHandlersToNode(parentNode, methodInt, handlers...)
		return nil
	}

	sectionRaw := pathSections[0]
	sectionPretty := sectionRaw
	// Case sensitive routing, all to lowercase
	if !isCaseSensitive {
		sectionPretty = utils.ToLower(sectionPretty)
	}
	// We should remove trailing slashes when the either of the conditions below
	// is true:
	// 1. this is not the last path section, strict routing enabled or not.
	// 2. this is the last path section, and strict routing is disabled.
	if len(sectionPretty) > 1 {
		if len(pathSections) > 1 || !isStrictRouting {
			sectionPretty = utils.TrimRight(sectionPretty, '/')
		}
	}

	// Use `sectionPretty` as the key for children nodes, so that we always use
	// the same node for routes with the same `sectionPretty`.
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
		methodInt,
		isCaseSensitive,
		isStrictRouting,
		handlers...,
	)

	return currentRouteNode
}

func addHandlersToNode(node *RouteNode, methodInt int, handlers ...Handler) {
	// lazy initialize `MethodHandlers` since some node might not have any
	// handlers associated with.
	if node.MethodHandlers == nil {
		node.MethodHandlers = make([][]Handler, len(intMethod))
	}

	node.MethodHandlers[methodInt] = append(
		node.MethodHandlers[methodInt],
		handlers...,
	)
}

// findHandlers returns the handlers on a node that matches `pathSections`.
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

		// loop children nodes, and see there are matches.
		// TODO(larrylv): add index support
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

	// If the initial `delimiterPos` is `-1`, we could avoid storing anything in
	// `node.Segments` since `match` will just use `node.pathPretty` directly.
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
