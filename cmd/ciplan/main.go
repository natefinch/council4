// Command ciplan selects read-only validation from the actual changed paths.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type plan struct {
	Code          bool `json:"code"`
	Compiler      bool `json:"compiler"`
	Preservation  bool `json:"preservation"`
	Slow          bool `json:"slow"`
	CorpusUpgrade bool `json:"corpus_upgrade"`
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	flags := flag.NewFlagSet("ciplan", flag.ContinueOnError)
	base := flags.String("base", "", "base Git revision (empty/zero means initial push)")
	head := flags.String("head", "HEAD", "head Git revision")
	all := flags.Bool("all", false, "explicitly select all acceptance checks")
	if err := flags.Parse(args); err != nil {
		return err
	}
	command := []string{"ls-files", "-z"}
	if *base != "" && strings.Trim(*base, "0") != "" {
		command = []string{"diff", "--name-only", "-z", *base, *head, "--"}
	}
	output, err := exec.CommandContext(context.Background(), "git", command...).Output()
	if err != nil {
		return fmt.Errorf("reading changed paths: %w", err)
	}
	selection := selectPlan(strings.Split(string(output), "\x00"))
	if *all {
		selection = plan{Code: true, Compiler: true, Preservation: true, Slow: true}
	}
	data, err := json.MarshalIndent(selection, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding CI plan: %w", err)
	}
	_, _ = fmt.Fprintln(os.Stdout, string(data))
	if path := os.Getenv("GITHUB_OUTPUT"); path != "" {
		//nolint:gosec // GITHUB_OUTPUT is the runner-owned file, not a path from event payloads.
		file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o600)
		if err != nil {
			return fmt.Errorf("opening GitHub outputs: %w", err)
		}
		defer file.Close()
		if _, err := fmt.Fprintf(file, "code=%t\ncompiler=%t\npreservation=%t\nslow=%t\ncorpus_upgrade=%t\n",
			selection.Code, selection.Compiler, selection.Preservation, selection.Slow, selection.CorpusUpgrade); err != nil {
			return fmt.Errorf("writing GitHub outputs: %w", err)
		}
	}
	return nil
}

func selectPlan(paths []string) plan {
	var selection plan
	pinChanged, baselineChanged, upgradeOnly := false, false, true
	for _, path := range paths {
		if path == "" {
			continue
		}
		if path == "testdata/card-preservation/corpus.json" {
			pinChanged = true
		}
		if path == "testdata/card-preservation/baseline.json" {
			baselineChanged = true
		}
		if !strings.HasSuffix(path, ".md") &&
			path != "testdata/card-preservation/corpus.json" &&
			path != "testdata/card-preservation/baseline.json" {
			upgradeOnly = false
		}
		if strings.HasSuffix(path, ".md") {
			continue
		}
		selection.Code = true
		if strings.HasSuffix(path, ".go") && !strings.HasPrefix(path, "mtg/cards/") {
			selection.Preservation, selection.Slow = true, true
		}
		if strings.HasPrefix(path, "cardgen/") || strings.HasPrefix(path, "internal/preservation/") ||
			strings.HasPrefix(path, "magefiles/") || strings.HasPrefix(path, "testdata/card-preservation/") ||
			path == "go.mod" || path == "go.sum" || strings.HasPrefix(path, ".github/workflows/") {
			selection.Compiler, selection.Preservation, selection.Slow = true, true, true
		}
		if strings.HasPrefix(path, "mtg/") || strings.HasPrefix(path, "cmd/") {
			selection.Preservation = true
			if !strings.HasPrefix(path, "mtg/cards/") {
				selection.Slow = true
			}
		}
	}
	selection.CorpusUpgrade = pinChanged && baselineChanged && upgradeOnly
	return selection
}
