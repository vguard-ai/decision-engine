// Package config implements a generic, format-agnostic configuration
// loader with mandatory checksum validation and load-once immutability.
//
// Governance compliance (Sprint B3-003):
//   - EG-001: The decoder (Unmarshaler) is injected via Load's parameters
//     rather than imported directly by this package — no service locator,
//     no hidden global registry of decoders.
//   - EG-002: Provider itself is a leaf value holder, not the stateless
//     Decision Engine core; nonetheless it upholds immutability strictly
//     (see Get doc comment) so it is safe to share as a long-lived
//     singleton dependency injected into stateless consumers.
//   - EG-003: This package has no awareness of what the configuration
//     MEANS (fraud rules, thresholds, etc.) — it is a generic T loader.
//   - EG-004: No YAML/JSON library is imported directly by this file; the
//     caller supplies the decoding function, keeping this package
//     dependency-free and framework-agnostic. See yaml_provider.go (in
//     the deployment package, added separately) for the production YAML
//     wiring.
package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// ErrEmptyPath is returned when Load is called with an empty path.
var ErrEmptyPath = errors.New("config: path must not be empty")

// ErrNilUnmarshaler is returned when Load is called with a nil decoder.
var ErrNilUnmarshaler = errors.New("config: unmarshal function must not be nil")

// Unmarshaler decodes raw bytes into the value pointed to by v. This is
// the seam that keeps this package free of any dependency on a specific
// serialization format — inject yaml.Unmarshal for production YAML
// configs, json.Unmarshal for JSON/tests, or any other compatible decoder.
type Unmarshaler func(data []byte, v any) error

// Provider loads a configuration value of type T exactly once, verifies
// its integrity via SHA-256 checksum, and exposes it thereafter as an
// immutable value via Get. There is intentionally NO Reload/Watch method
// on this type — hot reload is out of scope by design; a new deployment
// (and therefore a new process, and therefore a new Load call) is the
// only supported way to pick up configuration changes.
type Provider[T any] struct {
	path      string
	checksum  string
	raw       []byte
	unmarshal Unmarshaler
}

// Load reads the file at path, verifies its SHA-256 checksum against
// expectedChecksum (produced via Checksum or ChecksumFile), and validates
// that it decodes successfully with unmarshal. It returns an error —
// never a panic — for any I/O, checksum, or decode failure.
//
// Passing an empty expectedChecksum skips checksum verification; this is
// intended only for local development against files that change
// frequently, never for production deployments.
func Load[T any](path string, expectedChecksum string, unmarshal Unmarshaler) (*Provider[T], error) {
	if path == "" {
		return nil, ErrEmptyPath
	}
	if unmarshal == nil {
		return nil, ErrNilUnmarshaler
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("config: reading %q: %w", path, err)
	}

	actual := Checksum(raw)
	if expectedChecksum != "" && !strings.EqualFold(actual, expectedChecksum) {
		return nil, &ErrChecksumMismatch{Path: path, Expected: expectedChecksum, Actual: actual}
	}

	// Decode once up front purely to validate the file is well-formed —
	// fail Load() immediately rather than deferring a decode error to
	// the first Get() call, which would violate "fail fast".
	var probe T
	if err := unmarshal(raw, &probe); err != nil {
		return nil, fmt.Errorf("config: decoding %q: %w", path, err)
	}

	return &Provider[T]{
		path:      path,
		checksum:  actual,
		raw:       raw,
		unmarshal: unmarshal,
	}, nil
}

// LoadWithChecksumFile is a convenience wrapper around Load that reads
// the expected checksum from a sidecar file (convention: "<path>.sha256",
// compatible with both a bare hex digest and `sha256sum`-style output)
// instead of requiring the caller to pass the digest explicitly.
func LoadWithChecksumFile[T any](path string, checksumPath string, unmarshal Unmarshaler) (*Provider[T], error) {
	expected, err := readChecksumFile(checksumPath)
	if err != nil {
		return nil, err
	}
	return Load[T](path, expected, unmarshal)
}

// Get returns the loaded configuration value. It re-decodes from the
// Provider's retained immutable raw bytes on every call, rather than
// returning a single stored struct by reference. This guarantees true
// immutability regardless of what mutable structures (slices, maps,
// pointers) T contains: a caller can freely mutate the value returned by
// Get without ever affecting the Provider's internal state or any other
// caller's independently-decoded copy.
//
// The decode error path is unreachable in normal operation, because Load
// already proved raw decodes successfully into T; it is preserved here so
// Get remains panic-free even under adversarial use (e.g. a
// non-deterministic custom Unmarshaler), consistent with "fail fast, no
// panic".
func (p *Provider[T]) Get() (T, error) {
	var value T
	if err := p.unmarshal(p.raw, &value); err != nil {
		var zero T
		return zero, fmt.Errorf("config: re-decoding cached config for %q: %w", p.path, err)
	}
	return value, nil
}

// Checksum returns the verified SHA-256 checksum of the loaded file.
func (p *Provider[T]) Checksum() string {
	return p.checksum
}

// Path returns the filesystem path this Provider was loaded from.
func (p *Provider[T]) Path() string {
	return p.path
}
