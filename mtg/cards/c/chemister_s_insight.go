package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ChemisterSInsight is the card definition for Chemister's Insight.
//
// Type: Instant
// Cost: {3}{U}
//
// Oracle text:
//
//	Draw two cards.
//	Jump-start (You may cast this card from your graveyard by discarding a card in addition to paying its other costs. Then exile this card.)
var ChemisterSInsight = newChemisterSInsight

func newChemisterSInsight() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Chemister's Insight",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Instant},
			StaticAbilities: []game.StaticAbility{
				game.JumpStartStaticBody,
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
			Jump-start (You may cast this card from your graveyard by discarding a card in addition to paying its other costs. Then exile this card.)
		`,
		},
	}
}
