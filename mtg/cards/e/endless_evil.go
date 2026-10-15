package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// EndlessEvil is the card definition for Endless Evil.
//
// Type: Enchantment — Aura
// Cost: {2}{U}
//
// Oracle text:
//
//	Enchant creature you control
//	At the beginning of your upkeep, create a token that's a copy of enchanted creature, except the token is 1/1.
//	When enchanted creature dies, if that creature was a Horror, return this card to its owner's hand.
var EndlessEvil = newEndlessEvil

func newEndlessEvil() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Endless Evil",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
			}),
			Colors:   []color.Color{color.Blue},
			Types:    []types.Card{types.Enchantment},
			Subtypes: []types.Sub{types.Aura},
			StaticAbilities: []game.StaticAbility{
				game.EnchantStaticAbility(&game.TargetSpec{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "creature you control",
					Allow:      game.TargetAllowPermanent,
					Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Controller: game.ControllerYou}),
				}),
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "At the beginning of your upkeep, create a token that's a copy of enchanted creature, except the token is 1/1.",
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
											Amount: game.Fixed(1),
											Source: game.TokenCopyOf(game.TokenCopySpec{
												Source:       game.TokenCopySourceObject,
												Object:       game.SourceAttachedPermanentReference(),
												SetPower:     opt.Val(game.PT{Value: 1}),
												SetToughness: opt.Val(game.PT{Value: 1}),
											}),
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
				game.TriggeredAbility{
					Text: "When enchanted creature dies, if that creature was a Horror, return this card to its owner's hand.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:            game.EventPermanentDied,
							Source:           game.TriggerSourceAttachedPermanent,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
							StackObjectKind:  game.StackSpell,
							CounterKind:      counter.PlusOnePlusOne,
						},
						InterveningIf: "if that creature was a Horror",
						InterveningCondition: opt.Val(game.Condition{
							Text:              "if that creature was a Horror",
							Object:            opt.Val(game.EventPermanentReference()),
							ObjectMatches:     opt.Val(game.Selection{SubtypesAny: []types.Sub{types.Horror}}),
							SourceCounterKind: counter.PlusOnePlusOne,
						}),
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.MoveCard{
											Card: game.CardReference{
												Kind: game.CardReferenceSource,
											},
											FromZone:    zone.Graveyard,
											Destination: zone.Hand,
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
			Enchant creature you control
			At the beginning of your upkeep, create a token that's a copy of enchanted creature, except the token is 1/1.
			When enchanted creature dies, if that creature was a Horror, return this card to its owner's hand.
		`,
		},
	}
}
