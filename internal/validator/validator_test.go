package validator

import (
	"testing"
	"time"

	"github.com/vguard-ai/decision-engine/contracts"
)

// fixedClock returns a deterministic clock function pinned to t, so tests
// never depend on wall-clock time (required for 100% deterministic tests).
func fixedClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

// validRequest returns a DecisionRequest that satisfies every schema rule.
// Individual test cases mutate a copy of this to isolate exactly one
// violation at a time (plus a few combined-violation cases).
func validRequest(now time.Time) contracts.DecisionRequest {
	return contracts.DecisionRequest{
		CorrelationID: "a1b2c3d4-e5f6-47a8-b9c0-d1e2f3a4b5c6",
		TransactionID: "TRX-00123",
		StoreID:       "STORE-01",
		CashierID:     "CASHIER-07",
		Amount:        150000,
		Currency:      "IDR",
		Items: []contracts.Item{
			{SKU: "SKU-001", Name: "Item A", Quantity: 2, UnitPrice: 50000},
			{SKU: "SKU-002", Name: "Item B", Quantity: 1, UnitPrice: 50000},
		},
		Timestamp: now.Add(-1 * time.Minute),
	}
}

func TestRequestValidator_Validate(t *testing.T) {
	now := time.Date(2026, 7, 13, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name       string
		mutate     func(contracts.DecisionRequest) contracts.DecisionRequest
		wantFields []string // field paths expected to have at least one error
		wantValid  bool
	}{
		{
			name:      "fully valid request produces no errors",
			mutate:    func(r contracts.DecisionRequest) contracts.DecisionRequest { return r },
			wantValid: true,
		},
		{
			name: "empty correlation_id",
			mutate: func(r contracts.DecisionRequest) contracts.DecisionRequest {
				r.CorrelationID = ""
				return r
			},
			wantFields: []string{"correlation_id"},
		},
		{
			name: "whitespace-only correlation_id treated as empty",
			mutate: func(r contracts.DecisionRequest) contracts.DecisionRequest {
				r.CorrelationID = "   "
				return r
			},
			wantFields: []string{"correlation_id"},
		},
		{
			name: "malformed correlation_id",
			mutate: func(r contracts.DecisionRequest) contracts.DecisionRequest {
				r.CorrelationID = "not-a-uuid"
				return r
			},
			wantFields: []string{"correlation_id"},
		},
		{
			name: "empty transaction_id",
			mutate: func(r contracts.DecisionRequest) contracts.DecisionRequest {
				r.TransactionID = ""
				return r
			},
			wantFields: []string{"transaction_id"},
		},
		{
			name: "transaction_id too long",
			mutate: func(r contracts.DecisionRequest) contracts.DecisionRequest {
				long := make([]byte, maxTransactionIDLength+1)
				for i := range long {
					long[i] = 'a'
				}
				r.TransactionID = string(long)
				return r
			},
			wantFields: []string{"transaction_id"},
		},
		{
			name: "empty store_id",
			mutate: func(r contracts.DecisionRequest) contracts.DecisionRequest {
				r.StoreID = ""
				return r
			},
			wantFields: []string{"store_id"},
		},
		{
			name: "store_id too long",
			mutate: func(r contracts.DecisionRequest) contracts.DecisionRequest {
				long := make([]byte, maxStoreIDLength+1)
				for i := range long {
					long[i] = 'a'
				}
				r.StoreID = string(long)
				return r
			},
			wantFields: []string{"store_id"},
		},
		{
			name: "empty cashier_id",
			mutate: func(r contracts.DecisionRequest) contracts.DecisionRequest {
				r.CashierID = ""
				return r
			},
			wantFields: []string{"cashier_id"},
		},
		{
			name: "cashier_id too long",
			mutate: func(r contracts.DecisionRequest) contracts.DecisionRequest {
				long := make([]byte, maxCashierIDLength+1)
				for i := range long {
					long[i] = 'a'
				}
				r.CashierID = string(long)
				return r
			},
			wantFields: []string{"cashier_id"},
		},
		{
			name: "zero amount",
			mutate: func(r contracts.DecisionRequest) contracts.DecisionRequest {
				r.Amount = 0
				return r
			},
			wantFields: []string{"amount"},
		},
		{
			name: "negative amount",
			mutate: func(r contracts.DecisionRequest) contracts.DecisionRequest {
				r.Amount = -100
				return r
			},
			wantFields: []string{"amount"},
		},
		{
			name: "NaN amount",
			mutate: func(r contracts.DecisionRequest) contracts.DecisionRequest {
				r.Amount = nan()
				return r
			},
			wantFields: []string{"amount"},
		},
		{
			name: "Inf amount",
			mutate: func(r contracts.DecisionRequest) contracts.DecisionRequest {
				r.Amount = inf()
				return r
			},
			wantFields: []string{"amount"},
		},
		{
			name: "amount exceeds schema ceiling",
			mutate: func(r contracts.DecisionRequest) contracts.DecisionRequest {
				r.Amount = maxAmount + 1
				return r
			},
			wantFields: []string{"amount"},
		},
		{
			name: "empty currency is allowed (optional field)",
			mutate: func(r contracts.DecisionRequest) contracts.DecisionRequest {
				r.Currency = ""
				return r
			},
			wantValid: true,
		},
		{
			name: "malformed currency",
			mutate: func(r contracts.DecisionRequest) contracts.DecisionRequest {
				r.Currency = "idr"
				return r
			},
			wantFields: []string{"currency"},
		},
		{
			name: "malformed currency wrong length",
			mutate: func(r contracts.DecisionRequest) contracts.DecisionRequest {
				r.Currency = "IDRX"
				return r
			},
			wantFields: []string{"currency"},
		},
		{
			name: "empty items slice",
			mutate: func(r contracts.DecisionRequest) contracts.DecisionRequest {
				r.Items = nil
				return r
			},
			wantFields: []string{"items"},
		},
		{
			name: "too many items",
			mutate: func(r contracts.DecisionRequest) contracts.DecisionRequest {
				items := make([]contracts.Item, maxItemsCount+1)
				for i := range items {
					items[i] = contracts.Item{SKU: "SKU", Quantity: 1, UnitPrice: 1}
				}
				r.Items = items
				return r
			},
			wantFields: []string{"items"},
		},
		{
			name: "item with empty sku",
			mutate: func(r contracts.DecisionRequest) contracts.DecisionRequest {
				r.Items[0].SKU = ""
				return r
			},
			wantFields: []string{"items[0].sku"},
		},
		{
			name: "item sku too long",
			mutate: func(r contracts.DecisionRequest) contracts.DecisionRequest {
				long := make([]byte, maxSKULength+1)
				for i := range long {
					long[i] = 'a'
				}
				r.Items[0].SKU = string(long)
				return r
			},
			wantFields: []string{"items[0].sku"},
		},
		{
			name: "item with zero quantity",
			mutate: func(r contracts.DecisionRequest) contracts.DecisionRequest {
				r.Items[0].Quantity = 0
				return r
			},
			wantFields: []string{"items[0].quantity"},
		},
		{
			name: "item with negative quantity",
			mutate: func(r contracts.DecisionRequest) contracts.DecisionRequest {
				r.Items[0].Quantity = -5
				return r
			},
			wantFields: []string{"items[0].quantity"},
		},
		{
			name: "item with negative unit_price",
			mutate: func(r contracts.DecisionRequest) contracts.DecisionRequest {
				r.Items[0].UnitPrice = -1
				return r
			},
			wantFields: []string{"items[0].unit_price"},
		},
		{
			name: "item with NaN unit_price",
			mutate: func(r contracts.DecisionRequest) contracts.DecisionRequest {
				r.Items[0].UnitPrice = nan()
				return r
			},
			wantFields: []string{"items[0].unit_price"},
		},
		{
			name: "item with zero unit_price is allowed (e.g. free promo item)",
			mutate: func(r contracts.DecisionRequest) contracts.DecisionRequest {
				r.Items[0].UnitPrice = 0
				return r
			},
			wantValid: true,
		},
		{
			name: "second item invalid produces indexed field path",
			mutate: func(r contracts.DecisionRequest) contracts.DecisionRequest {
				r.Items[1].Quantity = 0
				return r
			},
			wantFields: []string{"items[1].quantity"},
		},
		{
			name: "zero-valued timestamp",
			mutate: func(r contracts.DecisionRequest) contracts.DecisionRequest {
				r.Timestamp = time.Time{}
				return r
			},
			wantFields: []string{"timestamp"},
		},
		{
			name: "timestamp too far in the future",
			mutate: func(r contracts.DecisionRequest) contracts.DecisionRequest {
				r.Timestamp = now.Add(10 * time.Minute)
				return r
			},
			wantFields: []string{"timestamp"},
		},
		{
			name: "timestamp within allowed clock skew is valid",
			mutate: func(r contracts.DecisionRequest) contracts.DecisionRequest {
				r.Timestamp = now.Add(4 * time.Minute)
				return r
			},
			wantValid: true,
		},
		{
			name: "timestamp exactly at now is valid",
			mutate: func(r contracts.DecisionRequest) contracts.DecisionRequest {
				r.Timestamp = now
				return r
			},
			wantValid: true,
		},
		{
			name: "multiple simultaneous violations are all reported",
			mutate: func(r contracts.DecisionRequest) contracts.DecisionRequest {
				r.CorrelationID = ""
				r.Amount = -1
				r.Items = nil
				return r
			},
			wantFields: []string{"correlation_id", "amount", "items"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewRequestValidator(fixedClock(now))
			req := tt.mutate(validRequest(now))

			errs := v.Validate(req)

			if tt.wantValid {
				if errs.HasErrors() {
					t.Fatalf("expected no errors, got: %v", errs)
				}
				return
			}

			if !errs.HasErrors() {
				t.Fatalf("expected errors for fields %v, got none", tt.wantFields)
			}

			for _, wantField := range tt.wantFields {
				if !containsField(errs, wantField) {
					t.Errorf("expected error for field %q, got errors: %v", wantField, errs)
				}
			}
		})
	}
}

// TestRequestValidator_NeverPanics feeds a completely zero-valued request
// (worst case: nil slices, zero time, empty strings) to guarantee the
// validator never panics regardless of how malformed the input is.
func TestRequestValidator_NeverPanics(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Validate panicked on zero-value request: %v", r)
		}
	}()

	v := NewRequestValidator(nil) // exercise the nil-clock default path too
	_ = v.Validate(contracts.DecisionRequest{})
}

// TestRequestValidator_DefaultClock exercises NewRequestValidator(nil) to
// ensure the default time.Now clock path is covered.
func TestRequestValidator_DefaultClock(t *testing.T) {
	v := NewRequestValidator(nil)
	req := validRequest(time.Now())
	errs := v.Validate(req)
	if errs.HasErrors() {
		t.Fatalf("expected valid request to pass with default clock, got: %v", errs)
	}
}

func TestValidationErrors_ErrorString(t *testing.T) {
	var empty ValidationErrors
	if got := empty.Error(); got != "" {
		t.Errorf("expected empty string for no errors, got %q", got)
	}

	errs := ValidationErrors{
		{Field: "amount", Code: ErrCodeOutOfRange, Message: "must be greater than zero"},
	}
	got := errs.Error()
	want := "amount: must be greater than zero (OUT_OF_RANGE)"
	if got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

func TestValidationErrors_HasErrors(t *testing.T) {
	var empty ValidationErrors
	if empty.HasErrors() {
		t.Error("expected HasErrors() false for empty collection")
	}
	nonEmpty := ValidationErrors{{Field: "x", Code: ErrCodeRequired, Message: "required"}}
	if !nonEmpty.HasErrors() {
		t.Error("expected HasErrors() true for non-empty collection")
	}
}

// --- test helpers -----------------------------------------------------

func containsField(errs ValidationErrors, field string) bool {
	for _, e := range errs {
		if e.Field == field {
			return true
		}
	}
	return false
}

func nan() float64 {
	var zero float64
	return zero / zero //nolint:staticcheck // intentional NaN construction for test fixture
}

func inf() float64 {
	var zero float64
	one := 1.0
	return one / zero //nolint:staticcheck // intentional +Inf construction for test fixture
}
