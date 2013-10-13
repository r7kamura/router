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

func (router *Router) Get(pattern string, handlerOrFunc interface{}) {
	router.AppendRoute("GET", pattern, handlerOrFunc)
}

func (router *Router) Post(pattern string, handlerOrFunc interface{}) {
	router.AppendRoute("POST", pattern, handlerOrFunc)
}

func (router *Router) Put(pattern string, handlerOrFunc interface{}) {
	router.AppendRoute("PUT", pattern, handlerOrFunc)
}

func (router *Router) Delete(pattern string, handlerOrFunc interface{}) {
	router.AppendRoute("DELETE", pattern, handlerOrFunc)
}

func (router *Router) Any(pattern string, handlerOrFunc interface{}) {
	router.AppendRoute("ANY", pattern, handlerOrFunc)
}

func (router *Router) Handle(handlerOrFunc interface{}) {
	router.Routes["ANY"] = append(router.Routes["ANY"], NewEmptyRoute(handlerOrFunc))
}

func (router *Router) AppendRoute(method, pattern string, handlerOrFunc interface{}) {
	router.Routes[method] = append(router.Routes[method], NewRoute(pattern, handlerOrFunc))
}

var (
	// Precompile Regexp to speed things up
	anythingMatcher *regexp.Regexp = regexp.MustCompile("")

	// Precompile Regexp to speed things up.
	placeholderMatcher *regexp.Regexp = regexp.MustCompile(`:(\w+)`)
)

type Route struct {
	Pattern *regexp.Regexp
	Keys []string
	Handler http.Handler
}

func NewRoute(pattern string, handlerOrFunc interface{}) *Route {
	regexp, keys := compilePattern(pattern)
	return &Route{regexp, keys, convertToHandler(handlerOrFunc)}
}

func NewEmptyRoute(handlerOrFunc interface{}) *Route {
	return &Route{anythingMatcher, make([]string, 0), convertToHandler(handlerOrFunc)}
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

// Converts interface{} to http.Handler so that router can take Handler or HandlerFunc.
func convertToHandler(handlerOrFunc interface{}) (handler http.Handler) {
	if _, ok := handlerOrFunc.(http.Handler); ok {
		handler = handlerOrFunc.(http.Handler)
	} else {
		handler = http.HandlerFunc(handlerOrFunc.(func(http.ResponseWriter, *http.Request)))
	}
	return
}
