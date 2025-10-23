package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegisterMiddleware(t *testing.T) {
	t.Run("registers middleware successfully", func(t *testing.T) {
		ClearMiddlewares()
		mw := func(next http.HandlerFunc) http.HandlerFunc { return next }
		err := RegisterMiddleware("auth", mw)
		assert.NoError(t, err)
		retrieved, err := GetMiddleware("auth")
		assert.NoError(t, err)
		assert.NotNil(t, retrieved)
	})

	t.Run("returns error for duplicate registration", func(t *testing.T) {
		ClearMiddlewares()
		mw := func(next http.HandlerFunc) http.HandlerFunc { return next }
		RegisterMiddleware("logging", mw)
		err := RegisterMiddleware("logging", mw)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already registered")
	})
}

func TestGetMiddleware(t *testing.T) {
	t.Run("retrieves existing middleware", func(t *testing.T) {
		ClearMiddlewares()
		mw := func(next http.HandlerFunc) http.HandlerFunc { return next }
		RegisterMiddleware("cors", mw)
		retrieved, err := GetMiddleware("cors")
		assert.NoError(t, err)
		assert.NotNil(t, retrieved)
	})

	t.Run("returns error for non-existent middleware", func(t *testing.T) {
		ClearMiddlewares()
		retrieved, err := GetMiddleware("notfound")
		assert.Error(t, err)
		assert.Nil(t, retrieved)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("middleware modifies request", func(t *testing.T) {
		ClearMiddlewares()
		mw := func(next http.HandlerFunc) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("X-Custom", "test")
				next(w, r)
			}
		}
		RegisterMiddleware("custom", mw)
		retrieved, _ := GetMiddleware("custom")
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}
		wrapped := retrieved(handler)
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()
		wrapped(w, req)
		assert.Equal(t, "test", w.Header().Get("X-Custom"))
	})
}

func TestGetMiddlewares(t *testing.T) {
	t.Run("retrieves multiple middlewares", func(t *testing.T) {
		ClearMiddlewares()
		mw1 := func(next http.HandlerFunc) http.HandlerFunc { return next }
		mw2 := func(next http.HandlerFunc) http.HandlerFunc { return next }
		RegisterMiddleware("first", mw1)
		RegisterMiddleware("second", mw2)
		retrieved, err := GetMiddlewares([]string{"first", "second"})
		assert.NoError(t, err)
		assert.Len(t, retrieved, 2)
	})

	t.Run("returns error if any middleware missing", func(t *testing.T) {
		ClearMiddlewares()
		mw := func(next http.HandlerFunc) http.HandlerFunc { return next }
		RegisterMiddleware("exists", mw)
		retrieved, err := GetMiddlewares([]string{"exists", "missing"})
		assert.Error(t, err)
		assert.Nil(t, retrieved)
	})

	t.Run("handles empty list", func(t *testing.T) {
		ClearMiddlewares()
		retrieved, err := GetMiddlewares([]string{})
		assert.NoError(t, err)
		assert.Empty(t, retrieved)
	})

	t.Run("middlewares chain correctly", func(t *testing.T) {
		ClearMiddlewares()
		order := []string{}
		mw1 := func(next http.HandlerFunc) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				order = append(order, "first")
				next(w, r)
			}
		}
		mw2 := func(next http.HandlerFunc) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				order = append(order, "second")
				next(w, r)
			}
		}
		RegisterMiddleware("first", mw1)
		RegisterMiddleware("second", mw2)
		mws, _ := GetMiddlewares([]string{"first", "second"})
		handler := func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "handler")
		}
		wrapped := handler
		for i := len(mws) - 1; i >= 0; i-- {
			wrapped = mws[i](wrapped)
		}
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()
		wrapped(w, req)
		assert.Equal(t, []string{"first", "second", "handler"}, order)
	})
}

func TestClearMiddlewares(t *testing.T) {
	t.Run("removes all middlewares", func(t *testing.T) {
		mw := func(next http.HandlerFunc) http.HandlerFunc { return next }
		RegisterMiddleware("test1", mw)
		RegisterMiddleware("test2", mw)
		ClearMiddlewares()
		_, err := GetMiddleware("test1")
		assert.Error(t, err)
		_, err = GetMiddleware("test2")
		assert.Error(t, err)
	})

	t.Run("allows re-registration after clear", func(t *testing.T) {
		mw := func(next http.HandlerFunc) http.HandlerFunc { return next }
		RegisterMiddleware("reuse", mw)
		ClearMiddlewares()
		err := RegisterMiddleware("reuse", mw)
		assert.NoError(t, err)
	})
}

func TestApplyMiddlewares(t *testing.T) {
	t.Run("applies middlewares in order", func(t *testing.T) {
		ClearMiddlewares()
		order := []string{}
		mw1 := func(next http.HandlerFunc) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				order = append(order, "mw1")
				next(w, r)
			}
		}
		mw2 := func(next http.HandlerFunc) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				order = append(order, "mw2")
				next(w, r)
			}
		}
		RegisterMiddleware("mw1", mw1)
		RegisterMiddleware("mw2", mw2)
		handler := func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "handler")
		}
		op := RestOperation{Middlewares: []string{"mw1", "mw2"}}
		wrapped, err := ApplyMiddlewares(handler, op)
		require.NoError(t, err)
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()
		wrapped(w, req)
		assert.Equal(t, []string{"mw1", "mw2", "handler"}, order)
	})

	t.Run("returns error for missing middleware", func(t *testing.T) {
		ClearMiddlewares()
		handler := func(w http.ResponseWriter, r *http.Request) {}
		op := RestOperation{Middlewares: []string{"missing"}}
		wrapped, err := ApplyMiddlewares(handler, op)
		assert.Error(t, err)
		assert.Nil(t, wrapped)
	})

	t.Run("handles empty middleware list", func(t *testing.T) {
		ClearMiddlewares()
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}
		op := RestOperation{Middlewares: []string{}}
		wrapped, err := ApplyMiddlewares(handler, op)
		require.NoError(t, err)
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()
		wrapped(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestThreadSafety(t *testing.T) {
	t.Run("concurrent operations", func(t *testing.T) {
		ClearMiddlewares()
		done := make(chan bool, 10)
		for i := 0; i < 10; i++ {
			go func(n int) {
				mw := func(next http.HandlerFunc) http.HandlerFunc { return next }
				name := string(rune('a' + n))
				RegisterMiddleware(name, mw)
				GetMiddleware(name)
				done <- true
			}(i)
		}
		for i := 0; i < 10; i++ {
			<-done
		}
	})
}
