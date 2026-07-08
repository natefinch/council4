package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// VivienReid is the card definition for Vivien Reid.
//
// Type: Legendary Planeswalker — Vivien
// Cost: {3}{G}{G}
//
// Oracle text:
//
//	+1: Look at the top four cards of your library. You may reveal a creature or land card from among them and put it into your hand. Put the rest on the bottom of your library in a random order.
//	−3: Destroy target artifact, enchantment, or creature with flying.
//	−8: You get an emblem with "Creatures you control get +2/+2 and have vigilance, trample, and indestructible."
var VivienReid = newVivienReid

func newVivienReid() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Vivien Reid",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.G,
				cost.G,
			}),
			Colors:     []color.Color{color.Green},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Planeswalker},
			Subtypes:   []types.Sub{types.Vivien},
			Loyalty:    opt.Val(5),
			LoyaltyAbilities: []game.LoyaltyAbility{
				game.LoyaltyAbility{
					LoyaltyCost: 1,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Dig{
									Player:    game.ControllerReference(),
									Look:      game.Fixed(4),
									Take:      game.Fixed(1),
									Remainder: game.DigRemainderLibraryBottom,
									Filter:    opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature, types.Land}}),
									TakeUpTo:  true,
									Reveal:    true,
								},
							},
						},
					}.Ability(),
				},
				game.LoyaltyAbility{
					LoyaltyCost: -3,
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{AnyOf: []game.Selection{game.Selection{RequiredTypesAny: []types.Card{types.Artifact}}, game.Selection{RequiredTypesAny: []types.Card{types.Enchantment}}, game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Keyword: game.Flying}}}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.Destroy{
									Object: game.TargetPermanentReference(0),
								},
							},
						},
					}.Ability(),
				},
				game.LoyaltyAbility{
					LoyaltyCost: -8,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateEmblem{
									EmblemAbilities: []game.Ability{
										new(game.StaticAbility{
											ContinuousEffects: []game.ContinuousEffect{
												game.ContinuousEffect{
													Layer:          game.LayerPowerToughnessModify,
													Group:          game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}}),
													PowerDelta:     2,
													ToughnessDelta: 2,
												},
												game.ContinuousEffect{
													Layer: game.LayerAbility,
													Group: game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}}),
													AddKeywords: []game.Keyword{
														game.Vigilance,
														game.Trample,
														game.Indestructible,
													},
												},
											},
										}),
									},
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			+1: Look at the top four cards of your library. You may reveal a creature or land card from among them and put it into your hand. Put the rest on the bottom of your library in a random order.
			−3: Destroy target artifact, enchantment, or creature with flying.
			−8: You get an emblem with "Creatures you control get +2/+2 and have vigilance, trample, and indestructible."
		`,
		},
	}
}
