package main

import (
	"bufio"
	"fmt"
	"os"
	"slices"
	"strings"
)

// generatedSubsetViolations returns the names of generated cards that the parser
// did not classify as parser-complete. A card the lowerer can fully generate must
// be parser-complete, so any violation indicates the recognized-span union is too
// strict.
func generatedSubsetViolations(generatedPath string, cards []cardResult) (generatedCount int, violations []string, err error) {
	generated, records, err := readGeneratedNames(generatedPath)
	if err != nil {
		return 0, nil, err
	}
	complete := map[string]bool{}
	for i := range cards {
		if cards[i].eligible && cards[i].complete {
			complete[cards[i].name] = true
		}
	}
	for name := range generated {
		if !complete[name] {
			violations = append(violations, name)
		}
	}
	slices.Sort(violations)
	return records, violations, nil
}

// readGeneratedNames reads a supported-card Markdown list, returning the set of
// distinct card names from its "- <name>" bullet lines and the total bullet count
// (records). The corpus repeats some token names across printings, so records can
// exceed the number of distinct names; the subset check is name-based.
func readGeneratedNames(path string) (names map[string]bool, records int, err error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, 0, fmt.Errorf("opening generated card list: %w", err)
	}
	defer file.Close()

	names = map[string]bool{}
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "- ") {
			continue
		}
		name := strings.TrimSpace(strings.TrimPrefix(line, "- "))
		if name != "" {
			names[name] = true
			records++
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, 0, fmt.Errorf("reading generated card list: %w", err)
	}
	return names, records, nil
}

func buildValidation(generatedCount int, violations []string) *validationResult {
	return &validationResult{
		GeneratedCards: generatedCount,
		Violations:     len(violations),
		ViolationNames: violations,
	}
}

func reportSubsetViolations(violations []string) {
	if len(violations) == 0 {
		_, _ = fmt.Fprintln(os.Stderr, "generated ⊆ parser-complete: OK (0 violations)")
		return
	}
	_, _ = fmt.Fprintf(os.Stderr, "generated ⊆ parser-complete: %d VIOLATIONS\n", len(violations))
	for _, name := range violations {
		_, _ = fmt.Fprintln(os.Stderr, "  -", name)
	}
}
