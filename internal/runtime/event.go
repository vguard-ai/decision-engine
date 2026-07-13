// Package audit defines the audit event contract and publisher interface
// used by the Decision Engine to emit an immutable trail of everything it
// does, without coupling the core engine to any specific transport
// (Kafka, HTTP, file, etc.) — per EG-004.
package audit

import "time"

// EventType is a closed enumeration of audit event kinds the Decision
// Engine core is aware of. Downstream consumers (e.g. a Kafka-backed
// Publisher implementation) may recognize additional types, but the core
// only ever emits these.
type EventType string

const (
	EventValidationFailed EventType = "VALIDATION_FAILED"
	EventValidationPassed EventType = "VALIDATION_PASSED"
	EventDecisionMade     EventType = "DECISION_MADE"
	EventEngineError      EventType = "ENGINE_ERROR"
)

// Event is a single immutable audit record. It carries no behavior — pure
// data, safe to serialize, safe to share across goroutines once
// constructed (fields should not be mutated after creation).
type Event struct {
	// CorrelationID ties this event back to a single end-to-end request,
	// consistent with the correlation_id used across the wider V-Guard
	// platform (POS intake, Fraud Processing Engine, dashboards).
	CorrelationID string
	// TransactionID identifies the underlying business transaction, when
	// known. May be empty for events that occur before a transaction ID
	// is available (e.g. a request that fails validation before parsing).
	TransactionID string
	// Type is the kind of event being recorded.
	Type EventType
	// Payload carries event-specific structured detail. Callers are
	// responsible for ensuring payload values are themselves safe for
	// concurrent read access (e.g. avoid storing pointers to mutable
	// state); prefer plain values or already-serialized data.
	Payload map[string]any
	// OccurredAt is the timestamp the event was generated. Set by the
	// caller (not the publisher) so it reflects when the fact became
	// true, not when it was transported/persisted.
	OccurredAt time.Time
}
