package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/wellscui/go-rest-annotation/internal/person"
	rest "github.com/wellscui/go-rest-annotation/internal/http"

)



func main() {
	handler := &person.Handler{}
	router := mux.NewRouter()
	rest.RegisterMiddleware("PersonMiddleWare", person.PersonMiddleWare)
	err := rest.RegisterRoutes(router, handler, "./internal/person/handler.go")
	if err != nil {
		log.Fatalf("Failed to register routes: %v", err)
	}
	log.Println("Server starting on :8089")
	log.Println("Try: curl http://localhost:8089/person/123")
	log.Fatal(http.ListenAndServe(":8089", router))
}
