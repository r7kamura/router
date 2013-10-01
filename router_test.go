package router

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"fmt"
)

// Utility type for table-testing at TestRouteMatch
type routeTestExample struct {
	pattern string
	path string
	match bool
}

func TestRouteMatch(t *testing.T) {
	examples := []routeTestExample{
		{"/:a", "/", false},
		{"/:a", "/a", true},
		{"/:a", "/a.html", true},
		{"/:a/", "/a/", true},
		{"/:a/:b", "/a/", false},
		{"/:a/:b", "/a/b", true},
		{"/:a/:b", "/a/b/c", false},
		{"/:a/b/:c", "/a/b/c", true},
		{"/a/:b", "/a", false},
		{"/a/:b", "/a/", false},
		{"/a/:b", "/a/b", true},
		{"/a/:b", "/a/b/", false},
		{"/a/:b", "/a/b/c", false},
		{"/a/:b/c", "/a/b/c", true},
		{"/a/:b/c", "/a/b/c/d", false},
	}
	for _, example := range examples {
		if NewRoute(example.pattern, emptyHandler).Match(example.path) != example.match {
			if example.match {
				t.Errorf("%s should match %s", example.pattern, example.path)
			} else {
				t.Errorf("%s should not match %s", example.pattern, example.path)
			}
		}
	}
}

// Utility function to create an instance GET request.
func get(router http.Handler, path string) *httptest.ResponseRecorder {
	return request(router, "GET", path)
}

// Utility function to create an instance POST request.
func post(router http.Handler, path string) *httptest.ResponseRecorder {
	return request(router, "POST", path)
}

// Utility function to create an instance request.
func request(router http.Handler, method, path string) *httptest.ResponseRecorder {
	request, _ := http.NewRequest(method, path, nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	return recorder
}

// Utility function to create a http.Handler object from a callback function.
func handler(callback func()) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		callback()
	})
}

// Utility object as empty http.Handler
var emptyHandler http.Handler = handler(func() {})

func TestRouterDoesNotMatchUnrelatedPath(t *testing.T) {
	router := NewRouter()
	router.Get("/error", emptyHandler)
	if get(router, "/a").Code != 404 {
		t.Error("Router should not match unrelated path")
	}
}

func TestRouterCallsDefaultNotFoundHandler(t *testing.T) {
	router := NewRouter()
	response := get(router, "/a")
	if response.Code != 404 || response.Body.String() != "Not Found\n" {
		t.Error("Router should return 404 in non-matched case")
	}
}

func TestRouterCallsCustomNotFoundHandler(t *testing.T) {
	router := NewRouter()
	router.NotFoundHandler = http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		http.Error(writer, "Custom NotFoundHandler", 403)
	})
	if get(router, "/a").Code != 403 {
		t.Error("Router should use custom NotFoundHandler")
	}
}


func TestRouterDoesNotMatchUnrelatedMethod(t *testing.T) {
	router := NewRouter()
	router.Post("/a", handler(func() {
		t.Error("Router should not match unrelated method")
	}))
	router.Get("/a", emptyHandler)
	get(router, "/a")
}

func TestRouterMatchesRelatedRoute(t *testing.T) {
	router := NewRouter()
	router.Get("/a", emptyHandler)
	if get(router, "/a").Code != 200 {
		t.Error("Router should match related route")
	}
}

func TestRouterMatchesFirstDefinedRoute(t *testing.T) {
	router := NewRouter()
	router.Get("/:any", emptyHandler)
	router.Get("/:a", handler(func() {
		t.Error("Router should match first defined route")
	}))
	get(router, "/a")
}

func TestRouterAddsQueryStringWithPathParams(t *testing.T) {
	router := NewRouter()
	router.Get("/a/:b", http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		fmt.Fprint(writer, request.URL.RawQuery)
	}))
	if get(router, "/a/b").Body.String() != "b=b" {
		t.Error("Router should add query string with path params")
	}
}

func TestRouterExtendsQueryStringWithPathParams(t *testing.T) {
	router := NewRouter()
	router.Get("/a/:b", http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		fmt.Fprint(writer, request.URL.RawQuery)
	}))
	if get(router, "/a/b?b=c").Body.String() != "b=c&b=b" {
		t.Error("Router should extend query string with path params")
	}
}

func TestRouterMatchesAnyMethod(t *testing.T) {
	router := NewRouter()
	router.Any("/a", http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		fmt.Fprint(writer, request.Method)
	}))
	if get(router, "/a").Body.String() != "GET" {
		t.Error("Router#Any should match GET /a")
	}
	if post(router, "/a").Body.String() != "POST" {
		t.Error("Router#Any should match POST /a")
	}
	if get(router, "/b").Body.String() != "Not Found\n" {
		t.Error("Router#Any should not match GET /b")
	}
}

func TestRouterMatchesOnlyRelatedHost(t *testing.T) {
	router := NewRouter()
	router.Get("/a", emptyHandler)
	router.Host("example.com")
	if get(router, "/a").Code != 404 {
		t.Error("Router should not match unrelated host")
	}
	if get(router, "http://example.com/a").Code != 200 {
		t.Error("Router should match related host")
	}
}

func TestRouterMatchesManyHosts(t *testing.T) {
	mainRouter := NewRouter()
	mainRouter.Get("/b", emptyHandler)
	apiRouter := NewRouter()
	apiRouter.Host("api.example.com")
	apiRouter.Get("/a", emptyHandler)
	apiRouter.NotFoundHandler = mainRouter
	if get(apiRouter, "http://api.example.com/a").Code != 200 {
		t.Error("Router should match related host and path")
	}
	if get(apiRouter, "http://api.example.com/b").Code != 200 {
		t.Error("Router should match related host and path")
	}
	if get(apiRouter, "/b").Code != 200 {
		t.Error("Router should match related host and path")
	}
	if get(apiRouter, "/a").Code != 404 {
		t.Error("Router should not match unrelated host and path")
	}
}
