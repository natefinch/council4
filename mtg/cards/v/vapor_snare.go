package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// VaporSnare is the card definition for Vapor Snare.
//
// Type: Enchantment — Aura
// Cost: {4}{U}
//
// Oracle text:
//
//	Enchant creature
//	You control enchanted creature.
//	At the beginning of your upkeep, sacrifice this Aura unless you return a land you control to its owner's hand.
var VaporSnare = newVaporSnare

func newVaporSnare() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Vapor Snare",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.U,
			}),
			Colors:   []color.Color{color.Blue},
			Types:    []types.Card{types.Enchantment},
			Subtypes: []types.Sub{types.Aura},
			StaticAbilities: []game.StaticAbility{
				game.EnchantStaticAbility(&game.TargetSpec{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "creature",
					Allow:      game.TargetAllowPermanent,
					Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
				}),
				game.StaticAbility{
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer:         game.LayerControl,
							NewController: opt.Val(game.Player1),
							Group:         game.AttachedObjectGroup(game.SourcePermanentReference()),
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event:      game.EventBeginningOfStep,
							Controller: game.TriggerControllerYou,
							Step:       game.StepUpkeep,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Pay{
									Payment: game.ResolutionPayment{
										Prompt: "Return a land you control to its owner's hand?",
										AdditionalCosts: []cost.Additional{
											{
												Kind:               cost.AdditionalReturnToHand,
												Text:               "return a land you control to its owner's hand",
												Amount:             1,
												MatchPermanentType: true,
												PermanentType:      types.Land,
											},
										},
									},
								},
								PublishResult: game.ResultKey("sacrifice-unless-paid"),
							},
							{
								Primitive: game.Sacrifice{
									Object: game.SourcePermanentReference(),
								},
								ResultGate: opt.Val(game.InstructionResultGate{
									Key:       "sacrifice-unless-paid",
									Succeeded: game.TriFalse,
								}),
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Enchant creature
			You control enchanted creature.
			At the beginning of your upkeep, sacrifice this Aura unless you return a land you control to its owner's hand.
		`,
		},
	}
}
