// ⚡️ Fiber is an Express inspired web framework written in Go with ☕️
// 📝 Github Repository: https://github.com/gofiber/fiber
// 📌 API Documentation: https://docs.gofiber.io

package fiber

import (
	"fmt"
	"testing"

	utils "github.com/gofiber/utils"
)

type testparams struct {
	url          string
	method       string
	params       []string
	match        bool
	partialCheck bool
}

// go test -race -run Test_RouteNode_matchParams
func Test_RouteNode_matchParams(t *testing.T) {
	t.Parallel()
	testCase := func(r string, method string, cases []testparams) {
		app := New()
		app.buildRouteNode(method, r, func(ctx *Ctx) {})
		for _, c := range cases {
			caseMethod := c.method
			if caseMethod == "" {
				caseMethod = method
			}
			handlers := app.findHandlers(c.url, methodInt(caseMethod))
			utils.AssertEqual(t, c.match, nil != handlers, fmt.Sprintf("route: '%s', url: '%s', method: '%s'", r, c.url, caseMethod))
		}
	}
	// testCase("/api/v1/:param/*", MethodGet, []testparams{
	// 	{url: "/api/v1/entity", params: []string{"entity", ""}, match: true},
	// 	{url: "/api/v1/entity/", params: []string{"entity", ""}, match: true},
	// 	{url: "/api/v1/entity/1", params: []string{"entity", "1"}, match: true},
	// 	{url: "/api/v", params: nil, match: false},
	// 	{url: "/api/v2", params: nil, match: false},
	// 	{url: "/api/v1/", params: nil, match: false},
	// })
	// testCase("/api/v1/:param?", MethodGet, []testparams{
	// 	{url: "/api/v1", params: []string{""}, match: true},
	// 	{url: "/api/v1/", params: []string{""}, match: true},
	// 	{url: "/api/v1/optional", params: []string{"optional"}, match: true},
	// 	{url: "/api/v", params: nil, match: false},
	// 	{url: "/api/v2", params: nil, match: false},
	// 	{url: "/api/xyz", params: nil, match: false},
	// })
	// testCase("/api/v1/*", MethodGet, []testparams{
	// 	{url: "/api/v1", params: []string{""}, match: true},
	// 	{url: "/api/v1/", params: []string{""}, match: true},
	// 	{url: "/api/v1/entity", params: []string{"entity"}, match: true},
	// 	{url: "/api/v1/entity/1/2", params: []string{"entity/1/2"}, match: true},
	// 	{url: "/api/v1/Entity/1/2", params: []string{"Entity/1/2"}, match: true},
	// 	{url: "/api/v", params: nil, match: false},
	// 	{url: "/api/v2", params: nil, match: false},
	// 	{url: "/api/abc", params: nil, match: false},
	// })
	// testCase("/api/v1/:param", MethodGet, []testparams{
	// 	{url: "/api/v1/entity", params: []string{"entity"}, match: true},
	// 	{url: "/api/v1/entity/8728382", params: nil, match: false},
	// 	{url: "/api/v1", params: nil, match: false},
	// 	{url: "/api/v1/", params: nil, match: false},
	// })
	// testCase("/api/v1/:param-:param2", MethodGet, []testparams{
	// 	{url: "/api/v1/entity-entity2", params: []string{"entity", "entity2"}, match: true},
	// 	{url: "/api/v1/entity/8728382", params: nil, match: false},
	// 	{url: "/api/v1/entity-8728382", params: []string{"entity", "8728382"}, match: true},
	// 	{url: "/api/v1", params: nil, match: false},
	// 	{url: "/api/v1/", params: nil, match: false},
	// })
	// testCase("/api/v1/:filename.:extension", MethodGet, []testparams{
	// 	{url: "/api/v1/test.pdf", params: []string{"test", "pdf"}, match: true},
	// 	{url: "/api/v1/test/pdf", params: nil, match: false},
	// 	{url: "/api/v1/test-pdf", params: nil, match: false},
	// 	{url: "/api/v1/test_pdf", params: nil, match: false},
	// 	{url: "/api/v1", params: nil, match: false},
	// 	{url: "/api/v1/", params: nil, match: false},
	// })
	testCase("/api/v1/const", MethodGet, []testparams{
		{url: "/api/v1/const", params: []string{}, match: true},
		{url: "/api/v1/const/", params: []string{}, match: true},
		{url: "/api/v1", params: nil, match: false},
		{url: "/api/v1/", params: nil, match: false},
		{url: "/api/v1/something", params: nil, match: false},
	})
	testCase("/api/v1/const-route", MethodGet, []testparams{
		{url: "/api/v1/const-route", params: []string{}, match: true},
		{url: "/api/v1/const-route/", params: []string{}, match: true},
		{url: "/api/v1/const", params: nil, match: false},
		{url: "/api/v1/const-/", params: nil, match: false},
		{url: "/api/v1/const-", params: nil, match: false},
		{url: "/api/v1/something", params: nil, match: false},
	})
	// testCase("/api/v1/:param/abc/*", MethodGet, []testparams{
	// 	{url: "/api/v1/well/abc/wildcard", params: []string{"well", "wildcard"}, match: true},
	// 	{url: "/api/v1/well/abc/", params: []string{"well", ""}, match: true},
	// 	{url: "/api/v1/well/abc", params: []string{"well", ""}, match: true},
	// 	{url: "/api/v1/well/ttt", params: nil, match: false},
	// })
	// testCase("/api/:day/:month?/:year?", MethodGet, []testparams{
	// 	{url: "/api/1", params: []string{"1", "", ""}, match: true},
	// 	{url: "/api/1/", params: []string{"1", "", ""}, match: true},
	// 	{url: "/api/1//", params: []string{"1", "", ""}, match: true},
	// 	{url: "/api/1/-/", params: []string{"1", "-", ""}, match: true},
	// 	{url: "/api/1-", params: []string{"1-", "", ""}, match: true},
	// 	{url: "/api/1.", params: []string{"1.", "", ""}, match: true},
	// 	{url: "/api/1/2", params: []string{"1", "2", ""}, match: true},
	// 	{url: "/api/1/2/3", params: []string{"1", "2", "3"}, match: true},
	// 	{url: "/api/", params: nil, match: false},
	// })
	// testCase("/api/:day.:month?.:year?", MethodGet, []testparams{
	// 	{url: "/api/1", params: []string{"1", "", ""}, match: true},
	// 	{url: "/api/1/", params: nil, match: false},
	// 	{url: "/api/1.", params: []string{"1", "", ""}, match: true},
	// 	{url: "/api/1.2", params: []string{"1", "2", ""}, match: true},
	// 	{url: "/api/1.2.3", params: []string{"1", "2", "3"}, match: true},
	// 	{url: "/api/", params: nil, match: false},
	// })
	// testCase("/api/:day-:month?-:year?", MethodGet, []testparams{
	// 	{url: "/api/1", params: []string{"1", "", ""}, match: true},
	// 	{url: "/api/1/", params: nil, match: false},
	// 	{url: "/api/1-", params: []string{"1", "", ""}, match: true},
	// 	{url: "/api/1-/", params: nil, match: false},
	// 	{url: "/api/1-/-", params: nil, match: false},
	// 	{url: "/api/1-2", params: []string{"1", "2", ""}, match: true},
	// 	{url: "/api/1-2-3", params: []string{"1", "2", "3"}, match: true},
	// 	{url: "/api/", params: nil, match: false},
	// })
	// testCase("/api/*", MethodGet, []testparams{
	// 	{url: "/api/", params: []string{""}, match: true},
	// 	{url: "/api/joker", params: []string{"joker"}, match: true},
	// 	{url: "/api", params: []string{""}, match: true},
	// 	{url: "/api/v1/entity", params: []string{"v1/entity"}, match: true},
	// 	{url: "/api2/v1/entity", params: nil, match: false},
	// 	{url: "/api_ignore/v1/entity", params: nil, match: false},
	// })
	// testCase("/api/*/:param?", MethodGet, []testparams{
	// 	{url: "/api/", params: []string{"", ""}, match: true},
	// 	{url: "/api/joker", params: []string{"joker", ""}, match: true},
	// 	{url: "/api/joker/batman", params: []string{"joker", "batman"}, match: true},
	// 	{url: "/api/joker//batman", params: []string{"joker/", "batman"}, match: true},
	// 	{url: "/api/joker/batman/robin", params: []string{"joker/batman", "robin"}, match: true},
	// 	{url: "/api/joker/batman/robin/1", params: []string{"joker/batman/robin", "1"}, match: true},
	// 	{url: "/api/joker/batman/robin/1/", params: []string{"joker/batman/robin/1", ""}, match: true},
	// 	{url: "/api/joker-batman/robin/1", params: []string{"joker-batman/robin", "1"}, match: true},
	// 	{url: "/api/joker-batman-robin/1", params: []string{"joker-batman-robin", "1"}, match: true},
	// 	{url: "/api/joker-batman-robin-1", params: []string{"joker-batman-robin-1", ""}, match: true},
	// 	{url: "/api", params: []string{"", ""}, match: true},
	// })
	// testCase("/api/*/:param", MethodGet, []testparams{
	// 	{url: "/api/test/abc", params: []string{"test", "abc"}, match: true},
	// 	{url: "/api/joker/batman", params: []string{"joker", "batman"}, match: true},
	// 	{url: "/api/joker/batman/robin", params: []string{"joker/batman", "robin"}, match: true},
	// 	{url: "/api/joker/batman/robin/1", params: []string{"joker/batman/robin", "1"}, match: true},
	// 	{url: "/api/joker/batman-robin/1", params: []string{"joker/batman-robin", "1"}, match: true},
	// 	{url: "/api/joker-batman-robin-1", params: nil, match: false},
	// 	{url: "/api", params: nil, match: false},
	// })
	// testCase("/api/*/:param/:param2", MethodGet, []testparams{
	// 	{url: "/api/test/abc/1", params: []string{"test", "abc", "1"}, match: true},
	// 	{url: "/api/joker/batman", params: nil, match: false},
	// 	{url: "/api/joker/batman/robin", params: []string{"joker", "batman", "robin"}, match: true},
	// 	{url: "/api/joker/batman/robin/1", params: []string{"joker/batman", "robin", "1"}, match: true},
	// 	{url: "/api/joker/batman/robin/2/1", params: []string{"joker/batman/robin", "2", "1"}, match: true},
	// 	{url: "/api/joker/batman-robin/1", params: []string{"joker", "batman-robin", "1"}, match: true},
	// 	{url: "/api/joker-batman-robin-1", params: nil, match: false},
	// 	{url: "/api", params: nil, match: false},
	// })
	// testCase("/partialCheck/foo/bar/:param", MethodGet, []testparams{
	// 	{url: "/partialCheck/foo/bar/test", params: []string{"test"}, match: true, partialCheck: true},
	// 	{url: "/partialCheck/foo/bar/test/test2", params: []string{"test"}, match: true, partialCheck: true},
	// 	{url: "/partialCheck/foo/bar", params: nil, match: false, partialCheck: true},
	// 	{url: "/partiaFoo", params: nil, match: false, partialCheck: true},
	// })
	// testCase("/api/*/:param/:param2", MethodGet, []testparams{
	// 	{url: "/api/test/abc", params: nil, match: false},
	// 	{url: "/api/joker/batman", params: nil, match: false},
	// 	{url: "/api/joker/batman/robin", params: []string{"joker", "batman", "robin"}, match: true},
	// 	{url: "/api/joker/batman/robin/1", params: []string{"joker/batman", "robin", "1"}, match: true},
	// 	{url: "/api/joker/batman/robin/1/2", params: []string{"joker/batman/robin", "1", "2"}, match: true},
	// 	{url: "/api", params: nil, match: false},
	// 	{url: "/api/:test", params: nil, match: false},
	// })
	testCase("/", MethodGet, []testparams{
		{url: "/api", params: nil, match: false},
		{url: "", params: []string{}, match: true},
		{url: "/", params: []string{}, match: true},
	})
	// testCase("/config/abc.json", MethodGet, []testparams{
	// 	{url: "/config/abc.json", params: []string{}, match: true},
	// 	{url: "config/abc.json", params: nil, match: false},
	// 	{url: "/config/efg.json", params: nil, match: false},
	// 	{url: "/config", params: nil, match: false},
	// })
	// testCase("/config/*.json", MethodGet, []testparams{
	// 	{url: "/config/abc.json", params: []string{"abc"}, match: true},
	// 	{url: "/config/efg.json", params: []string{"efg"}, match: true},
	// 	{url: "/config/efg.csv", params: nil, match: false},
	// 	{url: "config/abc.json", params: nil, match: false},
	// 	{url: "/config", params: nil, match: false},
	// })
	testCase("/xyz", MethodGet, []testparams{
		{url: "xyz", params: nil, match: false},
		{url: "xyz/", params: nil, match: false},
	})
}

// go test -race -run Test_RouteNode_StrictRouting
func Test_RouteNode_StrinctRouting(t *testing.T) {
	testCase := func(r string, method string, cases []testparams) {
		app := New(&Settings{StrictRouting: true})
		app.buildRouteNode(method, r, func(ctx *Ctx) {})
		for _, c := range cases {
			caseMethod := c.method
			if caseMethod == "" {
				caseMethod = method
			}
			handlers := app.findHandlers(c.url, methodInt(caseMethod))
			utils.AssertEqual(t, c.match, nil != handlers, fmt.Sprintf("route: '%s', url: '%s', method: '%s'", r, c.url, c.method))
		}
	}
	testCase("/api/v1/const", MethodGet, []testparams{
		{url: "/api/v1/const", method: MethodGet, match: true},
		{url: "/api/v1/const/", method: MethodGet, match: false},
		{url: "/api/v1/const", method: MethodPost, match: false},
		{url: "/api/v1", method: MethodGet, match: false},
		{url: "/api/v1/", method: MethodGet, match: false},
	})
}

// go test -race -run Test_RouteNode_CaseSensitive
func Test_RouteNode_CaseSensitive(t *testing.T) {
	testCase := func(r string, method string, cases []testparams) {
		app := New(&Settings{CaseSensitive: true})
		app.buildRouteNode(method, r, func(ctx *Ctx) {})
		for _, c := range cases {
			caseMethod := c.method
			if caseMethod == "" {
				caseMethod = method
			}
			handlers := app.findHandlers(c.url, methodInt(caseMethod))
			utils.AssertEqual(t, c.match, nil != handlers, fmt.Sprintf("route: '%s', url: '%s', method: '%s'", r, c.url, caseMethod))
		}
	}
	testCase("/api/v1/const", MethodGet, []testparams{
		{url: "/api/v1/const", method: MethodGet, match: true},
		{url: "/api/v1/const/", method: MethodGet, match: true},
		{url: "/api/v1/Const/", method: MethodGet, match: false},
		{url: "/api/v1/CONST/", method: MethodGet, match: false},
		{url: "/api/v1/const", method: MethodPost, match: false},
	})
}

var testBenchmarkRoutes = []string{
	"/user",
	"/user/k",
	"/user/k/1234",
	"/user/ke",
	"/user/ke/1234",
	"/user/key",
	"/user/key/1234",
	"/user/keys",
	"/user/keys/1234",
}

// go test -v ./... -run=^$ -bench=Benchmark_OldRouter -benchmem -count=4
func Benchmark_OldRouter(b *testing.B) {
	var match bool

	var routes []*Route
	for _, routePath := range testBenchmarkRoutes {
		parsed := parseRoute(routePath)
		route := &Route{
			use:         false,
			root:        false,
			star:        false,
			routeParser: parsed,
			path:        routePath,

			Path:   routePath,
			Method: "GET",
		}
		route.Handlers = append(route.Handlers, func(c *Ctx) {})
		routes = append(routes, route)
	}
	for n := 0; n < b.N; n++ {
		match = false
		for _, route := range routes {
			match, _ = route.match("/user/keys/1234", "/user/keys/1234")
			if match {
				break
			}
		}
		utils.AssertEqual(b, true, match)
	}
}

// go test -v ./... -run=^$ -bench=Benchmark_RadixTreeBasedRouter -benchmem -count=4
func Benchmark_RadixTreeBasedRouter(b *testing.B) {
	app := New()
	for _, routePath := range testBenchmarkRoutes {
		app.buildRouteNode("GET", routePath, func(ctx *Ctx) {})
	}
	for n := 0; n < b.N; n++ {
		handlers := app.findHandlers("/user/keys/1234", methodInt("GET"))
		utils.AssertEqual(b, true, handlers != nil)
	}
}
