package config

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
)

// ErrChecksumMismatch is returned when the computed checksum of a config
// file does not match the expected checksum supplied by the caller.
type ErrChecksumMismatch struct {
	Path     string
	Expected string
	Actual   string
}

func (e *ErrChecksumMismatch) Error() string {
	return fmt.Sprintf("config: checksum mismatch for %q: expected %s, got %s", e.Path, e.Expected, e.Actual)
}

// Checksum computes the lowercase hex-encoded SHA-256 digest of raw bytes.
// This is the canonical checksum format used throughout this package —
// callers producing an expected checksum (e.g. as part of a release
// pipeline) should use this same function to generate it.
func Checksum(raw []byte) string {
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

// ChecksumFile computes the Checksum of the file at path. Intended as a
// build/release-time helper (e.g. a small `go run` tool that prints the
// checksum of a config file to commit alongside it) rather than for use
// in the hot request path.
func ChecksumFile(path string) (string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("config: reading file for checksum %q: %w", path, err)
	}
	return Checksum(raw), nil
}

// readChecksumFile reads a sidecar checksum file (convention:
// "<config>.sha256") and returns the trimmed hex digest it contains.
// The file may contain either a bare digest or the common
// "<digest>  <filename>" format produced by tools like sha256sum; both
// are accepted by taking only the first whitespace-delimited token.
func readChecksumFile(path string) (string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("config: reading checksum file %q: %w", path, err)
	}
	fields := strings.Fields(string(raw))
	if len(fields) == 0 {
		return "", fmt.Errorf("config: checksum file %q is empty", path)
	}
	return strings.ToLower(fields[0]), nil
}
