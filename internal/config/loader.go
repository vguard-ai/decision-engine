// Package config provides the configuration loader bootstrap for the
// V-Guard Decision Engine runtime.
//
// Scope (B3-004): loader construction and env-override resolution only.
// No business config keys (thresholds, policies, rules) live here.
package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"
)

// Source enumerates where a config value was resolved from.
type Source string

const (
	SourceDefault Source = "default"
	SourceEnv     Source = "env"
)

// Entry holds a resolved configuration value with provenance.
type Entry struct {
	Value  string
	Source Source
}

// Loader resolves configuration values from environment variables with
// typed defaults. It is immutable after construction (all defaults are
// baked in at New time) and safe for concurrent reads.
type Loader struct {
	entries map[string]Entry
}

// New constructs a Loader by resolving each key against the environment.
// defaults maps key → default value string; env vars of the same name
// (uppercased) override the default.
func New(defaults map[string]string) *Loader {
	entries := make(map[string]Entry, len(defaults))
	for k, def := range defaults {
		envKey := toEnvKey(k)
		if v, ok := os.LookupEnv(envKey); ok {
			entries[k] = Entry{Value: v, Source: SourceEnv}
		} else {
			entries[k] = Entry{Value: def, Source: SourceDefault}
		}
	}
	return &Loader{entries: entries}
}

// Get returns the resolved Entry for key. Returns an error if key was not
// registered in defaults.
func (l *Loader) Get(key string) (Entry, error) {
	e, ok := l.entries[key]
	if !ok {
		return Entry{}, fmt.Errorf("config: unknown key %q", key)
	}
	return e, nil
}

// String returns the string value for key, or the fallback if the key is
// unknown.
func (l *Loader) String(key, fallback string) string {
	e, err := l.Get(key)
	if err != nil {
		return fallback
	}
	return e.Value
}

// Int returns the integer value for key. Returns an error if the value
// cannot be parsed as a base-10 int.
func (l *Loader) Int(key string, fallback int) (int, error) {
	e, err := l.Get(key)
	if err != nil {
		return fallback, nil
	}
	v, err := strconv.Atoi(e.Value)
	if err != nil {
		return fallback, fmt.Errorf("config: key %q value %q is not an integer: %w", key, e.Value, err)
	}
	return v, nil
}

// Duration returns the time.Duration for key, parsed via time.ParseDuration.
func (l *Loader) Duration(key string, fallback time.Duration) (time.Duration, error) {
	e, err := l.Get(key)
	if err != nil {
		return fallback, nil
	}
	d, err := time.ParseDuration(e.Value)
	if err != nil {
		return fallback, fmt.Errorf("config: key %q value %q is not a duration: %w", key, e.Value, err)
	}
	return d, nil
}

// Validate checks that all required keys are non-empty. Returns a joined
// error listing every missing key.
func (l *Loader) Validate(required ...string) error {
	var errs []error
	for _, k := range required {
		e, err := l.Get(k)
		if err != nil || e.Value == "" {
			errs = append(errs, fmt.Errorf("config: required key %q is missing or empty", k))
		}
	}
	return errors.Join(errs...)
}

// toEnvKey converts "some.key" → "SOME_KEY".
func toEnvKey(key string) string {
	out := make([]byte, len(key))
	for i := 0; i < len(key); i++ {
		c := key[i]
		if c == '.' || c == '-' {
			out[i] = '_'
		} else if c >= 'a' && c <= 'z' {
			out[i] = c - 32
		} else {
			out[i] = c
		}
	}
	return string(out)
}
