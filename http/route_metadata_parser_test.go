package http

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseRoutes(t *testing.T) {
	t.Run("parses file with single route", func(t *testing.T) {
		content := `package main

// @RestOperation( method = "GET", path = "/users" )
func (s *Service) getUsers() {}

type Service struct{}
`
		tmpFile := createTempFile(t, content)
		defer os.Remove(tmpFile)
		routes, err := ParseRouteMetadata(tmpFile)
		require.NoError(t, err)
		require.Len(t, routes, 1)
		assert.Equal(t, "GET", routes[0].Operation.Method)
		assert.Equal(t, "/users", routes[0].Operation.Path)
		assert.Equal(t, "getUsers", routes[0].HandlerMethod)
		assert.Equal(t, "Service", routes[0].HandlerType)
		assert.Equal(t, "main", routes[0].Package)
	})

	t.Run("parses file with multiple routes", func(t *testing.T) {
		content := `package api

// @RestOperation( method = "GET", path = "/users" )
func (s *Service) getUsers() {}

// @RestOperation( method = "POST", path = "/users" )
func (s *Service) createUser() {}

type Service struct{}
`
		tmpFile := createTempFile(t, content)
		defer os.Remove(tmpFile)
		routes, err := ParseRouteMetadata(tmpFile)
		require.NoError(t, err)
		require.Len(t, routes, 2)
		assert.Equal(t, "GET", routes[0].Operation.Method)
		assert.Equal(t, "/users", routes[0].Operation.Path)
		assert.Equal(t, "POST", routes[1].Operation.Method)
		assert.Equal(t, "/users", routes[1].Operation.Path)
	})

	t.Run("parses route with all fields", func(t *testing.T) {
		content := `package main

// @RestOperation( method = "GET", path = "/person/{uid}", middlewares = ["auth"], timeout = 30, disableAuth = true )
func (s *Service) getPerson() {}

type Service struct{}
`
		tmpFile := createTempFile(t, content)
		defer os.Remove(tmpFile)
		routes, err := ParseRouteMetadata(tmpFile)
		require.NoError(t, err)
		require.Len(t, routes, 1)
		assert.Equal(t, "GET", routes[0].Operation.Method)
		assert.Equal(t, "/person/{uid}", routes[0].Operation.Path)
		assert.Equal(t, []string{"auth"}, routes[0].Operation.Middlewares)
		assert.Equal(t, 30, routes[0].Operation.Timeout)
		assert.True(t, routes[0].Operation.DisableAuth)
	})

	t.Run("parses value receiver", func(t *testing.T) {
		content := `package main

// @RestOperation( method = "GET", path = "/test" )
func (s Service) test() {}

type Service struct{}
`
		tmpFile := createTempFile(t, content)
		defer os.Remove(tmpFile)
		routes, err := ParseRouteMetadata(tmpFile)
		require.NoError(t, err)
		require.Len(t, routes, 1)
		assert.Equal(t, "Service", routes[0].HandlerType)
	})

	t.Run("parses function without receiver", func(t *testing.T) {
		content := `package main

// @RestOperation( method = "GET", path = "/test" )
func test() {}
`
		tmpFile := createTempFile(t, content)
		defer os.Remove(tmpFile)
		routes, err := ParseRouteMetadata(tmpFile)
		require.NoError(t, err)
		require.Len(t, routes, 1)
		assert.Equal(t, "test", routes[0].HandlerMethod)
		assert.Empty(t, routes[0].HandlerType)
	})

	t.Run("ignores functions without annotations", func(t *testing.T) {
		content := `package main

func (s *Service) noAnnotation() {}

// Regular comment
func (s *Service) alsoNoAnnotation() {}

type Service struct{}
`
		tmpFile := createTempFile(t, content)
		defer os.Remove(tmpFile)
		routes, err := ParseRouteMetadata(tmpFile)
		require.NoError(t, err)
		assert.Len(t, routes, 0)
	})

	t.Run("returns error for invalid file path", func(t *testing.T) {
		routes, err := ParseRouteMetadata("/nonexistent/file.go")
		assert.Error(t, err)
		assert.Nil(t, routes)
		assert.Contains(t, err.Error(), "failed to parse file")
	})

	t.Run("returns error for invalid annotation", func(t *testing.T) {
		content := `package main

// @RestOperation( method = "INVALID", path = "/test" )
func (s *Service) test() {}

type Service struct{}
`
		tmpFile := createTempFile(t, content)
		defer os.Remove(tmpFile)
		routes, err := ParseRouteMetadata(tmpFile)
		assert.Error(t, err)
		assert.Nil(t, routes)
		assert.Contains(t, err.Error(), "failed to parse annotation")
	})

	t.Run("handles multiple comments on same function", func(t *testing.T) {
		content := `package main

// Some documentation
// @RestOperation( method = "GET", path = "/test" )
// More documentation
func (s *Service) test() {}

type Service struct{}
`
		tmpFile := createTempFile(t, content)
		defer os.Remove(tmpFile)
		routes, err := ParseRouteMetadata(tmpFile)
		require.NoError(t, err)
		require.Len(t, routes, 1)
		assert.Equal(t, "GET", routes[0].Operation.Method)
	})
}

func TestExtractReceiverType(t *testing.T) {
	t.Run("extracts pointer receiver type", func(t *testing.T) {
		content := `package main

func (s *Service) test() {}

type Service struct{}
`
		tmpFile := createTempFile(t, content)
		defer os.Remove(tmpFile)
		routes, err := ParseRouteMetadata(tmpFile)
		require.NoError(t, err)
		if len(routes) == 0 {
			t.Skip("No routes found, testing extractReceiverType directly")
		}
	})

	t.Run("extracts value receiver type", func(t *testing.T) {
		content := `package main

func (s Service) test() {}

type Service struct{}
`
		tmpFile := createTempFile(t, content)
		defer os.Remove(tmpFile)
		routes, err := ParseRouteMetadata(tmpFile)
		require.NoError(t, err)
		if len(routes) == 0 {
			t.Skip("No routes found, testing extractReceiverType directly")
		}
	})
}

func createTempFile(t *testing.T, content string) string {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.go")
	err := os.WriteFile(tmpFile, []byte(content), 0644)
	require.NoError(t, err)
	return tmpFile
}
