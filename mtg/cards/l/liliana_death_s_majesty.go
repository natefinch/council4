package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// LilianaDeathSMajesty is the card definition for Liliana, Death's Majesty.
//
// Type: Legendary Planeswalker — Liliana
// Cost: {3}{B}{B}
//
// Oracle text:
//
//	+1: Create a 2/2 black Zombie creature token. Mill two cards.
//	−3: Return target creature card from your graveyard to the battlefield. That creature is a black Zombie in addition to its other colors and types.
//	−7: Destroy all non-Zombie creatures.
var LilianaDeathSMajesty = newLilianaDeathSMajesty()

func newLilianaDeathSMajesty() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Liliana, Death's Majesty",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.B,
				cost.B,
			}),
			Colors:     []color.Color{color.Black},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Planeswalker},
			Subtypes:   []types.Sub{types.Liliana},
			Loyalty:    opt.Val(5),
			LoyaltyAbilities: []game.LoyaltyAbility{
				game.LoyaltyAbility{
					LoyaltyCost: 1,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(1),
									Source: game.TokenDef(lilianaDeathSMajestyToken),
								},
							},
							{
								Primitive: game.Mill{
									Amount: game.Fixed(2),
									Player: game.ControllerReference(),
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
								Constraint: "target creature card from your graveyard",
								Allow:      game.TargetAllowCard,
								TargetZone: zone.Graveyard,
								Selection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.PutOnBattlefield{
									Source:        game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget}),
									PublishLinked: game.LinkedKey("leave-bf-exile-1"),
								},
							},
							{
								Primitive: game.ApplyContinuous{
									Object: opt.Val(game.LinkedObjectReference("leave-bf-exile-1")),
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Layer:     game.LayerColor,
											AddColors: []color.Color{color.Black},
										},
										game.ContinuousEffect{
											Layer:       game.LayerType,
											AddSubtypes: []types.Sub{types.Sub("Zombie")},
										},
									},
									Duration: game.DurationPermanent,
								},
							},
						},
					}.Ability(),
				},
				game.LoyaltyAbility{
					LoyaltyCost: -7,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Destroy{
									Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, ExcludedSubtype: types.Sub("Zombie")}),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			+1: Create a 2/2 black Zombie creature token. Mill two cards.
			−3: Return target creature card from your graveyard to the battlefield. That creature is a black Zombie in addition to its other colors and types.
			−7: Destroy all non-Zombie creatures.
		`,
		},
	}
}

var lilianaDeathSMajestyToken = newLilianaDeathSMajestyToken()

func newLilianaDeathSMajestyToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Zombie",
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Zombie},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
		},
	}
}
