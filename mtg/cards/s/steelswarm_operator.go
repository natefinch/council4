package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SteelswarmOperator is the card definition for Steelswarm Operator.
//
// Type: Artifact Creature — Robot Soldier
// Cost: {1}{U}
//
// Oracle text:
//
//	Flying
//	{T}: Add {U}. Spend this mana only to cast an artifact spell.
//	{T}: Add {U}{U}. Spend this mana only to activate abilities of artifact sources.
var SteelswarmOperator = newSteelswarmOperator()

func newSteelswarmOperator() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Steelswarm Operator",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Robot, types.Soldier},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			ManaAbilities: []game.ManaAbility{
				game.ManaAbility{
					AdditionalCosts: cost.Tap,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.AddMana{
									Amount:    game.Fixed(1),
									ManaColor: mana.U,
									SpendRider: opt.Val(game.ManaSpendRider{
										Condition:   game.ManaSpendCastArtifactSpellOnly,
										Restriction: game.ManaSpendRestrictedToCondition,
									}),
								},
							},
						},
					}.Ability(),
				},
				game.ManaAbility{
					AdditionalCosts: cost.Tap,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.AddMana{
									Amount:    game.Fixed(1),
									ManaColor: mana.U,
									SpendRider: opt.Val(game.ManaSpendRider{
										Condition:   game.ManaSpendActivateArtifactAbility,
										Restriction: game.ManaSpendRestrictedToCondition,
									}),
								},
							},
							{
								Primitive: game.AddMana{
									Amount:    game.Fixed(1),
									ManaColor: mana.U,
									SpendRider: opt.Val(game.ManaSpendRider{
										Condition:   game.ManaSpendActivateArtifactAbility,
										Restriction: game.ManaSpendRestrictedToCondition,
									}),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Flying
			{T}: Add {U}. Spend this mana only to cast an artifact spell.
			{T}: Add {U}{U}. Spend this mana only to activate abilities of artifact sources.
		`,
		},
	}
}
