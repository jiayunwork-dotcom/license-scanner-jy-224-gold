package scan

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScan(t *testing.T) {
	root := t.TempDir()
	write := func(rel, content string) {
		p := filepath.Join(root, rel)
		os.MkdirAll(filepath.Dir(p), 0o755)
		os.WriteFile(p, []byte(content), 0o644)
	}
	write("LICENSE", "MIT License\nPermission is hereby granted, and the above copyright notice")
	write("lib/code.go", "// Apache License\n// Version 2.0\n// Apache Software Foundation\npackage lib")
	write("cmd/main.go", "package main\nfunc main() {}\n") // no license -> skipped

	findings, err := Scan(root, "Apache-2.0")
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 2 {
		t.Fatalf("want 2 findings, got %d: %+v", len(findings), findings)
	}
	byPath := map[string]Finding{}
	for _, f := range findings {
		byPath[f.Path] = f
	}
	mit := byPath["LICENSE"]
	if mit.License != "MIT" || !mit.Compatible {
		t.Errorf("LICENSE finding wrong: %+v", mit)
	}
	ap := byPath["lib/code.go"]
	if ap.License != "Apache-2.0" || !ap.Compatible {
		t.Errorf("code.go finding wrong: %+v", ap)
	}
}

func TestScanIncompatible(t *testing.T) {
	root := t.TempDir()
	p := filepath.Join(root, "LICENSE")
	os.WriteFile(p, []byte("GNU General Public License\nVersion 3"), 0o644)
	findings, err := Scan(root, "MIT")
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 {
		t.Fatalf("want 1, got %d", len(findings))
	}
	if findings[0].Compatible {
		t.Errorf("MIT target should be incompatible with GPL-3.0: %+v", findings[0])
	}
}

func TestScanUnreadableFile(t *testing.T) {
	root := t.TempDir()
	p := filepath.Join(root, "LICENSE")
	if err := os.WriteFile(p, []byte("MIT License\nPermission is hereby granted, and the above copyright notice"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(p, 0); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(p, 0o644) })

	_, err := Scan(root, "Apache-2.0")
	if err == nil {
		t.Fatal("expected error when LICENSE is unreadable, got nil")
	}
}

func TestIncompatibleOwnsBackingArray(t *testing.T) {
	in := []Finding{
		{Path: "ok.go", License: "MIT", Compatible: true},
		{Path: "bad.go", License: "GPL-3.0", Compatible: false},
		{Path: "also.go", License: "Apache-2.0", Compatible: true},
	}
	// Keep a snapshot of the input to detect in-place mutation / shared backing.
	origSecond := in[1]

	out := Incompatible(in)
	if len(out) != 1 || out[0].Path != "bad.go" {
		t.Fatalf("want only bad.go, got %+v", out)
	}
	if in[0].Path != "ok.go" || in[1] != origSecond || in[2].Path != "also.go" {
		t.Fatalf("Incompatible must not mutate input, got %+v", in)
	}
	// Append within the returned slice's capacity; a shared-backing filter
	// would overwrite leftover input slots.
	_ = append(out, Finding{Path: "hijack", License: "X", Compatible: false})
	if in[1].Path == "hijack" {
		t.Fatal("append to result corrupted input — shared backing array")
	}
}
