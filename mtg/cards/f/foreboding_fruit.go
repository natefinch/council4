package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ForebodingFruit is the card definition for Foreboding Fruit.
//
// Type: Sorcery
// Cost: {2}{B}
//
// Oracle text:
//
//	Target player draws two cards and loses 2 life.
//	Adamant — If at least three black mana was spent to cast this spell, create a Food token. (It's an artifact with "{2}, {T}, Sacrifice this token: You gain 3 life.")
var ForebodingFruit = newForebodingFruit

func newForebodingFruit() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Foreboding Fruit",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target player",
						Allow:      game.TargetAllowPlayer,
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Draw{
							Amount: game.Fixed(2),
							Player: game.TargetPlayerReference(0),
						},
					},
					{
						Primitive: game.LoseLife{
							Amount: game.Fixed(2),
							Player: game.TargetPlayerReference(0),
						},
					},
					{
						Primitive: game.CreateToken{
							Amount: game.Fixed(1),
							Source: game.TokenDef(forebodingFruitToken),
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								SpellColorManaSpent: game.ColorManaSpendThreshold{Color: color.Black, Count: 3},
							}),
						}),
					},
				},
			}.Ability()),
			OracleText: `
			Target player draws two cards and loses 2 life.
			Adamant — If at least three black mana was spent to cast this spell, create a Food token. (It's an artifact with "{2}, {T}, Sacrifice this token: You gain 3 life.")
		`,
		},
	}
}

var forebodingFruitToken = newForebodingFruitToken()

func newForebodingFruitToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:     "Food",
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Food},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:     "{2}, {T}, Sacrifice this artifact: You gain 3 life.",
					ManaCost: opt.Val(cost.Mana{cost.O(2)}),
					AdditionalCosts: []cost.Additional{
						{
							Kind: cost.AdditionalTap,
						},
						{
							Kind:               cost.AdditionalSacrificeSource,
							Text:               "Sacrifice this artifact",
							Amount:             1,
							MatchPermanentType: true,
							PermanentType:      types.Artifact,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.GainLife{
									Amount: game.Fixed(3),
									Player: game.ControllerReference(),
								},
							},
						},
					}.Ability(),
				},
			},
		},
	}
}
