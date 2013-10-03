package main

import (
	"fmt"
	"github.com/r7kamura/router"
	"net/http"
)

func root(writer http.ResponseWriter, request *http.Request) {
	fmt.Fprint(writer, "Welcome")
}

func entry(writer http.ResponseWriter, request *http.Request) {
	fmt.Fprint(writer, "Entry: " + request.URL.Query().Get("id") + "\n")
}

func main() {
	router := router.NewRouter()
	router.Get("/", http.HandlerFunc(root))
	router.Get("/entries/:id", http.HandlerFunc(entry))
	http.ListenAndServe(":3000", router)
}
