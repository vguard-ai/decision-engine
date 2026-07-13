package audit

import (
	"context"
	"errors"
	"sync"
)

// ErrPublisherClosed is returned by Publish and Close when called on a
// Publisher that has already been closed.
var ErrPublisherClosed = errors.New("audit: publisher is closed")

// Publisher is the boundary the Decision Engine core depends on to emit
// audit events. It is intentionally minimal and transport-agnostic — a
// real deployment might implement this backed by Kafka, a message queue,
// or direct database writes, but none of that leaks into this interface
// or into the core engine (EG-004). The core only ever imports this
// interface, never a concrete transport.
type Publisher interface {
	// Publish emits a single audit event. Implementations must be safe
	// for concurrent use by multiple goroutines.
	Publish(ctx context.Context, event Event) error
	// Close releases any resources held by the publisher and rejects
	// further Publish calls. Close must be safe to call at most once
	// meaningfully; subsequent calls return ErrPublisherClosed.
	Close() error
}

// InMemoryPublisher is a stub Publisher implementation for local
// development, unit tests, and as a safe default at the composition root
// before a real transport is wired in. It performs NO I/O — no stdout, no
// stderr, no network, no disk — and never panics.
//
// It is safe for concurrent use (EG-002 adjacent: while Publisher itself
// is allowed to hold state, unlike the stateless Decision Engine core,
// that state must be correctly synchronized, which this type guarantees
// via an internal mutex).
type InMemoryPublisher struct {
	mu     sync.RWMutex
	events []Event
	closed bool
}

// NewInMemoryPublisher constructs a ready-to-use InMemoryPublisher.
func NewInMemoryPublisher() *InMemoryPublisher {
	return &InMemoryPublisher{events: make([]Event, 0)}
}

// Publish appends event to the in-memory log. Returns ctx.Err() if the
// context is already canceled/expired, and ErrPublisherClosed if Close
// has already been called. Never blocks and never performs I/O.
func (p *InMemoryPublisher) Publish(ctx context.Context, event Event) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return ErrPublisherClosed
	}
	p.events = append(p.events, event)
	return nil
}

// Close marks the publisher as closed. Further Publish calls will return
// ErrPublisherClosed. Calling Close more than once returns
// ErrPublisherClosed on the second and subsequent calls.
func (p *InMemoryPublisher) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return ErrPublisherClosed
	}
	p.closed = true
	return nil
}

// Events returns a snapshot copy of every event published so far. Safe to
// call concurrently with Publish/Close. Intended for test assertions and
// local debugging — not part of the Publisher interface itself, since
// production transports (e.g. Kafka) would not support reading back their
// own history this way.
func (p *InMemoryPublisher) Events() []Event {
	p.mu.RLock()
	defer p.mu.RUnlock()

	out := make([]Event, len(p.events))
	copy(out, p.events)
	return out
}

// compile-time assertion that InMemoryPublisher satisfies Publisher.
var _ Publisher = (*InMemoryPublisher)(nil)
