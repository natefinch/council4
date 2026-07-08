package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// CleanupCrew is the card definition for Cleanup Crew.
//
// Type: Creature — Human Citizen
// Cost: {4}{G}{G}
//
// Oracle text:
//
//	When this creature enters, choose one —
//	• Destroy target artifact.
//	• Destroy target enchantment.
//	• Exile target card from a graveyard.
//	• You gain 4 life.
var CleanupCrew = newCleanupCrew

func newCleanupCrew() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Cleanup Crew",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.G,
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Citizen},
			Power:     opt.Val(game.PT{Value: 6}),
			Toughness: opt.Val(game.PT{Value: 6}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:  game.EventPermanentEnteredBattlefield,
							Source: game.TriggerSourceSelf,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Text: "Destroy target artifact.",
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target artifact",
										Allow:      game.TargetAllowPermanent,
										Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Artifact}}),
									},
								},
								Sequence: []game.Instruction{
									{
										Primitive: game.Destroy{
											Object: game.TargetPermanentReference(0),
										},
									},
								},
							},
							game.Mode{
								Text: "Destroy target enchantment.",
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target enchantment",
										Allow:      game.TargetAllowPermanent,
										Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Enchantment}}),
									},
								},
								Sequence: []game.Instruction{
									{
										Primitive: game.Destroy{
											Object: game.TargetPermanentReference(0),
										},
									},
								},
							},
							game.Mode{
								Text: "Exile target card from a graveyard.",
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target card from a graveyard",
										Allow:      game.TargetAllowCard,
										TargetZone: zone.Graveyard,
										Selection:  opt.Val(game.Selection{}),
									},
								},
								Sequence: []game.Instruction{
									{
										Primitive: game.MoveCard{
											Card:        game.CardReference{Kind: game.CardReferenceTarget},
											FromZone:    zone.Graveyard,
											Destination: zone.Exile,
										},
									},
								},
							},
							game.Mode{
								Text: "You gain 4 life.",
								Sequence: []game.Instruction{
									{
										Primitive: game.GainLife{
											Amount: game.Fixed(4),
											Player: game.ControllerReference(),
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
			When this creature enters, choose one —
			• Destroy target artifact.
			• Destroy target enchantment.
			• Exile target card from a graveyard.
			• You gain 4 life.
		`,
		},
	}
}
