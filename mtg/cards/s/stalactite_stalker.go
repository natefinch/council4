package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// StalactiteStalker is the card definition for Stalactite Stalker.
//
// Type: Creature — Goblin Rogue
// Cost: {B}
//
// Oracle text:
//
//	Menace
//	At the beginning of your end step, if you descended this turn, put a +1/+1 counter on this creature. (You descended if a permanent card was put into your graveyard from anywhere.)
//	{2}{B}, Sacrifice this creature: Target creature gets -X/-X until end of turn, where X is this creature's power.
var StalactiteStalker = newStalactiteStalker()

func newStalactiteStalker() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Stalactite Stalker",
			ManaCost: opt.Val(cost.Mana{
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Goblin, types.Rogue},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.MenaceStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:     "{2}{B}, Sacrifice this creature: Target creature gets -X/-X until end of turn, where X is this creature's power.",
					ManaCost: opt.Val(cost.Mana{cost.O(2), cost.B}),
					AdditionalCosts: []cost.Additional{
						{
							Kind:   cost.AdditionalSacrificeSource,
							Text:   "Sacrifice this creature",
							Amount: 1,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
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
								Primitive: game.ModifyPT{
									Object: game.TargetPermanentReference(0),
									PowerDelta: game.Dynamic(game.DynamicAmount{
										Kind:       game.DynamicAmountObjectPower,
										Multiplier: -1,
										Object:     game.SourcePermanentReference(),
									}),
									ToughnessDelta: game.Dynamic(game.DynamicAmount{
										Kind:       game.DynamicAmountObjectPower,
										Multiplier: -1,
										Object:     game.SourcePermanentReference(),
									}),
									Duration: game.DurationUntilEndOfTurn,
								},
							},
						},
					}.Ability(),
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event:      game.EventBeginningOfStep,
							Controller: game.TriggerControllerYou,
							Step:       game.StepEnd,
						},
						InterveningIf: "if you descended this turn",
						InterveningCondition: opt.Val(game.Condition{
							EventHistory: opt.Val(game.EventHistoryCondition{Pattern: game.TriggerPattern{
								Event:            game.EventZoneChanged,
								Player:           game.TriggerPlayerYou,
								MatchToZone:      true,
								ToZone:           zone.Graveyard,
								SubjectSelection: game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Battle, types.Creature, types.Enchantment, types.Land, types.Planeswalker}, NonToken: true},
							}, Window: game.EventHistoryCurrentTurn}),
						}),
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.AddCounter{
									Amount:      game.Fixed(1),
									Object:      game.SourcePermanentReference(),
									CounterKind: counter.PlusOnePlusOne,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Menace
			At the beginning of your end step, if you descended this turn, put a +1/+1 counter on this creature. (You descended if a permanent card was put into your graveyard from anywhere.)
			{2}{B}, Sacrifice this creature: Target creature gets -X/-X until end of turn, where X is this creature's power.
		`,
		},
	}
}
