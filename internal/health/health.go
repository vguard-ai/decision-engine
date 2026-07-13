// Package health provides the HealthChecker interface and its default
// in-process implementation used by the Runtime.
//
// This package is infrastructure-only. It carries no business logic.
package health

import (
	"sync"
)

// Status represents the health status of a single named component.
type Status uint8

const (
	// StatusUnknown is the zero value: component has not reported yet.
	StatusUnknown Status = iota
	// StatusHealthy: component reported healthy.
	StatusHealthy
	// StatusUnhealthy: component reported unhealthy or has been stopped.
	StatusUnhealthy
)

func (s Status) String() string {
	switch s {
	case StatusHealthy:
		return "HEALTHY"
	case StatusUnhealthy:
		return "UNHEALTHY"
	default:
		return "UNKNOWN"
	}
}

// ComponentHealth holds the last-known status of one subsystem.
type ComponentHealth struct {
	Name   string
	Status Status
}

// HealthChecker is the interface used by the Runtime to track and query
// the health of registered subsystems.
type HealthChecker interface {
	MarkHealthy(name string)
	MarkUnhealthy(name string)
	// IsHealthy returns true only when every registered component is healthy.
	IsHealthy() bool
	// Components returns a snapshot of all component statuses.
	Components() []ComponentHealth
}

// InProcessHealthChecker is the default HealthChecker implementation.
// It is safe for concurrent use.
type InProcessHealthChecker struct {
	mu         sync.RWMutex
	components map[string]Status
}

// New returns a ready-to-use InProcessHealthChecker.
func New() *InProcessHealthChecker {
	return &InProcessHealthChecker{
		components: make(map[string]Status),
	}
}

// MarkHealthy records the named component as healthy.
func (h *InProcessHealthChecker) MarkHealthy(name string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.components[name] = StatusHealthy
}

// MarkUnhealthy records the named component as unhealthy.
func (h *InProcessHealthChecker) MarkUnhealthy(name string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.components[name] = StatusUnhealthy
}

// IsHealthy returns true when every tracked component is StatusHealthy.
// Returns true (vacuously) when no components have been registered yet.
func (h *InProcessHealthChecker) IsHealthy() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, s := range h.components {
		if s != StatusHealthy {
			return false
		}
	}
	return true
}

// Components returns a stable-order snapshot of all component statuses.
func (h *InProcessHealthChecker) Components() []ComponentHealth {
	h.mu.RLock()
	defer h.mu.RUnlock()
	out := make([]ComponentHealth, 0, len(h.components))
	for name, status := range h.components {
		out = append(out, ComponentHealth{Name: name, Status: status})
	}
	return out
}
