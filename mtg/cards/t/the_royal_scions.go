package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// TheRoyalScions is the card definition for The Royal Scions.
//
// Type: Legendary Planeswalker — Will Rowan
// Cost: {1}{U}{R}
//
// Oracle text:
//
//	+1: Draw a card, then discard a card.
//	+1: Target creature gets +2/+0 and gains first strike and trample until end of turn.
//	−8: Draw four cards. When you do, The Royal Scions deals damage to any target equal to the number of cards in your hand.
var TheRoyalScions = newTheRoyalScions

func newTheRoyalScions() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Red),
		CardFace: game.CardFace{
			Name: "The Royal Scions",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
				cost.R,
			}),
			Colors:     []color.Color{color.Red, color.Blue},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Planeswalker},
			Subtypes:   []types.Sub{types.Will, types.Rowan},
			Loyalty:    opt.Val(5),
			LoyaltyAbilities: []game.LoyaltyAbility{
				game.LoyaltyAbility{
					Text:        "+1: Draw a card, then discard a card.",
					LoyaltyCost: 1,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Draw{
											Amount: game.Fixed(1),
											Player: game.ControllerReference(),
										},
									},
									game.Instruction{
										Primitive: game.Discard{
											Amount: game.Fixed(1),
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
				game.LoyaltyAbility{
					Text:        "+1: Target creature gets +2/+0 and gains first strike and trample until end of turn.",
					LoyaltyCost: 1,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target creature",
										Allow:      game.TargetAllowPermanent,
										Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.ApplyContinuous{
											Object: opt.Val(game.TargetPermanentReference(0)),
											ContinuousEffects: []game.ContinuousEffect{
												game.ContinuousEffect{
													Duration:   game.DurationPermanent,
													Layer:      game.LayerPowerToughnessModify,
													PowerDelta: 2,
												},
												game.ContinuousEffect{
													Duration:    game.DurationPermanent,
													Layer:       game.LayerAbility,
													AddKeywords: []game.Keyword{game.FirstStrike, game.Trample},
												},
											},
											Duration: game.DurationUntilEndOfTurn,
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
				game.LoyaltyAbility{
					Text:        "−8: Draw four cards. When you do, The Royal Scions deals damage to any target equal to the number of cards in your hand.",
					LoyaltyCost: -8,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "any target",
										Allow:      game.TargetAllowPermanent | game.TargetAllowPlayer,
										Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature, types.Planeswalker, types.Battle}}),
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Draw{
											Amount: game.Fixed(4),
											Player: game.ControllerReference(),
										},
									},
									game.Instruction{
										Primitive: game.Damage{
											Amount: game.Dynamic(game.DynamicAmount{
												Kind:       game.DynamicAmountCountCardsInZone,
												Multiplier: 1,
												Player:     func() *game.PlayerReference { ref := game.ControllerReference(); return &ref }(),
												CardZone:   zone.Hand,
												Selection:  &game.Selection{},
											}),
											Recipient:    game.AnyTargetDamageRecipient(0),
											DamageSource: opt.Val(game.SourcePermanentReference()),
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
			+1: Draw a card, then discard a card.
			+1: Target creature gets +2/+0 and gains first strike and trample until end of turn.
			−8: Draw four cards. When you do, The Royal Scions deals damage to any target equal to the number of cards in your hand.
		`,
		},
	}
}
