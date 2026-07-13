package audit

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

func TestInMemoryPublisher_Publish(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(p *InMemoryPublisher)
		ctx     func() context.Context
		event   Event
		wantErr error
	}{
		{
			name: "publishes successfully on an open publisher",
			ctx:  context.Background,
			event: Event{
				CorrelationID: "corr-1",
				TransactionID: "txn-1",
				Type:          EventValidationPassed,
				OccurredAt:    time.Now(),
			},
			wantErr: nil,
		},
		{
			name: "returns ErrPublisherClosed after Close",
			setup: func(p *InMemoryPublisher) {
				if err := p.Close(); err != nil {
					t.Fatalf("setup Close failed: %v", err)
				}
			},
			ctx: context.Background,
			event: Event{
				CorrelationID: "corr-2",
				Type:          EventEngineError,
			},
			wantErr: ErrPublisherClosed,
		},
		{
			name: "returns context error when context already canceled",
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			},
			event: Event{
				CorrelationID: "corr-3",
				Type:          EventDecisionMade,
			},
			wantErr: context.Canceled,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewInMemoryPublisher()
			if tt.setup != nil {
				tt.setup(p)
			}

			err := p.Publish(tt.ctx(), tt.event)

			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				if len(p.Events()) != 1 {
					t.Fatalf("expected 1 event recorded, got %d", len(p.Events()))
				}
				return
			}

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected error %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestInMemoryPublisher_Close(t *testing.T) {
	p := NewInMemoryPublisher()

	if err := p.Close(); err != nil {
		t.Fatalf("first Close() should succeed, got %v", err)
	}

	err := p.Close()
	if !errors.Is(err, ErrPublisherClosed) {
		t.Fatalf("second Close() should return ErrPublisherClosed, got %v", err)
	}
}

func TestInMemoryPublisher_Events_ReturnsSnapshotCopy(t *testing.T) {
	p := NewInMemoryPublisher()
	ctx := context.Background()

	if err := p.Publish(ctx, Event{CorrelationID: "corr-1", Type: EventDecisionMade}); err != nil {
		t.Fatalf("Publish failed: %v", err)
	}

	snapshot := p.Events()
	if len(snapshot) != 1 {
		t.Fatalf("expected 1 event, got %d", len(snapshot))
	}

	// Mutating the returned slice must not affect internal state.
	snapshot[0].CorrelationID = "mutated"

	fresh := p.Events()
	if fresh[0].CorrelationID != "corr-1" {
		t.Fatalf("internal state was mutated via returned snapshot: got %q", fresh[0].CorrelationID)
	}
}

// TestInMemoryPublisher_ConcurrentPublish exercises the publisher from
// many goroutines simultaneously. Run with -race to verify thread-safety
// (EG requirement: "Thread-safe").
func TestInMemoryPublisher_ConcurrentPublish(t *testing.T) {
	p := NewInMemoryPublisher()
	ctx := context.Background()

	const goroutines = 50
	const eventsPerGoroutine = 20

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for g := 0; g < goroutines; g++ {
		go func(id int) {
			defer wg.Done()
			for i := 0; i < eventsPerGoroutine; i++ {
				_ = p.Publish(ctx, Event{
					CorrelationID: "concurrent-test",
					Type:          EventValidationPassed,
					OccurredAt:    time.Now(),
				})
			}
		}(g)
	}

	wg.Wait()

	got := len(p.Events())
	want := goroutines * eventsPerGoroutine
	if got != want {
		t.Fatalf("expected %d events after concurrent publish, got %d", want, got)
	}
}

// TestInMemoryPublisher_ConcurrentPublishAndClose ensures Close racing
// with Publish never panics and always resolves deterministically to
// either success or ErrPublisherClosed for each call.
func TestInMemoryPublisher_ConcurrentPublishAndClose(t *testing.T) {
	p := NewInMemoryPublisher()
	ctx := context.Background()

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			_ = p.Publish(ctx, Event{CorrelationID: "race-test", Type: EventEngineError})
		}
	}()

	go func() {
		defer wg.Done()
		_ = p.Close()
	}()

	wg.Wait()
	// No assertion beyond "did not panic and did not deadlock" — the
	// race detector (go test -race) is what actually proves correctness.
}

func TestInMemoryPublisher_ImplementsPublisherInterface(t *testing.T) {
	var _ Publisher = NewInMemoryPublisher()
}
