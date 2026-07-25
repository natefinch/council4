package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestDistributiveRemovalCoversUnwrittenScopeVerbCombinations is the payoff test
// for collapsing the distributive removal family (epic #3172, step 3).
//
// Before the collapse, each scope-and-verb combination needed its own parser
// boolean, its own recognizer, its own primitive, its own runtime handler, and
// its own lowering matcher - so a card gained support only if someone had
// written that exact combination. "For each opponent, destroy ..." and "For each
// player, exile ... then draw" are wordings no real card had when the collapse
// landed, and no code in this repository names them.
//
// They lower anyway, because scope and verb are now parameters. If this test
// starts failing, the family has been re-specialized and the throughput problem
// the epic exists to fix has come back.
func TestDistributiveRemovalCoversUnwrittenScopeVerbCombinations(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		typeLine   string
		layout     string
		oracleText string
		abilities  string
		scope      game.PlayerGroupReference
		removal    game.DistributiveRemoval
	}{
		{
			name:     "destroy for each opponent with a token payoff",
			typeLine: "Enchantment — Saga",
			layout:   "saga",
			oracleText: "(As this Saga enters and after your draw step, add a lore counter. Sacrifice after III.)\n" +
				"I, II, III — For each opponent, destroy up to one target creature that player controls. " +
				"For each creature destroyed this way, its controller creates a 3/3 green Mutant creature token with deathtouch.",
			abilities: "ChapterAbilities[0]",
			scope:     game.OpponentsReference(),
			removal:   game.DistributiveRemovalDestroy,
		},
		{
			name:     "exile for each player with a draw payoff",
			typeLine: "Creature — Human Wizard",
			layout:   "normal",
			oracleText: "When this creature enters, for each player, exile up to one target permanent that player controls. " +
				"For each permanent exiled this way, its controller draws a card.",
			abilities: "TriggeredAbilities[0]",
			scope:     game.AllPlayersReference(),
			removal:   game.DistributiveRemovalExile,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			card := &ScryfallCard{
				Name:       "Distributive Test Creature",
				Layout:     test.layout,
				ManaCost:   "{2}{G}{W}",
				TypeLine:   test.typeLine,
				OracleText: test.oracleText,
				Colors:     []string{"G", "W"},
			}
			prefix := test.abilities + ".Content.Modes[0].Sequence[0].Primitive.(game.ForEachPlayer)"
			assertCardPaths(t, card,
				prefix+".Scope.Kind = "+enumPathValue(test.scope.Kind),
				prefix+".Removal = "+enumPathValue(test.removal),
			)
		})
	}
}

// enumPathValue spells an enum the way the CardDef path dumper does, so a test
// can name an expected constant instead of hard-coding its rendering.
func enumPathValue(value any) string {
	spelling, ok := enumSpelling(value)
	if !ok {
		panic("enumPathValue: no spelling for value")
	}
	return spelling
}
