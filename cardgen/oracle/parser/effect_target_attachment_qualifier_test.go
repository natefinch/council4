package parser

import "testing"

// TestParseAttachmentQualifiedTargetsAreExact verifies that the leading
// "modified", "enchanted", and "equipped" permanent adjectives are captured
// onto the target's selection rather than wiping it, and that each target
// still round-trips to an exact, lowerable production.
func TestParseAttachmentQualifiedTargetsAreExact(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		source        string
		wantKind      SelectionKind
		wantModified  bool
		wantEnchanted bool
		wantEquipped  bool
		wantMin       int
		wantMax       int
	}{
		{
			name:          "enchanted permanent",
			source:        "Destroy target enchanted permanent.",
			wantKind:      SelectionPermanent,
			wantEnchanted: true,
			wantMin:       1,
			wantMax:       1,
		},
		{
			name:          "enchanted creature",
			source:        "Destroy target enchanted creature.",
			wantKind:      SelectionCreature,
			wantEnchanted: true,
			wantMin:       1,
			wantMax:       1,
		},
		{
			name:         "modified creature you control",
			source:       "Destroy target modified creature you control.",
			wantKind:     SelectionCreature,
			wantModified: true,
			wantMin:      1,
			wantMax:      1,
		},
		{
			name:         "equipped creature",
			source:       "Destroy target equipped creature.",
			wantKind:     SelectionCreature,
			wantEquipped: true,
			wantMin:      1,
			wantMax:      1,
		},
		{
			name:          "plural enchanted permanents",
			source:        "Destroy up to two target enchanted permanents.",
			wantKind:      SelectionPermanent,
			wantEnchanted: true,
			wantMin:       0,
			wantMax:       2,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			targets := firstDestroyTargets(t, test.source)
			if len(targets) != 1 {
				t.Fatalf("targets = %d, want 1", len(targets))
			}
			target := targets[0]
			if !target.Exact {
				t.Fatalf("target = %#v, want exact", target)
			}
			if target.Selection.Kind != test.wantKind {
				t.Fatalf("kind = %v, want %v", target.Selection.Kind, test.wantKind)
			}
			if target.Cardinality.Min != test.wantMin || target.Cardinality.Max != test.wantMax {
				t.Fatalf("cardinality = [%d,%d], want [%d,%d]", target.Cardinality.Min, target.Cardinality.Max, test.wantMin, test.wantMax)
			}
			selection := target.Selection
			if selection.Modified != test.wantModified ||
				selection.Enchanted != test.wantEnchanted ||
				selection.Equipped != test.wantEquipped {
				t.Fatalf("selection attachment flags = {modified:%v enchanted:%v equipped:%v}, want {modified:%v enchanted:%v equipped:%v}",
					selection.Modified, selection.Enchanted, selection.Equipped,
					test.wantModified, test.wantEnchanted, test.wantEquipped)
			}
		})
	}
}

// TestAttachmentQualifiedTargetIsNotAFreeReference verifies that the leading
// attachment adjective does not leak a spurious semantic reference.
func TestAttachmentQualifiedTargetIsNotAFreeReference(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse(
		"Target enchanted creature gains trample until end of turn.",
		Context{InstantOrSorcery: true},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if len(document.Abilities) != 1 {
		t.Fatalf("abilities = %#v", document.Abilities)
	}
	if len(document.Abilities[0].SemanticReferences) != 0 {
		t.Fatalf("semantic references = %#v, want none", document.Abilities[0].SemanticReferences)
	}
}
