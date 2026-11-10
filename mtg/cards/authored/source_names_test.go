package authored

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

// TestSourceFileNamesAreASCII keeps authored file names ASCII even when a
// card's printed name has diacritics; derive them with cardname.FileStem
// (go run ./cmd/cardauthor does so by default).
func TestSourceFileNamesAreASCII(t *testing.T) {
	fileName := regexp.MustCompile(`^[_A-Za-z0-9][_A-Za-z0-9-]*\.go$`)
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if name := entry.Name(); !entry.IsDir() && strings.HasSuffix(name, ".go") && !fileName.MatchString(name) {
			t.Errorf("file name %q must be ASCII; use cardname.FileStem", name)
		}
	}
}
