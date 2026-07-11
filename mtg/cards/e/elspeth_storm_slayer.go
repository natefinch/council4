package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ElspethStormSlayer is the card definition for Elspeth, Storm Slayer.
//
// Type: Legendary Planeswalker — Elspeth
// Cost: {3}{W}{W}
//
// Oracle text:
//
//	If one or more tokens would be created under your control, twice that many of those tokens are created instead.
//	+1: Create a 1/1 white Soldier creature token.
//	0: Put a +1/+1 counter on each creature you control. Those creatures gain flying until your next turn.
//	−3: Destroy target creature an opponent controls with mana value 3 or greater.
var ElspethStormSlayer = newElspethStormSlayer

func newElspethStormSlayer() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Elspeth, Storm Slayer",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
				cost.W,
			}),
			Colors:     []color.Color{color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Planeswalker},
			Subtypes:   []types.Sub{types.Elspeth},
			Loyalty:    opt.Val(5),
			LoyaltyAbilities: []game.LoyaltyAbility{
				game.LoyaltyAbility{
					LoyaltyCost: 1,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(1),
									Source: game.TokenDef(elspethStormSlayerToken),
								},
							},
						},
					}.Ability(),
				},
				game.LoyaltyAbility{
					LoyaltyCost: 0,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.AddCounter{
									Amount:      game.Fixed(1),
									Group:       game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
									CounterKind: counter.PlusOnePlusOne,
								},
							},
							{
								Primitive: game.ApplyContinuous{
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Layer: game.LayerAbility,
											Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
											AddKeywords: []game.Keyword{
												game.Flying,
											},
										},
									},
									Duration: game.DurationUntilYourNextTurn,
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
								Constraint: "target creature an opponent controls with mana value 3 or greater",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Controller: game.ControllerOpponent, ManaValue: opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 3})}),
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
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.TokenCreationReplacement("If one or more tokens would be created under your control, twice that many of those tokens are created instead.", 2, game.TriggerControllerYou),
			},
			OracleText: `
			If one or more tokens would be created under your control, twice that many of those tokens are created instead.
			+1: Create a 1/1 white Soldier creature token.
			0: Put a +1/+1 counter on each creature you control. Those creatures gain flying until your next turn.
			−3: Destroy target creature an opponent controls with mana value 3 or greater.
		`,
		},
	}
}

var elspethStormSlayerToken = newElspethStormSlayerToken()

func newElspethStormSlayerToken() *game.CardDef {
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
