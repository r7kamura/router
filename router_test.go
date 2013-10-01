package router

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"fmt"
)

var dummyHandler = http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {})

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
		if NewRoute(example.pattern, dummyHandler).Match(example.path) != example.match {
			if example.match {
				t.Errorf("%s should match %s", example.pattern, example.path)
			} else {
				t.Errorf("%s should not match %s", example.pattern, example.path)
			}
		}
	}
}

func TestRouterDoesNotMatchUnrelatedRoute(t *testing.T) {
	router := NewRouter()
	router.Get("/error", http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		t.Error("GET /a should not match this route")
	}))
	request, _ := http.NewRequest("GET", "/a", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
}

func TestRouterCallsDefaultNotFoundHandler(t *testing.T) {
	router := NewRouter()
	request, _ := http.NewRequest("GET", "/a", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	if recorder.Code != 404 {
		t.Errorf("Status code should be 404 but %v", recorder.Code)
	}
	if recorder.Body.String() != "Not Found\n" {
		t.Errorf("Response body should be Not Found but %v", recorder.Body.String())
	}
}

func TestRouterCallsCustomNotFoundHandler(t *testing.T) {
	router := NewRouter()
	router.NotFoundHandler = http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		http.Error(writer, "Custom NotFoundHandler", 403)
	})
	request, _ := http.NewRequest("GET", "/a", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	if recorder.Code != 403 {
		t.Errorf("Status code should be 403 but %v", recorder.Code)
	}
	if recorder.Body.String() != "Custom NotFoundHandler\n" {
		t.Errorf("Response body should be Custom NotFoundHandler but %v", recorder.Body.String())
	}
}


func TestRouterDoesNotMatchUnrelatedMethod(t *testing.T) {
	router := NewRouter()
	router.Post("/a", http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		t.Error("GET /a should not match this route")
	}))
	router.Get("/a", http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
	}))
	request, _ := http.NewRequest("GET", "/a", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
}

func TestRouterMatchesRelatedRoute(t *testing.T) {
	router := NewRouter()
	router.Get("/a", http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		fmt.Fprint(writer, "matched")
	}))
	request, _ := http.NewRequest("GET", "/a", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	if recorder.Body.String() != "matched" {
		t.Error("GET /a should match the defined route")
	}
	if recorder.Code != 200 {
		t.Errorf("GET /a should return status code 200 but %v", recorder.Code)
	}
}

func TestRouterMatchesFirstDefinedRoute(t *testing.T) {
	router := NewRouter()
	router.Get("/:any", http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
	}))
	router.Get("/a", http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		t.Error("GET /a should not match this route")
	}))
	request, _ := http.NewRequest("GET", "/a", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
}

func TestRouterAddsQueryStringWithPathParams(t *testing.T) {
	router := NewRouter()
	router.Get("/a/:b", http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		fmt.Fprint(writer, request.URL.RawQuery)
	}))
	request, _ := http.NewRequest("GET", "/a/b", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	if recorder.Body.String() != "b=b" {
		t.Errorf("Query string should be b=b but %v", recorder.Body)
	}
}

func TestRouterExtendsQueryStringWithPathParams(t *testing.T) {
	router := NewRouter()
	router.Get("/a/:b", http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		fmt.Fprint(writer, request.URL.RawQuery)
	}))
	request, _ := http.NewRequest("GET", "/a/b?b=c", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	if recorder.Body.String() != "b=c&b=b" {
		t.Errorf("Query string should be b=c&b=b but %v", recorder.Body)
	}
}
