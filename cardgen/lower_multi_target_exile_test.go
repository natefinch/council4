package cardgen

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestLowerMultiTargetExile proves plural and optional permanent exile wordings
// lower to a single multi-target spec carrying the cardinality range and one
// Exile instruction per slot, each addressing its own target index.
func TestLowerMultiTargetExile(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		oracleText string
		minTargets int
		maxTargets int
		permType   types.Card
	}{
		{
			name:       "fixed two",
			oracleText: "Exile two target artifacts.",
			minTargets: 2,
			maxTargets: 2,
			permType:   types.Artifact,
		},
		{
			name:       "up to two",
			oracleText: "Exile up to two target creatures.",
			minTargets: 0,
			maxTargets: 2,
			permType:   types.Creature,
		},
		{
			name:       "up to three",
			oracleText: "Exile up to three target enchantments.",
			minTargets: 0,
			maxTargets: 3,
			permType:   types.Enchantment,
		},
		{
			name:       "up to one",
			oracleText: "Exile up to one target permanent.",
			minTargets: 0,
			maxTargets: 1,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Multi Exile",
				Layout:     "normal",
				TypeLine:   "Sorcery",
				OracleText: test.oracleText,
			})
			mode := face.SpellAbility.Val.Modes[0]
			if len(mode.Targets) != 1 {
				t.Fatalf("targets = %#v, want one spec", mode.Targets)
			}
			spec := mode.Targets[0]
			if spec.MinTargets != test.minTargets || spec.MaxTargets != test.maxTargets {
				t.Fatalf("cardinality = {%d,%d}, want {%d,%d}", spec.MinTargets, spec.MaxTargets, test.minTargets, test.maxTargets)
			}
			if spec.Allow != game.TargetAllowPermanent {
				t.Fatalf("allow = %v, want TargetAllowPermanent", spec.Allow)
			}
			if test.permType != "" {
				if len(spec.Selection.Val.RequiredTypesAny) != 1 || spec.Selection.Val.RequiredTypesAny[0] != test.permType {
					t.Fatalf("predicate types = %v, want [%v]", spec.Selection.Val.RequiredTypesAny, test.permType)
				}
			}
			if len(mode.Sequence) != test.maxTargets {
				t.Fatalf("sequence len = %d, want %d", len(mode.Sequence), test.maxTargets)
			}
			for i := range mode.Sequence {
				exile, ok := mode.Sequence[i].Primitive.(game.Exile)
				if !ok {
					t.Fatalf("sequence[%d] = %T, want game.Exile", i, mode.Sequence[i].Primitive)
				}
				if exile.Object != game.TargetPermanentReference(i) {
					t.Fatalf("sequence[%d] object = %v, want TargetPermanentReference(%d)", i, exile.Object, i)
				}
			}
		})
	}
}

// TestLowerMultiTargetExileSingleTargetUnchanged proves the single-target exile
// path is untouched: it still lowers to one spec with one Exile instruction.
func TestLowerMultiTargetExileSingleTargetUnchanged(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Single Exile",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Exile target creature.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 1 || mode.Targets[0].MinTargets != 1 || mode.Targets[0].MaxTargets != 1 {
		t.Fatalf("targets = %#v, want one {1,1} spec", mode.Targets)
	}
	if len(mode.Sequence) != 1 {
		t.Fatalf("sequence len = %d, want 1", len(mode.Sequence))
	}
	if exile, ok := mode.Sequence[0].Primitive.(game.Exile); !ok || exile.Object != game.TargetPermanentReference(0) {
		t.Fatalf("sequence[0] = %#v, want Exile of TargetPermanentReference(0)", mode.Sequence[0].Primitive)
	}
}

// TestLowerMultiTargetExileFailClosed proves shapes the executable backend
// cannot represent exactly stay rejected with a diagnostic and no partial card.
func TestLowerMultiTargetExileFailClosed(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		oracleText string
	}{
		{name: "graveyard zone", oracleText: "Exile up to two target cards from a single graveyard."},
		{name: "subtype qualifier", oracleText: "Exile up to two target Goblin creatures."},
		{name: "tapped qualifier", oracleText: "Exile two target tapped creatures."},
		{name: "unbounded any number", oracleText: "Exile any number of target creatures."},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
				Name:       "Test Reject Exile",
				Layout:     "normal",
				TypeLine:   "Sorcery",
				OracleText: test.oracleText,
			}, "t")
			if err != nil {
				t.Fatal(err)
			}
			if source != "" {
				t.Fatalf("source = %q, want no partial card", source)
			}
			if len(diagnostics) == 0 {
				t.Fatal("expected unsupported diagnostic")
			}
		})
	}
}

// TestLowerUnionExile proves an Oxford-comma card-type union and a subtype union
// lower to one single-target spec whose predicate carries every union member,
// driving a single Exile of that target.
func TestLowerUnionExile(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		oracleText string
		types      []types.Card
		subtypes   []types.Sub
	}{
		{
			name:       "card-type union",
			oracleText: "Exile target artifact, creature, or enchantment.",
			types:      []types.Card{types.Artifact, types.Creature, types.Enchantment},
		},
		{
			name:       "card-type union with land",
			oracleText: "Exile target artifact, creature, or land.",
			types:      []types.Card{types.Artifact, types.Creature, types.Land},
		},
		{
			name:       "subtype union",
			oracleText: "Exile target Skeleton, Vampire, or Zombie.",
			subtypes:   []types.Sub{types.Sub("Skeleton"), types.Sub("Vampire"), types.Sub("Zombie")},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Union Exile",
				Layout:     "normal",
				TypeLine:   "Sorcery",
				OracleText: test.oracleText,
			})
			mode := face.SpellAbility.Val.Modes[0]
			if len(mode.Targets) != 1 {
				t.Fatalf("targets = %#v, want one spec", mode.Targets)
			}
			spec := mode.Targets[0]
			if spec.MinTargets != 1 || spec.MaxTargets != 1 || spec.Allow != game.TargetAllowPermanent {
				t.Fatalf("spec = %#v, want one {1,1} permanent target", spec)
			}
			if !slices.Equal(spec.Selection.Val.RequiredTypesAny, test.types) {
				t.Fatalf("predicate types = %v, want %v", spec.Selection.Val.RequiredTypesAny, test.types)
			}
			if !slices.Equal(spec.Selection.Val.SubtypesAny, test.subtypes) {
				t.Fatalf("predicate subtypes = %v, want %v", spec.Selection.Val.SubtypesAny, test.subtypes)
			}
			if len(mode.Sequence) != 1 {
				t.Fatalf("sequence len = %d, want 1", len(mode.Sequence))
			}
			exile, ok := mode.Sequence[0].Primitive.(game.Exile)
			if !ok || exile.Object != game.TargetPermanentReference(0) {
				t.Fatalf("sequence[0] = %#v, want Exile of TargetPermanentReference(0)", mode.Sequence[0].Primitive)
			}
		})
	}
}
