package n

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// NaduWingedWisdom is the card definition for Nadu, Winged Wisdom.
//
// Type: Legendary Creature — Bird Wizard
// Cost: {1}{G}{U}
//
// Oracle text:
//
//	Flying
//	Creatures you control have "Whenever this creature becomes the target of a spell or ability, reveal the top card of your library. If it's a land card, put it onto the battlefield. Otherwise, put it into your hand. This ability triggers only twice each turn."
var NaduWingedWisdom = newNaduWingedWisdom

func newNaduWingedWisdom() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Green),
		CardFace: game.CardFace{
			Name: "Nadu, Winged Wisdom",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
				cost.U,
			}),
			Colors:     []color.Color{color.Green, color.Blue},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Bird, types.Wizard},
			Power:      opt.Val(game.PT{Value: 3}),
			Toughness:  opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.StaticAbility{
					Text: "Creatures you control have \"Whenever this creature becomes the target of a spell or ability, reveal the top card of your library. If it's a land card, put it onto the battlefield. Otherwise, put it into your hand. This ability triggers only twice each turn.\"",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration: game.DurationPermanent,
							Group:    game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}}),
							Layer:    game.LayerAbility,
							AddAbilities: []game.Ability{
								&game.TriggeredAbility{
									Text: "Whenever this creature becomes the target of a spell or ability, reveal the top card of your library. If it's a land card, put it onto the battlefield. Otherwise, put it into your hand. This ability triggers only twice each turn.",
									Trigger: game.TriggerCondition{
										Type: game.TriggerWhenever,
										Pattern: game.TriggerPattern{
											Event:           game.EventObjectBecameTarget,
											Source:          game.TriggerSourceSelf,
											StackObjectKind: game.StackSpell,
											CounterKind:     counter.PlusOnePlusOne,
										},
									},
									MaxTriggersPerTurn: 2,
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
																Text:              "If it's a land card",
																Object:            opt.Val(game.LinkedObjectReference("sequence-effect-0-product")),
																ObjectMatches:     opt.Val(game.Selection{RequiredTypes: []types.Card{types.Land}}),
																SourceCounterKind: counter.PlusOnePlusOne,
															}),
														}),
														PublishCondition: game.ConditionKey("condition-0"),
													},
													game.Instruction{
														Primitive: game.MoveCard{
															Card: game.CardReference{
																Kind:   game.CardReferenceLinked,
																LinkID: "sequence-effect-0-product",
															},
															FromZone:    zone.Library,
															Destination: zone.Hand,
														},
														ConditionGate:       game.ConditionKey("condition-0"),
														ConditionGateNegate: true,
													},
												},
											},
										},
										MinModes: 1,
										MaxModes: 1,
									},
								},
							},
						},
					},
				},
			},
			OracleText: `
			Flying
			Creatures you control have "Whenever this creature becomes the target of a spell or ability, reveal the top card of your library. If it's a land card, put it onto the battlefield. Otherwise, put it into your hand. This ability triggers only twice each turn."
		`,
		},
	}
}
