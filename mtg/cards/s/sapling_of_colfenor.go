package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// SaplingOfColfenor is the card definition for Sapling of Colfenor.
//
// Type: Legendary Creature — Treefolk Shaman
// Cost: {3}{B/G}{B/G}
//
// Oracle text:
//
//	Indestructible
//	Whenever Sapling of Colfenor attacks, reveal the top card of your library. If it's a creature card, you gain life equal to that card's toughness, lose life equal to its power, then put it into your hand.
var SaplingOfColfenor = newSaplingOfColfenor

func newSaplingOfColfenor() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Green),
		CardFace: game.CardFace{
			Name: "Sapling of Colfenor",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.HybridMana(mana.B, mana.G),
				cost.HybridMana(mana.B, mana.G),
			}),
			Colors:     []color.Color{color.Black, color.Green},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Treefolk, types.Shaman},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.IndestructibleStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever Sapling of Colfenor attacks, reveal the top card of your library. If it's a creature card, you gain life equal to that card's toughness, lose life equal to its power, then put it into your hand.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:           game.EventAttackerDeclared,
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
										Primitive: game.Reveal{
											Amount:        game.Fixed(1),
											Player:        game.ControllerReference(),
											PublishLinked: game.LinkedKey("sequence-effect-0-product"),
											PublishCharacteristics: game.LibraryCardCharacteristics{
												Power:     game.ResultKey("sequence-effect-0-power"),
												Toughness: game.ResultKey("sequence-effect-0-toughness"),
											},
										},
										LocalProducts: game.LocalProducts{
											Results: []game.ResultKey{game.ResultKey("sequence-effect-0-toughness"), game.ResultKey("sequence-effect-0-power")},
											Links:   []game.LinkedKey{game.LinkedKey("sequence-effect-0-product")},
										},
									},
									game.Instruction{
										Primitive: game.GainLife{
											Amount: game.Dynamic(game.DynamicAmount{
												Kind:       game.DynamicAmountPreviousEffectResult,
												Multiplier: 1,
												ResultKey:  game.ResultKey("sequence-effect-0-toughness"),
											}),
											Player: game.ControllerReference(),
										},
										Condition: opt.Val(game.EffectCondition{
											Condition: opt.Val(game.Condition{
												Text:              "If it's a creature card",
												Object:            opt.Val(game.LinkedObjectReference("sequence-effect-0-product")),
												ObjectMatches:     opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
												SourceCounterKind: counter.PlusOnePlusOne,
											}),
										}),
										PublishCondition: game.ConditionKey("condition-0"),
										ResultGate: opt.Val(game.InstructionResultGate{
											Key:             game.ResultKey("sequence-effect-0-toughness"),
											AmountAvailable: true,
										}),
									},
									game.Instruction{
										Primitive: game.LoseLife{
											Amount: game.Dynamic(game.DynamicAmount{
												Kind:       game.DynamicAmountPreviousEffectResult,
												Multiplier: 1,
												ResultKey:  game.ResultKey("sequence-effect-0-power"),
											}),
											Player: game.ControllerReference(),
										},
										ConditionGate: game.ConditionKey("condition-0"),
										ResultGate: opt.Val(game.InstructionResultGate{
											Key:             game.ResultKey("sequence-effect-0-power"),
											AmountAvailable: true,
										}),
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
										ConditionGate: game.ConditionKey("condition-0"),
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
			Indestructible
			Whenever Sapling of Colfenor attacks, reveal the top card of your library. If it's a creature card, you gain life equal to that card's toughness, lose life equal to its power, then put it into your hand.
		`,
		},
	}
}
