# router
Provides a simple router for net/http.

## Install
```
go get github.com/r7kamura/router
```

## Usage
router has a great regard for http.Handler interface.

* router is a http.Handler
* router.{Get,Post,Put,Delete} takes a http.Handler
* router.{Get,Post,Put,Delete} takes a sinatra-like URL path pattern too

```go
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

	http.Handle("/", router)
	http.ListenAndServe(":3000", nil)
}
```
