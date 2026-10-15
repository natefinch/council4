package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// VisionSkeins is the card definition for Vision Skeins.
//
// Type: Instant
// Cost: {1}{U}
//
// Oracle text:
//
//	Each player draws two cards.
var VisionSkeins = newVisionSkeins

func newVisionSkeins() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Vision Skeins",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Draw{
							Amount:      game.Fixed(2),
							PlayerGroup: game.AllPlayersReference(),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Each player draws two cards.
		`,
		},
	}
}
