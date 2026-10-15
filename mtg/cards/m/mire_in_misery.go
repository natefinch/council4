package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MireInMisery is the card definition for Mire in Misery.
//
// Type: Sorcery
// Cost: {1}{B}
//
// Oracle text:
//
//	Each opponent sacrifices a creature or enchantment of their choice.
var MireInMisery = newMireInMisery

func newMireInMisery() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Mire in Misery",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.SacrificePermanents{
							PlayerGroup: game.OpponentsReference(),
							Amount:      game.Fixed(1),
							Selection:   game.Selection{RequiredTypesAny: []types.Card{types.Creature, types.Enchantment}},
						},
					},
				},
			}.Ability()),
			OracleText: `
			Each opponent sacrifices a creature or enchantment of their choice.
		`,
		},
	}
}
