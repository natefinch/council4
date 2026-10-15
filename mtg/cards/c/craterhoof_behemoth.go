package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CraterhoofBehemoth is the card definition for Craterhoof Behemoth.
//
// Type: Creature — Beast
// Cost: {5}{G}{G}{G}
//
// Oracle text:
//
//	Haste
//	When this creature enters, creatures you control gain trample and get +X/+X until end of turn, where X is the number of creatures you control.
var CraterhoofBehemoth = newCraterhoofBehemoth

func newCraterhoofBehemoth() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Craterhoof Behemoth",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.G,
				cost.G,
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Beast},
			Power:     opt.Val(game.PT{Value: 5}),
			Toughness: opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.HasteStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When this creature enters, creatures you control gain trample and get +X/+X until end of turn, where X is the number of creatures you control.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:           game.EventPermanentEnteredBattlefield,
							Source:          game.TriggerSourceSelf,
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
													Duration: game.DurationPermanent,
													Group:    game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
													Layer:    game.LayerPowerToughnessModify,
													PowerDeltaDynamic: opt.Val(game.DynamicAmount{
														Kind:        game.DynamicAmountCountSelector,
														Multiplier:  1,
														CounterKind: counter.PlusOnePlusOne,
														Group:       game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
													}),
													ToughnessDeltaDynamic: opt.Val(game.DynamicAmount{
														Kind:        game.DynamicAmountCountSelector,
														Multiplier:  1,
														CounterKind: counter.PlusOnePlusOne,
														Group:       game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
													}),
												},
												game.ContinuousEffect{
													Duration:    game.DurationPermanent,
													Group:       game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
													Layer:       game.LayerAbility,
													AddKeywords: []game.Keyword{game.Trample},
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
			OracleText: `
			Haste
			When this creature enters, creatures you control gain trample and get +X/+X until end of turn, where X is the number of creatures you control.
		`,
		},
	}
}
