package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Gush is the card definition for Gush.
//
// Type: Instant
// Cost: {4}{U}
//
// Oracle text:
//
//	You may return two Islands you control to their owner's hand rather than pay this spell's mana cost.
//	Draw two cards.
var Gush = newGush

func newGush() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Gush",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Instant},
			AlternativeCosts: []cost.Alternative{
				cost.Alternative{
					Label: "Return two Islands you control to their owner's hand",
					AdditionalCosts: []cost.Additional{
						{
							Kind:        cost.AdditionalReturnToHand,
							Text:        "return two Islands you control to their owner's hand",
							Amount:      2,
							SubtypesAny: cost.SubtypeSet{types.Island},
						},
					},
				},
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
			You may return two Islands you control to their owner's hand rather than pay this spell's mana cost.
			Draw two cards.
		`,
		},
	}
}
