// Package license defines a small set of well-known open-source licenses,
// detects them from file text, and evaluates compatibility against a target
// license chosen by the project.
package license

import "strings"

// Category groups licenses by how they affect redistribution.
const (
	CatPermissive   = "permissive"   // MIT, BSD, ISC, Apache-2.0
	CatWeakCopyleft = "weak-copyleft" // LGPL
	CatCopyleft     = "copyleft"     // GPL-2.0, GPL-3.0
	CatUnknown      = "unknown"
)

// License is a known license with detection signatures.
type License struct {
	SPDX     string
	Name     string
	Category string
	// Signatures are lower-cased substrings that, when all present in a file,
	// indicate this license.
	Signatures []string
}

// KnownLicenses is the built-in detection database.
var KnownLicenses = []License{
	{
		SPDX: "MIT", Name: "MIT License", Category: CatPermissive,
		Signatures: []string{"mit license", "permission is hereby granted", "the above copyright notice"},
	},
	{
		SPDX: "Apache-2.0", Name: "Apache License 2.0", Category: CatPermissive,
		Signatures: []string{"apache license", "version 2.0", "apache software foundation"},
	},
	{
		SPDX: "BSD-3-Clause", Name: "BSD 3-Clause", Category: CatPermissive,
		Signatures: []string{"redistribution and use in source and binary forms", "bsd"},
	},
	{
		SPDX: "ISC", Name: "ISC License", Category: CatPermissive,
		Signatures: []string{"isc license", "permission to use, copy, modify"},
	},
	{
		SPDX: "LGPL-2.1", Name: "GNU Lesser General Public License v2.1", Category: CatWeakCopyleft,
		Signatures: []string{"gnu lesser general public license", "gnu library general public license"},
	},
	{
		SPDX: "GPL-2.0", Name: "GNU General Public License v2.0", Category: CatCopyleft,
		Signatures: []string{"gnu general public license", "version 2"},
	},
	{
		SPDX: "GPL-3.0", Name: "GNU General Public License v3.0", Category: CatCopyleft,
		Signatures: []string{"gnu general public license", "version 3"},
	},
}

// Detect returns the SPDX identifiers whose signatures all appear in text.
func Detect(text string) []string {
	lower := strings.ToLower(text)
	var found []string
	for _, l := range KnownLicenses {
		ok := true
		for _, sig := range l.Signatures {
			if !strings.Contains(lower, sig) {
				ok = false
				break
			}
		}
		if ok {
			found = append(found, l.SPDX)
		}
	}
	return found
}

// Category returns the category of an SPDX id, or CatUnknown.
func Category(spdx string) string {
	for _, l := range KnownLicenses {
		if l.SPDX == spdx {
			return l.Category
		}
	}
	return CatUnknown
}

// strength orders licenses by the obligations they impose on redistributors:
// permissive (0) < weak-copyleft (1) < copyleft (2). Returns -1 for unknown.
func strength(spdx string) int {
	switch Category(spdx) {
	case CatPermissive:
		return 0
	case CatWeakCopyleft:
		return 1
	case CatCopyleft:
		return 2
	default:
		return -1
	}
}

// Compatible reports whether a dependency under `other` may be incorporated
// into a project licensed `target`. The check is directional: a dependency's
// copyleft strength must not exceed the project's, otherwise the whole
// project would have to adopt the stronger license. So: can I use `other`
// code in my `target` project?
func Compatible(target, other string) (bool, string) {
	if other == "" || other == CatUnknown {
		return false, "unknown/unverified license cannot be cleared"
	}
	tS, oS := strength(target), strength(other)
	if tS < 0 {
		return false, "unknown target license"
	}
	if oS < 0 {
		return false, "unknown dependency license"
	}
	if oS <= tS {
		return true, other + " obligations do not exceed " + target
	}
	return false, other + " is stronger copyleft than target " + target
}

// Lookup returns the known license record for an SPDX id, or nil.
func Lookup(spdx string) *License {
	for i := range KnownLicenses {
		if KnownLicenses[i].SPDX == spdx {
			return &KnownLicenses[i]
		}
	}
	return nil
}

// NameOf returns the human-readable name for an SPDX id, or "" if unknown.
func NameOf(spdx string) string {
	l := Lookup(spdx)
	if l == nil {
		return ""
	}
	return l.Name
}
