package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// HieroglyphicIllumination is the card definition for Hieroglyphic Illumination.
//
// Type: Instant
// Cost: {3}{U}
//
// Oracle text:
//
//	Draw two cards.
//	Cycling {U} ({U}, Discard this card: Draw a card.)
var HieroglyphicIllumination = newHieroglyphicIllumination

func newHieroglyphicIllumination() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Hieroglyphic Illumination",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Instant},
			ActivatedAbilities: []game.ActivatedAbility{
				game.CyclingActivatedAbility(cost.Mana{cost.U}),
			},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Draw{
							Amount: game.Fixed(2),
							Player: game.ControllerReference(),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Draw two cards.
			Cycling {U} ({U}, Discard this card: Draw a card.)
		`,
		},
	}
}
