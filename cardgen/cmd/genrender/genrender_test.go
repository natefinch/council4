package main

import (
	"os"
	"path/filepath"
	"testing"
)

// repoRoot is this command's path relative to the repository root, used to find
// the committed generated file from the package directory.
const repoRoot = "../../.."

// TestGeneratedFileIsCurrent regenerates the literal tables and compares them
// with the committed file.
//
// This is the drift guard for the whole renderer literal layer. It replaces the
// hand-maintained coverage tests that previously asserted, one enum at a time,
// that someone had remembered to add a case: those could only catch drift in the
// types someone had thought to list, and several enums had no such test at all.
func TestGeneratedFileIsCurrent(t *testing.T) {
	if testing.Short() {
		t.Skip("type-checking mtg/game from source takes several seconds")
	}
	want, err := generate(repoRoot)
	if err != nil {
		t.Fatalf("generating literals: %v", err)
	}
	path := filepath.Join(repoRoot, generatedFile)
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading %s: %v", path, err)
	}
	if string(got) != string(want) {
		t.Errorf("%s is out of date; run go generate ./cardgen/...", generatedFile)
	}
}

// TestBitmaskTypesAreFlagShaped asserts that every type named in bitmaskTypes is
// reachable and really is a set of independent power-of-two flags, so the list
// cannot silently describe a type that has stopped being a bitmask.
func TestBitmaskTypesAreFlagShaped(t *testing.T) {
	if testing.Short() {
		t.Skip("type-checking mtg/game from source takes several seconds")
	}
	graph, err := load(repoRoot)
	if err != nil {
		t.Fatalf("loading type graph: %v", err)
	}
	if err := checkBitmaskTypes(graph.Enums); err != nil {
		t.Fatal(err)
	}
	for _, e := range graph.Enums {
		if !isBitmask(e) {
			continue
		}
		if _, err := flags(e); err != nil {
			t.Errorf("%s: %v", e.Ref(), err)
		}
	}
}

// TestRootTypesAreReachable asserts the walk finds the type graph it is supposed
// to. A walk that silently stopped early would emit fewer tables and quietly
// reintroduce the gaps this generator exists to close, so this pins the shapes
// the renderer must be able to emit.
func TestRootTypesAreReachable(t *testing.T) {
	if testing.Short() {
		t.Skip("type-checking mtg/game from source takes several seconds")
	}
	graph, err := load(repoRoot)
	if err != nil {
		t.Fatalf("loading type graph: %v", err)
	}
	for _, want := range []string{"game.Primitive", "game.Ability", "game.KeywordAbility"} {
		impls := graph.Implementations[want]
		if len(impls) == 0 {
			t.Errorf("interface %s has no recorded implementations", want)
		}
	}
	// game.ManaSpendRider is reachable only through opt.V[ManaSpendRider], so
	// it is the regression test for walking generic type arguments.
	for _, want := range []string{"game.CardDef", "game.CardFace", "game.ManaSpendRider"} {
		if !containsNamed(graph.Structs, want) {
			t.Errorf("struct %s is not reachable from the root types", want)
		}
	}
	for _, want := range []string{"game.Step", "game.EventKind", "zone.Type", "game.ManaSpendConditionKind"} {
		if !containsEnum(graph.Enums, want) {
			t.Errorf("enum %s is not reachable from the root types", want)
		}
	}
}
