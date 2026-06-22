package magefiles

import (
	"context"
	"fmt"
	"net/http"
	"os"
)

// compileCheckOutput is the ignored scratch directory the generated corpus is
// written to for the compile check. It lives under .cardwork so it is never
// committed and is removed after the check.
const compileCheckOutput = ".cardwork/compile-check"

// CompileCheck generates the full card corpus and compiles every generated
// package, failing if any card does not build or fails go vet. Card-generation
// unit tests cover individual cards, but a renderer that emits a package
// reference without registering its import produces source that still fails to
// compile; this catches that class of codegen regression on the introducing
// change rather than downstream when the corpus is regenerated. The generated
// packages live under the module, so they build in-module without a replace
// directive.
func CompileCheck(ctx context.Context) error {
	corpusPath, err := oracleCardsCachePath()
	if err != nil {
		return err
	}
	if err := ensureOracleCards(ctx, http.DefaultClient, scryfallOracleCardsMetadataURL, corpusPath); err != nil {
		return err
	}
	if err := os.RemoveAll(compileCheckOutput); err != nil {
		return fmt.Errorf("removing previous compile-check tree: %w", err)
	}
	defer func() { _ = os.RemoveAll(compileCheckOutput) }()
	if err := runCommand(ctx, "go", "run", "./cardgen/oracle/cmd/compilecards",
		"-in", corpusPath,
		"-out", compileCheckOutput,
		"-report", os.DevNull,
	); err != nil {
		return err
	}
	if err := runCommand(ctx, "go", "build", "./"+compileCheckOutput+"/..."); err != nil {
		return err
	}
	return runCommand(ctx, "go", "vet", "./"+compileCheckOutput+"/...")
}
