package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// WaterspoutDjinn is the card definition for Waterspout Djinn.
//
// Type: Creature — Djinn
// Cost: {2}{U}{U}
//
// Oracle text:
//
//	Flying
//	At the beginning of your upkeep, sacrifice this creature unless you return an untapped Island you control to its owner's hand.
var WaterspoutDjinn = newWaterspoutDjinn()

func newWaterspoutDjinn() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Waterspout Djinn",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Djinn},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
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
										Prompt: "Return an untapped Island you control to its owner's hand?",
										AdditionalCosts: []cost.Additional{
											{
												Kind:            cost.AdditionalReturnToHand,
												Text:            "return an untapped Island you control to its owner's hand",
												Amount:          1,
												RequireUntapped: true,
												SubtypesAny:     cost.SubtypeSet{types.Island},
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
			Flying
			At the beginning of your upkeep, sacrifice this creature unless you return an untapped Island you control to its owner's hand.
		`,
		},
	}
}
