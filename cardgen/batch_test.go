package cardgen

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseCardList(t *testing.T) {
	input := `
# comment
// Commander
1 Atraxa, Praetors' Voice

MAIN:
Sol Ring // ramp
2 Forest # basics
4x Fire // Ice
`

	items, err := ParseCardList(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ParseCardList error: %v", err)
	}

	want := []CardListItem{
		{InputName: "Atraxa, Praetors' Voice", Quantity: 1, Section: "Commander", Line: 4},
		{InputName: "Sol Ring", Quantity: 1, Section: "Main", Line: 7},
		{InputName: "Forest", Quantity: 2, Section: "Main", Line: 8},
		{InputName: "Fire // Ice", Quantity: 4, Section: "Main", Line: 9},
	}
	if len(items) != len(want) {
		t.Fatalf("len(items) = %d, want %d: %+v", len(items), len(want), items)
	}
	for i := range want {
		if items[i] != want[i] {
			t.Fatalf("items[%d] = %+v, want %+v", i, items[i], want[i])
		}
	}
}

func TestNewManifestFromItemsAggregatesBySectionAndName(t *testing.T) {
	items := []CardListItem{
		{InputName: "Sol Ring", Quantity: 1, Section: "Main", Line: 1},
		{InputName: "sol ring", Quantity: 1, Section: "Main", Line: 2},
		{InputName: "Sol Ring", Quantity: 1, Section: "Commander", Line: 3},
	}

	manifest := NewManifestFromItems(items)

	if len(manifest.Cards) != 2 {
		t.Fatalf("len(cards) = %d, want 2: %+v", len(manifest.Cards), manifest.Cards)
	}
	if manifest.Cards[0].Quantity != 2 || manifest.Cards[0].FirstLine != 1 || manifest.Cards[0].Status != BatchStatusPending {
		t.Fatalf("first manifest card = %+v", manifest.Cards[0])
	}
	if manifest.Cards[1].Section != "Commander" {
		t.Fatalf("second manifest card section = %q, want Commander", manifest.Cards[1].Section)
	}
}

func TestMarkExistingFiles(t *testing.T) {
	dir := t.TempDir()
	manifest := Manifest{Version: ManifestVersion, Cards: []ManifestCard{
		{InputName: "Lightning Bolt", CanonicalName: "Lightning Bolt", Status: BatchStatusFetched},
		{InputName: "Missing Card", Status: BatchStatusFetched},
	}}

	existingPath := ExpectedCardFile("Lightning Bolt")
	writeTestFile(t, dir, existingPath)
	MarkExistingFiles(&manifest, dir)

	if manifest.Cards[0].FileStatus != BatchFileStatusExisting || manifest.Cards[0].FilePath != existingPath {
		t.Fatalf("existing card = %+v", manifest.Cards[0])
	}
	if manifest.Cards[1].FileStatus != BatchFileStatusMissing {
		t.Fatalf("missing card file status = %q, want %q", manifest.Cards[1].FileStatus, BatchFileStatusMissing)
	}
}

func writeTestFile(t *testing.T, root, relPath string) {
	t.Helper()
	path := filepath.Join(root, relPath)
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("package test\n"), 0o600); err != nil {
		t.Fatal(err)
	}
}
