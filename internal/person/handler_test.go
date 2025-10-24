package person

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	rest "github.com/wellscui/go-rest-annotation/internal/http"
)

func TestRegisterRoutes(t *testing.T) {
	t.Run("registers single route and handles request", func(t *testing.T) {
		router := mux.NewRouter()
		handler := &Handler{}
		rest.RegisterMiddleware("PersonMiddleWare", PersonMiddleWare)
		err := rest.RegisterRoutes(router, handler, "./handler.go")
		require.NoError(t, err)
		require.NotNil(t, router.Get("person.Handler.GetPersonHTTP"))
		req := httptest.NewRequest("GET", "/person/bill", nil)
		res := httptest.NewRecorder()
		router.ServeHTTP(res, req)
		assert.Equal(t, http.StatusOK, res.Code)
	})
}