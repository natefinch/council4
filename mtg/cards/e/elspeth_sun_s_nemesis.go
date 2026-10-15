package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ElspethSunSNemesis is the card definition for Elspeth, Sun's Nemesis.
//
// Type: Legendary Planeswalker — Elspeth
// Cost: {2}{W}{W}
//
// Oracle text:
//
//	−1: Up to two target creatures you control each get +2/+1 until end of turn.
//	−2: Create two 1/1 white Human Soldier creature tokens.
//	−3: You gain 5 life.
//	Escape—{4}{W}{W}, Exile four other cards from your graveyard. (You may cast this card from your graveyard for its escape cost.)
var ElspethSunSNemesis = newElspethSunSNemesis

func newElspethSunSNemesis() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Elspeth, Sun's Nemesis",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
				cost.W,
			}),
			Colors:     []color.Color{color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Planeswalker},
			Subtypes:   []types.Sub{types.Elspeth},
			Loyalty:    opt.Val(5),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					KeywordAbilities: []game.KeywordAbility{
						game.SimpleKeyword{
							Kind: game.Escape,
						},
					},
				},
			},
			LoyaltyAbilities: []game.LoyaltyAbility{
				game.LoyaltyAbility{
					Text:        "−1: Up to two target creatures you control each get +2/+1 until end of turn.",
					LoyaltyCost: -1,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MaxTargets: 2,
										Constraint: "up to two target creatures you control",
										Allow:      game.TargetAllowPermanent,
										Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Controller: game.ControllerYou}),
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.ModifyPT{
											Object:         game.TargetPermanentReference(0),
											PowerDelta:     game.Fixed(2),
											ToughnessDelta: game.Fixed(1),
											Duration:       game.DurationUntilEndOfTurn,
										},
									},
									game.Instruction{
										Primitive: game.ModifyPT{
											Object:         game.TargetPermanentReference(1),
											PowerDelta:     game.Fixed(2),
											ToughnessDelta: game.Fixed(1),
											Duration:       game.DurationUntilEndOfTurn,
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
					Text:        "−2: Create two 1/1 white Human Soldier creature tokens.",
					LoyaltyCost: -2,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateToken{
											Amount: game.Fixed(2),
											Source: game.TokenDef(elspethSunSNemesisToken),
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
					Text:        "−3: You gain 5 life.",
					LoyaltyCost: -3,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.GainLife{
											Amount: game.Fixed(5),
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
			AlternativeCosts: []cost.Alternative{
				cost.Alternative{
					Label:    "Escape",
					ManaCost: opt.Val(cost.Mana{cost.O(4), cost.W, cost.W}),
					AdditionalCosts: []cost.Additional{
						{
							Kind:          cost.AdditionalExile,
							Text:          "Exile four other cards from your graveyard",
							Amount:        4,
							Source:        zone.Graveyard,
							ExcludeSource: true,
						},
					},
					Mechanic: cost.AlternativeMechanicEscape,
				},
			},
			OracleText: `
			−1: Up to two target creatures you control each get +2/+1 until end of turn.
			−2: Create two 1/1 white Human Soldier creature tokens.
			−3: You gain 5 life.
			Escape—{4}{W}{W}, Exile four other cards from your graveyard. (You may cast this card from your graveyard for its escape cost.)
		`,
		},
	}
}

var elspethSunSNemesisToken = newElspethSunSNemesisToken()

func newElspethSunSNemesisToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Human Soldier",
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Soldier},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
