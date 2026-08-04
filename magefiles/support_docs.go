package magefiles

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
)

// SupportDocs regenerates the committed card-support documentation set
// (supported.md, unsupported.md, unsupported-reasons.md, the README summary, and
// card-backlog.md) from the current Oracle corpus without rewriting the
// generated card tree.
//
// Card-support changes are expected to refresh this documentation, but that step
// is easy to forget; CI runs this target on every pull request and commits the
// result so the docs always reflect what the engine actually supports. In
// particular unsupported-reasons.md carries the greedy unblock roadmap that ranks
// which capability gaps block the most cards, and card-backlog.md routes each of
// those cards to the parser or lowering layer that blocks it — both are only
// useful to whoever is choosing the next card-support task if they are current.
//
// The generated card sources are written to a throwaway directory — only the
// documentation is updated in place.
func SupportDocs(ctx context.Context) error {
	corpusPath, err := oracleCardsCachePath()
	if err != nil {
		return err
	}
	if err := ensureOracleCards(ctx, http.DefaultClient, scryfallOracleCardsMetadataURL, corpusPath); err != nil {
		return err
	}
	scratch, err := os.MkdirTemp("", "support-docs")
	if err != nil {
		return fmt.Errorf("creating support-docs scratch directory: %w", err)
	}
	defer func() { _ = os.RemoveAll(scratch) }()
	compileReportPath := filepath.Join(scratch, "compile-report.json")
	if err := runCommand(ctx, "go", supportDocsArgs(corpusPath, scratch, compileReportPath)...); err != nil {
		return err
	}
	return regenerateCardBacklog(ctx, corpusPath, compileReportPath, os.DevNull)
}

// supportDocsArgs builds the compilecards invocation that regenerates the
// documentation set. scratch receives the (discarded) generated card sources;
// compileReportPath receives the real compile report so the caller can join it
// with the parser signal to regenerate card-backlog.md without compiling the
// corpus a second time. Every document text itself is still rendered from the
// in-memory compile results, not read back from the report file.
func supportDocsArgs(corpusPath, scratch, compileReportPath string) []string {
	args := []string{
		"run", "./cardgen/oracle/cmd/compilecards",
		"-in", corpusPath,
		"-out", scratch,
		"-report", compileReportPath,
	}
	return append(args, documentationArgs()...)
}
