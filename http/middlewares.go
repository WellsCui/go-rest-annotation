package http

import (
	"fmt"
	"net/http"
	"sync"
)

type MiddlewareFunc func(http.HandlerFunc) http.HandlerFunc

var (
	middlewares = make(map[string]MiddlewareFunc)
	mu          sync.RWMutex
)

// RegisterMiddleware registers a middleware with a given name.
func RegisterMiddleware(name string, middleware MiddlewareFunc) error {
	mu.Lock()
	defer mu.Unlock()
	if _, exists := middlewares[name]; exists {
		return fmt.Errorf("middleware %s already registered", name)
	}
	middlewares[name] = middleware
	return nil
}

// GetMiddleware retrieves a middleware by name.
func GetMiddleware(name string) (MiddlewareFunc, error) {
	mu.RLock()
	defer mu.RUnlock()
	middleware, exists := middlewares[name]
	if !exists {
		return nil, fmt.Errorf("middleware %s not found", name)
	}
	return middleware, nil
}

// GetMiddlewares retrieves multiple middlewares by names.
func GetMiddlewares(names []string) ([]MiddlewareFunc, error) {
	mu.RLock()
	defer mu.RUnlock()
	result := make([]MiddlewareFunc, 0, len(names))
	for _, name := range names {
		middleware, exists := middlewares[name]
		if !exists {
			return nil, fmt.Errorf("middleware %s not found", name)
		}
		result = append(result, middleware)
	}
	return result, nil
}

// ClearMiddlewares clears all registered middlewares (useful for testing).
func ClearMiddlewares() {
	mu.Lock()
	defer mu.Unlock()
	middlewares = make(map[string]MiddlewareFunc)
}

func ApplyMiddlewares(handler http.HandlerFunc, restOperation RestOperation) (http.HandlerFunc, error) {
	// Apply middlewares in the order they were registered
	for i:=len(restOperation.Middlewares)-1; i>=0; i-- {
		middleware, err := GetMiddleware(restOperation.Middlewares[i])
		if err != nil {
			return nil, err
		}
		handler = middleware(handler)
	}
	return handler, nil
}

