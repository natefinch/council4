package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TheVision is the card definition for The Vision.
//
// Type: Legendary Artifact Creature — Robot Hero
// Cost: {4}
//
// Oracle text:
//
//	Flying, vigilance
//	Whenever you cast a noncreature spell, choose one that hasn't been chosen this turn —
//	• Solar Beam — The Vision gains double strike until end of turn.
//	• Density Control — The Vision gains indestructible until end of turn.
//	• Technopathy — Draw a card.
var TheVision = newTheVision

func newTheVision() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "The Vision",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
			}),
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Artifact, types.Creature},
			Subtypes:   []types.Sub{types.Robot, types.Hero},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.VigilanceStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever you cast a noncreature spell, choose one that hasn't been chosen this turn —\n• Solar Beam — The Vision gains double strike until end of turn.\n• Density Control — The Vision gains indestructible until end of turn.\n• Technopathy — Draw a card.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:           game.EventSpellCast,
							Controller:      game.TriggerControllerYou,
							CardSelection:   game.Selection{ExcludedTypes: []types.Card{types.Creature}},
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Text: "Solar Beam — The Vision gains double strike until end of turn.",
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.ApplyContinuous{
											Object: opt.Val(game.SourcePermanentReference()),
											ContinuousEffects: []game.ContinuousEffect{
												game.ContinuousEffect{
													Duration:    game.DurationPermanent,
													Layer:       game.LayerAbility,
													AddKeywords: []game.Keyword{game.DoubleStrike},
												},
											},
											Duration: game.DurationUntilEndOfTurn,
										},
									},
								},
							},
							game.Mode{
								Text: "Density Control — The Vision gains indestructible until end of turn.",
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.ApplyContinuous{
											Object: opt.Val(game.SourcePermanentReference()),
											ContinuousEffects: []game.ContinuousEffect{
												game.ContinuousEffect{
													Duration:    game.DurationPermanent,
													Layer:       game.LayerAbility,
													AddKeywords: []game.Keyword{game.Indestructible},
												},
											},
											Duration: game.DurationUntilEndOfTurn,
										},
									},
								},
							},
							game.Mode{
								Text: "Technopathy — Draw a card.",
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Draw{
											Amount: game.Fixed(1),
											Player: game.ControllerReference(),
										},
									},
								},
							},
						},
						MinModes:           1,
						MaxModes:           1,
						ModesUniquePerTurn: true,
					},
				},
			},
			OracleText: `
			Flying, vigilance
			Whenever you cast a noncreature spell, choose one that hasn't been chosen this turn —
			• Solar Beam — The Vision gains double strike until end of turn.
			• Density Control — The Vision gains indestructible until end of turn.
			• Technopathy — Draw a card.
		`,
		},
	}
}
