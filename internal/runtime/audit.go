package runtime

import (
	"context"
	"fmt"
	"time"
)

// EventKind is a stable enumeration of runtime lifecycle audit events.
type EventKind string

const (
	EventRuntimeStarting   EventKind = "RUNTIME_STARTING"
	EventRuntimeRunning    EventKind = "RUNTIME_RUNNING"
	EventRuntimeStopping   EventKind = "RUNTIME_STOPPING"
	EventRuntimeStopped    EventKind = "RUNTIME_STOPPED"
	EventSubsystemStarting EventKind = "SUBSYSTEM_STARTING"
	EventSubsystemStarted  EventKind = "SUBSYSTEM_STARTED"
	EventSubsystemStopping EventKind = "SUBSYSTEM_STOPPING"
	EventSubsystemStopped  EventKind = "SUBSYSTEM_STOPPED"
	EventSubsystemFailed   EventKind = "SUBSYSTEM_FAILED"
)

// AuditEvent is a single structured lifecycle event emitted by the Runtime.
type AuditEvent struct {
	Kind      EventKind
	Subsystem string // empty for runtime-level events
	Err       error  // nil unless Kind is a failure event
	At        time.Time
}

// AuditLogger is the interface the Runtime uses to emit structured audit
// events. Implementations must be safe for concurrent use.
type AuditLogger interface {
	Log(ctx context.Context, event AuditEvent)
}

// NoopAuditLogger discards all events. Useful in tests that do not need
// audit output.
type NoopAuditLogger struct{}

func (NoopAuditLogger) Log(_ context.Context, _ AuditEvent) {}

// StdoutAuditLogger writes events to stdout in a structured line format.
// Intended for development and CI environments.
type StdoutAuditLogger struct{}

func (StdoutAuditLogger) Log(_ context.Context, e AuditEvent) {
	at := e.At
	if at.IsZero() {
		at = time.Now().UTC()
	}
	if e.Err != nil {
		fmt.Printf("[AUDIT] %s kind=%s subsystem=%q err=%q\n",
			at.Format(time.RFC3339), e.Kind, e.Subsystem, e.Err.Error())
	} else {
		fmt.Printf("[AUDIT] %s kind=%s subsystem=%q\n",
			at.Format(time.RFC3339), e.Kind, e.Subsystem)
	}
}
