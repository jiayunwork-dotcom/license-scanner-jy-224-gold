// Package scan walks a source tree, detects license declarations and judges
// their compatibility against a project target license.
package scan

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"license-scanner/internal/license"
)

// Finding is one detected license in one file.
type Finding struct {
	Path       string `json:"path"`
	License    string `json:"license"`
	Compatible bool   `json:"compatible"`
	Reason     string `json:"reason"`
}

var licenseFileNames = map[string]bool{
	"license": true, "license.md": true, "license.txt": true,
	"copying": true, "copying.md": true, "copying.txt": true,
}

var sourceExts = map[string]bool{
	".go": true, ".py": true, ".js": true, ".ts": true, ".java": true,
	".c": true, ".h": true, ".cpp": true, ".rb": true, ".rs": true,
}

// Scan returns a finding for every file that declares a known license.
func Scan(root, target string) ([]Finding, error) {
	var findings []Finding
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		name := strings.ToLower(info.Name())
		isLicenseFile := licenseFileNames[name]
		isSource := sourceExts[strings.ToLower(filepath.Ext(path))]
		if !isLicenseFile && !isSource {
			return nil
		}
		f, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer f.Close()
		buf, err := io.ReadAll(io.LimitReader(f, 1<<20))
		if err != nil {
			return err
		}
		det := license.Detect(string(buf))
		if len(det) == 0 {
			if isLicenseFile {
				findings = append(findings, Finding{
					Path: path, License: "unknown", Compatible: false,
					Reason: "LICENSE file present but no known signature",
				})
			}
			return nil
		}
		lic := det[0]
		ok, reason := license.Compatible(target, lic)
		rel, rerr := filepath.Rel(root, path)
		if rerr != nil {
			rel = path
		}
		findings = append(findings, Finding{
			Path: rel, License: lic, Compatible: ok, Reason: reason,
		})
		return nil
	})
	return findings, err
}

// Incompatible returns a new slice of findings that are not compatible with
// the scan target. The input slice is left unchanged.
func Incompatible(in []Finding) []Finding {
	out := make([]Finding, 0, len(in))
	for _, f := range in {
		if !f.Compatible {
			out = append(out, f)
		}
	}
	return out
}
