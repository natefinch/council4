package y

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// YunaSDecision is the card definition for Yuna's Decision.
//
// Type: Sorcery
// Cost: {3}{G}
//
// Oracle text:
//
//	Choose one —
//	• Continue the Pilgrimage — Sacrifice a creature. If you do, draw a card, then you may put a creature card and/or a land card from your hand onto the battlefield.
//	• Find Another Way — Return one or two target permanent cards from your graveyard to your hand.
var YunaSDecision = newYunaSDecision

func newYunaSDecision() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Yuna's Decision",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					game.Mode{
						Text: "Continue the Pilgrimage — Sacrifice a creature. If you do, draw a card, then you may put a creature card and/or a land card from your hand onto the battlefield.",
						Sequence: []game.Instruction{
							{
								Primitive: game.SacrificePermanents{
									Player:    game.ControllerReference(),
									Amount:    game.Fixed(1),
									Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
								},
								PublishResult: game.ResultKey("if-you-do"),
								LocalProducts: game.LocalProducts{
									Results: []game.ResultKey{game.ResultKey("if-you-do")},
								},
							},
							{
								Primitive: game.Draw{
									Amount: game.Fixed(1),
									Player: game.ControllerReference(),
								},
								ResultGate: opt.Val(game.InstructionResultGate{
									Key:       "if-you-do",
									Succeeded: game.TriTrue,
								}),
							},
							{
								Primitive: game.ChooseFromZone{
									Player:     game.ControllerReference(),
									SourceZone: zone.Hand,
									Filter:     game.Selection{RequiredTypesAny: []types.Card{types.Creature, types.Land}},
									Quantity:   game.Fixed(1),
									Destination: game.ChooseDestination{
										Zone: zone.Battlefield,
									},
									Prompt: "Choose a card to put onto the battlefield",
								},
								ResultGate: opt.Val(game.InstructionResultGate{
									Key:       "if-you-do",
									Succeeded: game.TriTrue,
								}),
								Optional: true,
							},
						},
					},
					game.Mode{
						Text: "Find Another Way — Return one or two target permanent cards from your graveyard to your hand.",
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 2,
								Constraint: "one or two target permanent cards from your graveyard",
								Allow:      game.TargetAllowCard,
								TargetZone: zone.Graveyard,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Creature, types.Enchantment, types.Land, types.Planeswalker, types.Battle}, Controller: game.ControllerYou}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.MoveCard{
									Card: game.CardReference{
										Kind: game.CardReferenceTarget,
									},
									FromZone:    zone.Graveyard,
									Destination: zone.Hand,
								},
							},
							{
								Primitive: game.MoveCard{
									Card: game.CardReference{
										Kind:        game.CardReferenceTarget,
										TargetIndex: 1,
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
			}),
			OracleText: `
			Choose one —
			• Continue the Pilgrimage — Sacrifice a creature. If you do, draw a card, then you may put a creature card and/or a land card from your hand onto the battlefield.
			• Find Another Way — Return one or two target permanent cards from your graveyard to your hand.
		`,
		},
	}
}
