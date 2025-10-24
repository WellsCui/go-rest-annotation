package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type Handler struct{}

// @RestOperation( method = "GET", path = "/person/{uid}", middlewares = ["TestMiddleWare"], timeout = 30, disableAuth = true )
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

// @RestOperation( method = "POST", path = "/person", middlewares = ["TestMiddleWare"], timeout = 30, disableAuth = true )
func (h *Handler) Post(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

// handler2 for testing invalid method
type handler2 struct{}

// @RestOperation( method = "GET", path = "/person", middlewares = ["TestMiddleWare"], timeout = 30, disableAuth = true )
func (h *handler2) Get() {
	
}

func testMiddleWare(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Header.Set("middleware", "TestMiddleWare")
		next(w, r)
	}
}

func TestRegisterRoutes(t *testing.T) {
	t.Run("registers single route and handles request", func(t *testing.T) {
		router := mux.NewRouter()
		handler := &Handler{}
		RegisterMiddleware("TestMiddleWare", testMiddleWare)
		err := RegisterRoutes(router, handler, "./route_register_test.go")
		require.NoError(t, err)
		require.NotNil(t, router.Get("http.Handler.Get"))
		require.NotNil(t, router.Get("http.Handler.Post"))
		req := httptest.NewRequest("GET", "/person/bill", nil)
		res := httptest.NewRecorder()
		router.ServeHTTP(res, req)
		assert.Equal(t, http.StatusOK, res.Code)
		req = httptest.NewRequest("POST", "/person", nil)
		res = httptest.NewRecorder()
		router.ServeHTTP(res, req)
		assert.Equal(t, http.StatusOK, res.Code)
	})

	t.Run("returns error when handler is not pointer of struct", func(t *testing.T) {
		router := mux.NewRouter()
		handler := Handler{}
		err := RegisterRoutes(router, handler, "./route_register_test.go")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "handler must be a pointer to struct")
	})

	t.Run("returns error when file does not exist", func(t *testing.T) {
		router := mux.NewRouter()
		handler := &Handler{}
		err := RegisterRoutes(router, handler, "nonexistent.go")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse routes")
	})

	t.Run("returns error when go file does not match", func(t *testing.T) {
		router := mux.NewRouter()
		handler := &Handler{}
		err := RegisterRoutes(router, handler, "../example/person/handler.go")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "method GetPersonHTTP not found in handler Handler")
	})

	t.Run("returns error when method is not a http.HandlerFunc", func(t *testing.T) {
		router := mux.NewRouter()
		handler := &handler2{}
		err := RegisterRoutes(router, handler, "./route_register_test.go")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "method is not a http.HandlerFunc")
	})

	t.Run("returns error when middleware not found", func(t *testing.T) {
		router := mux.NewRouter()
		handler := &Handler{}
		ClearMiddlewares()
		err := RegisterRoutes(router, handler, "./route_register_test.go")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to apply middlewares")
	})
}
