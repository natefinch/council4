package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// HornedCheetah is the card definition for Horned Cheetah.
//
// Type: Creature — Cat
// Cost: {2}{G}{W}
//
// Oracle text:
//
//	Whenever this creature deals damage, you gain that much life.
var HornedCheetah = newHornedCheetah

func newHornedCheetah() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Green),
		CardFace: game.CardFace{
			Name: "Horned Cheetah",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
				cost.W,
			}),
			Colors:    []color.Color{color.Green, color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Cat},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:   game.EventDamageDealt,
							Source:  game.TriggerSourceSelf,
							Subject: game.TriggerSubjectDamageSource,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.GainLife{
									Amount: game.Dynamic(game.DynamicAmount{
										Kind:       game.DynamicAmountEventDamage,
										Multiplier: 1,
									}),
									Player: game.ControllerReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Whenever this creature deals damage, you gain that much life.
		`,
		},
	}
}
