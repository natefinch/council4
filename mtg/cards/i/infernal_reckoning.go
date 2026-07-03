package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// InfernalReckoning is the card definition for Infernal Reckoning.
//
// Type: Instant
// Cost: {B}
//
// Oracle text:
//
//	Exile target colorless creature. You gain life equal to its power.
var InfernalReckoning = newInfernalReckoning()

func newInfernalReckoning() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Infernal Reckoning",
			ManaCost: opt.Val(cost.Mana{
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target colorless creature",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Colorless: true}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Exile{
							Object:         game.TargetPermanentReference(0),
							ExileLinkedKey: game.LinkedKey("life-rider-1"),
						},
					},
					{
						Primitive: game.GainLife{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind:       game.DynamicAmountObjectPower,
								Multiplier: 1,
								Object:     game.LinkedObjectReference("life-rider-1"),
							}),
							Player: game.ControllerReference(),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Exile target colorless creature. You gain life equal to its power.
		`,
		},
	}
}
