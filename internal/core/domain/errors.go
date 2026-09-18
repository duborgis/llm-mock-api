package domain

import "errors"

var (
	ErrModelNotFound  = errors.New("model not found")
	ErrNotFound       = errors.New("resource not found")
	ErrInvalidRequest = errors.New("invalid request")
	ErrScenarioFailed = errors.New("scenario error injection")
)

// InjectedError wraps an ErrorInjection so adapters can translate it into a wire-format error response.
type InjectedError struct {
	Injection ErrorInjection
}

func (e *InjectedError) Error() string {
	if e.Injection.Message != "" {
		return e.Injection.Message
	}
	return "injected error"
}
