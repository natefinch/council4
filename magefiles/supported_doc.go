package magefiles

import (
	"context"
	"fmt"
	"net/http"
	"os"
)

// SupportedDoc regenerates the committed supported-card list (supported.md)
// from the current Oracle corpus without rewriting the generated card tree.
//
// Card-support changes are expected to refresh supported.md, but that step is
// easy to forget; CI runs this target on every pull request and commits the
// result so the list always reflects what the engine actually supports. The
// generated card sources are written to a throwaway directory — only
// supported.md is updated in place.
func SupportedDoc(ctx context.Context) error {
	corpusPath, err := oracleCardsCachePath()
	if err != nil {
		return err
	}
	if err := ensureOracleCards(ctx, http.DefaultClient, scryfallOracleCardsMetadataURL, corpusPath); err != nil {
		return err
	}
	scratch, err := os.MkdirTemp("", "supported-doc")
	if err != nil {
		return fmt.Errorf("creating supported-doc scratch directory: %w", err)
	}
	defer func() { _ = os.RemoveAll(scratch) }()
	return runCommand(ctx, "go", supportedDocArgs(corpusPath, scratch)...)
}

// supportedDocArgs builds the compilecards invocation that regenerates only
// supported.md. scratch receives the (discarded) generated card sources.
func supportedDocArgs(corpusPath, scratch string) []string {
	return []string{
		"run", "./cardgen/oracle/cmd/compilecards",
		"-in", corpusPath,
		"-out", scratch,
		"-report", os.DevNull,
		"-supported", "supported.md",
	}
}
