package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerLeadingResolvingGatePreservesGroup proves the end-to-end payoff of the
// parser stripping a leading resolving gate from a pump clause: "If this spell was
// kicked, creatures you control get +2/+2 until end of turn" lowers to a single
// group ApplyContinuous over the controlled-creatures group (not the resolving
// spell) that is gated on the SpellWasKicked condition. Before the parser change
// the leading condition suppressed the group subject, so lowering had no group to
// buff; this confirms the recovered subject and the retained gate survive lowering
// together.
func TestLowerLeadingResolvingGatePreservesGroup(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Kicked Group Pump",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		ManaCost:   "{2}{G}",
		OracleText: "Kicker {2} (You may pay an additional {2} as you cast this spell.)\nIf this spell was kicked, creatures you control get +2/+2 until end of turn.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("no spell ability lowered")
	}
	sequence := face.SpellAbility.Val.Modes[0].Sequence
	if len(sequence) != 1 {
		t.Fatalf("sequence length = %d, want one gated group pump", len(sequence))
	}
	instruction := sequence[0]
	if !instruction.Condition.Exists || !instruction.Condition.Val.Condition.Exists ||
		!instruction.Condition.Val.Condition.Val.SpellWasKicked {
		t.Fatalf("instruction condition = %#v, want a SpellWasKicked gate", instruction.Condition)
	}
	apply, ok := instruction.Primitive.(game.ApplyContinuous)
	if !ok {
		t.Fatalf("primitive = %#v, want group ApplyContinuous", instruction.Primitive)
	}
	if apply.Object.Exists {
		t.Fatalf("ApplyContinuous.Object = %#v, want unset (group form, not a single object)", apply.Object)
	}
	if len(apply.ContinuousEffects) != 1 {
		t.Fatalf("continuous effects = %d, want one", len(apply.ContinuousEffects))
	}
	if apply.ContinuousEffects[0].AffectedSource {
		t.Fatal("continuous effect AffectedSource = true, want false (buffs the group, not the resolving spell)")
	}
}

// TestLowerFinaleHeadlineRiderFailsClosed proves the literal Finale of Devastation
// pump rider stays fail closed at lowering even though the parser now recovers its
// group subject: the executable backend supports only fixed group power/toughness
// changes and single-target keyword grants, so the dynamic "+X/+X" group pump and
// the group keyword grant on a resolving spell are both unsupported. The parser
// fix is text-blind and composes the leading gate with any downstream effect;
// where the downstream effect is unsupported the card correctly fails closed
// rather than shipping a partial ability.
func TestLowerFinaleHeadlineRiderFailsClosed(t *testing.T) {
	t.Parallel()
	face := lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
		Name:       "Finale Headline Rider",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		ManaCost:   "{X}{G}{G}",
		OracleText: "If X is 10 or more, creatures you control get +X/+X and gain haste until end of turn.",
	})
	if face.SpellAbility.Exists {
		t.Fatal("unsupported dynamic group pump rider produced a spell ability, want none")
	}
}

// TestLowerSpellSourceBackReferenceKeywordGrantFailsClosed proves a resolving
// spell whose keyword grant back-references a creature it cannot tie to its target
// or group antecedent fails closed rather than granting the keyword to the spell
// itself. Heroic Charge's "those creatures also gain trample" and Arrester's Zeal's
// split "that creature gains flying" both fall back to the spell source, which the
// lowering guard rejects.
func TestLowerSpellSourceBackReferenceKeywordGrantFailsClosed(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name     string
		typeLine string
		cost     string
		oracle   string
	}{
		{
			name:     "group back-reference",
			typeLine: "Instant",
			cost:     "{2}{W}{W}",
			oracle:   "Kicker {1}{R} (You may pay an additional {1}{R} as you cast this spell.)\nCreatures you control get +2/+1 until end of turn. If this spell was kicked, those creatures also gain trample until end of turn.",
		},
		{
			name:     "split addendum target back-reference",
			typeLine: "Instant",
			cost:     "{W}",
			oracle:   "Target creature gets +2/+2 until end of turn.\nAddendum — If you cast this spell during your main phase, that creature gains flying until end of turn.",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
				Name:       "Back Reference Grant",
				Layout:     "normal",
				TypeLine:   tc.typeLine,
				ManaCost:   tc.cost,
				OracleText: tc.oracle,
			})
			if face.SpellAbility.Exists {
				t.Fatal("unsupported back-reference keyword grant produced a spell ability, want none")
			}
		})
	}
}

// TestLowerPermanentSourceKeywordGrantStillLowers proves the spell-only guard does
// not over-reach: a permanent's activated ability that grants a keyword to its own
// source via a referenced-object back-reference ("It gains hexproof until end of
// turn") still lowers, because a permanent's source is a real battlefield object.
// This is the non-regression counterpart to the spell fail-closed shapes.
func TestLowerPermanentSourceKeywordGrantStillLowers(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Counter Beast",
		Layout:     "normal",
		TypeLine:   "Creature — Beast",
		ManaCost:   "{2}{G}",
		OracleText: "{1}{G}: Put a +1/+1 counter on this creature. It gains hexproof until end of turn.",
	})
	if len(face.ActivatedAbilities) == 0 {
		t.Fatal("activated ability did not lower")
	}
}
