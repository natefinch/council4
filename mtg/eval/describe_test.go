package eval

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestDescribeGlossesCostsThenEffects(t *testing.T) {
	ability := ScorableAbility{
		Costs: []cost.Additional{
			{Kind: cost.AdditionalSacrifice, MatchPermanentType: true, PermanentType: types.Creature},
		},
		Effect: []EffectAtom{
			{Kind: EffectCardsDrawn, Amount: 1, Affected: AffectedYou},
		},
	}
	if got, want := Describe(ability), "sacrifice a creature, draw a card"; got != want {
		t.Fatalf("Describe = %q, want %q", got, want)
	}
}

func TestDescribeEffectPhrases(t *testing.T) {
	cases := []struct {
		name string
		atom EffectAtom
		want string
	}{
		{"draw two", EffectAtom{Kind: EffectCardsDrawn, Amount: 2, Affected: AffectedYou}, "draw 2 cards"},
		{"lose cards", EffectAtom{Kind: EffectCardsLost, Amount: 3, Affected: AffectedYou}, "lose 3 cards"},
		{"opponent loses cards", EffectAtom{Kind: EffectCardsLost, Amount: 1, Affected: AffectedEachOpponent}, "each opponent loses a card"},
		{"gain life", EffectAtom{Kind: EffectLifeGained, Amount: 4, Affected: AffectedYou}, "gain 4 life"},
		{"damage", EffectAtom{Kind: EffectDamageDealt, Amount: 3}, "deal 3 damage"},
		{"removal", EffectAtom{Kind: EffectPermanentRemoved, Affected: AffectedTarget}, "remove a permanent"},
		{"tap", EffectAtom{Kind: EffectPermanentTapped, Affected: AffectedTarget}, "tap a permanent"},
		{"mana", EffectAtom{Kind: EffectManaAdded, Amount: 2}, "add 2 mana"},
		{"token", EffectAtom{Kind: EffectTokenCreated, Amount: 1}, "create a token"},
		{"counter", EffectAtom{Kind: EffectCounterAdded, Amount: 2}, "put 2 counters"},
		{"tutor", EffectAtom{Kind: EffectCardTutored, Amount: 1}, "search your library"},
		{"dynamic draw", EffectAtom{Kind: EffectCardsDrawn, IsDynamic: true, Affected: AffectedYou}, "draw X cards"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := Describe(ScorableAbility{Effect: []EffectAtom{c.atom}}); got != c.want {
				t.Fatalf("Describe = %q, want %q", got, c.want)
			}
		})
	}
}

func TestDescribeOmitsUnmodeledEffects(t *testing.T) {
	if got := Describe(ScorableAbility{Effect: []EffectAtom{{Kind: EffectNeutral}}}); got != "" {
		t.Fatalf("Describe of neutral-only ability = %q, want empty", got)
	}
}

func TestDescribePrefersPreservedCostText(t *testing.T) {
	ability := ScorableAbility{
		Costs: []cost.Additional{{Kind: cost.AdditionalSacrifice, Text: "Sacrifice an artifact or creature"}},
	}
	if got, want := Describe(ability), "sacrifice an artifact or creature"; got != want {
		t.Fatalf("Describe = %q, want %q", got, want)
	}
}
