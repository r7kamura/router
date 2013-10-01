package router

import (
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

type Router struct {
	Routes map[string][]*Route
	NotFoundHandler http.Handler
	host string
}

var notFoundHandler http.Handler = http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
	http.Error(writer, "Not Found", 404)
})

func NewRouter() *Router {
	return &Router{
		Routes: make(map[string][]*Route),
		NotFoundHandler: notFoundHandler,
	}
}

func (router *Router) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	if router.Match(request) {
		for _, route := range append(router.Routes[request.Method], router.Routes["ANY"]...) {
			if route.Match(request.URL.Path) {
				route.ServeHTTP(writer, request)
				return
			}
		}
	}
	router.NotFoundHandler.ServeHTTP(writer, request)
}

func (router *Router) Host(host string) {
	router.host = host
}

func (router *Router) Match(request *http.Request) bool {
	return router.MatchHost(request.URL.Host)
}

func (router *Router) MatchHost(host string) bool {
	return router.host == "" || router.host == strings.Split(host, ":")[0]
}

func (router *Router) Get(pattern string, handler http.Handler) {
	router.AppendRoute("GET", pattern, handler)
}

func (router *Router) Post(pattern string, handler http.Handler) {
	router.AppendRoute("POST", pattern, handler)
}

func (router *Router) Put(pattern string, handler http.Handler) {
	router.AppendRoute("PUT", pattern, handler)
}

func (router *Router) Delete(pattern string, handler http.Handler) {
	router.AppendRoute("DELETE", pattern, handler)
}

func (router *Router) Any(pattern string, handler http.Handler) {
	router.AppendRoute("ANY", pattern, handler)
}

func (router *Router) Handle(handler http.Handler) {
	router.Routes["ANY"] = append(router.Routes["ANY"], NewEmptyRoute(handler))
}

func (router *Router) AppendRoute(method, pattern string, handler http.Handler) {
	router.Routes[method] = append(router.Routes[method], NewRoute(pattern, handler))
}

type Route struct {
	Pattern *regexp.Regexp
	Keys []string
	Handler http.Handler
}

func NewRoute(pattern string, handler http.Handler) *Route {
	regexp, keys := compilePattern(pattern)
	return &Route{regexp, keys, handler}
}

// Precompile Regexp to speed things up
var anythingMatcher *regexp.Regexp = regexp.MustCompile("")

func NewEmptyRoute(handler http.Handler) *Route {
	return &Route{anythingMatcher, make([]string, 0), handler}
}

func (route *Route) Match(path string) bool {
	return route.Pattern.MatchString(path)
}

func (route *Route) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	params := request.URL.Query()
	for key, values := range route.extractParams(request.URL.Path) {
		params[key] = append(params[key], values...)
	}
	request.URL.RawQuery = params.Encode()
	route.Handler.ServeHTTP(writer, request)
}

func (route *Route) extractParams(path string) url.Values {
	params := make(url.Values)
	for i, param := range route.Pattern.FindStringSubmatch(path)[1:] {
		params[route.Keys[i]] = append(params[route.Keys[i]], param)
	}
	return params
}

// Precompile Regexp to speed things up.
var placeholderMatcher *regexp.Regexp = regexp.MustCompile(`:(\w+)`)

// compilePattern("/hello/:world") => ^\/hello\/([^#?/]+)$, ["world"]
func compilePattern(pattern string) (*regexp.Regexp, []string) {
	var segments, keys []string
	for _, segment := range strings.Split(pattern, "/") {
		if strings := placeholderMatcher.FindStringSubmatch(segment); strings != nil {
			keys = append(keys, strings[1])
			segments = append(segments, placeholderMatcher.ReplaceAllString(segment, "([^#?/]+)"))
		} else {
			segments = append(segments, segment)
		}
	}
	return regexp.MustCompile(`^` + strings.Join(segments, `\/`) + "$"), keys
}
