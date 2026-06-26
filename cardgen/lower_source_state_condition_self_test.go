package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerSourceStateConditionSelfStatic proves that self-statics gated on a
// source-permanent-state condition expressed with the bare pronoun subject
// ("as long as it's attacking", "... it's untapped", "... it's equipped",
// "... it's enchanted") lower onto an ObjectMatches selection bound to the
// source permanent, with the matching combat/tap/attachment filter set.
func TestLowerSourceStateConditionSelfStatic(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		oracleText string
		selection  game.Selection
		power      int
		toughness  int
		keywords   []game.Keyword
	}{
		"pronoun attacking keyword": {
			oracleText: "This creature has first strike as long as it's attacking.",
			selection:  game.Selection{CombatState: game.CombatStateAttacking},
			keywords:   []game.Keyword{game.FirstStrike},
		},
		"pronoun untapped pt": {
			oracleText: "This creature gets +0/+3 as long as it's untapped.",
			selection:  game.Selection{Tapped: game.TriFalse},
			toughness:  3,
		},
		"pronoun equipped keyword": {
			oracleText: "This creature has flying as long as it's equipped.",
			selection:  game.Selection{MatchEquipped: true},
			keywords:   []game.Keyword{game.Flying},
		},
		"pronoun enchanted keyword": {
			oracleText: "This creature has flying as long as it's enchanted.",
			selection:  game.Selection{MatchEnchanted: true},
			keywords:   []game.Keyword{game.Flying},
		},
		"explicit subject equipped keyword": {
			oracleText: "This creature has flying as long as this creature is equipped.",
			selection:  game.Selection{MatchEquipped: true},
			keywords:   []game.Keyword{game.Flying},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ability := lowerSelfStatic(t, test.oracleText)
			if !ability.Condition.Exists {
				t.Fatalf("condition missing: %#v", ability)
			}
			condition := ability.Condition.Val
			if !condition.Object.Exists || condition.Object.Val != game.SourcePermanentReference() {
				t.Fatalf("condition Object = %#v, want source permanent reference", condition.Object)
			}
			if !condition.ObjectMatches.Exists {
				t.Fatalf("ObjectMatches missing: %#v", condition)
			}
			got := condition.ObjectMatches.Val
			if got.CombatState != test.selection.CombatState ||
				got.Tapped != test.selection.Tapped ||
				got.MatchEquipped != test.selection.MatchEquipped ||
				got.MatchEnchanted != test.selection.MatchEnchanted {
				t.Fatalf("selection = %#v, want %#v", got, test.selection)
			}
			assertSelfContinuous(t, &ability, test.power, test.toughness, test.keywords)
		})
	}
}
