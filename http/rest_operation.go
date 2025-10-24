package http

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type RestOperation struct {
	Method      string
	Path        string
	Middlewares []string
	Timeout     int
	DisableAuth bool
}

var (
	ErrInvalidFormat    = errors.New("invalid @RestOperation format")
	ErrMissingMethod    = errors.New("method is required")
	ErrMissingPath      = errors.New("path is required")
	ErrInvalidMethod    = errors.New("invalid HTTP method")
	ErrInvalidTimeout   = errors.New("timeout must be positive")
	validMethods        = map[string]bool{"GET": true, "POST": true, "PUT": true, "DELETE": true, "PATCH": true}
	annotationRegex     = regexp.MustCompile(`@RestOperation\s*\((.*)\)`)
	paramRegex          = regexp.MustCompile(`(\w+)\s*=\s*([^,]+)`)
)

// ParseRestOperation parses a @RestOperation annotation string into a RestOperation struct.
func ParseRestOperation(annotation string) (*RestOperation, error) {
	matches := annotationRegex.FindStringSubmatch(annotation)
	if len(matches) < 2 {
		return nil, ErrInvalidFormat
	}
	paramsStr := matches[1]
	op := &RestOperation{Timeout: 30}
	paramMatches := paramRegex.FindAllStringSubmatch(paramsStr, -1)
	for _, match := range paramMatches {
		if len(match) < 3 {
			continue
		}
		key := strings.TrimSpace(match[1])
		value := strings.TrimSpace(match[2])
		switch key {
		case "method":
			op.Method = strings.Trim(value, `"`)
		case "path":
			op.Path = strings.Trim(value, `"`)
		case "middlewares":
			op.Middlewares = parseArray(value)
		case "timeout":
			t, err := strconv.Atoi(value)
			if err != nil {
				return nil, fmt.Errorf("invalid timeout value: %w", err)
			}
			op.Timeout = t
		case "disableAuth":
			op.DisableAuth = value == "true"
		}
	}
	if err := op.Validate(); err != nil {
		return nil, err
	}
	return op, nil
}

// Validate checks if the RestOperation has valid values.
func (r *RestOperation) Validate() error {
	if r.Method == "" {
		return ErrMissingMethod
	}
	if !validMethods[r.Method] {
		return ErrInvalidMethod
	}
	if r.Path == "" {
		return ErrMissingPath
	}
	if r.Timeout < 0 {
		return ErrInvalidTimeout
	}
	return nil
}

func parseArray(s string) []string {
	s = strings.Trim(s, "[]")
	if s == "" {
		return nil
	}
	parts := strings.Split(s, " ")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		part = strings.Trim(part, `"`)
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}
