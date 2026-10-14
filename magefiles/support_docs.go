package magefiles

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// SupportDocs regenerates the committed card-support documentation set
// (supported.md, unsupported.md, unsupported-reasons.md, the README summary, and
// card-backlog.md) from the pinned Oracle corpus without rewriting the
// generated card tree.
//
// This is an explicit legacy-maintenance command, not an every-issue or CI
// write-back gate. These reports measure compiler coverage, not authored behavior.
//
// The generated card sources are written to a throwaway directory — only the
// documentation is updated in place.
func SupportDocs(ctx context.Context) error {
	corpusPath, err := pinnedOracleCards(ctx)
	if err != nil {
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
