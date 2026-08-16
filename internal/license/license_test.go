package license

import "testing"

func TestDetect(t *testing.T) {
	if got := Detect("Apache License\nVersion 2.0\nApache Software Foundation"); !contains(got, "Apache-2.0") {
		t.Errorf("expected Apache-2.0, got %v", got)
	}
	if got := Detect("MIT License\nPermission is hereby granted, and the above copyright notice"); !contains(got, "MIT") {
		t.Errorf("expected MIT, got %v", got)
	}
	if got := Detect("GNU General Public License\nVersion 3"); !contains(got, "GPL-3.0") {
		t.Errorf("expected GPL-3.0, got %v", got)
	}
	if got := Detect("totally unrelated text"); len(got) != 0 {
		t.Errorf("expected no detection, got %v", got)
	}
}

func TestCompatible(t *testing.T) {
	cases := []struct {
		target, other string
		want          bool
	}{
		{"Apache-2.0", "MIT", true},
		{"Apache-2.0", "GPL-3.0", false}, // permissive target cannot absorb copyleft
		{"GPL-3.0", "MIT", true},
		{"GPL-3.0", "Apache-2.0", true},
		{"MIT", "GPL-3.0", false},     // permissive target can't take GPL
		{"LGPL-2.1", "GPL-3.0", false}, // weak-copyleft can't take strong copyleft
		{"Apache-2.0", "unknown", false},
		{"Apache-2.0", "Apache-2.0", true},
	}
	for _, c := range cases {
		ok, _ := Compatible(c.target, c.other)
		if ok != c.want {
			t.Errorf("Compatible(%s,%s)=%v want %v", c.target, c.other, ok, c.want)
		}
	}
}

func TestCompatibleSameStrength(t *testing.T) {
	// Same-category (both permissive) must be allowed: obligations do not exceed.
	ok, reason := Compatible("Apache-2.0", "MIT")
	if !ok {
		t.Fatalf("Apache-2.0 project should accept MIT dependency, got ok=%v reason=%q", ok, reason)
	}
}

func TestDetectCaseInsensitive(t *testing.T) {
	text := "MIT LICENSE\nPERMISSION IS HEREBY GRANTED\nTHE ABOVE COPYRIGHT NOTICE"
	got := Detect(text)
	if !contains(got, "MIT") {
		t.Fatalf("expected MIT from uppercase license text, got %v", got)
	}
}

func TestCompatibleApacheRejectsLGPL(t *testing.T) {
	ok, reason := Compatible("Apache-2.0", "LGPL-2.1")
	if ok {
		t.Fatalf("Apache-2.0 must not absorb LGPL-2.1, got ok=true reason=%q", reason)
	}
}

func TestNameOfUnknown(t *testing.T) {
	if got := NameOf("NOT-A-REAL-SPDX"); got != "" {
		t.Fatalf("unknown SPDX should yield empty name, got %q", got)
	}
	if got := NameOf("MIT"); got == "" {
		t.Fatal("MIT should have a non-empty display name")
	}
}

func contains(s []string, v string) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}
