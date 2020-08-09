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

// RouteNode is a radix tree node that holds metadata for a segment of a
// registered route. Route segments are separated by `/`. `App` should hold a
// pointer of the root RouteNode, and use it to find the handler given a url.
type RouteNode struct {
	pathPretty     string                // Create a stripped path in-case sensitive / trailing slashes
	Path           string                // Original registered route path
	MethodHandlers [][]Handler           // key is the http method
	ChildrenNodes  map[string]*RouteNode // key is the full segment
	Segments       []routeSegment        // Each route has its own segments, separated by hyphen `-` and colon `:`. TODO(larrylv): add regexp support
	Params         []string              // Case sensitive param keys
}

var routeSegmentDelimiter = ".-/"

func (app *App) findHandlers(path string, methodInt int) []Handler {
	segments := strings.SplitAfter(path, "/")
	currentNode := app.rootRouteNode
	if path == "/" && currentNode.pathPretty == "" { // it's the root node
		handlers := currentNode.MethodHandlers[methodInt]
		return handlers
	}

	return currentNode.findHandlers(segments, path, methodInt)
}

func (app *App) buildRouteNode(method, path string, handlers ...Handler) {
	segments := strings.SplitAfter(path, "/")
	// The first segment should always be `/` since callers should preappend a
	// `/` to the path if the path doesn't start with `/`. This check makes sure
	// we tolerate if callers make a mistake, and does nothing but return.
	if len(segments) == 0 || segments[0] != "/" {
		return
	}

	buildChildRouteNode(
		app.rootRouteNode,
		segments[1:],
		method,
		app.Settings.CaseSensitive,
		app.Settings.StrictRouting,
		handlers...,
	)
}

// buildChildRouteNode tries to build the child node for the passed parent node.
// If the child node is nil, it adds the handlers to parent node. Otherwise, it
// saves the child node on the `ChildrennNodes` field on parent node.
func buildChildRouteNode(parentNode *RouteNode, pathSegments []string, method string, isCaseSensitive bool, isStrictRouting bool, handlers ...Handler) *RouteNode {
	// When there is only an empty string in the `pathSegments`, that means we
	// reach the end of the url path, and the last node is the leaf node.
	if len(pathSegments) == 0 || (len(pathSegments) == 1 && pathSegments[0] == "") {
		addHandlersToNode(parentNode, method, handlers...)
		return nil
	}

	pathRaw := pathSegments[0]
	pathPretty := pathRaw
	// Case sensitive routing, all to lowercase
	if !isCaseSensitive {
		pathPretty = utils.ToLower(pathPretty)
	}
	// We should remove trailing slashes when the current segment is not `/`, and
	// either of the conditions below is true:
	// 1. this is not the last path segment
	// 2. this is the last path segment, and strict routing is disabled.
	if len(pathPretty) > 1 {
		if len(pathSegments) > 1 || !isStrictRouting {
			pathPretty = utils.TrimRight(pathPretty, '/')
		}
	}

	currentRouteNode, ok := parentNode.ChildrenNodes[pathPretty]
	// there isn't a current RouteNode for `pathSegments[0]`, so let's create one
	if !ok {
		currentRouteNode = &RouteNode{
			pathPretty:    pathPretty,
			Path:          pathRaw,
			ChildrenNodes: make(map[string]*RouteNode),
		}
		currentRouteNode.build()
	}

	childNode := buildChildRouteNode(
		currentRouteNode,
		pathSegments[1:],
		method,
		isCaseSensitive,
		isStrictRouting,
		handlers...,
	)

	if childNode != nil {
		parentNode.ChildrenNodes[pathPretty] = childNode
	}

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

func (node *RouteNode) findHandlers(segments []string, path string, methodInt int) []Handler {
	if len(segments) == 0 {
		return nil
	}

	currentSegment := segments[0]
	if node.match(currentSegment) {
		if len(segments) == 1 {
			return node.MethodHandlers[methodInt]
		}

		for _, childNode := range node.ChildrenNodes {
			handlers := childNode.findHandlers(segments[1:], path, methodInt)
			if handlers != nil {
				return handlers
			}
		}
	} else {
		return nil
	}

	return nil
}

func (node *RouteNode) match(s string) bool {
	if node.pathPretty == s {
		return true
	}

	return false
}

func (node *RouteNode) build() {
	pattern := node.Path
	part, delimiterPos := "", 0

	for len(pattern) > 0 && delimiterPos != -1 {
		delimiterPos = findNextRouteSegmentDelimiter(pattern)
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
	}
	if len(node.Segments) > 0 {
		node.Segments[len(node.Segments)-1].IsLast = true
	}
}

// findNextRouteSegmentDelimiter searches in the route for the next end position for a segment
func findNextRouteSegmentDelimiter(search string) int {
	return strings.IndexAny(search, routeSegmentDelimiter)
}
