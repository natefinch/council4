package magefiles

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// ParserCoverage regenerates the parser-only round-trip coverage report. It
// measures how completely the Oracle parser represents the eligible Scryfall
// corpus as typed syntax, independently of the compiler and lowering. It reads
// existing pinned input only and is not a validation gate for authored cards.
func ParserCoverage(ctx context.Context) error {
	corpusPath, err := pinnedOracleCards()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(".cardwork", 0o750); err != nil {
		return fmt.Errorf("creating cardgen work directory: %w", err)
	}
	args := []string{
		"run", "./cardgen/oracle/cmd/parsercoverage",
		"-in", corpusPath,
		"-out", "parser-coverage.md",
		"-report", filepath.FromSlash(".cardwork/parser-coverage-report.json"),
	}
	return runCommand(ctx, "go", args...)
}
