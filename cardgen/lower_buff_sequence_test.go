package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerMultiTargetCombinedBuffSpell covers the multi-target Overrun-style
// combined buff "Up to two target creatures each get +1/+1 and gain lifelink
// until end of turn." Each chosen target slot receives its own ApplyContinuous
// carrying both the power/toughness and keyword layers, addressing its own
// target index.
func TestLowerMultiTargetCombinedBuffSpell(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Cutthroat",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Up to two target creatures each get +1/+1 and gain lifelink until end of turn.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability not lowered")
	}
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 1 {
		t.Fatalf("targets = %d, want exactly 1 target spec", len(mode.Targets))
	}
	if mode.Targets[0].MinTargets != 0 || mode.Targets[0].MaxTargets != 2 {
		t.Fatalf("cardinality = [%d,%d], want [0,2]", mode.Targets[0].MinTargets, mode.Targets[0].MaxTargets)
	}
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence = %d, want one ApplyContinuous per target slot", len(mode.Sequence))
	}
	for i := range mode.Sequence {
		apply, ok := mode.Sequence[i].Primitive.(game.ApplyContinuous)
		if !ok {
			t.Fatalf("sequence[%d] primitive = %T, want game.ApplyContinuous", i, mode.Sequence[i].Primitive)
		}
		if !apply.Object.Exists || apply.Object.Val.TargetIndex() != i {
			t.Fatalf("sequence[%d] object = %+v, want target index %d", i, apply.Object, i)
		}
		if apply.Duration != game.DurationUntilEndOfTurn {
			t.Fatalf("sequence[%d] duration = %v, want until end of turn", i, apply.Duration)
		}
		if len(apply.ContinuousEffects) != 2 {
			t.Fatalf("sequence[%d] continuous effects = %d, want 2 (P/T + keyword)", i, len(apply.ContinuousEffects))
		}
		pt := apply.ContinuousEffects[0]
		if pt.Layer != game.LayerPowerToughnessModify || pt.PowerDelta != 1 || pt.ToughnessDelta != 1 {
			t.Fatalf("sequence[%d] P/T layer = %+v, want +1/+1", i, pt)
		}
		kw := apply.ContinuousEffects[1]
		if kw.Layer != game.LayerAbility || len(kw.AddKeywords) != 1 || kw.AddKeywords[0] != game.Lifelink {
			t.Fatalf("sequence[%d] keyword layer = %+v, want lifelink", i, kw)
		}
	}
}

// TestLowerBuffThenDemonstrativeReference covers a buff clause followed by a
// second clause whose object is the singular demonstrative "that creature"
// referring back to the buffed target. Both the untap form ("…Untap that
// creature.") and the tap form ("…Tap that creature.") must address the same
// target index as the buff.
func TestLowerBuffThenDemonstrativeReference(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name       string
		oracle     string
		wantSecond func(game.Instruction) bool
	}{
		{
			name:   "untap",
			oracle: "Target creature gets +2/+2 until end of turn. Untap that creature.",
			wantSecond: func(in game.Instruction) bool {
				untap, ok := in.Primitive.(game.Untap)
				return ok && untap.Object.Kind() == game.ObjectReferenceTargetPermanent && untap.Object.TargetIndex() == 0
			},
		},
		{
			name:   "tap",
			oracle: "Target creature gets -1/-1 until end of turn. Tap that creature.",
			wantSecond: func(in game.Instruction) bool {
				tap, ok := in.Primitive.(game.Tap)
				return ok && tap.Object.Kind() == game.ObjectReferenceTargetPermanent && tap.Object.TargetIndex() == 0
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Demonstrative",
				Layout:     "normal",
				TypeLine:   "Instant",
				OracleText: tc.oracle,
			})
			if !face.SpellAbility.Exists {
				t.Fatal("spell ability not lowered")
			}
			mode := face.SpellAbility.Val.Modes[0]
			if len(mode.Targets) != 1 {
				t.Fatalf("targets = %d, want exactly 1 (shared by both clauses)", len(mode.Targets))
			}
			if len(mode.Sequence) != 2 {
				t.Fatalf("sequence = %d, want 2", len(mode.Sequence))
			}
			modify, ok := mode.Sequence[0].Primitive.(game.ModifyPT)
			if !ok || modify.Object.TargetIndex() != 0 {
				t.Fatalf("sequence[0] = %+v, want ModifyPT on target 0", mode.Sequence[0].Primitive)
			}
			if !tc.wantSecond(mode.Sequence[1]) {
				t.Fatalf("sequence[1] = %+v, did not match expected back-reference clause", mode.Sequence[1].Primitive)
			}
		})
	}
}
