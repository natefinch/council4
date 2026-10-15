package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ImposingSovereign is the card definition for Imposing Sovereign.
//
// Type: Creature — Human Noble
// Cost: {1}{W}
//
// Oracle text:
//
//	Creatures your opponents control enter tapped.
var ImposingSovereign = newImposingSovereign

func newImposingSovereign() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Imposing Sovereign",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Noble},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedGroupReplacement("Creatures your opponents control enter tapped.", game.TriggerControllerOpponent, types.Creature),
			},
			OracleText: `
			Creatures your opponents control enter tapped.
		`,
		},
	}
}
