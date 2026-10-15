package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// PerilousResearch is the card definition for Perilous Research.
//
// Type: Instant
// Cost: {1}{U}
//
// Oracle text:
//
//	Draw two cards, then sacrifice a permanent.
var PerilousResearch = newPerilousResearch

func newPerilousResearch() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Perilous Research",
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
							Amount: game.Fixed(2),
							Player: game.ControllerReference(),
						},
					},
					{
						Primitive: game.SacrificePermanents{
							Player: game.ControllerReference(),
							Amount: game.Fixed(1),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Draw two cards, then sacrifice a permanent.
		`,
		},
	}
}
