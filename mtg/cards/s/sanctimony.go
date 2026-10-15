package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Sanctimony is the card definition for Sanctimony.
//
// Type: Enchantment
// Cost: {1}{W}
//
// Oracle text:
//
//	Whenever an opponent taps a Mountain for mana, you may gain 1 life.
var Sanctimony = newSanctimony

func newSanctimony() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Sanctimony",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Enchantment},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever an opponent taps a Mountain for mana, you may gain 1 life.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:                game.EventManaProduced,
							Controller:           game.TriggerControllerOpponent,
							SubjectSelection:     game.Selection{SubtypesAny: []types.Sub{types.Mountain}},
							StackObjectKind:      game.StackSpell,
							CounterKind:          counter.PlusOnePlusOne,
							RequireTappedForMana: true,
						},
					},
					Optional: true,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.GainLife{
											Amount: game.Fixed(1),
											Player: game.ControllerReference(),
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
			},
			OracleText: `
			Whenever an opponent taps a Mountain for mana, you may gain 1 life.
		`,
		},
	}
}
