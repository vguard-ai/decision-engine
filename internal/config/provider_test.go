package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// testConfig is a stand-in configuration shape used only by these tests.
// Its content is irrelevant — what matters is that Provider correctly
// loads, checksums, and immutably serves ANY type T.
type testConfig struct {
	Name    string   `json:"name"`
	Version int      `json:"version"`
	Tags    []string `json:"tags"`
}

// jsonUnmarshaler adapts encoding/json to the Unmarshaler signature. Using
// JSON here (rather than YAML) keeps this test suite 100% dependency-free
// and runnable fully offline — Provider's logic is decoder-agnostic by
// design (see provider.go), so this proves the same guarantees Load/Get
// will provide in production with a YAML decoder injected instead.
func jsonUnmarshaler(data []byte, v any) error {
	return json.Unmarshal(data, v)
}

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}
	return path
}

const validConfigJSON = `{"name":"vguard","version":2,"tags":["fraud","enterprise"]}`

func TestLoad(t *testing.T) {
	t.Run("loads successfully with no checksum verification requested", func(t *testing.T) {
		path := writeTempConfig(t, validConfigJSON)

		p, err := Load[testConfig](path, "", jsonUnmarshaler)
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}

		got, err := p.Get()
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}
		if got.Name != "vguard" || got.Version != 2 || len(got.Tags) != 2 {
			t.Fatalf("unexpected decoded value: %+v", got)
		}
	})

	t.Run("loads successfully with correct checksum", func(t *testing.T) {
		path := writeTempConfig(t, validConfigJSON)
		expected := Checksum([]byte(validConfigJSON))

		p, err := Load[testConfig](path, expected, jsonUnmarshaler)
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}
		if p.Checksum() != expected {
			t.Errorf("Checksum() = %q, want %q", p.Checksum(), expected)
		}
		if p.Path() != path {
			t.Errorf("Path() = %q, want %q", p.Path(), path)
		}
	})

	t.Run("checksum comparison is case-insensitive", func(t *testing.T) {
		path := writeTempConfig(t, validConfigJSON)
		expected := Checksum([]byte(validConfigJSON))

		_, err := Load[testConfig](path, upper(expected), jsonUnmarshaler)
		if err != nil {
			t.Fatalf("expected uppercase checksum to still match, got: %v", err)
		}
	})

	t.Run("fails with ErrChecksumMismatch on wrong checksum", func(t *testing.T) {
		path := writeTempConfig(t, validConfigJSON)

		_, err := Load[testConfig](path, "0000000000000000000000000000000000000000000000000000000000000000", jsonUnmarshaler)
		if err == nil {
			t.Fatal("expected checksum mismatch error, got nil")
		}
		var mismatch *ErrChecksumMismatch
		if !errors.As(err, &mismatch) {
			t.Fatalf("expected *ErrChecksumMismatch, got %T: %v", err, err)
		}
	})

	t.Run("fails with ErrEmptyPath when path is empty", func(t *testing.T) {
		_, err := Load[testConfig]("", "", jsonUnmarshaler)
		if !errors.Is(err, ErrEmptyPath) {
			t.Fatalf("expected ErrEmptyPath, got %v", err)
		}
	})

	t.Run("fails with ErrNilUnmarshaler when decoder is nil", func(t *testing.T) {
		path := writeTempConfig(t, validConfigJSON)
		_, err := Load[testConfig](path, "", nil)
		if !errors.Is(err, ErrNilUnmarshaler) {
			t.Fatalf("expected ErrNilUnmarshaler, got %v", err)
		}
	})

	t.Run("fails when file does not exist", func(t *testing.T) {
		_, err := Load[testConfig]("/nonexistent/path/config.json", "", jsonUnmarshaler)
		if err == nil {
			t.Fatal("expected error for nonexistent file, got nil")
		}
	})

	t.Run("fails fast on malformed content at Load time, not Get time", func(t *testing.T) {
		path := writeTempConfig(t, `{not valid json`)
		_, err := Load[testConfig](path, "", jsonUnmarshaler)
		if err == nil {
			t.Fatal("expected decode error at Load time, got nil")
		}
	})

	t.Run("never panics on any input combination", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("Load panicked: %v", r)
			}
		}()
		_, _ = Load[testConfig]("", "", nil)
		_, _ = Load[testConfig]("/does/not/exist", "deadbeef", jsonUnmarshaler)
	})
}

func TestLoadWithChecksumFile(t *testing.T) {
	t.Run("loads using digest from sidecar file", func(t *testing.T) {
		path := writeTempConfig(t, validConfigJSON)
		expected := Checksum([]byte(validConfigJSON))

		checksumPath := path + ".sha256"
		if err := os.WriteFile(checksumPath, []byte(expected+"\n"), 0o600); err != nil {
			t.Fatalf("failed to write checksum file: %v", err)
		}

		p, err := LoadWithChecksumFile[testConfig](path, checksumPath, jsonUnmarshaler)
		if err != nil {
			t.Fatalf("LoadWithChecksumFile failed: %v", err)
		}
		if p.Checksum() != expected {
			t.Errorf("Checksum() = %q, want %q", p.Checksum(), expected)
		}
	})

	t.Run("accepts sha256sum-style output with trailing filename", func(t *testing.T) {
		path := writeTempConfig(t, validConfigJSON)
		expected := Checksum([]byte(validConfigJSON))

		checksumPath := path + ".sha256"
		content := expected + "  config.json\n"
		if err := os.WriteFile(checksumPath, []byte(content), 0o600); err != nil {
			t.Fatalf("failed to write checksum file: %v", err)
		}

		_, err := LoadWithChecksumFile[testConfig](path, checksumPath, jsonUnmarshaler)
		if err != nil {
			t.Fatalf("expected sha256sum-style checksum file to work, got: %v", err)
		}
	})

	t.Run("fails when checksum file does not exist", func(t *testing.T) {
		path := writeTempConfig(t, validConfigJSON)
		_, err := LoadWithChecksumFile[testConfig](path, path+".missing.sha256", jsonUnmarshaler)
		if err == nil {
			t.Fatal("expected error for missing checksum file, got nil")
		}
	})

	t.Run("fails when checksum file is empty", func(t *testing.T) {
		path := writeTempConfig(t, validConfigJSON)
		checksumPath := path + ".sha256"
		if err := os.WriteFile(checksumPath, []byte("   \n"), 0o600); err != nil {
			t.Fatalf("failed to write checksum file: %v", err)
		}
		_, err := LoadWithChecksumFile[testConfig](path, checksumPath, jsonUnmarshaler)
		if err == nil {
			t.Fatal("expected error for empty checksum file, got nil")
		}
	})
}

// TestProvider_Get_IsImmutable proves that mutating a value returned by
// Get never affects the Provider's internal state or subsequent Get
// calls — the core "immutable after load" guarantee.
func TestProvider_Get_IsImmutable(t *testing.T) {
	path := writeTempConfig(t, validConfigJSON)

	p, err := Load[testConfig](path, "", jsonUnmarshaler)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	first, err := p.Get()
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	// Mutate the returned value's slice and scalar fields aggressively.
	first.Name = "mutated"
	first.Tags[0] = "mutated-tag"
	first.Tags = append(first.Tags, "extra")

	second, err := p.Get()
	if err != nil {
		t.Fatalf("second Get failed: %v", err)
	}

	if second.Name != "vguard" {
		t.Errorf("Provider state was mutated: Name = %q, want %q", second.Name, "vguard")
	}
	if second.Tags[0] != "fraud" {
		t.Errorf("Provider state was mutated: Tags[0] = %q, want %q", second.Tags[0], "fraud")
	}
	if len(second.Tags) != 2 {
		t.Errorf("Provider state was mutated: len(Tags) = %d, want 2", len(second.Tags))
	}
}

// TestProvider_NoReloadMethod is a compile-time-flavored regression guard:
// there is intentionally no Reload/Watch method on Provider. This test
// exists primarily as executable documentation; if such a method is ever
// added, the governance review for this package should be re-triggered.
func TestProvider_NoReloadMethod(t *testing.T) {
	path := writeTempConfig(t, validConfigJSON)
	p, err := Load[testConfig](path, "", jsonUnmarshaler)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Modify the underlying file after loading...
	if err := os.WriteFile(path, []byte(`{"name":"changed","version":99,"tags":["x"]}`), 0o600); err != nil {
		t.Fatalf("failed to overwrite config file: %v", err)
	}

	// ...and confirm Get() still returns the ORIGINALLY loaded value,
	// proving there is no hot reload / no re-read from disk.
	got, err := p.Get()
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if got.Name != "vguard" {
		t.Fatalf("expected Provider to retain originally loaded value, got Name=%q (hot reload should not happen)", got.Name)
	}
}

// TestProvider_Get_DefensiveErrorPath covers the defensive branch in
// Get() where re-decoding fails even though the initial Load succeeded.
// This is unreachable with a well-behaved decoder in normal operation
// (Load already proved raw decodes cleanly), but Get must still handle it
// without panicking — e.g. a decoder with non-deterministic or
// call-count-dependent behavior. This test injects exactly such a
// decoder to exercise that path deterministically.
func TestProvider_Get_DefensiveErrorPath(t *testing.T) {
	path := writeTempConfig(t, validConfigJSON)

	calls := 0
	flakyUnmarshal := func(data []byte, v any) error {
		calls++
		if calls == 1 {
			// First call happens inside Load — must succeed so Load
			// itself does not reject the file.
			return json.Unmarshal(data, v)
		}
		// Every subsequent call (i.e. from Get) fails, to exercise
		// Get's defensive error-handling path.
		return errors.New("simulated decode failure on re-decode")
	}

	p, err := Load[testConfig](path, "", flakyUnmarshal)
	if err != nil {
		t.Fatalf("Load should succeed on first decode, got: %v", err)
	}

	_, err = p.Get()
	if err == nil {
		t.Fatal("expected Get to surface the simulated re-decode error, got nil")
	}
}

func TestChecksumFile(t *testing.T) {
	path := writeTempConfig(t, validConfigJSON)

	got, err := ChecksumFile(path)
	if err != nil {
		t.Fatalf("ChecksumFile failed: %v", err)
	}
	want := Checksum([]byte(validConfigJSON))
	if got != want {
		t.Errorf("ChecksumFile() = %q, want %q", got, want)
	}
}

func TestChecksumFile_NonexistentFile(t *testing.T) {
	_, err := ChecksumFile("/nonexistent/path/config.json")
	if err == nil {
		t.Fatal("expected error for nonexistent file, got nil")
	}
}

func TestErrChecksumMismatch_Error(t *testing.T) {
	err := &ErrChecksumMismatch{Path: "config.yaml", Expected: "aaa", Actual: "bbb"}
	msg := err.Error()
	if msg == "" {
		t.Fatal("expected non-empty error message")
	}
}

// upper is a tiny test-local helper avoiding an extra import of strings
// in the test table just for one uppercase conversion.
func upper(s string) string {
	b := []byte(s)
	for i, c := range b {
		if c >= 'a' && c <= 'z' {
			b[i] = c - 32
		}
	}
	return string(b)
}
