package preservation

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPinRequiresTheActualInput(t *testing.T) {
	t.Parallel()
	pin := testManifest().Input
	path := filepath.Join(t.TempDir(), "corpus.json")
	if err := os.WriteFile(path, []byte("different bytes"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := pin.Verify(path); err == nil {
		t.Fatal("accepted a declared digest without verifying the input")
	}
}

func TestReadRejectsTrailingAndUnknownData(t *testing.T) {
	t.Parallel()
	for _, source := range []string{`{} {}`, `{"unknown":true}`} {
		path := filepath.Join(t.TempDir(), "pin.json")
		if err := os.WriteFile(path, []byte(source), 0o600); err != nil {
			t.Fatal(err)
		}
		if _, err := Read[Pin](path); err == nil {
			t.Fatal("accepted ambiguous pin data")
		}
	}
}

func TestValidateRefusesPartialAcceptedDefinitions(t *testing.T) {
	t.Parallel()
	for _, record := range []Record{
		{Key: "id", Name: "Card", DefinitionSHA256: "short"},
		{Key: "id", Name: "Card", DefinitionSHA256: strings.Repeat("c", 64), Refusals: []string{"partial"}},
	} {
		manifest := testManifest()
		manifest.Compiler = []Record{record}
		if err := manifest.Validate(); err == nil {
			t.Fatal("partially supported record was accepted")
		}
	}
}
