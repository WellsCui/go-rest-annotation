package http

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wellscui/go-rest-annotation/internal/person"
)



func TestRegisterRoutes(t *testing.T) {
	t.Run("registers single route and handles request", func(t *testing.T) {
		router := mux.NewRouter()
		handler := &person.Handler{}
		RegisterMiddleware("PersonMiddleWare", person.PersonMiddleWare)
		err := RegisterRoutes(router, handler, "../person/handler.go")
		require.NoError(t, err)
		require.NotNil(t, router.Get("person.Handler.GetPersonHTTP"))
		req := httptest.NewRequest("GET", "/person/bill", nil)
		res := httptest.NewRecorder()
		router.ServeHTTP(res, req)
		assert.Equal(t, http.StatusOK, res.Code)
	})

	t.Run("returns error when file does not exist", func(t *testing.T) {
		router := mux.NewRouter()
		handler := &person.Handler{}
		err := RegisterRoutes(router, handler, "nonexistent.go")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse routes")
	})

	t.Run("returns error when method not found", func(t *testing.T) {
		router := mux.NewRouter()
		type Handler struct{}
		handler := &Handler{}
		err := RegisterRoutes(router, handler, "../person/handler.go")
		require.Equal(t, errors.New("method GetPersonHTTP not found in handler Handler"), err)
	})

	t.Run("returns error when middleware not found", func(t *testing.T) {
		router := mux.NewRouter()
		handler := &person.Handler{}
		ClearMiddlewares()
		err := RegisterRoutes(router, handler, "../person/handler.go")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to apply middlewares")
	})
}
