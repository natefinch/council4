package authored

import (
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

const riveteersDecoyOracleID = "1c4918e6-3810-458e-a11b-949bbc43e7cc"

var RiveteersDecoy = cardset.Entry{
	OracleID: riveteersDecoyOracleID,
	Name:     "Riveteers Decoy",
	New:      newRiveteersDecoy,
	Verification: "mtg/rules/riveteers_decoy_test.go: " +
		"TestAuthoredRiveteersDecoyScenario",
}

func newRiveteersDecoy() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name:       "Riveteers Decoy",
			ManaCost:   opt.Val(cost.Mana{cost.O(1), cost.G}),
			Colors:     []color.Color{color.Green},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Warrior},
			Power:      opt.Val(game.PT{Value: 3}),
			Toughness:  opt.Val(game.PT{Value: 1}),
			OracleText: "This creature must be blocked if able.\nBlitz {3}{G} (If you cast this spell for its blitz cost, it gains haste and \"When this creature dies, draw a card.\" Sacrifice it at the beginning of the next end step.)",
			AlternativeCosts: []cost.Alternative{
				cost.BlitzAlternative(cost.Mana{cost.O(3), cost.G}),
			},
			StaticAbilities: []game.StaticAbility{
				game.BlitzStaticAbility(),
				game.MustBeBlockedStaticBody,
			},
		},
	}
}
