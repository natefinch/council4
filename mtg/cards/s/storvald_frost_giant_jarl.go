package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// StorvaldFrostGiantJarl is the card definition for Storvald, Frost Giant Jarl.
//
// Type: Legendary Creature — Giant
// Cost: {4}{G}{W}{U}
//
// Oracle text:
//
//	Ward {3}
//	Other creatures you control have ward {3}.
//	Whenever Storvald enters or attacks, choose one or both —
//	• Target creature has base power and toughness 7/7 until end of turn.
//	• Target creature has base power and toughness 1/1 until end of turn.
var StorvaldFrostGiantJarl = newStorvaldFrostGiantJarl

func newStorvaldFrostGiantJarl() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue, color.Green),
		CardFace: game.CardFace{
			Name: "Storvald, Frost Giant Jarl",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.G,
				cost.W,
				cost.U,
			}),
			Colors:     []color.Color{color.Green, color.Blue, color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Giant},
			Power:      opt.Val(game.PT{Value: 7}),
			Toughness:  opt.Val(game.PT{Value: 7}),
			StaticAbilities: []game.StaticAbility{
				game.WardStaticAbility(cost.Mana{cost.O(3)}),
				game.StaticAbility{
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer: game.LayerAbility,
							Group: game.ObjectControlledGroupExcluding(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}}, game.SourcePermanentReference()),
							AddAbilities: []game.Ability{
								new(game.WardStaticAbility(cost.Mana{cost.O(3)})),
							},
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:      game.EventPermanentEnteredBattlefield,
							Source:     game.TriggerSourceSelf,
							UnionEvent: game.EventAttackerDeclared,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Text: "Target creature has base power and toughness 7/7 until end of turn.",
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target creature",
										Allow:      game.TargetAllowPermanent,
										Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
									},
								},
								Sequence: []game.Instruction{
									{
										Primitive: game.ApplyContinuous{
											Object: opt.Val(game.TargetPermanentReference(0)),
											ContinuousEffects: []game.ContinuousEffect{
												game.ContinuousEffect{
													Layer:        game.LayerPowerToughnessSet,
													SetPower:     opt.Val(game.PT{Value: 7}),
													SetToughness: opt.Val(game.PT{Value: 7}),
												},
											},
											Duration: game.DurationUntilEndOfTurn,
										},
									},
								},
							},
							game.Mode{
								Text: "Target creature has base power and toughness 1/1 until end of turn.",
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target creature",
										Allow:      game.TargetAllowPermanent,
										Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
									},
								},
								Sequence: []game.Instruction{
									{
										Primitive: game.ApplyContinuous{
											Object: opt.Val(game.TargetPermanentReference(0)),
											ContinuousEffects: []game.ContinuousEffect{
												game.ContinuousEffect{
													Layer:        game.LayerPowerToughnessSet,
													SetPower:     opt.Val(game.PT{Value: 1}),
													SetToughness: opt.Val(game.PT{Value: 1}),
												},
											},
											Duration: game.DurationUntilEndOfTurn,
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 2,
					},
				},
			},
			OracleText: `
			Ward {3}
			Other creatures you control have ward {3}.
			Whenever Storvald enters or attacks, choose one or both —
			• Target creature has base power and toughness 7/7 until end of turn.
			• Target creature has base power and toughness 1/1 until end of turn.
		`,
		},
	}
}
