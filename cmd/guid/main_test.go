package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/schigh/guid"
)

var binaryPath string

func TestMain(m *testing.M) {
	tmp, err := os.MkdirTemp("", "guid-test")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tmp)

	binaryPath = filepath.Join(tmp, "guid")
	cmd := exec.Command("go", "build", "-o", binaryPath, "github.com/schigh/guid/cmd/guid")
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		panic("failed to build binary: " + err.Error())
	}

	os.Exit(m.Run())
}

func runBinary(t *testing.T, args ...string) (stdout, stderr string, exitCode int) {
	t.Helper()
	cmd := exec.Command(binaryPath, args...)
	var outBuf, errBuf strings.Builder
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err := cmd.Run()
	exitCode = 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			t.Fatalf("failed to run binary: %v", err)
		}
	}
	return outBuf.String(), errBuf.String(), exitCode
}

func TestDefaultGeneration(t *testing.T) {
	stdout, _, code := runBinary(t)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	// should produce exactly one GUID
	stdout = strings.TrimSpace(stdout)
	if len(stdout) != 28 {
		t.Fatalf("expected GUID length %d, got %d: %q", 28, len(stdout), stdout)
	}
	// should be parseable
	if _, err := guid.ParseString(stdout); err != nil {
		t.Fatalf("output is not a valid GUID: %v", err)
	}
}

func TestMultipleGeneration(t *testing.T) {
	stdout, _, code := runBinary(t, "-n", "5")
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	if len(lines) != 5 {
		t.Fatalf("expected 5 GUIDs, got %d: %q", len(lines), stdout)
	}
	seen := make(map[string]struct{})
	for _, line := range lines {
		if _, err := guid.ParseString(line); err != nil {
			t.Fatalf("invalid GUID %q: %v", line, err)
		}
		if _, exists := seen[line]; exists {
			t.Fatalf("duplicate GUID: %s", line)
		}
		seen[line] = struct{}{}
	}
}

func TestSerialGeneration(t *testing.T) {
	stdout, _, code := runBinary(t, "-n", "3", "-serial")
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 GUIDs, got %d", len(lines))
	}
	for _, line := range lines {
		if _, err := guid.ParseString(line); err != nil {
			t.Fatalf("invalid GUID %q: %v", line, err)
		}
	}
}

func TestCustomSeparator(t *testing.T) {
	stdout, _, code := runBinary(t, "-n", "3", "-sep", ",")
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	parts := strings.Split(stdout, ",")
	if len(parts) != 3 {
		t.Fatalf("expected 3 comma-separated GUIDs, got %d: %q", len(parts), stdout)
	}
	for _, part := range parts {
		if _, err := guid.ParseString(part); err != nil {
			t.Fatalf("invalid GUID %q: %v", part, err)
		}
	}
}

func TestSlugOutput(t *testing.T) {
	stdout, _, code := runBinary(t, "-slug")
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	stdout = strings.TrimSpace(stdout)
	if len(stdout) != 12 {
		t.Fatalf("expected slug length 12, got %d: %q", len(stdout), stdout)
	}
}

func TestOutputToFile(t *testing.T) {
	tmp, err := os.CreateTemp("", "guid-output-*.txt")
	if err != nil {
		t.Fatal(err)
	}
	tmp.Close()
	defer os.Remove(tmp.Name())

	_, _, code := runBinary(t, "-n", "2", "-o", tmp.Name())
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}

	data, err := os.ReadFile(tmp.Name())
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 GUIDs in file, got %d", len(lines))
	}
	for _, line := range lines {
		if _, err := guid.ParseString(line); err != nil {
			t.Fatalf("invalid GUID in file %q: %v", line, err)
		}
	}
}

func TestScanJSON(t *testing.T) {
	// first generate a GUID
	genOut, _, code := runBinary(t, "-n", "1")
	if code != 0 {
		t.Fatalf("generation failed with exit code %d", code)
	}
	guidStr := strings.TrimSpace(genOut)

	// scan it with JSON output
	stdout, _, code := runBinary(t, "-scan", guidStr, "-json")
	if code != 0 {
		t.Fatalf("scan failed with exit code %d", code)
	}

	var result map[string]string
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("invalid JSON output: %v\nraw: %q", err, stdout)
	}

	for _, key := range []string{"prefix", "timestamp", "fingerprint", "counter", "random"} {
		if _, ok := result[key]; !ok {
			t.Fatalf("missing key %q in JSON output", key)
		}
	}
}

func TestScanText(t *testing.T) {
	genOut, _, code := runBinary(t, "-n", "1")
	if code != 0 {
		t.Fatal("generation failed")
	}
	guidStr := strings.TrimSpace(genOut)

	// scan outputs to stderr in text mode
	_, stderr, code := runBinary(t, "-scan", guidStr)
	if code != 0 {
		t.Fatalf("scan failed with exit code %d", code)
	}

	for _, label := range []string{"PREFIX", "TIMESTAMP", "FINGERPRINT", "COUNTER", "RANDOM"} {
		if !strings.Contains(stderr, label) {
			t.Fatalf("expected %q in scan output, got: %q", label, stderr)
		}
	}
}

func TestScanInvalidGUID(t *testing.T) {
	_, _, code := runBinary(t, "-scan", "not-a-guid")
	if code == 0 {
		t.Fatal("expected non-zero exit code for invalid GUID scan")
	}
}

func TestZeroCount(t *testing.T) {
	// -n 0 should still produce 1 GUID
	stdout, _, code := runBinary(t, "-n", "0")
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	stdout = strings.TrimSpace(stdout)
	if len(stdout) != 28 {
		t.Fatalf("expected 1 GUID, got output length %d", len(stdout))
	}
}
