package nordigen

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/antonioplacerda/requests"
)

var (
	ErrInvalidToken   = errors.New("invalid token")
	ErrRateLimitError = errors.New("daily request limit set by the Institution has been exceeded")
	ErrNotFound       = errors.New("not found")
)

func getError(err error) error {
	var re *requests.Error
	if errors.As(err, &re) {
		e := &Error{err: re}
		err := json.Unmarshal([]byte(re.Data), &e)
		if err != nil {
			return err
		}
		return e.Unwrap()
	}

	return nil
}

// Error handles the errors received from the provider.
//
// Some specific errors messages are unwrapped into errors. All the other
// received error messages that are not recognised should be sent back to the
// caller.
type Error struct {
	err        *requests.Error `json:"-'"`
	Summary    string          `json:"summary"`
	Detail     string          `json:"detail"`
	Type       string          `json:"type"`
	StatusCode int             `json:"status_code"`
}

func (e *Error) Error() string {
	if e.Summary != "" && e.Detail != "" {
		return fmt.Sprintf("nordigen error: %d: %s - %s", e.StatusCode, e.Summary, e.Detail)
	}
	if e.Summary != "" {
		return fmt.Sprintf("nordigen error: %d, %s", e.StatusCode, e.Summary)
	}
	if e.Detail != "" {
		return fmt.Sprintf("nordigen error: %d, %s", e.StatusCode, e.Detail)
	}
	return e.err.Error()
}

var (
	reOrderNotExists = regexp.MustCompile(`^OrderID \d+ doesn't exist$`)
)

func (e *Error) Unwrap() error {
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

	switch e.err.HttpStatusCode {
	case http.StatusNotFound:
		return ErrNotFound
	}
	return fmt.Errorf(e.Error())
}

func (e *Error) UnmarshalJSON(data []byte) error {
	type Clone Error

	tmp := struct {
		*Clone
	}{
		Clone: (*Clone)(e),
	}
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	e.err.Data = string(data)
	return nil
}
