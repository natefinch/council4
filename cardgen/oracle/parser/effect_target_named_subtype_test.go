package parser

import "testing"

// firstFightTargets parses a single fight sentence and returns its two targets,
// requiring the parse to be diagnostic-free and shaped as one fight effect.
func firstFightTargets(t *testing.T, source string) []TargetSyntax {
	t.Helper()
	document, diagnostics := Parse(source, Context{InstantOrSorcery: true, CardName: "Test"})
	if len(diagnostics) != 0 {
		t.Fatalf("Parse(%q) diagnostics = %#v", source, diagnostics)
	}
	if len(document.Abilities) != 1 || len(document.Abilities[0].Sentences) != 1 {
		t.Fatalf("Parse(%q) shape = %#v", source, document.Abilities)
	}
	effects := document.Abilities[0].Sentences[0].Effects
	if len(effects) != 1 {
		t.Fatalf("Parse(%q) effects = %#v", source, effects)
	}
	return effects[0].Targets
}

// TestParseNamedTargetIsExact verifies a "named <Name>" target survives the
// unsupported-qualifier check and keeps its required name while round-tripping
// to an exact production (The Curse of Fenric III).
func TestParseNamedTargetIsExact(t *testing.T) {
	t.Parallel()
	targets := firstFightTargets(t, "Target creature you control fights another target creature named Fenric.")
	if len(targets) != 2 {
		t.Fatalf("targets = %d, want 2", len(targets))
	}
	named := targets[1]
	if named.Selection.Kind != SelectionCreature {
		t.Fatalf("named target kind = %v, want SelectionCreature", named.Selection.Kind)
	}
	if named.Selection.RequiredName != "Fenric" {
		t.Fatalf("named target RequiredName = %q, want %q", named.Selection.RequiredName, "Fenric")
	}
	if !named.Selection.Another {
		t.Fatal("named target Another = false, want true")
	}
	if !named.Exact {
		t.Fatal("named target Exact = false, want true")
	}
}

// TestParseBareSubtypeTargetIsExact verifies a bare creature-subtype target
// ("Target Mutant") records the subtype without a card-type word and still
// round-trips to an exact production, so downstream lowering can treat it as a
// creature target (The Curse of Fenric III).
func TestParseBareSubtypeTargetIsExact(t *testing.T) {
	t.Parallel()
	targets := firstFightTargets(t, "Target Mutant fights target creature you don't control.")
	if len(targets) != 2 {
		t.Fatalf("targets = %d, want 2", len(targets))
	}
	subtype := targets[0]
	if subtype.Selection.Kind != SelectionUnknown {
		t.Fatalf("bare subtype target kind = %v, want SelectionUnknown", subtype.Selection.Kind)
	}
	if len(subtype.Selection.SubtypesAny) != 1 || subtype.Selection.SubtypesAny[0] != "Mutant" {
		t.Fatalf("bare subtype target SubtypesAny = %#v, want [Mutant]", subtype.Selection.SubtypesAny)
	}
	if !subtype.Exact {
		t.Fatal("bare subtype target Exact = false, want true")
	}
}
