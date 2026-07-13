// Package runtime provides the lifecycle infrastructure for the V-Guard
// Decision Engine. It is scope-limited to bootstrap, start, stop, and
// health surface — no business logic lives here.
//
// Sprint B3-004 scope (execution-only, per Engineering Governor Directive):
//   - Runtime bootstrap
//   - Runtime lifecycle (Start / Stop)
//   - Health checker integration
//   - Graceful shutdown
//   - Configuration loader bootstrap
//   - Audit runtime bootstrap
//
// Out of scope (do NOT add here):
//   - Policy Resolver
//   - Evidence Evaluator
//   - Risk Evaluator
//   - Action Resolver
package runtime

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

// State represents the lifecycle state of the Runtime.
type State uint8

const (
	// StateIdle is the zero value: Runtime is constructed but not started.
	StateIdle State = iota
	// StateStarting: Start has been called; bootstrap is in progress.
	StateStarting
	// StateRunning: all subsystems are healthy and serving.
	StateRunning
	// StateStopping: Stop has been called; graceful drain is in progress.
	StateStopping
	// StateStopped: shutdown is complete.
	StateStopped
)

func (s State) String() string {
	switch s {
	case StateIdle:
		return "IDLE"
	case StateStarting:
		return "STARTING"
	case StateRunning:
		return "RUNNING"
	case StateStopping:
		return "STOPPING"
	case StateStopped:
		return "STOPPED"
	default:
		return fmt.Sprintf("UNKNOWN(%d)", s)
	}
}

// ShutdownTimeout is the maximum time the runtime will wait for in-flight
// work to drain before forcibly terminating subsystems.
const ShutdownTimeout = 30 * time.Second

// Subsystem is the interface every engine subsystem must implement to
// participate in the managed lifecycle. Implementations must be safe for
// concurrent use; Start and Stop may be called from any goroutine.
type Subsystem interface {
	// Name returns the stable identifier used in logs and health output.
	Name() string
	// Start initialises the subsystem. It blocks until the subsystem is
	// ready to serve or ctx is cancelled. A non-nil error aborts startup.
	Start(ctx context.Context) error
	// Stop drains and releases all resources. It must return within
	// ShutdownTimeout when ctx is cancelled.
	Stop(ctx context.Context) error
}

// Runtime orchestrates the startup and shutdown of all registered
// subsystems. It is the single entry point for lifecycle management.
//
// Runtime is NOT safe for concurrent Start/Stop calls; callers must
// serialise lifecycle transitions externally (e.g. signal handler loop).
type Runtime struct {
	mu         sync.RWMutex
	state      State
	subsystems []Subsystem
	health     HealthChecker
	audit      AuditLogger
	stopOnce   sync.Once
	stopCh     chan struct{}
}

// New constructs a Runtime with the supplied subsystems registered in
// start order. health and audit must not be nil.
func New(health HealthChecker, audit AuditLogger, subsystems ...Subsystem) (*Runtime, error) {
	if health == nil {
		return nil, errors.New("runtime: HealthChecker must not be nil")
	}
	if audit == nil {
		return nil, errors.New("runtime: AuditLogger must not be nil")
	}
	return &Runtime{
		state:      StateIdle,
		subsystems: subsystems,
		health:     health,
		audit:      audit,
		stopCh:     make(chan struct{}),
	}, nil
}

// State returns the current lifecycle state. Safe for concurrent reads.
func (r *Runtime) State() State {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.state
}

// Start transitions the runtime from Idle → Starting → Running.
// It starts subsystems in registration order; the first failure aborts
// startup and attempts to stop already-started subsystems in reverse order.
func (r *Runtime) Start(ctx context.Context) error {
	r.mu.Lock()
	if r.state != StateIdle {
		r.mu.Unlock()
		return fmt.Errorf("runtime: Start called in state %s (want IDLE)", r.state)
	}
	r.state = StateStarting
	r.mu.Unlock()

	r.audit.Log(ctx, AuditEvent{Kind: EventRuntimeStarting})

	started := make([]Subsystem, 0, len(r.subsystems))
	for _, s := range r.subsystems {
		r.audit.Log(ctx, AuditEvent{Kind: EventSubsystemStarting, Subsystem: s.Name()})
		if err := s.Start(ctx); err != nil {
			r.audit.Log(ctx, AuditEvent{Kind: EventSubsystemFailed, Subsystem: s.Name(), Err: err})
			// Best-effort rollback of already-started subsystems.
			_ = r.stopSubsystems(ctx, started)
			r.mu.Lock()
			r.state = StateStopped
			r.mu.Unlock()
			return fmt.Errorf("runtime: subsystem %q failed to start: %w", s.Name(), err)
		}
		r.audit.Log(ctx, AuditEvent{Kind: EventSubsystemStarted, Subsystem: s.Name()})
		r.health.MarkHealthy(s.Name())
		started = append(started, s)
	}

	r.mu.Lock()
	r.state = StateRunning
	r.mu.Unlock()

	r.audit.Log(ctx, AuditEvent{Kind: EventRuntimeRunning})
	return nil
}

// Stop transitions the runtime from Running → Stopping → Stopped.
// Subsystems are stopped in reverse registration order. Stop is
// idempotent: subsequent calls are no-ops.
func (r *Runtime) Stop(ctx context.Context) error {
	var err error
	r.stopOnce.Do(func() {
		r.mu.Lock()
		if r.state != StateRunning {
			r.mu.Unlock()
			return
		}
		r.state = StateStopping
		r.mu.Unlock()

		r.audit.Log(ctx, AuditEvent{Kind: EventRuntimeStopping})

		// Stop in reverse order.
		reversed := make([]Subsystem, len(r.subsystems))
		copy(reversed, r.subsystems)
		for i, j := 0, len(reversed)-1; i < j; i, j = i+1, j-1 {
			reversed[i], reversed[j] = reversed[j], reversed[i]
		}

		err = r.stopSubsystems(ctx, reversed)

		r.mu.Lock()
		r.state = StateStopped
		r.mu.Unlock()

		r.audit.Log(ctx, AuditEvent{Kind: EventRuntimeStopped, Err: err})
		close(r.stopCh)
	})
	return err
}

// Done returns a channel that is closed when the runtime has fully stopped.
func (r *Runtime) Done() <-chan struct{} {
	return r.stopCh
}

// stopSubsystems calls Stop on each subsystem, collecting errors.
func (r *Runtime) stopSubsystems(ctx context.Context, subsystems []Subsystem) error {
	var errs []error
	for _, s := range subsystems {
		r.audit.Log(ctx, AuditEvent{Kind: EventSubsystemStopping, Subsystem: s.Name()})
		r.health.MarkUnhealthy(s.Name())
		if err := s.Stop(ctx); err != nil {
			r.audit.Log(ctx, AuditEvent{Kind: EventSubsystemFailed, Subsystem: s.Name(), Err: err})
			errs = append(errs, fmt.Errorf("subsystem %q: %w", s.Name(), err))
		} else {
			r.audit.Log(ctx, AuditEvent{Kind: EventSubsystemStopped, Subsystem: s.Name()})
		}
	}
	return errors.Join(errs...)
}
