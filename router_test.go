package router

import (
	. "github.com/r7kamura/gospel"
	"net/http"
	"net/http/httptest"
	"testing"
	"fmt"
)

// Utility type for table-testing at TestRoute
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
	Describe(t, "router.Route#Match", func() {
		for _, example := range examples {
			var verb string
			if example.match {
				verb = " matches "
			} else {
				verb = " does not match "
			}
			It(example.pattern + verb + example.path, func() {
				Expect(NewRoute(example.pattern, dummyHandler).Match(example.path)).To(Equal, example.match)
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

func dummyHandler(writer http.ResponseWriter, request *http.Request) {
	fmt.Fprintf(writer, request.URL.Path + "?" + request.URL.RawQuery)
}

func TestRouter(t *testing.T) {
	Describe(t, "router.Router", func() {
		router := NewRouter()
		router.Get("/a", dummyHandler)
		router.Get("/:any", dummyHandler)
		router.Get("/b", dummyHandler)
		router.Any("/c/d", dummyHandler)

		Context("with unrelated path request", func() {
			It("does not match & return default 404 response", func() {
				response := get(router, "/a/b/c")
				Expect(response.Code).To(Equal, 404)
				Expect(response.Body.String()).To(Equal, "Not Found\n")
			})
		})

		Context("with custom 404 handler", func() {
			customNotFoundRouter := NewRouter()
			customNotFoundRouter.NotFoundHandler = http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				http.Error(writer, "Custom NotFoundHandler", 403)
			})
			It("returns custom response", func() {
				Expect(get(customNotFoundRouter, "/404").Code).To(Equal, 403)
			})
		})

		Context("with unrelated method request", func() {
			It("does not match", func() {
				Expect(post(router, "/a").Code).To(Equal, 404)
			})
		})

		Context("with related method request", func() {
			It("does not match", func() {
				Expect(get(router, "/a").Code).To(Equal, 200)
			})
		})

		Context("with 2 related pattern", func() {
			It("matches the first", func() {
				Expect(get(router, "/a").Body.String()).To(Equal, "/a?")
			})
		})

		Context("with path params pattern", func() {
			It("addes query string with path params", func() {
				Expect(get(router, "/c").Body.String()).To(Equal, "/c?any=c")
			})
		})

		Context("with duplicated params", func() {
			It("merged query string with path params", func() {
				Expect(get(router, "/c?any=d").Body.String()).To(Equal, "/c?any=d&any=c")
			})
		})

		Context("with Any route", func() {
			It("matches any method", func() {
				Expect(get(router, "/c/d").Code).To(Equal, 200)
				Expect(post(router, "/c/d").Code).To(Equal, 200)
			})
		})

		Context("with host route", func() {
			hostRouter := NewRouter()
			hostRouter.Host("example.com")
			hostRouter.Get("/a", dummyHandler)
			It("matches related host", func() {
				Expect(get(hostRouter, "/a").Code).To(Equal, 404)
				Expect(get(hostRouter, "http://example.com/a").Code).To(Equal, 200)
			})
		})

		Context("with many host routes", func() {
			mainRouter := NewRouter()
			mainRouter.Get("/a", dummyHandler)
			apiRouter := NewRouter()
			apiRouter.Host("api.example.com")
			apiRouter.Get("/b", dummyHandler)
			apiRouter.NotFoundHandler = mainRouter
			It("matches related host", func() {
				Expect(get(apiRouter, "/a").Code).To(Equal, 200)
				Expect(get(apiRouter, "/b").Code).To(Equal, 404)
				Expect(get(apiRouter, "http://api.example.com/a").Code).To(Equal, 200)
				Expect(get(apiRouter, "http://api.example.com/b").Code).To(Equal, 200)
			})
		})

		Context("with Handle handler", func() {
			anyRouter := NewRouter()
			anyRouter.Handle(dummyHandler)
			It("matches any request", func() {
				Expect(get(anyRouter, "/a").Code).To(Equal, 200)
				Expect(get(anyRouter, "/b").Code).To(Equal, 200)
				Expect(post(anyRouter, "/b").Code).To(Equal, 200)
			})
		})
	})
}
