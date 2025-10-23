package http

import (
	"fmt"
	"log"
	"net/http"
	"reflect"

	"github.com/gorilla/mux"
)

// RegisterRoutes parses annotations from handlerFile and registers routes on the router using reflection.
func RegisterRoutes(router *mux.Router, handler any, handlerFile string) error {
	routes, err := ParseRouteMetadata(handlerFile)
	if err != nil {
		return fmt.Errorf("failed to parse routes: %w", err)
	}
	handlerValue := reflect.ValueOf(handler)
	handlerType := reflect.TypeOf(handler)
	if handlerType.Kind() == reflect.Ptr {
		handlerType = handlerType.Elem()
	}
	for _, route := range routes {
		if route.HandlerType != "" && route.HandlerType != handlerType.Name() {
			continue
		}
		method := handlerValue.MethodByName(route.HandlerMethod)
		if !method.IsValid() {
			return fmt.Errorf("method %s not found on handler %s", route.HandlerMethod, route.HandlerType)
		}
		httpHandler := createHTTPHandler(method, route)
		httpHandler, err := ApplyMiddlewares(httpHandler, *route.Operation)
		if err != nil {
			return fmt.Errorf("failed to apply middlewares to handler: %w", err)
		}
		router.HandleFunc(route.Operation.Path, httpHandler).Methods(route.Operation.Method)
		log.Printf("Registered route: %s %s -> %s", route.Operation.Method, route.Operation.Path, route.HandlerMethod)
	}
	return nil
}

func createHTTPHandler(method reflect.Value, route *RouteMetadata) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Prepare arguments as a slice of reflect.Value
		args := []reflect.Value{
			reflect.ValueOf(w),
			reflect.ValueOf(r),
		}
		method.Call(args)
	}
}
