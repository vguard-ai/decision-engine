// Package contracts defines the canonical wire contracts for the V-Guard
// Decision Engine.
//
// # Contract Governance
//
// All types in this package are governed by VG-CONTRACT-001 and the
// Production Architecture Baseline (PAB-1). No field, tag, or type may
// be changed without a new ADR and an explicit version bump of
// ContractVersion.
//
// # Immutability
//
// Contract types are value types. Once a DecisionRequest is constructed
// and handed to the Decision Engine, callers MUST NOT mutate any field.
// The engine treats every incoming request as an immutable snapshot.
//
// # Semantic Versioning
//
// ContractVersion follows semver (MAJOR.MINOR.PATCH):
//   - MAJOR: breaking change (field removed, type changed, tag renamed).
//   - MINOR: backwards-compatible addition (new optional field).
//   - PATCH: documentation or comment change only — no wire impact.
package contracts

import "time"

// ContractName is the stable identifier for this contract.
// Referenced by ADR-017, ADR-018, ADR-019, ADR-020 and VG-CONTRACT-001.
const (
	ContractName    = "DecisionRequest"
	ContractVersion = "1.0.0"
)

// DecisionRequest is the canonical input to the Decision Engine.
//
// # Immutability
//
// DecisionRequest is an immutable snapshot of a transaction at the moment
// it enters the Decision Engine. Callers must treat every field as
// read-only after the request is submitted. In particular, the Items slice
// must not be appended to, truncated, or have its elements mutated once the
// request has been handed off — doing so produces undefined behaviour in
// concurrent evaluation pipelines.
//
// # Versioning
//
// This type is frozen at ContractVersion 1.0.0 per VG-CONTRACT-001
// (Status: FROZEN). Any structural change requires a new ADR and a
// ContractVersion bump before merging.
//
// # Serialization
//
// All fields carry explicit json struct tags (VG-SERIALIZATION-001).
// time.Time fields serialize to RFC 3339 UTC. float64 fields must be
// finite (non-NaN, non-Inf) — the validator enforces this before the
// request reaches the engine.
//
// It is intentionally free of any business-logic fields (risk scores,
// thresholds, etc.) — those are produced BY the engine, never received.
type DecisionRequest struct {
	CorrelationID string    `json:"correlation_id"`
	TransactionID string    `json:"transaction_id"`
	StoreID       string    `json:"store_id"`
	CashierID     string    `json:"cashier_id"`
	Amount        float64   `json:"amount"`
	Currency      string    `json:"currency"`
	// Items is an immutable snapshot of the line items present in the
	// transaction at intake time. Callers must not modify the slice or
	// any of its elements after passing the request to the engine.
	Items     []Item    `json:"items"`
	Timestamp time.Time `json:"timestamp"`
}

// Item represents a single line item within a transaction.
//
// Like DecisionRequest, Item is immutable after construction.
// UnitPrice must be a finite, non-negative float64.
type Item struct {
	SKU       string  `json:"sku"`
	Name      string  `json:"name"`
	Quantity  int     `json:"quantity"`
	UnitPrice float64 `json:"unit_price"`
}
