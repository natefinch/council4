package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// CatharticReunion is the card definition for Cathartic Reunion.
//
// Type: Sorcery
// Cost: {1}{R}
//
// Oracle text:
//
//	As an additional cost to cast this spell, discard two cards.
//	Draw three cards.
var CatharticReunion = newCatharticReunion

func newCatharticReunion() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Cathartic Reunion",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Sorcery},
			AdditionalCosts: []cost.Additional{
				{
					Kind:   cost.AdditionalDiscard,
					Text:   "discard two cards",
					Amount: 2,
					Source: zone.Hand,
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Draw{
							Amount: game.Fixed(3),
							Player: game.ControllerReference(),
						},
					},
				},
			}.Ability()),
			OracleText: `
			As an additional cost to cast this spell, discard two cards.
			Draw three cards.
		`,
		},
	}
}
