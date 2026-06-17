package magefiles

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
)

// CardBacklog regenerates the card-support backlog. It first runs compilecards to
// produce the authoritative generated/unsupported report, then joins that
// lowering signal with the parser-only coverage signal for every eligible corpus
// card and routes each unsupported card to the layer that blocks it, emitting two
// ranked task queues (a lowering queue and a parser queue) plus a headline that
// partitions the eligible corpus. cardbacklog reconciles its own per-card
// recompile against the compilecards report and fails if they diverge.
func CardBacklog(ctx context.Context) error {
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

	compileReportPath := filepath.FromSlash(".cardwork/card-backlog-compile-report.json")
	generatedRoot := filepath.FromSlash(".cardwork/card-backlog-generated")
	if err := os.RemoveAll(generatedRoot); err != nil {
		return fmt.Errorf("removing previous card-backlog generated tree: %w", err)
	}
	if err := runCommand(ctx,
		"go", "run", "./cardgen/oracle/cmd/compilecards",
		"-in", corpusPath,
		"-out", generatedRoot,
		"-report", compileReportPath,
	); err != nil {
		return err
	}

	return runCommand(ctx,
		"go", "run", "./cardgen/oracle/cmd/cardbacklog",
		"-in", corpusPath,
		"-out", "card-backlog.md",
		"-report", filepath.FromSlash(".cardwork/card-backlog-report.json"),
		"-compile-report", compileReportPath,
	)
}
