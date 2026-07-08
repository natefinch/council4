// Command gencardlist scans a card sub-package directory for its card builder
// functions and generates a cards.go file registering each as a cardset.Entry
// (a name paired with its constructor) for lazy loading.
//
// Usage (typically via go generate):
//
//	go run github.com/natefinch/council4/cardgen/cmd/gencardlist
//
// Run from the letter sub-package directory (e.g., mtg/cards/l/).
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/natefinch/council4/cardgen/cardlist"
)

func main() {
	dir, err := os.Getwd()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	source, err := cardlist.Generate(dir, filepath.Base(dir), "gencardlist")
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error generating cards.go: %v\n", err)
		os.Exit(1)
	}

	outPath := filepath.Join(dir, "cards.go")
	if err := os.WriteFile(outPath, source, 0o600); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error writing %s: %v\n", outPath, err)
		os.Exit(1)
	}

	_, _ = fmt.Fprintf(os.Stderr, "Generated %s\n", outPath)
}
