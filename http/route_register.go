package http

import (
	"fmt"
	"log"
	"net/http"
	"reflect"
	"errors"

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
	if handlerType.Kind() != reflect.Ptr || handlerType.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("handler must be a pointer to struct")
	}
	handlerType = handlerType.Elem()
	for _, route := range routes {
		if route.HandlerType != "" && route.HandlerType != handlerType.Name() {
			continue
		}
		method := handlerValue.MethodByName(route.HandlerMethod)
		if !method.IsValid() {
			return fmt.Errorf("method %s not found in handler %s", route.HandlerMethod, route.HandlerType)
		}

		httpHandler, err := createHTTPHandler(method)
		if err != nil {
			return fmt.Errorf("failed to create http handler for %s.%s : %w", route.HandlerType, route.HandlerMethod,  err)
		}

		httpHandler, err = ApplyMiddlewares(httpHandler, *route.Operation)
		if err != nil {
			return fmt.Errorf("failed to apply middlewares to handler: %w", err)
		}
		routeName:=fmt.Sprintf("%s.%s.%s", route.PackagePath, route.HandlerType, route.HandlerMethod)
		router.HandleFunc(route.Operation.Path, httpHandler).Methods(route.Operation.Method).Name(routeName)
		log.Printf("Registered route %s: %s %s -> %s", routeName, route.Operation.Method, route.Operation.Path, route.HandlerMethod)
	}
	return nil
}

func createHTTPHandler(method reflect.Value) (http.HandlerFunc, error ){
	if !isHTTPHandlerFunc(method) {
			return nil, errors.New("method is not a http.HandlerFunc")
	}
	return func(w http.ResponseWriter, r *http.Request) {
		// Prepare arguments as a slice of reflect.Value
		args := []reflect.Value{
			reflect.ValueOf(w),
			reflect.ValueOf(r),
		}
		method.Call(args)
	}, nil
}

// isHTTPHandlerFunc checks if a reflect.Value is an http.HandlerFunc
func isHTTPHandlerFunc(funcValue reflect.Value) bool {
	f:=func(w http.ResponseWriter, r *http.Request) {}
	return funcValue.Type()==reflect.TypeOf(f)
}
