package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunCheckDetectsUnformatted(t *testing.T) {
	dir := t.TempDir()
	name := filepath.Join(dir, "t.md")
	if err := os.WriteFile(name, []byte("|a|b|\n|-|-|\n|1|2|\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if code := run([]string{"-c", "-q", name}); code != 1 {
		t.Fatalf("-c on unformatted file = %d, want 1", code)
	}
	if code := run([]string{"-c", "-w", name}); code != 2 {
		t.Fatalf("-c -w = %d, want 2", code)
	}
	if code := run([]string{"-w", "-q", name}); code != 0 {
		t.Fatalf("-w = %d, want 0", code)
	}
	data, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	if want := "| a   | b   |\n| --- | --- |\n| 1   | 2   |\n"; string(data) != want {
		t.Fatalf("file after -w = %q, want %q", data, want)
	}
	if code := run([]string{"-c", "-q", name}); code != 0 {
		t.Fatalf("-c on formatted file = %d, want 0", code)
	}
}

func TestRunWriteRequiresFile(t *testing.T) {
	if code := run([]string{"-w"}); code != 2 {
		t.Fatalf("run(-w) = %d, want 2", code)
	}
}

func TestRunMissingFile(t *testing.T) {
	if code := run([]string{filepath.Join(t.TempDir(), "absent.md")}); code != 2 {
		t.Fatalf("run(missing) = %d, want 2", code)
	}
}

func TestRunNegativeSortColumn(t *testing.T) {
	if code := run([]string{"-s", "-1"}); code != 2 {
		t.Fatalf("run(-s -1) = %d, want 2", code)
	}
}

func TestRunVersion(t *testing.T) {
	if code := run([]string{"-V"}); code != 0 {
		t.Fatalf("run(-V) = %d, want 0", code)
	}
}
