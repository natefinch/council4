package u

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// UltraMagnusTactician is the card definition for Ultra Magnus, Tactician // Ultra Magnus, Armored Carrier.
//
// Type: Legendary Artifact Creature — Robot // Legendary Artifact — Vehicle
// Face: Ultra Magnus, Armored Carrier — Legendary Artifact — Vehicle
//
// Oracle text:
//
//	More Than Meets the Eye {2}{R}{G}{W} (You may cast this card converted for {2}{R}{G}{W}.)
//	Ward {2}
//	Whenever Ultra Magnus attacks, you may put an artifact creature card from your hand onto the battlefield tapped and attacking. If you do, convert Ultra Magnus at end of combat.
var UltraMagnusTactician = newUltraMagnusTactician

func newUltraMagnusTactician() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Red, color.Green),
		CardFace: game.CardFace{
			Name: "Ultra Magnus, Tactician",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.R,
				cost.G,
				cost.W,
			}),
			Colors:     []color.Color{color.Green, color.Red, color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Artifact, types.Creature},
			Subtypes:   []types.Sub{types.Robot},
			Power:      opt.Val(game.PT{Value: 7}),
			Toughness:  opt.Val(game.PT{Value: 7}),
			StaticAbilities: []game.StaticAbility{
				game.WardStaticAbility(cost.Mana{cost.O(2)}),
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:  game.EventAttackerDeclared,
							Source: game.TriggerSourceSelf,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.ChooseFromZone{
									Player:     game.ControllerReference(),
									SourceZone: zone.Hand,
									Filter:     game.Selection{RequiredTypes: []types.Card{types.Artifact, types.Creature}},
									Quantity:   game.Fixed(1),
									Destination: game.ChooseDestination{
										Zone: zone.Battlefield,
									},
									Riders: game.ChooseRiders{
										EntersTapped:    true,
										EntersAttacking: true,
									},
									Prompt: "Choose a card to put onto the battlefield",
								},
								Optional:      true,
								PublishResult: game.ResultKey("if-you-do"),
							},
							{
								Primitive: game.CreateDelayedTrigger{
									Trigger: game.DelayedTriggerDef{
										Timing: game.DelayedAtEndOfCombat,
										Content: game.Mode{
											Sequence: []game.Instruction{
												{
													Primitive: game.Transform{
														Object: game.SourceCardPermanentReference(),
													},
												},
											},
										}.Ability(),
									},
								},
								ResultGate: opt.Val(game.InstructionResultGate{
									Key:       "if-you-do",
									Succeeded: game.TriTrue,
								}),
							},
						},
					}.Ability(),
				},
			},
			AlternativeCosts: []cost.Alternative{
				cost.Alternative{
					Label:    "More Than Meets the Eye",
					ManaCost: opt.Val(cost.Mana{cost.O(2), cost.R, cost.G, cost.W}),
					Mechanic: cost.AlternativeMechanicMoreThanMeetsTheEye,
				},
			},
			OracleText: `
			More Than Meets the Eye {2}{R}{G}{W} (You may cast this card converted for {2}{R}{G}{W}.)
			Ward {2}
			Whenever Ultra Magnus attacks, you may put an artifact creature card from your hand onto the battlefield tapped and attacking. If you do, convert Ultra Magnus at end of combat.
		`,
		},
		Layout: game.LayoutTransform,
		Back: opt.Val(game.CardFace{
			Name:       "Ultra Magnus, Armored Carrier",
			Colors:     []color.Color{color.Green, color.Red, color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Artifact},
			Subtypes:   []types.Sub{types.Vehicle},
			Power:      opt.Val(game.PT{Value: 4}),
			Toughness:  opt.Val(game.PT{Value: 7}),
			StaticAbilities: []game.StaticAbility{
				game.LivingMetalStaticBody,
				game.HasteStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:  game.EventAttackerDeclared,
							Source: game.TriggerSourceSelf,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.ApplyContinuous{
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Layer: game.LayerAbility,
											Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou, CombatState: game.CombatStateAttacking}),
											AddKeywords: []game.Keyword{
												game.Indestructible,
											},
										},
									},
									Duration: game.DurationUntilEndOfTurn,
								},
							},
							{
								Primitive: game.Transform{
									Object: game.SourcePermanentReference(),
								},
								Condition: opt.Val(game.EffectCondition{
									Condition: opt.Val(game.Condition{
										ControlsMatching: opt.Val(game.SelectionCount{
											Selection:  game.Selection{RequiredTypes: []types.Card{types.Creature}, CombatState: game.CombatStateAttacking},
											TotalPower: opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 8}),
										}),
									}),
								}),
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Living metal (During your turn, this Vehicle is also a creature.)
			Haste
			Formidable — Whenever Ultra Magnus attacks, attacking creatures you control gain indestructible until end of turn. If those creatures have total power 8 or greater, convert Ultra Magnus.
		`,
		}),
	}
}
