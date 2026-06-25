package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestLowerMultiDistinctTargetDestroy proves the multi-distinct-typed-target
// destroy "Destroy target artifact, target creature, target enchantment, and
// target land." (Decimate) and its shorter dual-target forms lower to one
// single-target spec per "target <type>" clause in Oracle order and one Destroy
// per slot, each addressing its own target index. Each spec carries the type
// predicate of its own clause, so the four chosen permanents are independent
// single targets of distinct types.
func TestLowerMultiDistinctTargetDestroy(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		oracleText string
		permTypes  []types.Card
	}{
		{
			name:       "decimate four distinct types",
			oracleText: "Destroy target artifact, target creature, target enchantment, and target land.",
			permTypes:  []types.Card{types.Artifact, types.Creature, types.Enchantment, types.Land},
		},
		{
			name:       "artifact and creature",
			oracleText: "Destroy target artifact and target creature.",
			permTypes:  []types.Card{types.Artifact, types.Creature},
		},
		{
			name:       "creature and land",
			oracleText: "Destroy target creature and target land.",
			permTypes:  []types.Card{types.Creature, types.Land},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Decimate",
				Layout:     "normal",
				ManaCost:   "{2}{R}{G}",
				TypeLine:   "Sorcery",
				OracleText: test.oracleText,
			})
			if !face.SpellAbility.Exists {
				t.Fatal("spell ability not lowered")
			}
			mode := face.SpellAbility.Val.Modes[0]
			if len(mode.Targets) != len(test.permTypes) {
				t.Fatalf("targets = %d, want %d specs", len(mode.Targets), len(test.permTypes))
			}
			if len(mode.Sequence) != len(test.permTypes) {
				t.Fatalf("sequence = %d, want %d", len(mode.Sequence), len(test.permTypes))
			}
			for i := range mode.Targets {
				spec := mode.Targets[i]
				if spec.MinTargets != 1 || spec.MaxTargets != 1 {
					t.Fatalf("spec[%d] cardinality = {%d,%d}, want {1,1}", i, spec.MinTargets, spec.MaxTargets)
				}
				if spec.Allow != game.TargetAllowPermanent {
					t.Fatalf("spec[%d] allow = %v, want TargetAllowPermanent", i, spec.Allow)
				}
				if len(spec.Selection.Val.RequiredTypesAny) != 1 || spec.Selection.Val.RequiredTypesAny[0] != test.permTypes[i] {
					t.Fatalf("spec[%d] types = %v, want [%v]", i, spec.Selection.Val.RequiredTypesAny, test.permTypes[i])
				}
				destroy, ok := mode.Sequence[i].Primitive.(game.Destroy)
				if !ok || destroy.Object != game.TargetPermanentReference(i) {
					t.Fatalf("sequence[%d] = %#v, want Destroy of TargetPermanentReference(%d)", i, mode.Sequence[i].Primitive, i)
				}
			}
		})
	}
}

// TestLowerMultiDistinctTargetExile proves the exile counterpart "Exile target
// artifact and target creature." lowers to one single-target spec per clause and
// one Exile per slot, mirroring the destroy form.
func TestLowerMultiDistinctTargetExile(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Dual Exile",
		Layout:     "normal",
		ManaCost:   "{2}{W}{B}",
		TypeLine:   "Sorcery",
		OracleText: "Exile target artifact and target creature.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability not lowered")
	}
	mode := face.SpellAbility.Val.Modes[0]
	wantTypes := []types.Card{types.Artifact, types.Creature}
	if len(mode.Targets) != len(wantTypes) || len(mode.Sequence) != len(wantTypes) {
		t.Fatalf("targets = %d, sequence = %d, want %d each", len(mode.Targets), len(mode.Sequence), len(wantTypes))
	}
	for i := range mode.Targets {
		if len(mode.Targets[i].Selection.Val.RequiredTypesAny) != 1 || mode.Targets[i].Selection.Val.RequiredTypesAny[0] != wantTypes[i] {
			t.Fatalf("spec[%d] types = %v, want [%v]", i, mode.Targets[i].Selection.Val.RequiredTypesAny, wantTypes[i])
		}
		exile, ok := mode.Sequence[i].Primitive.(game.Exile)
		if !ok || exile.Object != game.TargetPermanentReference(i) {
			t.Fatalf("sequence[%d] = %#v, want Exile of TargetPermanentReference(%d)", i, mode.Sequence[i].Primitive, i)
		}
	}
}

// TestLowerMultiDistinctTargetSingleUnchanged proves the single-target destroy
// form stays on the single-target path: the multi-distinct recognizer requires
// two or more targets, so "Destroy target creature." still lowers to one spec.
func TestLowerMultiDistinctTargetSingleUnchanged(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Single Destroy",
		Layout:     "normal",
		ManaCost:   "{1}{B}",
		TypeLine:   "Sorcery",
		OracleText: "Destroy target creature.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability not lowered")
	}
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 1 || len(mode.Sequence) != 1 {
		t.Fatalf("targets = %d, sequence = %d, want 1 each", len(mode.Targets), len(mode.Sequence))
	}
}
