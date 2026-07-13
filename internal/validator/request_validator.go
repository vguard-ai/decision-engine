// Package validator implements schema-only validation for the canonical
// DecisionRequest contract.
//
// Governance compliance (Sprint B3-003):
//   - EG-001: RequestValidator takes all dependencies via its constructor
//     (NewRequestValidator). No package-level state, no service locator.
//   - EG-002: RequestValidator holds no runtime/mutable state. The only
//     field is an injected pure function (clock) used for deterministic
//     timestamp checks; Validate never mutates the validator itself.
//   - EG-003: This package validates SHAPE ONLY (required fields, formats,
//     numeric ranges). It never references fraud rules, thresholds, or any
//     configuration — that is explicitly out of scope here.
//   - EG-004: No HTTP/Kafka/framework imports. Standard library only.
package validator

import (
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/vguard-ai/decision-engine/contracts"
)

// uuidPattern matches canonical UUID v1-v5 textual representation.
// Kept as a package-level compiled regexp (immutable, not runtime state)
// for performance; this is a constant, not a global variable holding state.
var uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// currencyPattern matches a 3-letter uppercase ISO-4217-shaped code.
// Format-only check — does NOT validate against a real currency list,
// since that would be a business/reference-data concern, not schema.
var currencyPattern = regexp.MustCompile(`^[A-Z]{3}$`)

const (
	maxTransactionIDLength = 128
	maxStoreIDLength       = 64
	maxCashierIDLength     = 64
	maxItemsCount          = 500
	maxSKULength           = 128
	maxAmount              = 1_000_000_000 // defensive schema ceiling, not a business threshold
	maxFutureClockSkew     = 5 * time.Minute
)

// Validator is the contract every request validator implementation must
// satisfy. Defined here (consumer side) so callers can depend on the
// interface and inject any implementation — real or mock — per EG-001.
type Validator interface {
	Validate(req contracts.DecisionRequest) ValidationErrors
}

// RequestValidator is the production implementation of Validator for
// contracts.DecisionRequest. It is safe for concurrent use: it holds no
// mutable state (EG-002), so a single instance can be shared across
// goroutines/requests without synchronization.
type RequestValidator struct {
	now func() time.Time
}

// NewRequestValidator constructs a RequestValidator. Pass nil for now to
// use time.Now; providing an explicit clock function is recommended in
// tests for deterministic timestamp-window assertions.
func NewRequestValidator(now func() time.Time) *RequestValidator {
	if now == nil {
		now = time.Now
	}
	return &RequestValidator{now: now}
}

// Validate performs deterministic, side-effect-free schema validation of
// req. It never panics: all field access is nil/zero-value safe. Every
// applicable violation is collected and returned together (rather than
// stopping at the first failure) so a caller can surface a complete list
// to the client in a single round trip; "fail fast" here means the method
// does no I/O and no expensive work — it fails fast into a returned value,
// not via early-return-on-first-error control flow.
//
// A nil/empty return value (zero-length ValidationErrors) means req is
// schema-valid.
func (v *RequestValidator) Validate(req contracts.DecisionRequest) ValidationErrors {
	var errs ValidationErrors

	v.validateCorrelationID(req, &errs)
	v.validateTransactionID(req, &errs)
	v.validateStoreID(req, &errs)
	v.validateCashierID(req, &errs)
	v.validateAmount(req, &errs)
	v.validateCurrency(req, &errs)
	v.validateItems(req, &errs)
	v.validateTimestamp(req, &errs)

	return errs
}

func (v *RequestValidator) validateCorrelationID(req contracts.DecisionRequest, errs *ValidationErrors) {
	id := strings.TrimSpace(req.CorrelationID)
	if id == "" {
		errs.add("correlation_id", ErrCodeRequired, "correlation_id must not be empty")
		return
	}
	if !uuidPattern.MatchString(id) {
		errs.add("correlation_id", ErrCodeInvalidFormat, "correlation_id must be a valid UUID")
	}
}

func (v *RequestValidator) validateTransactionID(req contracts.DecisionRequest, errs *ValidationErrors) {
	id := strings.TrimSpace(req.TransactionID)
	if id == "" {
		errs.add("transaction_id", ErrCodeRequired, "transaction_id must not be empty")
		return
	}
	if len(id) > maxTransactionIDLength {
		errs.add("transaction_id", ErrCodeOutOfRange, "transaction_id exceeds maximum length")
	}
}

func (v *RequestValidator) validateStoreID(req contracts.DecisionRequest, errs *ValidationErrors) {
	id := strings.TrimSpace(req.StoreID)
	if id == "" {
		errs.add("store_id", ErrCodeRequired, "store_id must not be empty")
		return
	}
	if len(id) > maxStoreIDLength {
		errs.add("store_id", ErrCodeOutOfRange, "store_id exceeds maximum length")
	}
}

func (v *RequestValidator) validateCashierID(req contracts.DecisionRequest, errs *ValidationErrors) {
	id := strings.TrimSpace(req.CashierID)
	if id == "" {
		errs.add("cashier_id", ErrCodeRequired, "cashier_id must not be empty")
		return
	}
	if len(id) > maxCashierIDLength {
		errs.add("cashier_id", ErrCodeOutOfRange, "cashier_id exceeds maximum length")
	}
}

func (v *RequestValidator) validateAmount(req contracts.DecisionRequest, errs *ValidationErrors) {
	if math.IsNaN(req.Amount) || math.IsInf(req.Amount, 0) {
		errs.add("amount", ErrCodeInvalidValue, "amount must be a finite number")
		return
	}
	if req.Amount <= 0 {
		errs.add("amount", ErrCodeOutOfRange, "amount must be greater than zero")
		return
	}
	if req.Amount > maxAmount {
		errs.add("amount", ErrCodeOutOfRange, "amount exceeds maximum allowed schema value")
	}
}

func (v *RequestValidator) validateCurrency(req contracts.DecisionRequest, errs *ValidationErrors) {
	// Currency is optional at the schema level; only validate format if present.
	if req.Currency == "" {
		return
	}
	if !currencyPattern.MatchString(req.Currency) {
		errs.add("currency", ErrCodeInvalidFormat, "currency must be a 3-letter uppercase ISO-4217-shaped code")
	}
}

func (v *RequestValidator) validateItems(req contracts.DecisionRequest, errs *ValidationErrors) {
	if len(req.Items) == 0 {
		errs.add("items", ErrCodeRequired, "items must contain at least one entry")
		return
	}
	if len(req.Items) > maxItemsCount {
		errs.add("items", ErrCodeOutOfRange, "items exceeds maximum allowed count")
	}

	for i, item := range req.Items {
		prefix := "items[" + strconv.Itoa(i) + "]"

		sku := strings.TrimSpace(item.SKU)
		if sku == "" {
			errs.add(prefix+".sku", ErrCodeRequired, "sku must not be empty")
		} else if len(sku) > maxSKULength {
			errs.add(prefix+".sku", ErrCodeOutOfRange, "sku exceeds maximum length")
		}

		if item.Quantity <= 0 {
			errs.add(prefix+".quantity", ErrCodeOutOfRange, "quantity must be greater than zero")
		}

		if math.IsNaN(item.UnitPrice) || math.IsInf(item.UnitPrice, 0) {
			errs.add(prefix+".unit_price", ErrCodeInvalidValue, "unit_price must be a finite number")
		} else if item.UnitPrice < 0 {
			errs.add(prefix+".unit_price", ErrCodeOutOfRange, "unit_price must not be negative")
		}
	}
}

func (v *RequestValidator) validateTimestamp(req contracts.DecisionRequest, errs *ValidationErrors) {
	if req.Timestamp.IsZero() {
		errs.add("timestamp", ErrCodeRequired, "timestamp must not be zero-valued")
		return
	}
	now := v.now()
	if req.Timestamp.After(now.Add(maxFutureClockSkew)) {
		errs.add("timestamp", ErrCodeOutOfRange, "timestamp must not be more than 5 minutes in the future")
	}
}
