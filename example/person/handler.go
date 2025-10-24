package person

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func (s *Handler) getPerson(ctx context.Context, uid string) (*Person, error) {
	return &Person{
		UID:  uid,
		Name: "John Doe",
	}, nil
}

// Handler represents your application service
type Handler struct{}

// GetPersonHTTP is the HTTP handler wrapper for getPerson
// @RestOperation( method = "GET", path = "/person/{uid}", middlewares = ["PersonMiddleWare"], timeout = 30, disableAuth = true )
func (s *Handler) GetPersonHTTP(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uid := vars["uid"]
	person, err := s.getPerson(r.Context(), uid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Println("Getting Address in Header")
	person.Addr = r.Header.Get("Address")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(person)
}

// Person represents a data model
type Person struct {
	UID  string
	Name string
	Addr string
}

func PersonMiddleWare(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Setting Address in context")
		r.Header.Set("Address", "123 Main St")
		next(w, r)
	}
}