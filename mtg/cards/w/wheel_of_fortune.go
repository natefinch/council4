package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// WheelOfFortune is the card definition for Wheel of Fortune.
//
// Type: Sorcery
// Cost: {2}{R}
//
// Oracle text:
//
//	Each player discards their hand, then draws seven cards.
var WheelOfFortune = newWheelOfFortune

func newWheelOfFortune() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Wheel of Fortune",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Discard{
							PlayerGroup: game.AllPlayersReference(),
							EntireHand:  true,
						},
					},
					{
						Primitive: game.Draw{
							Amount:      game.Fixed(7),
							PlayerGroup: game.AllPlayersReference(),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Each player discards their hand, then draws seven cards.
		`,
		},
	}
}
