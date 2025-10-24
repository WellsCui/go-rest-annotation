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
