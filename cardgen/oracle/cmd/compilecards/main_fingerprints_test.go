package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/natefinch/council4/cardgen"
	"github.com/natefinch/council4/cardgen/oracle/shared"
)

// fingerprintLines returns the manifest's card lines, dropping the leading
// comment header.
func fingerprintLines(t *testing.T, results []result) []string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "card-fingerprints.txt")
	if err := writeFingerprints(path, results); err != nil {
		t.Fatalf("writeFingerprints: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	var lines []string
	for line := range strings.SplitSeq(string(data), "\n") {
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		lines = append(lines, line)
	}
	return lines
}

func supportedResult(name, source string) result {
	return result{card: cardgen.ScryfallCard{Name: name}, source: source}
}

// TestWriteFingerprintsRecordsOnlySupportedCards proves the manifest tracks
// exactly the set supported.md lists. A card that failed to compile, produced
// diagnostics, or was excluded by corpus policy has no generated source, so
// recording it would put a digest of the empty string in the manifest and make
// every such card look identical.
func TestWriteFingerprintsRecordsOnlySupportedCards(t *testing.T) {
	t.Parallel()
	results := []result{
		supportedResult("Supported", "package a"),
		{card: cardgen.ScryfallCard{Name: "Excluded"}, exclusion: cardgen.CorpusExclusionReason("digital")},
		{card: cardgen.ScryfallCard{Name: "Diagnosed"}, diagnostics: []shared.Diagnostic{{Summary: "unsupported"}}},
		{card: cardgen.ScryfallCard{Name: "Failed"}, err: os.ErrNotExist},
	}
	lines := fingerprintLines(t, results)
	if len(lines) != 1 {
		t.Fatalf("manifest lines = %v, want only the supported card", lines)
	}
	if !strings.HasSuffix(lines[0], "  Supported") {
		t.Fatalf("manifest line = %q, want the supported card", lines[0])
	}
}

// TestWriteFingerprintsDistinguishesSourceChanges is the property the whole
// manifest exists for: two cards whose generated source differs must get
// different digests, and regenerating unchanged source must reproduce the same
// digest. Without both halves the diff would either hide real changes or churn
// on every run, and in either case it would stop being a usable blast-radius
// readout.
func TestWriteFingerprintsDistinguishesSourceChanges(t *testing.T) {
	t.Parallel()
	first := fingerprintLines(t, []result{supportedResult("Card", "package a")})
	same := fingerprintLines(t, []result{supportedResult("Card", "package a")})
	changed := fingerprintLines(t, []result{supportedResult("Card", "package b")})
	if first[0] != same[0] {
		t.Fatalf("identical source produced %q then %q, want a stable digest", first[0], same[0])
	}
	if first[0] == changed[0] {
		t.Fatalf("changed source produced the same line %q, want a different digest", first[0])
	}
}

// TestWriteFingerprintsOrdersDuplicateNamesByDigest keeps the manifest stable
// for the 149 names the corpus supports more than once, almost all of them
// tokens. Ordering by name alone would leave those entries free to swap
// positions between runs, and the resulting phantom diff would train reviewers
// to ignore the file.
func TestWriteFingerprintsOrdersDuplicateNamesByDigest(t *testing.T) {
	t.Parallel()
	forward := fingerprintLines(t, []result{
		supportedResult("Angel", "package a"),
		supportedResult("Angel", "package b"),
		supportedResult("angel", "package c"),
	})
	reversed := fingerprintLines(t, []result{
		supportedResult("angel", "package c"),
		supportedResult("Angel", "package b"),
		supportedResult("Angel", "package a"),
	})
	if len(forward) != 3 {
		t.Fatalf("manifest lines = %v, want three entries", forward)
	}
	for i := range forward {
		if forward[i] != reversed[i] {
			t.Fatalf("line %d = %q from one input order and %q from another, want identical", i, forward[i], reversed[i])
		}
	}
	// Case-insensitive name ordering groups the printings of one name
	// together, matching how supported.md sorts.
	for i, line := range forward {
		if !strings.HasSuffix(strings.ToLower(line), "  angel") {
			t.Fatalf("line %d = %q, want an Angel printing", i, line)
		}
	}
}
