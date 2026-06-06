// Command card-impl generates a partial CardDef Go source file from Scryfall
// data. It fetches a card by exact name, produces the mechanical fields in the
// canonical card-source format, and leaves categorized abilities for completion.
//
// Usage:
//
//	go run .agents/skills/card-impl/main.go "Lightning Bolt"
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/natefinch/council4/cardgen"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: go run .agents/skills/card-impl/main.go \"Card Name\"")
		os.Exit(1)
	}

	cardName := os.Args[1]

	fmt.Fprintf(os.Stderr, "Fetching %q from Scryfall...\n", cardName)
	card, err := cardgen.FetchCard(cardName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	letter := cardgen.CardNameToPackageLetter(card.Name)
	fileName := cardgen.CardNameToFileName(card.Name) + ".go"
	dir := filepath.Join("mtg", "cards", letter)

	if err := os.MkdirAll(dir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "error creating directory %s: %v\n", dir, err)
		os.Exit(1)
	}

	// Create doc.go with go:generate directive if this is a new letter directory.
	docPath := filepath.Join(dir, "doc.go")
	if _, err := os.Stat(docPath); os.IsNotExist(err) {
		docContent := fmt.Sprintf("// Package %s contains card definitions for cards starting with %s.\npackage %s\n\n//go:generate go run github.com/natefinch/council4/cardgen/cmd/gencardlist\n", letter, strings.ToUpper(letter), letter)
		if err := os.WriteFile(docPath, []byte(docContent), 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "error writing %s: %v\n", docPath, err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "Created %s\n", docPath)
	}

	source, err := cardgen.GenerateCardSource(card, letter)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error generating source: %v\n", err)
		os.Exit(1)
	}
	outPath := filepath.Join(dir, fileName)

	if err := os.WriteFile(outPath, []byte(source), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "error writing %s: %v\n", outPath, err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "Generated %s\n", outPath)
	fmt.Println(outPath)
}
