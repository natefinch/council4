package magefiles

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
)

// CardBacklog regenerates the card-support backlog on its own: it first runs
// compilecards to produce a throwaway compile report, then delegates to
// regenerateCardBacklog. Callers that already hold a current compile report
// (CardSupport, SupportDocs) call regenerateCardBacklog directly instead, so the
// corpus is not compiled twice in the same run.
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

	return regenerateCardBacklog(ctx, corpusPath, compileReportPath, filepath.FromSlash(".cardwork/card-backlog-report.json"))
}

// regenerateCardBacklog joins the lowering signal in an already-produced
// compilecards report (compileReportPath) with the parser-only coverage signal
// for every eligible corpus card, and routes each unsupported card to the layer
// that blocks it, emitting two ranked task queues (a lowering queue and a parser
// queue) plus a headline that partitions the eligible corpus into the committed
// card-backlog.md. backlogReportPath receives cardbacklog's own JSON report (use
// os.DevNull to discard it); cardbacklog reconciles its own per-card recompile
// against compileReportPath and fails if they diverge.
//
// Sharing this helper (rather than each caller running its own compilecards
// pass) matters for the same reason documentationArgs does: card-backlog.md
// previously had no automatic regeneration path at all and went stale for
// months, invisible to whoever was choosing the next card-support task.
func regenerateCardBacklog(ctx context.Context, corpusPath, compileReportPath, backlogReportPath string) error {
	return runCommand(ctx, "go", cardBacklogArgs(corpusPath, compileReportPath, backlogReportPath)...)
}

// cardBacklogArgs builds the cardbacklog invocation that regenerates
// card-backlog.md from an already-produced compilecards report.
func cardBacklogArgs(corpusPath, compileReportPath, backlogReportPath string) []string {
	return []string{
		"run", "./cardgen/oracle/cmd/cardbacklog",
		"-in", corpusPath,
		"-out", "card-backlog.md",
		"-report", backlogReportPath,
		"-compile-report", compileReportPath,
	}
}
