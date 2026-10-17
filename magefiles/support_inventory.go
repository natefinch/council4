package magefiles

import (
	"context"
	"fmt"
	"os"
)

// SupportInventory explicitly reports production status using existing pinned
// input and compiler evidence. It never compiles, generates cards or downloads data.
func SupportInventory(ctx context.Context) error {
	corpusPath, err := pinnedOracleCards()
	if err != nil {
		return err
	}
	report := os.Getenv("COUNCIL4_COMPILER_REPORT")
	if report == "" {
		report = ".cardwork/preservation/compile-report.json"
	}
	if err := os.MkdirAll(".cardwork", 0o750); err != nil {
		return fmt.Errorf("creating report directory: %w", err)
	}
	return runCommand(ctx, "go", supportInventoryArgs(corpusPath, report)...)
}

func supportInventoryArgs(corpusPath, compilerPath string) []string {
	return []string{
		"run", "./cmd/supportinventory",
		"-corpus", corpusPath, "-compiler-report", compilerPath, "-docs",
	}
}
