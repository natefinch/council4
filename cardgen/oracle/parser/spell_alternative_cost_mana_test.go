package parser

import (
	"testing"
)

// parseManaAlternative parses a spell whose first line is a mana-only
// alternative cost clause and returns the recognized alternative cost, failing
// the test if the clause was not recognized as a typed alternative cost.
func parseManaAlternative(t *testing.T, source string) *SpellAlternativeCost {
	t.Helper()
	document, diagnostics := Parse(source, Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if len(document.Abilities) == 0 {
		t.Fatal("no abilities parsed")
	}
	ability := document.Abilities[0]
	if ability.Kind != AbilitySpellAlternativeCost || ability.AlternativeCost == nil {
		t.Fatalf("ability = %#v, want typed alternative spell cost", ability)
	}
	if ability.AlternativeCost.Kind != SpellAlternativeCostMana {
		t.Fatalf("alternative kind = %#v, want mana-only", ability.AlternativeCost.Kind)
	}
	return ability.AlternativeCost
}

func TestParseUnconditionalManaAlternativeCost(t *testing.T) {
	t.Parallel()
	alternative := parseManaAlternative(t,
		"You may pay {W}{U}{B}{R}{G} rather than pay this spell's mana cost.\nDraw a card.")
	if alternative.Condition != SpellAlternativeCostConditionUnknown {
		t.Fatalf("condition = %#v, want unconditional", alternative.Condition)
	}
	if alternative.ManaCost.String() != "{W}{U}{B}{R}{G}" {
		t.Fatalf("mana cost = %q, want {W}{U}{B}{R}{G}", alternative.ManaCost.String())
	}
}

func TestParseZeroManaAlternativeCostIsDistinct(t *testing.T) {
	t.Parallel()
	alternative := parseManaAlternative(t,
		"You may pay {0} rather than pay this spell's mana cost.\nDraw a card.")
	// {0} is a real, payable mana cost that must be preserved as an explicit
	// mana symbol, not collapsed into a "free/no cost" absence.
	if len(alternative.ManaCost) != 1 {
		t.Fatalf("mana cost = %#v, want a single explicit {0} symbol", alternative.ManaCost)
	}
	if alternative.ManaCost.String() != "{0}" {
		t.Fatalf("mana cost = %q, want {0}", alternative.ManaCost.String())
	}
	if alternative.WithoutPayingManaCost {
		t.Fatal("{0} alternative was flagged as without-paying-mana-cost")
	}
}

func TestParseOpponentGainedLifeManaAlternativeCost(t *testing.T) {
	t.Parallel()
	alternative := parseManaAlternative(t,
		"If an opponent gained life this turn, you may pay {B} rather than pay this spell's mana cost.\n"+
			"Target player loses 5 life and you gain 5 life.")
	if alternative.Condition != SpellAlternativeCostConditionOpponentGainedLifeThisTurn {
		t.Fatalf("condition = %#v, want opponent-gained-life", alternative.Condition)
	}
	if alternative.ManaCost.String() != "{B}" {
		t.Fatalf("mana cost = %q, want {B}", alternative.ManaCost.String())
	}
}

func TestParseOpponentCastSpellsManaAlternativeCost(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name  string
		text  string
		count int
	}{
		{
			name:  "two or more",
			text:  "If an opponent cast two or more spells this turn, you may pay {0} rather than pay this spell's mana cost.",
			count: 2,
		},
		{
			name:  "three or more (Mindbreak Trap)",
			text:  "If an opponent cast three or more spells this turn, you may pay {0} rather than pay this spell's mana cost.",
			count: 3,
		},
		{
			name:  "four or more",
			text:  "If an opponent cast four or more spells this turn, you may pay {1} rather than pay this spell's mana cost.",
			count: 4,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			alternative := parseManaAlternative(t, tc.text+"\nExile any number of target spells.")
			if alternative.Condition != SpellAlternativeCostConditionOpponentCastSpellsThisTurn {
				t.Fatalf("condition = %#v, want opponent-cast-spells-this-turn", alternative.Condition)
			}
			if alternative.ConditionCount != tc.count {
				t.Fatalf("condition count = %d, want %d", alternative.ConditionCount, tc.count)
			}
			// The threshold is an at-least gate ("N or more"), never an exact
			// match, so the option is offered for N and every higher count.
			if alternative.ConditionExactly {
				t.Fatal("opponent-cast-spells threshold was flagged exact, want at-least")
			}
		})
	}
}

func TestParseCreaturesAttackingManaAlternativeCost(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name    string
		text    string
		count   int
		exactly bool
		mana    string
	}{
		{
			name:    "three or more",
			text:    "If three or more creatures are attacking, you may pay {U} rather than pay this spell's mana cost.",
			count:   3,
			exactly: false,
			mana:    "{U}",
		},
		{
			name:    "four or more",
			text:    "If four or more creatures are attacking, you may pay {1}{W} rather than pay this spell's mana cost.",
			count:   4,
			exactly: false,
			mana:    "{1}{W}",
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
			alternative := parseManaAlternative(t, tc.text+"\nDraw a card.")
			if alternative.Condition != SpellAlternativeCostConditionCreaturesAttacking {
				t.Fatalf("condition = %#v, want creatures-attacking", alternative.Condition)
			}
			if alternative.ConditionCount != tc.count || alternative.ConditionExactly != tc.exactly {
				t.Fatalf("count/exactly = %d/%t, want %d/%t",
					alternative.ConditionCount, alternative.ConditionExactly, tc.count, tc.exactly)
			}
			if alternative.ManaCost.String() != tc.mana {
				t.Fatalf("mana cost = %q, want %q", alternative.ManaCost.String(), tc.mana)
			}
		})
	}
}

// TestParseManaAlternativeCostFailsClosed proves the parser refuses to recognize
// a mana-only alternative cost whenever the leading condition, the trailing
// condition, or the replacement payment is not a shape this backend can model
// correctly. Each of these must be left for another family or reported as
// unsupported rather than approximated.
func TestParseManaAlternativeCostFailsClosed(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name string
		text string
	}{
		{
			// The opponent-cast-spells gate is specific to spells an
			// opponent cast; the same wording about the caster's own spells
			// must not be approximated onto it.
			name: "unmodeled leading condition",
			text: "If you cast two or more spells this turn, you may pay {0} rather than pay this spell's mana cost.",
		},
		{
			name: "trailing condition",
			text: "You may pay {0} rather than pay this spell's mana cost if an opponent cast a spell this turn.",
		},
		{
			name: "non-mana payment",
			text: "You may pay 4 life rather than pay this spell's mana cost.",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			document, _ := Parse(tc.text+"\nDraw a card.", Context{InstantOrSorcery: true})
			for _, ability := range document.Abilities {
				if ability.AlternativeCost != nil &&
					ability.AlternativeCost.Kind == SpellAlternativeCostMana {
					t.Fatalf("wording was wrongly recognized as a mana-only alternative cost: %#v",
						ability.AlternativeCost)
				}
			}
		})
	}
}
