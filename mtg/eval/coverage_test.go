package eval

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

// knownPrimitiveCount is the number of game.Primitive implementations the
// translator was last reconciled against. ScorableEffect classifies the
// value-dominant primitives and treats the rest as value-neutral; when this
// guard fails because a primitive was added or removed, review the new
// primitive in appendPrimitiveAtoms (give it a value atom or confirm neutral)
// and update this constant.
const knownPrimitiveCount = 104

// TestPrimitiveCountIsReconciled keeps a newly added resolution primitive from
// silently falling through the translator: adding one trips this guard so its
// value classification is considered.
func TestPrimitiveCountIsReconciled(t *testing.T) {
	gameDir := filepath.Join("..", "game")
	entries, err := os.ReadDir(gameDir)
	if err != nil {
		t.Fatalf("reading %s: %v", gameDir, err)
	}
	pattern := regexp.MustCompile(`func \([A-Za-z]+\) isPrimitive\(\)`)
	seen := map[string]bool{}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".go" {
			continue
		}
		source, err := os.ReadFile(filepath.Join(gameDir, entry.Name()))
		if err != nil {
			t.Fatalf("reading %s: %v", entry.Name(), err)
		}
		for _, match := range pattern.FindAllString(string(source), -1) {
			seen[match] = true
		}
	}
	if len(seen) != knownPrimitiveCount {
		t.Fatalf("found %d game.Primitive implementations, knownPrimitiveCount = %d; "+
			"reconcile appendPrimitiveAtoms with the change and update the constant",
			len(seen), knownPrimitiveCount)
	}
}
