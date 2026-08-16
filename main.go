// Command license-scanner scans a source tree for license declarations and
// reports their compatibility with a target project license.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"license-scanner/internal/scan"
)

func main() {
	root := flag.String("path", ".", "root directory to scan")
	target := flag.String("target", "Apache-2.0", "project target license (SPDX id)")
	format := flag.String("format", "text", "output format: text | json")
	flag.Parse()

	findings, err := scan.Scan(*root, *target)
	if err != nil {
		fatal("scan: %v", err)
	}

	if *format == "json" {
		b, err := json.MarshalIndent(findings, "", "  ")
		if err != nil {
			fatal("marshal: %v", err)
		}
		fmt.Println(string(b))
	} else {
		var bad int
		for _, f := range findings {
			status := "OK"
			if !f.Compatible {
				status = "INCOMPATIBLE"
				bad++
			}
			fmt.Printf("%-40s [%s] %-12s %s\n", f.Path, f.License, status, f.Reason)
		}
		fmt.Printf("\n%d file(s), %d incompatible with %s\n", len(findings), bad, *target)
		if bad > 0 {
			os.Exit(1)
		}
	}
}

func fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "license-scanner: "+format+"\n", args...)
	os.Exit(1)
}
