package magefiles

import (
	"context"
	"fmt"
	"os"
)

// PreserveCards checks compiler and registry inventories on the pinned corpus.
func PreserveCards(ctx context.Context) error {
	path, err := pinnedOracleCards(ctx)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(".cardwork/preservation", 0o750); err != nil {
		return fmt.Errorf("creating preservation scratch space: %w", err)
	}
	if err := runCommand(ctx, "go", "run", "./cardgen/oracle/cmd/compilecards",
		"-in", path, "-out", ".cardwork/preservation/generated",
		"-inventory", ".cardwork/preservation/candidate.json",
		"-report", ".cardwork/preservation/compile-report.json",
	); err != nil {
		return err
	}
	if err := runCommand(ctx, "go", "run", "./cmd/cardpreserve",
		"-baseline", preservationBaseline(),
		"-candidate", ".cardwork/preservation/candidate.json",
		"-report", ".cardwork/preservation/delta.json",
	); err != nil {
		return err
	}
	if err := runCommand(ctx, "go", "build", "./.cardwork/preservation/generated/..."); err != nil {
		return err
	}
	return runCommand(ctx, "go", "vet", "./.cardwork/preservation/generated/...")
}

// PreserveRegistry checks deployed definitions without downloading/recompiling the corpus.
func PreserveRegistry(ctx context.Context) error {
	if err := os.MkdirAll(".cardwork/preservation", 0o750); err != nil {
		return fmt.Errorf("creating preservation scratch space: %w", err)
	}
	return runCommand(ctx, "go", "run", "./cmd/cardpreserve",
		"-baseline", preservationBaseline(),
		"-registry-only", "-candidate", ".cardwork/preservation/registry.json",
		"-report", ".cardwork/preservation/registry-delta.json",
	)
}

func preservationBaseline() string {
	if path := os.Getenv("COUNCIL4_PRESERVATION_BASELINE"); path != "" {
		return path
	}
	return "testdata/card-preservation/baseline.json"
}

// RefreshPreservation explicitly replaces the pinned baseline; it never commits or pushes.
// Inspect the input pin, support changes and rules scenarios before accepting its output.
func RefreshPreservation(ctx context.Context) error {
	path, err := pinnedOracleCards(ctx)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(".cardwork/preservation", 0o750); err != nil {
		return fmt.Errorf("creating preservation scratch space: %w", err)
	}
	return runCommand(ctx, "go", "run", "./cardgen/oracle/cmd/compilecards",
		"-in", path, "-out", ".cardwork/preservation/refresh-generated",
		"-inventory", "testdata/card-preservation/baseline.json",
		"-report", ".cardwork/preservation/refresh-report.json",
	)
}
