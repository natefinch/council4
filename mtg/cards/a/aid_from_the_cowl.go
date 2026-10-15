package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// AidFromTheCowl is the card definition for Aid from the Cowl.
//
// Type: Enchantment
// Cost: {3}{G}{G}
//
// Oracle text:
//
//	Revolt — At the beginning of your end step, if a permanent left the battlefield under your control this turn, reveal the top card of your library. If it's a permanent card, you may put it onto the battlefield. Otherwise, you may put it on the bottom of your library.
var AidFromTheCowl = newAidFromTheCowl

func newAidFromTheCowl() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Aid from the Cowl",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.G,
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Enchantment},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Revolt — At the beginning of your end step, if a permanent left the battlefield under your control this turn, reveal the top card of your library. If it's a permanent card, you may put it onto the battlefield. Otherwise, you may put it on the bottom of your library.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event:           game.EventBeginningOfStep,
							Controller:      game.TriggerControllerYou,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
							Step:            game.StepEnd,
						},
						InterveningIf: "if a permanent left the battlefield under your control this turn",
						InterveningCondition: opt.Val(game.Condition{
							Text:              "if a permanent left the battlefield under your control this turn",
							SourceCounterKind: counter.PlusOnePlusOne,
							EventHistory: opt.Val(game.EventHistoryCondition{
								Pattern: game.TriggerPattern{
									Event:           game.EventZoneChanged,
									Controller:      game.TriggerControllerYou,
									MatchFromZone:   true,
									FromZone:        zone.Battlefield,
									StackObjectKind: game.StackSpell,
									CounterKind:     counter.PlusOnePlusOne,
								},
								Window: game.EventHistoryCurrentTurn,
							}),
						}),
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Reveal{
											Amount:        game.Fixed(1),
											Player:        game.ControllerReference(),
											PublishLinked: game.LinkedKey("sequence-effect-0-product"),
										},
										LocalProducts: game.LocalProducts{
											Links: []game.LinkedKey{game.LinkedKey("sequence-effect-0-product")},
										},
									},
									game.Instruction{
										Primitive: game.PutOnBattlefield{
											Source:    game.LinkedBattlefieldSource(game.LinkedKey("sequence-effect-0-product")),
											Recipient: opt.Val(game.ControllerReference()),
										},
										Condition: opt.Val(game.EffectCondition{
											Condition: opt.Val(game.Condition{
												Text:              "If it's a permanent card",
												Object:            opt.Val(game.LinkedObjectReference("sequence-effect-0-product")),
												ObjectMatches:     opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Battle, types.Creature, types.Enchantment, types.Land, types.Planeswalker}}),
												SourceCounterKind: counter.PlusOnePlusOne,
											}),
										}),
										PublishCondition: game.ConditionKey("condition-1"),
										Optional:         true,
									},
									game.Instruction{
										Primitive: game.MoveCard{
											Card: game.CardReference{
												Kind:   game.CardReferenceLinked,
												LinkID: "sequence-effect-0-product",
											},
											FromZone:          zone.Library,
											Destination:       zone.Library,
											DestinationBottom: true,
										},
										ConditionGate:       game.ConditionKey("condition-1"),
										ConditionGateNegate: true,
										Optional:            true,
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
			Revolt — At the beginning of your end step, if a permanent left the battlefield under your control this turn, reveal the top card of your library. If it's a permanent card, you may put it onto the battlefield. Otherwise, you may put it on the bottom of your library.
		`,
		},
	}
}
