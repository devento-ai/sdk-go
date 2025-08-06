package devento

import (
	"fmt"
)

type DeventoError struct {
	Message    string
	StatusCode int
	Code       string
}

func (e *DeventoError) Error() string {
	return e.Message
}

type AuthenticationError struct {
	DeventoError
}

func NewAuthenticationError(message string) *AuthenticationError {
	return &AuthenticationError{
		DeventoError: DeventoError{
			Message:    message,
			StatusCode: 401,
			Code:       "authentication_error",
		},
	}
}

type BoxNotFoundError struct {
	DeventoError
	BoxID string
}

func NewBoxNotFoundError(boxID string) *BoxNotFoundError {
	return &BoxNotFoundError{
		DeventoError: DeventoError{
			Message:    fmt.Sprintf("Box not found: %s", boxID),
			StatusCode: 404,
			Code:       "box_not_found",
		},
		BoxID: boxID,
	}
}

type CommandTimeoutError struct {
	DeventoError
	CommandID string
	Timeout   int // milliseconds
}

func NewCommandTimeoutError(commandID string, timeout int) *CommandTimeoutError {
	return &CommandTimeoutError{
		DeventoError: DeventoError{
			Message:    fmt.Sprintf("Command %s timed out after %dms", commandID, timeout),
			StatusCode: 408,
			Code:       "command_timeout",
		},
		CommandID: commandID,
		Timeout:   timeout,
	}
}

type BoxTimeoutError struct {
	DeventoError
	BoxID   string
	Timeout int // seconds
}

func NewBoxTimeoutError(boxID string, timeout int) *BoxTimeoutError {
	return &BoxTimeoutError{
		DeventoError: DeventoError{
			Message:    fmt.Sprintf("Box %s failed to become ready within %d seconds", boxID, timeout),
			StatusCode: 408,
			Code:       "box_timeout",
		},
		BoxID:   boxID,
		Timeout: timeout,
	}
}

type RateLimitError struct {
	DeventoError
	RetryAfter int // seconds
}

func NewRateLimitError(retryAfter int) *RateLimitError {
	return &RateLimitError{
		DeventoError: DeventoError{
			Message:    fmt.Sprintf("Rate limit exceeded. Retry after %d seconds", retryAfter),
			StatusCode: 429,
			Code:       "rate_limit",
		},
		RetryAfter: retryAfter,
	}
}

type ValidationError struct {
	DeventoError
	Field string
}

func NewValidationError(field, message string) *ValidationError {
	return &ValidationError{
		DeventoError: DeventoError{
			Message:    fmt.Sprintf("Validation error on field '%s': %s", field, message),
			StatusCode: 400,
			Code:       "validation_error",
		},
		Field: field,
	}
}

type InsufficientCreditsError struct {
	DeventoError
	Required  float64
	Available float64
}

func NewInsufficientCreditsError(required, available float64) *InsufficientCreditsError {
	return &InsufficientCreditsError{
		DeventoError: DeventoError{
			Message:    fmt.Sprintf("Insufficient credits. Required: %.2f, Available: %.2f", required, available),
			StatusCode: 402,
			Code:       "insufficient_credits",
		},
		Required:  required,
		Available: available,
	}
}

type APIError struct {
	DeventoError
}

func NewAPIError(statusCode int, message string) *APIError {
	return &APIError{
		DeventoError: DeventoError{
			Message:    message,
			StatusCode: statusCode,
			Code:       "api_error",
		},
	}
}

func parseError(statusCode int, errResp *errorResponse) error {
	message := errResp.Message
	if message == "" {
		message = errResp.Error
	}

	switch statusCode {
	case 401:
		return NewAuthenticationError(message)
	case 402:
		return NewAPIError(statusCode, message)
	case 404:
		if errResp.Code == "box_not_found" {
			return &BoxNotFoundError{
				DeventoError: DeventoError{
					Message:    message,
					StatusCode: statusCode,
					Code:       errResp.Code,
				},
			}
		}
		return NewAPIError(statusCode, message)
	case 429:
		return NewRateLimitError(0) // TODO: Parse Retry-After header
	case 400:
		if errResp.Code == "validation_error" {
			return NewValidationError("", message)
		}
		return NewAPIError(statusCode, message)
	default:
		return NewAPIError(statusCode, message)
	}
}
