package odoorpc

import (
	"errors"
	"fmt"
)

// Predefined errors.
var (
	ErrInvalidAuthResponse = errors.New("odoorpc: invalid authentication response")
)

// OdooError represents an error response from Odoo server
type OdooError struct {
	Code    int                    `json:"code"`
	Message string                 `json:"message"`
	Data    map[string]interface{} `json:"data"`
}

func (e *OdooError) Error() string {
	return fmt.Sprintf("odoo error %d: %s", e.Code, e.Message)
}

// OdooErrorType represents different categories of Odoo errors
type OdooErrorType string

const (
	ErrorTypeUserError       OdooErrorType = "odoo.exceptions.UserError"
	ErrorTypeValidationError OdooErrorType = "odoo.exceptions.ValidationError"
	ErrorTypeAccessError     OdooErrorType = "odoo.exceptions.AccessError"
	ErrorTypeMissingError    OdooErrorType = "odoo.exceptions.MissingError"
	ErrorTypeRedirectWarning OdooErrorType = "odoo.exceptions.RedirectWarning"
	ErrorTypeWarning         OdooErrorType = "odoo.exceptions.Warning"
	ErrorTypeExceptOlm       OdooErrorType = "odoo.exceptions.ExceptOlm"
	ErrorTypeCacheMiss       OdooErrorType = "odoo.exceptions.CacheMiss"
	ErrorTypeUserErrorORM    OdooErrorType = "odoo.exceptions.UserError (ORM)"
)

// OdooServerResponse represents a response from Odoo server that may contain nested errors
type OdooServerResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Error   *OdooError  `json:"error"`
	Result  interface{} `json:"result"`
}

// Specific error types for better error handling
type OdooValidationError struct {
	*OdooError
}

type OdooAccessError struct {
	*OdooError
}

type OdooUserError struct {
	*OdooError
}

func NewOdooError(code int, message string, data map[string]interface{}) error {
	baseErr := &OdooError{
		Code:    code,
		Message: message,
		Data:    data,
	}

	// Extract exception type from data if available
	if exceptionType, ok := data["name"].(string); ok {
		switch OdooErrorType(exceptionType) {
		case ErrorTypeValidationError:
			return &OdooValidationError{baseErr}
		case ErrorTypeAccessError:
			return &OdooAccessError{baseErr}
		case ErrorTypeUserError, ErrorTypeUserErrorORM:
			return &OdooUserError{baseErr}
		default:
			return baseErr
		}
	}

	return baseErr
}

// IsOdooError checks if error is an Odoo error
func IsOdooError(err error) bool {
	_, ok := err.(*OdooError)
	_, ok2 := err.(*OdooValidationError)
	_, ok3 := err.(*OdooAccessError)
	_, ok4 := err.(*OdooUserError)
	return ok || ok2 || ok3 || ok4
}

// GetOdooError extracts Odoo error from wrapped error
func GetOdooError(err error) (*OdooError, bool) {
	switch e := err.(type) {
	case *OdooError:
		return e, true
	case *OdooValidationError:
		return e.OdooError, true
	case *OdooAccessError:
		return e.OdooError, true
	case *OdooUserError:
		return e.OdooError, true
	default:
		return nil, false
	}
}
