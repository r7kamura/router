package router

import (
	. "github.com/r7kamura/gospel"
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

func TestRoute(t *testing.T) {
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
	Describe(t, "router.Route#Match", func(context Context, it It) {
		for _, example := range examples {
			var verb string
			if example.match {
				verb = " matches "
			} else {
				verb = " does not match "
			}
			it(example.pattern + verb + example.path, func(expect Expect) {
				expect(NewRoute(example.pattern, dummyHandler).Match(example.path)).ToEqual(example.match)
			})
		}
	})
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

// Utility object as empty http.Handler
var dummyHandler http.Handler = http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
	fmt.Fprintf(writer, request.URL.Path + "?" + request.URL.RawQuery)
})

func TestRouter(t *testing.T) {
	Describe(t, "router.Router", func(context Context, it It) {
		router := NewRouter()
		router.Get("/a", dummyHandler)
		router.Get("/:any", dummyHandler)
		router.Get("/b", dummyHandler)
		router.Any("/c/d", dummyHandler)

		context("with unrelated path request", func() {
			it("does not match & return default 404 response", func(expect Expect) {
				response := get(router, "/a/b/c")
				expect(response.Code).ToEqual(404)
				expect(response.Body.String()).ToEqual("Not Found\n")
			})
		})

		context("with custom 404 handler", func() {
			customNotFoundRouter := NewRouter()
			customNotFoundRouter.NotFoundHandler = http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				http.Error(writer, "Custom NotFoundHandler", 403)
			})
			it("returns custom response", func(expect Expect) {
				expect(get(customNotFoundRouter, "/404").Code).ToEqual(403)
			})
		})

		context("with unrelated method request", func() {
			it("does not match", func(expect Expect) {
				expect(post(router, "/a").Code).ToEqual(404)
			})
		})

		context("with related method request", func() {
			it("does not match", func(expect Expect) {
				expect(get(router, "/a").Code).ToEqual(200)
			})
		})

		context("with 2 related pattern", func() {
			it("matches the first", func(expect Expect) {
				expect(get(router, "/a").Body.String()).ToEqual("/a?")
			})
		})

		context("with path params pattern", func() {
			it("addes query string with path params", func(expect Expect) {
				expect(get(router, "/c").Body.String()).ToEqual("/c?any=c")
			})
		})

		context("with duplicated params", func() {
			it("merged query string with path params", func(expect Expect) {
				expect(get(router, "/c?any=d").Body.String()).ToEqual("/c?any=d&any=c")
			})
		})

		context("with Any route", func() {
			it("matches any method", func(expect Expect) {
				expect(get(router, "/c/d").Code).ToEqual(200)
				expect(post(router, "/c/d").Code).ToEqual(200)
			})
		})

		context("with host route", func() {
			hostRouter := NewRouter()
			hostRouter.Host("example.com")
			hostRouter.Get("/a", dummyHandler)
			it("matches related host", func(expect Expect) {
				expect(get(hostRouter, "/a").Code).ToEqual(404)
				expect(get(hostRouter, "http://example.com/a").Code).ToEqual(200)
			})
		})

		context("with many host routes", func() {
			mainRouter := NewRouter()
			mainRouter.Get("/a", dummyHandler)
			apiRouter := NewRouter()
			apiRouter.Host("api.example.com")
			apiRouter.Get("/b", dummyHandler)
			apiRouter.NotFoundHandler = mainRouter
			it("matches related host", func(expect Expect) {
				expect(get(apiRouter, "/a").Code).ToEqual(200)
				expect(get(apiRouter, "/b").Code).ToEqual(404)
				expect(get(apiRouter, "http://api.example.com/a").Code).ToEqual(200)
				expect(get(apiRouter, "http://api.example.com/b").Code).ToEqual(200)
			})
		})

		context("with Handle handler", func() {
			anyRouter := NewRouter()
			anyRouter.Handle(dummyHandler)
			it("matches any request", func(expect Expect) {
				expect(get(anyRouter, "/a").Code).ToEqual(200)
				expect(get(anyRouter, "/b").Code).ToEqual(200)
				expect(post(anyRouter, "/b").Code).ToEqual(200)
			})
		})
	})
}
