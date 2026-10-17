package preservation

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func fixturePin(t *testing.T, input string) Pin {
	t.Helper()
	digest, err := Digest(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	return Pin{
		DownloadURI:    "https://example.invalid/oracle.jsonl.gz",
		DownloadSHA256: strings.Repeat("a", 64), JSONSHA256: digest,
		UpdatedAt: "2026-10-13", RulesEdition: "2026-09-25", BaselineRevision: "baseline",
	}
}

func TestPinVerifiesActualInput(t *testing.T) {
	t.Parallel()
	const input = `[{"name":"Bear"}]`
	pin := fixturePin(t, input)
	path := filepath.Join(t.TempDir(), "corpus.json")
	if err := pin.Verify(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("missing corpus error = %v", err)
	}
	if err := os.WriteFile(path, []byte(input), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := pin.Verify(path); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("different bytes"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := pin.Verify(path); err == nil {
		t.Fatal("accepted a declared digest without verifying the input")
	}
}

func TestPinRejectsIncompleteDigests(t *testing.T) {
	t.Parallel()
	for _, digest := range []string{"", "short", strings.Repeat("z", 64)} {
		for _, download := range []bool{false, true} {
			pin := fixturePin(t, "[]")
			if download {
				pin.DownloadSHA256 = digest
			} else {
				pin.JSONSHA256 = digest
			}
			if err := pin.Validate(); err == nil {
				t.Fatalf("accepted invalid digest %q (download: %v)", digest, download)
			}
		}
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
