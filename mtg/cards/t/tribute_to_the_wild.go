package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TributeToTheWild is the card definition for Tribute to the Wild.
//
// Type: Instant
// Cost: {1}{G}
//
// Oracle text:
//
//	Each opponent sacrifices an artifact or enchantment of their choice.
var TributeToTheWild = newTributeToTheWild

func newTributeToTheWild() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Tribute to the Wild",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.SacrificePermanents{
							PlayerGroup: game.OpponentsReference(),
							Amount:      game.Fixed(1),
							Selection:   game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Enchantment}},
						},
					},
				},
			}.Ability()),
			OracleText: `
			Each opponent sacrifices an artifact or enchantment of their choice.
		`,
		},
	}
}
