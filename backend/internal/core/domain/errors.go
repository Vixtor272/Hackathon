package domain

import (
	"errors"
	"fmt"
)

// ErrorCode classifies a business error so adapters can map it to their own
// protocol (HTTP status, chat wording) without inspecting messages.
type ErrorCode string

const (
	CodeValidation           ErrorCode = "VALIDATION"
	CodeNotFound             ErrorCode = "NOT_FOUND"
	CodeInvalidState         ErrorCode = "INVALID_STATE"
	CodeInsufficientStock    ErrorCode = "INSUFFICIENT_STOCK"
	CodeRxIncreaseNotAllowed ErrorCode = "RX_INCREASE_NOT_ALLOWED"
	CodeReservationExpired   ErrorCode = "RESERVATION_EXPIRED"
	CodeEmptyCart            ErrorCode = "EMPTY_CART"
	CodePaymentInvalidated   ErrorCode = "PAYMENT_INVALIDATED"
	CodeOCRIllegible         ErrorCode = "OCR_ILLEGIBLE"
	CodePrescriptionInvalid  ErrorCode = "PRESCRIPTION_INVALID"
)

// Error is the single error type raised by the core. Messages are written in
// Spanish because they are shown to end users as-is.
type Error struct {
	Code    ErrorCode
	Message string
}

func (e *Error) Error() string { return fmt.Sprintf("%s: %s", e.Code, e.Message) }

// NewError builds a domain error with a formatted message.
func NewError(code ErrorCode, format string, args ...any) error {
	return &Error{Code: code, Message: fmt.Sprintf(format, args...)}
}

// CodeOf extracts the business code of an error, if it carries one.
func CodeOf(err error) (ErrorCode, bool) {
	var e *Error
	if errors.As(err, &e) {
		return e.Code, true
	}
	return "", false
}

// IsCode reports whether err carries the given business code.
func IsCode(err error, code ErrorCode) bool {
	got, ok := CodeOf(err)
	return ok && got == code
}

// NotFound builds the standard "not found" error for an entity.
func NotFound(entity, id string) error {
	return NewError(CodeNotFound, "%s %q no encontrado", entity, id)
}
