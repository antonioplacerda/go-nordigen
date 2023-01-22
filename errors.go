package go_nordigen

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

var (
	ErrInvalidToken   = errors.New("invalid token")
	ErrRateLimitError = errors.New("daily request limit set by the Institution has been exceeded")
	ErrNotFound       = errors.New("not found")
)

// RequestError handles the errors received from the provider.
//
// Some specific errors messages are unwrapped into errors. All the other
// received error messages that are not recognised should be sent back to the
// caller.
type RequestError struct {
	Summary        string `json:"summary"`
	Detail         string `json:"detail"`
	Type           string `json:"type"`
	StatusCode     int    `json:"status_code"`
	HttpStatusCode int    `json:"-"`
}

func (e *RequestError) Error() string {
	if e.Detail == "" {
		return fmt.Sprintf("nordigen error: %d, %s", e.StatusCode, e.Summary)
	}
	return fmt.Sprintf("nordigen error: %d: %s - %s", e.StatusCode, e.Summary, e.Detail)
}

var (
	reOrderNotExists = regexp.MustCompile(`^OrderID \d+ doesn't exist$`)
)

func (e *RequestError) Unwrap() error {
	switch strings.ToLower(e.Type) {
	case "RateLimitError":
		return ErrRateLimitError
	}

	switch strings.ToLower(e.Summary) {
	case "invalid token":
		return ErrInvalidToken
	}

	if e.Detail != "" {
		if reOrderNotExists.MatchString(e.Detail) {
			return ErrNotFound
		}
	}

	switch e.HttpStatusCode {
	case http.StatusNotFound:
		return ErrNotFound
	}
	return fmt.Errorf(e.Error())
}
