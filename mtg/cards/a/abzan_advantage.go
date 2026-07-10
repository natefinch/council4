package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AbzanAdvantage is the card definition for Abzan Advantage.
//
// Type: Instant
// Cost: {1}{W}
//
// Oracle text:
//
//	Target player sacrifices an enchantment of their choice. Bolster 1. (Choose a creature with the least toughness among creatures you control and put a +1/+1 counter on it.)
var AbzanAdvantage = newAbzanAdvantage

func newAbzanAdvantage() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Abzan Advantage",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "Target player",
						Allow:      game.TargetAllowPlayer,
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.SacrificePermanents{
							Amount:    game.Fixed(1),
							Player:    game.TargetPlayerReference(0),
							Selection: game.Selection{RequiredTypes: []types.Card{types.Enchantment}},
						},
					},
					{
						Primitive: game.Bolster{
							Amount: game.Fixed(1),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Target player sacrifices an enchantment of their choice. Bolster 1. (Choose a creature with the least toughness among creatures you control and put a +1/+1 counter on it.)
		`,
		},
	}
}
