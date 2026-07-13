package validator

import "strings"

// ErrorCode is a closed, contract-stable enumeration of validation failure
// reasons. Callers (HTTP handlers, gRPC servers, etc.) can safely switch on
// these codes without depending on Message text, which is human-oriented
// and may change wording over time.
type ErrorCode string

const (
	// ErrCodeRequired indicates a required field was empty/zero-valued.
	ErrCodeRequired ErrorCode = "REQUIRED"
	// ErrCodeInvalidFormat indicates a field did not match its expected
	// shape (e.g. malformed UUID, wrong currency code length).
	ErrCodeInvalidFormat ErrorCode = "INVALID_FORMAT"
	// ErrCodeOutOfRange indicates a numeric field was outside the
	// allowed schema bounds (e.g. negative amount, zero quantity).
	ErrCodeOutOfRange ErrorCode = "OUT_OF_RANGE"
	// ErrCodeInvalidValue indicates a structural/semantic schema issue
	// that doesn't fit REQUIRED/FORMAT/RANGE (e.g. empty items slice).
	ErrCodeInvalidValue ErrorCode = "INVALID_VALUE"
)

// ValidationError describes exactly one field-level schema violation.
// It intentionally carries no business-logic context (no risk scores,
// no rule references) — schema validation only, per EG-003.
type ValidationError struct {
	// Field is the dot-path to the offending field, e.g. "items[2].quantity".
	Field string `json:"field"`
	// Code is a stable machine-readable failure reason.
	Code ErrorCode `json:"code"`
	// Message is a human-readable explanation, safe to display to
	// operators but not guaranteed stable across versions.
	Message string `json:"message"`
}

// Error implements the standard library error interface so a single
// ValidationError can be used anywhere a plain error is expected.
func (e ValidationError) Error() string {
	return e.Field + ": " + e.Message + " (" + string(e.Code) + ")"
}

// ValidationErrors is an ordered collection of ValidationError. Order is
// deterministic (field declaration order in DecisionRequest), which is
// required for reproducible tests and stable client-facing output.
type ValidationErrors []ValidationError

// Error implements the standard library error interface, joining all
// individual field errors into one human-readable message. Returns an
// empty string when there are no errors — callers should prefer HasErrors()
// for control flow rather than checking Error() == "".
func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return ""
	}
	parts := make([]string, 0, len(e))
	for _, fe := range e {
		parts = append(parts, fe.Error())
	}
	return strings.Join(parts, "; ")
}

// HasErrors reports whether the collection contains at least one error.
// This is the preferred way to check validation outcome:
//
//	errs := v.Validate(req)
//	if errs.HasErrors() { ... }
func (e ValidationErrors) HasErrors() bool {
	return len(e) > 0
}

// add is an unexported convenience used only within this package to keep
// request_validator.go free of repetitive struct literals.
func (e *ValidationErrors) add(field string, code ErrorCode, message string) {
	*e = append(*e, ValidationError{Field: field, Code: code, Message: message})
}
