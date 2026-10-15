package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DecreeOfPain is the card definition for Decree of Pain.
//
// Type: Sorcery
// Cost: {6}{B}{B}
//
// Oracle text:
//
//	Destroy all creatures. They can't be regenerated. Draw a card for each creature destroyed this way.
//	Cycling {3}{B}{B} ({3}{B}{B}, Discard this card: Draw a card.)
//	When you cycle this card, all creatures get -2/-2 until end of turn.
var DecreeOfPain = newDecreeOfPain

func newDecreeOfPain() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Decree of Pain",
			ManaCost: opt.Val(cost.Mana{
				cost.O(6),
				cost.B,
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Sorcery},
			ActivatedAbilities: []game.ActivatedAbility{
				game.CyclingActivatedAbility(cost.Mana{cost.O(3), cost.B, cost.B}),
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When you cycle this card, all creatures get -2/-2 until end of turn.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:           game.EventCycled,
							Source:          game.TriggerSourceSelf,
							Player:          game.TriggerPlayerYou,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.ApplyContinuous{
											ContinuousEffects: []game.ContinuousEffect{
												game.ContinuousEffect{
													Duration:       game.DurationPermanent,
													Group:          game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
													Layer:          game.LayerPowerToughnessModify,
													PowerDelta:     -2,
													ToughnessDelta: -2,
												},
											},
											Duration: game.DurationUntilEndOfTurn,
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
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Destroy{
							Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
						},
						PublishResult: game.ResultKey("destroyed-this-way"),
					},
					{
						Primitive: game.Draw{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind:       game.DynamicAmountPreviousEffectResult,
								Multiplier: 1,
								ResultKey:  game.ResultKey("destroyed-this-way"),
							}),
							Player: game.ControllerReference(),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Destroy all creatures. They can't be regenerated. Draw a card for each creature destroyed this way.
			Cycling {3}{B}{B} ({3}{B}{B}, Discard this card: Draw a card.)
			When you cycle this card, all creatures get -2/-2 until end of turn.
		`,
		},
	}
}
