package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RadicalIdea is the card definition for Radical Idea.
//
// Type: Instant
// Cost: {1}{U}
//
// Oracle text:
//
//	Draw a card.
//	Jump-start (You may cast this card from your graveyard by discarding a card in addition to paying its other costs. Then exile this card.)
var RadicalIdea = newRadicalIdea

func newRadicalIdea() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Radical Idea",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
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
							Amount: game.Fixed(1),
							Player: game.ControllerReference(),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Draw a card.
			Jump-start (You may cast this card from your graveyard by discarding a card in addition to paying its other costs. Then exile this card.)
		`,
		},
	}
}
