package main

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/natefinch/council4/internal/preservation"
)

func TestRunWritesLossEvidenceAndFails(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	base := preservation.Manifest{
		SchemaVersion: 1,
		Input: preservation.Pin{
			DownloadURI: "https://example.test/frozen.gz", DownloadSHA256: strings.Repeat("a", 64),
			JSONSHA256: strings.Repeat("b", 64), UpdatedAt: "2026-10-13",
			RulesEdition: "2026-09-25", BaselineRevision: "baseline",
		},
		Compiler: []preservation.Record{{Key: "oracle-id", Name: "Card", DefinitionSHA256: strings.Repeat("c", 64)}},
	}
	candidate := base
	candidate.Compiler = nil
	for name, value := range map[string]preservation.Manifest{"base.json": base, "candidate.json": candidate} {
		if err := preservation.Write(filepath.Join(dir, name), value); err != nil {
			t.Fatal(err)
		}
	}
	exceptions := filepath.Join(dir, "exceptions.json")
	if err := preservation.Write(exceptions, []preservation.Exception{}); err != nil {
		t.Fatal(err)
	}
	report := filepath.Join(dir, "report.json")
	if err := run([]string{
		"-baseline", filepath.Join(dir, "base.json"), "-candidate", filepath.Join(dir, "candidate.json"),
		"-exceptions", exceptions, "-report", report,
	}); err == nil {
		t.Fatal("loss did not fail the command")
	}
	output, err := preservation.Read[struct {
		RegistryOnly bool                  `json:"registry_only"`
		Changes      []preservation.Change `json:"changes"`
		Error        string                `json:"error"`
	}](report)
	if err != nil {
		t.Fatal(err)
	}
	if output.Error == "" || len(output.Changes) != 1 || output.Changes[0].Kind != "loss" {
		t.Fatalf("loss evidence = %+v", output)
	}
}
