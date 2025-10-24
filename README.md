# go-rest-annotation
This Repo provides a way to register routes via adding annotation on handler functions.

## Register route via annotation
```go
// GetPersonHTTP is the HTTP handler for getting a person
// @RestOperation( method = "GET", path = "/person/{uid}", middlewares = ["PersonMiddleWare" "AnotherMiddleWare"], timeout = 30, disableAuth = true )
func (s *Handler) GetPersonHTTP(w http.ResponseWriter, r *http.Request) {
	
}

// PostPersonHTTP is the HTTP handler for posting a person
// @RestOperation( method = "POST", path = "/person", middlewares = ["PersonMiddleWare" "AnotherMiddleWare"], timeout = 30, disableAuth = true )
func (s *Handler) PostPersonHTTP(w http.ResponseWriter, r *http.Request) {
	
}


```

## Register middlewares and build routes
```go

package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/wellscui/go-rest-annotation/example/person"
	rest "github.com/wellscui/go-rest-annotation/http"

)

func main() {
	handler := &person.Handler{}
	router := mux.NewRouter()
	rest.RegisterMiddleware("PersonMiddleWare", person.PersonMiddleWare)
	err := rest.RegisterRoutes(router, handler, "./person/handler.go")
	if err != nil {
		log.Fatalf("Failed to register routes: %v", err)
	}
	log.Println("Server starting on :8080")
	log.Println("Try: curl http://localhost:8080/person/123")
	log.Fatal(http.ListenAndServe(":8080", router))
}
```
