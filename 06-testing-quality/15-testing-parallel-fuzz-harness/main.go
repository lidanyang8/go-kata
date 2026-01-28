package main

import (
	"errors"
	"fmt"
	"net/http"
)

var ErrInvalidHeaderKey = errors.New("invalid header key: contains forbidden characters")

func NormalizeHeaderKey(s string) (string, error) {
	if s == "" {
		return "", fmt.Errorf("%w: empty key", ErrInvalidHeaderKey)
	}

	for i := 0; i < len(s); i++ {
		c := s[i]
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-') {
			return "", fmt.Errorf("%w: '%c'", ErrInvalidHeaderKey, c)
		}
	}

	return http.CanonicalHeaderKey(s), nil
}

func main() {
	keys := []string{"content-type", "X-reQUEST-ID", "user-agent-123", "invalid_name", "bad@key"}
	for _, k := range keys {
		res, err := NormalizeHeaderKey(k)
		if err != nil {
			fmt.Printf("Input: %s | Error: %v\n", k, err)
		} else {
			fmt.Printf("Input: %s | Normalized: %s\n", k, res)
		}
	}
}
