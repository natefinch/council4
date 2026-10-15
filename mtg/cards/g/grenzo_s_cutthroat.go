package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GrenzoSCutthroat is the card definition for Grenzo's Cutthroat.
//
// Type: Creature — Goblin Rogue
// Cost: {1}{R}
//
// Oracle text:
//
//	First strike
//	Dethrone (Whenever this creature attacks the player with the most life or tied for most life, put a +1/+1 counter on it.)
var GrenzoSCutthroat = newGrenzoSCutthroat

func newGrenzoSCutthroat() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Grenzo's Cutthroat",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Goblin, types.Rogue},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.FirstStrikeStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.DethroneTriggeredBody,
			},
			OracleText: `
			First strike
			Dethrone (Whenever this creature attacks the player with the most life or tied for most life, put a +1/+1 counter on it.)
		`,
		},
	}
}
