package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// EnragedRevolutionary is the card definition for Enraged Revolutionary.
//
// Type: Creature — Human Warrior
// Cost: {2}{R}
//
// Oracle text:
//
//	Dethrone (Whenever this creature attacks the player with the most life or tied for most life, put a +1/+1 counter on it.)
var EnragedRevolutionary = newEnragedRevolutionary

func newEnragedRevolutionary() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Enraged Revolutionary",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Warrior},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.DethroneTriggeredBody,
			},
			OracleText: `
			Dethrone (Whenever this creature attacks the player with the most life or tied for most life, put a +1/+1 counter on it.)
		`,
		},
	}
}
