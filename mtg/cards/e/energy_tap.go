package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// EnergyTap is the card definition for Energy Tap.
//
// Type: Sorcery
// Cost: {U}
//
// Oracle text:
//
//	Tap target untapped creature you control. If you do, add an amount of {C} equal to that creature's mana value.
var EnergyTap = newEnergyTap

func newEnergyTap() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Energy Tap",
			ManaCost: opt.Val(cost.Mana{
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target untapped creature you control",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou, Tapped: game.TriFalse}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Tap{
							Object: game.TargetPermanentReference(0),
						},
						PublishResult: game.ResultKey("energy-tap-target"),
					},
					{
						Primitive: game.AddMana{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind:   game.DynamicAmountObjectManaValue,
								Object: game.TargetPermanentReference(0),
							}),
							ManaColor: mana.C,
						},
						ResultGate: opt.Val(game.InstructionResultGate{
							Key:       "energy-tap-target",
							Succeeded: game.TriTrue,
						}),
					},
				},
			}.Ability()),
			OracleText: `
			Tap target untapped creature you control. If you do, add an amount of {C} equal to that creature's mana value.
		`,
		},
	}
}
