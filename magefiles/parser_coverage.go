package magefiles

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
)

// ParserCoverage regenerates the parser-only round-trip coverage report. It
// measures how completely the Oracle parser represents the eligible Scryfall
// corpus as typed syntax, independently of the compiler and lowering, and asserts
// that every generated card (supported.md) is parser-complete.
func ParserCoverage(ctx context.Context) error {
	corpusPath, err := oracleCardsCachePath()
	if err != nil {
		return err
	}
	if err := ensureOracleCards(ctx, http.DefaultClient, scryfallOracleCardsMetadataURL, corpusPath); err != nil {
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
	if _, err := os.Stat("supported.md"); err == nil {
		args = append(args, "-generated", "supported.md")
	}
	return runCommand(ctx, "go", args...)
}
