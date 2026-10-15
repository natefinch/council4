package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Dredge is the card definition for Dredge.
//
// Type: Instant
// Cost: {B}
//
// Oracle text:
//
//	Sacrifice a creature or land.
//	Draw a card.
var Dredge = newDredge

func newDredge() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Dredge",
			ManaCost: opt.Val(cost.Mana{
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.SacrificePermanents{
							Player:    game.ControllerReference(),
							Amount:    game.Fixed(1),
							Selection: game.Selection{RequiredTypesAny: []types.Card{types.Creature, types.Land}},
						},
					},
					{
						Primitive: game.Draw{
							Amount: game.Fixed(1),
							Player: game.ControllerReference(),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Sacrifice a creature or land.
			Draw a card.
		`,
		},
	}
}
