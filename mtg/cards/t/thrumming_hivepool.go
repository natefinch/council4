package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ThrummingHivepool is the card definition for Thrumming Hivepool.
//
// Type: Artifact
// Cost: {6}
//
// Oracle text:
//
//	Affinity for Slivers (This spell costs {1} less to cast for each Sliver you control.)
//	Slivers you control have double strike and haste.
//	At the beginning of your upkeep, create two 1/1 colorless Sliver creature tokens.
var ThrummingHivepool = newThrummingHivepool

func newThrummingHivepool() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Thrumming Hivepool",
			ManaCost: opt.Val(cost.Mana{
				cost.O(6),
			}),
			Types: []types.Card{types.Artifact},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "This spell costs {1} less to cast for each Sliver you control.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCostModifier,
							AffectedSource: true,
							Duration:       game.DurationPermanent,
							CostModifier: game.CostModifier{
								Kind:               game.CostModifierSpell,
								PerObjectReduction: 1,
								CountSelection:     &game.Selection{SubtypesAny: []types.Sub{types.Sliver}, Controller: game.ControllerYou},
							},
						},
					},
				},
				game.StaticAbility{
					Text: "Slivers you control have double strike and haste.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:    game.DurationPermanent,
							Group:       game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{SubtypesAny: []types.Sub{types.Sub("Sliver")}}),
							Layer:       game.LayerAbility,
							AddKeywords: []game.Keyword{game.DoubleStrike, game.Haste},
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "At the beginning of your upkeep, create two 1/1 colorless Sliver creature tokens.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event:           game.EventBeginningOfStep,
							Controller:      game.TriggerControllerYou,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
							Step:            game.StepUpkeep,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateToken{
											Amount: game.Fixed(2),
											Source: game.TokenDef(thrummingHivepoolToken),
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
			Affinity for Slivers (This spell costs {1} less to cast for each Sliver you control.)
			Slivers you control have double strike and haste.
			At the beginning of your upkeep, create two 1/1 colorless Sliver creature tokens.
		`,
		},
	}
}

var thrummingHivepoolToken = newThrummingHivepoolToken()

func newThrummingHivepoolToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Sliver",
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Sliver},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
