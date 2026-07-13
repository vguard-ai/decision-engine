package runtime

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

// WaitForShutdown blocks until SIGINT or SIGTERM is received, then calls
// r.Stop with a timeout-bounded context derived from the parent ctx.
// This is the canonical signal-handler entry point for the Decision Engine
// process.
//
// Usage (in main):
//
//	if err := runtime.WaitForShutdown(ctx, rt); err != nil {
//	    log.Printf("shutdown error: %v", err)
//	}
func WaitForShutdown(ctx context.Context, r *Runtime) error {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	select {
	case <-sigCh:
	case <-ctx.Done():
	case <-r.Done():
		return nil
	}

	stopCtx, cancel := context.WithTimeout(context.Background(), ShutdownTimeout)
	defer cancel()
	return r.Stop(stopCtx)
}
