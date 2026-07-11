package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Ensnare is the card definition for Ensnare.
//
// Type: Instant
// Cost: {3}{U}
//
// Oracle text:
//
//	You may return two Islands you control to their owner's hand rather than pay this spell's mana cost.
//	Tap all creatures.
var Ensnare = newEnsnare

func newEnsnare() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Ensnare",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
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
						Primitive: game.Tap{
							Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
						},
					},
				},
			}.Ability()),
			OracleText: `
			You may return two Islands you control to their owner's hand rather than pay this spell's mana cost.
			Tap all creatures.
		`,
		},
	}
}
