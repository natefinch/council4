package magefiles

import (
	"context"
	"fmt"
	"net/http"
	"os"
)

// SupportDocs regenerates the committed card-support documentation set
// (supported.md, unsupported.md, unsupported-reasons.md, and the README summary)
// from the current Oracle corpus without rewriting the generated card tree.
//
// Card-support changes are expected to refresh this documentation, but that step
// is easy to forget; CI runs this target on every pull request and commits the
// result so the docs always reflect what the engine actually supports. In
// particular unsupported-reasons.md carries the greedy unblock roadmap that ranks
// which capability gaps block the most cards, which is only useful to whoever is
// choosing the next card-support task if it is current.
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
	return runCommand(ctx, "go", supportDocsArgs(corpusPath, scratch)...)
}

// supportDocsArgs builds the compilecards invocation that regenerates only the
// documentation set. scratch receives the (discarded) generated card sources, and
// the compile report is discarded because every document is rendered from the
// in-memory results rather than from the report file.
func supportDocsArgs(corpusPath, scratch string) []string {
	args := []string{
		"run", "./cardgen/oracle/cmd/compilecards",
		"-in", corpusPath,
		"-out", scratch,
		"-report", os.DevNull,
	}
	return append(args, documentationArgs()...)
}
