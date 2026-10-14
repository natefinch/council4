package magefiles

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/natefinch/council4/internal/preservation"
)

// ParserCoverage regenerates the parser-only round-trip coverage report. It
// measures how completely the Oracle parser represents the eligible Scryfall
// corpus as typed syntax, independently of the compiler and lowering, and asserts
// that the preserved legacy compiler output is parser-complete. Authored cards
// are not part of this translator-only invariant.
func ParserCoverage(ctx context.Context) error {
	corpusPath, err := pinnedOracleCards(ctx)
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
	baseline, err := preservation.Read[preservation.Manifest](preservationBaseline())
	if err != nil {
		return err
	}
	if err := baseline.Input.Verify(corpusPath); err != nil {
		return err
	}
	legacyList := filepath.FromSlash(".cardwork/legacy-compiler-supported.md")
	if err := os.WriteFile(legacyList, []byte(legacySupportedMarkdown(baseline.Compiler)), 0o600); err != nil {
		return fmt.Errorf("writing legacy compiler subset: %w", err)
	}
	args = append(args, "-generated", legacyList)
	return runCommand(ctx, "go", args...)
}

func legacySupportedMarkdown(records []preservation.Record) string {
	var text strings.Builder
	for _, record := range records {
		if record.DefinitionSHA256 != "" {
			_, _ = text.WriteString("- ")
			_, _ = text.WriteString(record.Name)
			_ = text.WriteByte('\n')
		}
	}
	return text.String()
}
