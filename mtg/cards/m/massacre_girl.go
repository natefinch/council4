package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MassacreGirl is the card definition for Massacre Girl.
//
// Type: Legendary Creature — Human Assassin
// Cost: {3}{B}{B}
//
// Oracle text:
//
//	Menace
//	When Massacre Girl enters, each other creature gets -1/-1 until end of turn. Whenever a creature dies this turn, each creature other than Massacre Girl gets -1/-1 until end of turn.
var MassacreGirl = newMassacreGirl

func newMassacreGirl() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Massacre Girl",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.B,
				cost.B,
			}),
			Colors:     []color.Color{color.Black},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Assassin},
			Power:      opt.Val(game.PT{Value: 4}),
			Toughness:  opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.MenaceStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:  game.EventPermanentEnteredBattlefield,
							Source: game.TriggerSourceSelf,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.ApplyContinuous{
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Layer:          game.LayerPowerToughnessModify,
											Group:          game.BattlefieldGroupExcluding(game.Selection{RequiredTypes: []types.Card{types.Creature}}, game.SourcePermanentReference()),
											PowerDelta:     -1,
											ToughnessDelta: -1,
										},
									},
									Duration: game.DurationUntilEndOfTurn,
								},
							},
							{
								Primitive: game.CreateDelayedTrigger{
									Trigger: game.DelayedTriggerDef{
										EventPattern: opt.Val(game.TriggerPattern{
											Event:            game.EventPermanentDied,
											SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
										}),
										Window: game.DelayedWindowThisTurn,
										Content: game.Mode{
											Sequence: []game.Instruction{
												{
													Primitive: game.ApplyContinuous{
														ContinuousEffects: []game.ContinuousEffect{
															game.ContinuousEffect{
																Layer:          game.LayerPowerToughnessModify,
																Group:          game.BattlefieldGroupExcluding(game.Selection{RequiredTypes: []types.Card{types.Creature}}, game.SourcePermanentReference()),
																PowerDelta:     -1,
																ToughnessDelta: -1,
															},
														},
														Duration: game.DurationUntilEndOfTurn,
													},
												},
											},
										}.Ability(),
									},
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Menace
			When Massacre Girl enters, each other creature gets -1/-1 until end of turn. Whenever a creature dies this turn, each creature other than Massacre Girl gets -1/-1 until end of turn.
		`,
		},
	}
}
