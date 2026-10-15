package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ElspethSunSChampion is the card definition for Elspeth, Sun's Champion.
//
// Type: Legendary Planeswalker — Elspeth
// Cost: {4}{W}{W}
//
// Oracle text:
//
//	+1: Create three 1/1 white Soldier creature tokens.
//	−3: Destroy all creatures with power 4 or greater.
//	−7: You get an emblem with "Creatures you control get +2/+2 and have flying."
var ElspethSunSChampion = newElspethSunSChampion

func newElspethSunSChampion() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Elspeth, Sun's Champion",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.W,
				cost.W,
			}),
			Colors:     []color.Color{color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Planeswalker},
			Subtypes:   []types.Sub{types.Elspeth},
			Loyalty:    opt.Val(4),
			LoyaltyAbilities: []game.LoyaltyAbility{
				game.LoyaltyAbility{
					Text:        "+1: Create three 1/1 white Soldier creature tokens.",
					LoyaltyCost: 1,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateToken{
											Amount: game.Fixed(3),
											Source: game.TokenDef(elspethSunSChampionToken),
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
					Text:        "−3: Destroy all creatures with power 4 or greater.",
					LoyaltyCost: -3,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Destroy{
											Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Power: opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 4})}),
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
					Text:        "−7: You get an emblem with \"Creatures you control get +2/+2 and have flying.\"",
					LoyaltyCost: -7,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateEmblem{
											EmblemAbilities: []game.Ability{
												&game.StaticAbility{
													Text: "Creatures you control get +2/+2 and have flying.",
													ContinuousEffects: []game.ContinuousEffect{
														game.ContinuousEffect{
															Duration:       game.DurationPermanent,
															Group:          game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}}),
															Layer:          game.LayerPowerToughnessModify,
															PowerDelta:     2,
															ToughnessDelta: 2,
														},
														game.ContinuousEffect{
															Duration:    game.DurationPermanent,
															Group:       game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}}),
															Layer:       game.LayerAbility,
															AddKeywords: []game.Keyword{game.Flying},
														},
													},
												},
											},
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
			+1: Create three 1/1 white Soldier creature tokens.
			−3: Destroy all creatures with power 4 or greater.
			−7: You get an emblem with "Creatures you control get +2/+2 and have flying."
		`,
		},
	}
}

var elspethSunSChampionToken = newElspethSunSChampionToken()

func newElspethSunSChampionToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Soldier",
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Soldier},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
