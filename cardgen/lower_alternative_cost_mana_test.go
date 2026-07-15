package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestLowerZeroManaAlternativeCost proves a {0} conditional alternative cost
// lowers to an explicit single-symbol {0} mana cost carried on the alternative,
// distinct from a without-paying-mana-cost ("free") alternative.
func TestLowerZeroManaAlternativeCost(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Zero Trap",
		Layout:   "normal",
		TypeLine: "Instant — Trap",
		ManaCost: "{4}{U}{U}",
		OracleText: "If an opponent gained life this turn, you may pay {0} rather than pay this spell's mana cost.\n" +
			"Draw two cards.",
	})
	if len(face.AlternativeCosts) != 1 {
		t.Fatalf("alternative costs = %#v, want one", face.AlternativeCosts)
	}
	alternative := face.AlternativeCosts[0]
	if !alternative.ManaCost.Exists {
		t.Fatalf("mana cost = %#v, want an explicit {0} cost", alternative.ManaCost)
	}
	if alternative.ManaCost.Val.String() != "{0}" {
		t.Fatalf("mana cost = %q, want {0}", alternative.ManaCost.Val.String())
	}
	if alternative.Condition != cost.AlternativeConditionOpponentGainedLifeThisTurn {
		t.Fatalf("condition = %#v, want opponent-gained-life", alternative.Condition)
	}
	if alternative.Label != "Pay {0}" {
		t.Fatalf("label = %q, want \"Pay {0}\"", alternative.Label)
	}
}

func TestLowerCreaturesAttackingManaAlternativeCost(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name    string
		text    string
		count   int
		exactly bool
		mana    string
	}{
		{
			name:    "N or more",
			text:    "If three or more creatures are attacking, you may pay {U} rather than pay this spell's mana cost.",
			count:   3,
			exactly: false,
			mana:    "{U}",
		},
		{
			name:    "exactly one",
			text:    "If exactly one creature is attacking, you may pay {W} rather than pay this spell's mana cost.",
			count:   1,
			exactly: true,
			mana:    "{W}",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Attack Trap",
				Layout:     "normal",
				TypeLine:   "Instant — Trap",
				ManaCost:   "{3}{W}",
				OracleText: tc.text + "\nDraw a card.",
			})
			if len(face.AlternativeCosts) != 1 {
				t.Fatalf("alternative costs = %#v, want one", face.AlternativeCosts)
			}
			alternative := face.AlternativeCosts[0]
			if alternative.Condition != cost.AlternativeConditionCreaturesAttacking {
				t.Fatalf("condition = %#v, want creatures-attacking", alternative.Condition)
			}
			if alternative.ConditionCount != tc.count || alternative.ConditionExactly != tc.exactly {
				t.Fatalf("count/exactly = %d/%t, want %d/%t",
					alternative.ConditionCount, alternative.ConditionExactly, tc.count, tc.exactly)
			}
			if !alternative.ManaCost.Exists || alternative.ManaCost.Val.String() != tc.mana {
				t.Fatalf("mana cost = %#v, want %q", alternative.ManaCost, tc.mana)
			}
		})
	}
}

// TestLowerManaAlternativeCostFailsClosed proves the lowering (via the parser it
// drives) refuses to emit a mana-only alternative cost for an unmodeled Trap
// condition, reporting the card as unsupported rather than approximating it.
func TestLowerManaAlternativeCostFailsClosed(t *testing.T) {
	t.Parallel()
	lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
		Name:     "Unmodeled Trap",
		Layout:   "normal",
		TypeLine: "Instant — Trap",
		ManaCost: "{2}{U}",
		OracleText: "If an opponent cast two or more spells this turn, you may pay {0} rather than pay this spell's mana cost.\n" +
			"Counter target spell.",
	})
}

// TestLowerPermanentsOnBattlefieldManaAlternativeCost proves Blasphemous Edict's
// board-state gate lowers to a cost.Alternative carrying the
// permanents-on-battlefield condition, its thirteen-permanent threshold, and the
// counted creature permanent type, so the payment planner can offer "Pay {B}"
// only when thirteen or more creatures are on the battlefield.
func TestLowerPermanentsOnBattlefieldManaAlternativeCost(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Board Trap",
		Layout:   "normal",
		TypeLine: "Sorcery",
		ManaCost: "{3}{B}{B}",
		OracleText: "You may pay {B} rather than pay this spell's mana cost if there are thirteen or more creatures on the battlefield.\n" +
			"Draw a card.",
	})
	if len(face.AlternativeCosts) != 1 {
		t.Fatalf("alternative costs = %#v, want one", face.AlternativeCosts)
	}
	alternative := face.AlternativeCosts[0]
	if alternative.Condition != cost.AlternativeConditionPermanentsOnBattlefield {
		t.Fatalf("condition = %#v, want permanents-on-battlefield", alternative.Condition)
	}
	if alternative.ConditionCount != 13 {
		t.Fatalf("count = %d, want 13", alternative.ConditionCount)
	}
	if alternative.ConditionPermanentType != types.Creature {
		t.Fatalf("permanent type = %#v, want creature", alternative.ConditionPermanentType)
	}
	if alternative.ConditionExactly {
		t.Fatal("board-state gate must never be an exact-count comparison")
	}
	if !alternative.ManaCost.Exists || alternative.ManaCost.Val.String() != "{B}" {
		t.Fatalf("mana cost = %#v, want {B}", alternative.ManaCost)
	}
	if alternative.Label != "Pay {B}" {
		t.Fatalf("label = %q, want \"Pay {B}\"", alternative.Label)
	}
}

// TestLowerPermanentsOnBattlefieldManaAlternativeCostFailsClosed proves a
// near-miss board-state gate (a non-permanent counted type) is left unsupported
// rather than lowered to an approximate condition.
func TestLowerPermanentsOnBattlefieldManaAlternativeCostFailsClosed(t *testing.T) {
	t.Parallel()
	lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
		Name:     "Bad Board Trap",
		Layout:   "normal",
		TypeLine: "Sorcery",
		ManaCost: "{3}{B}{B}",
		OracleText: "You may pay {B} rather than pay this spell's mana cost if there are thirteen or more instants on the battlefield.\n" +
			"Draw a card.",
	})
}
