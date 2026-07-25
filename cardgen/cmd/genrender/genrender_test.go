package main

import (
	"os"
	"path/filepath"
	"testing"
)

// repoRoot is this command's path relative to the repository root, used to find
// the committed generated file from the package directory.
const repoRoot = "../../.."

// TestGeneratedFileIsCurrent regenerates every generated file and compares it
// with the committed copy.
//
// This is the drift guard for the whole renderer literal layer. It replaces the
// hand-maintained coverage tests that previously asserted, one enum at a time,
// that someone had remembered to add a case: those could only catch drift in the
// types someone had thought to list, and several enums had no such test at all.
func TestGeneratedFileIsCurrent(t *testing.T) {
	if testing.Short() {
		t.Skip("type-checking mtg/game from source takes several seconds")
	}
	graph, err := load(repoRoot)
	if err != nil {
		t.Fatalf("loading type graph: %v", err)
	}
	for _, out := range generatedOutputs() {
		want, err := out.gen(graph)
		if err != nil {
			t.Errorf("generating %s: %v", out.path, err)
			continue
		}
		path := filepath.Join(repoRoot, filepath.FromSlash(out.path))
		got, err := os.ReadFile(path)
		if err != nil {
			t.Errorf("reading %s: %v", out.path, err)
			continue
		}
		if string(got) != string(want) {
			t.Errorf("%s is out of date; run go generate ./cardgen/...", out.path)
		}
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

// TestCuratedTablesAreCurrent asserts that every hand-maintained table the
// emitter consults still describes the real type graph.
//
// These tables are the only places hand-written knowledge survives in the
// renderer, and each one fails dangerously if it rots: a missing opaqueRenderers
// entry drops a value from every card that carries it, a stale openStringTypes
// entry turns a synthesized key into a render error, and a stale
// alwaysEmitEnums or alwaysEmitFields entry silently stops emitting a field
// whose zero names a real value, and a stale preRenderValidators or
// constructorRenderers entry wires a support gate or a constructor spelling to
// a type nothing renders any more. The check functions are the same ones generation runs, so this reports
// the rot as a test failure rather than as a broken corpus.
func TestCuratedTablesAreCurrent(t *testing.T) {
	if testing.Short() {
		t.Skip("type-checking mtg/game from source takes several seconds")
	}
	pkgs, err := loadPackages(repoRoot)
	if err != nil {
		t.Fatalf("loading packages: %v", err)
	}
	graph, err := load(repoRoot)
	if err != nil {
		t.Fatalf("loading type graph: %v", err)
	}
	if err := checkOpaqueRenderers(graph); err != nil {
		t.Error(err)
	}
	if err := checkOpenStringTypes(graph, pkgs); err != nil {
		t.Error(err)
	}
	if err := checkAlwaysEmit(graph, pkgs); err != nil {
		t.Error(err)
	}
	if err := checkPreRenderValidators(graph); err != nil {
		t.Error(err)
	}
}

// TestOpaqueTypesAllHaveRenderers pins the property that motivates
// opaqueRenderers: a type whose state is unexported cannot be written as a
// composite literal from cardgen, so the generator must refuse to emit one it
// has no hand-written renderer for rather than emitting an empty literal.
func TestOpaqueTypesAllHaveRenderers(t *testing.T) {
	if testing.Short() {
		t.Skip("type-checking mtg/game from source takes several seconds")
	}
	graph, err := load(repoRoot)
	if err != nil {
		t.Fatalf("loading type graph: %v", err)
	}
	if len(graph.Opaque) == 0 {
		t.Fatal("no opaque types found; the walk is not classifying unexported state")
	}
	for _, named := range graph.Opaque {
		ref := qualified(named)
		if _, ok := opaqueRenderers[ref]; !ok {
			t.Errorf("opaque type %s has no entry in opaqueRenderers", ref)
		}
	}
}
