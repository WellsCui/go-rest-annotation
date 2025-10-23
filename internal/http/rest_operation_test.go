package http

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseRestOperation(t *testing.T) {
	t.Run("parses valid annotation with all fields", func(t *testing.T) {
		annotation := `@RestOperation( method = "GET", path = "/person/{uid}", middlewares = ["auth" "logging"], timeout = 30, disableAuth = true )`
		op, err := ParseRestOperation(annotation)
		require.NoError(t, err)
		assert.Equal(t, "GET", op.Method)
		assert.Equal(t, "/person/{uid}", op.Path)
		assert.Equal(t, []string{"auth", "logging"}, op.Middlewares)
		assert.Equal(t, 30, op.Timeout)
		assert.True(t, op.DisableAuth)
	})

	t.Run("parses annotation with minimal fields", func(t *testing.T) {
		annotation := `@RestOperation( method = "POST", path = "/users" )`
		op, err := ParseRestOperation(annotation)
		require.NoError(t, err)
		assert.Equal(t, "POST", op.Method)
		assert.Equal(t, "/users", op.Path)
		assert.Nil(t, op.Middlewares)
		assert.Equal(t, 30, op.Timeout)
		assert.False(t, op.DisableAuth)
	})

	t.Run("parses annotation with empty middlewares", func(t *testing.T) {
		annotation := `@RestOperation( method = "PUT", path = "/items", middlewares = [] )`
		op, err := ParseRestOperation(annotation)
		require.NoError(t, err)
		assert.Equal(t, "PUT", op.Method)
		assert.Nil(t, op.Middlewares)
	})

	t.Run("parses annotation with disableAuth false", func(t *testing.T) {
		annotation := `@RestOperation( method = "DELETE", path = "/items/{id}", disableAuth = false )`
		op, err := ParseRestOperation(annotation)
		require.NoError(t, err)
		assert.False(t, op.DisableAuth)
	})

	t.Run("returns error for invalid format", func(t *testing.T) {
		annotation := `@RestOperation invalid format`
		op, err := ParseRestOperation(annotation)
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidFormat, err)
		assert.Nil(t, op)
	})

	t.Run("returns error for missing method", func(t *testing.T) {
		annotation := `@RestOperation( path = "/users" )`
		op, err := ParseRestOperation(annotation)
		assert.Error(t, err)
		assert.Equal(t, ErrMissingMethod, err)
		assert.Nil(t, op)
	})

	t.Run("returns error for missing path", func(t *testing.T) {
		annotation := `@RestOperation( method = "GET" )`
		op, err := ParseRestOperation(annotation)
		assert.Error(t, err)
		assert.Equal(t, ErrMissingPath, err)
		assert.Nil(t, op)
	})

	t.Run("returns error for invalid method", func(t *testing.T) {
		annotation := `@RestOperation( method = "INVALID", path = "/users" )`
		op, err := ParseRestOperation(annotation)
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidMethod, err)
		assert.Nil(t, op)
	})

	t.Run("returns error for invalid timeout", func(t *testing.T) {
		annotation := `@RestOperation( method = "GET", path = "/users", timeout = abc )`
		op, err := ParseRestOperation(annotation)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid timeout value")
		assert.Nil(t, op)
	})

	t.Run("parses all valid HTTP methods", func(t *testing.T) {
		methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH"}
		for _, method := range methods {
			annotation := `@RestOperation( method = "` + method + `", path = "/test" )`
			op, err := ParseRestOperation(annotation)
			require.NoError(t, err)
			assert.Equal(t, method, op.Method)
		}
	})

	t.Run("parses single middleware", func(t *testing.T) {
		annotation := `@RestOperation( method = "GET", path = "/test", middlewares = ["auth"] )`
		op, err := ParseRestOperation(annotation)
		require.NoError(t, err)
		assert.Equal(t, []string{"auth"}, op.Middlewares)
	})

	t.Run("handles whitespace in annotation", func(t *testing.T) {
		annotation := `@RestOperation(   method   =   "GET"  ,  path  =  "/test"   )`
		op, err := ParseRestOperation(annotation)
		require.NoError(t, err)
		assert.Equal(t, "GET", op.Method)
		assert.Equal(t, "/test", op.Path)
	})
}

func TestValidate(t *testing.T) {
	t.Run("validates correct operation", func(t *testing.T) {
		op := &RestOperation{
			Method:  "GET",
			Path:    "/users",
			Timeout: 30,
		}
		err := op.Validate()
		assert.NoError(t, err)
	})

	t.Run("returns error for empty method", func(t *testing.T) {
		op := &RestOperation{
			Path:    "/users",
			Timeout: 30,
		}
		err := op.Validate()
		assert.Equal(t, ErrMissingMethod, err)
	})

	t.Run("returns error for invalid method", func(t *testing.T) {
		op := &RestOperation{
			Method:  "INVALID",
			Path:    "/users",
			Timeout: 30,
		}
		err := op.Validate()
		assert.Equal(t, ErrInvalidMethod, err)
	})

	t.Run("returns error for empty path", func(t *testing.T) {
		op := &RestOperation{
			Method:  "GET",
			Timeout: 30,
		}
		err := op.Validate()
		assert.Equal(t, ErrMissingPath, err)
	})

	t.Run("returns error for negative timeout", func(t *testing.T) {
		op := &RestOperation{
			Method:  "GET",
			Path:    "/users",
			Timeout: -1,
		}
		err := op.Validate()
		assert.Equal(t, ErrInvalidTimeout, err)
	})

	t.Run("allows zero timeout", func(t *testing.T) {
		op := &RestOperation{
			Method:  "GET",
			Path:    "/users",
			Timeout: 0,
		}
		err := op.Validate()
		assert.NoError(t, err)
	})
}

func TestParseArray(t *testing.T) {
	t.Run("parses array with multiple elements", func(t *testing.T) {
		result := parseArray(`["auth" "logging" "metrics"]`)
		assert.Equal(t, []string{"auth", "logging", "metrics"}, result)
	})

	t.Run("parses array with single element", func(t *testing.T) {
		result := parseArray(`["auth"]`)
		assert.Equal(t, []string{"auth"}, result)
	})

	t.Run("parses empty array", func(t *testing.T) {
		result := parseArray(`[]`)
		assert.Nil(t, result)
	})

	t.Run("handles whitespace", func(t *testing.T) {
		result := parseArray(`[ "auth"  "logging" ]`)
		assert.Equal(t, []string{"auth", "logging"}, result)
	})

	t.Run("handles empty string", func(t *testing.T) {
		result := parseArray(``)
		assert.Nil(t, result)
	})
}
