package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TurnIntoAPumpkin is the card definition for Turn into a Pumpkin.
//
// Type: Instant
// Cost: {3}{U}
//
// Oracle text:
//
//	Return target nonland permanent to its owner's hand. Draw a card.
//	Adamant — If at least three blue mana was spent to cast this spell, create a Food token. (It's an artifact with "{2}, {T}, Sacrifice this token: You gain 3 life.")
var TurnIntoAPumpkin = newTurnIntoAPumpkin()

func newTurnIntoAPumpkin() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Turn into a Pumpkin",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target nonland permanent",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{ExcludedTypes: []types.Card{types.Land}}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Bounce{
							Object: game.TargetPermanentReference(0),
						},
					},
					{
						Primitive: game.Draw{
							Amount: game.Fixed(1),
							Player: game.ControllerReference(),
						},
					},
					{
						Primitive: game.CreateToken{
							Amount: game.Fixed(1),
							Source: game.TokenDef(turnIntoAPumpkinToken),
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								SpellColorManaSpent: game.ColorManaSpendThreshold{Color: color.Blue, Count: 3},
							}),
						}),
					},
				},
			}.Ability()),
			OracleText: `
			Return target nonland permanent to its owner's hand. Draw a card.
			Adamant — If at least three blue mana was spent to cast this spell, create a Food token. (It's an artifact with "{2}, {T}, Sacrifice this token: You gain 3 life.")
		`,
		},
	}
}

var turnIntoAPumpkinToken = newTurnIntoAPumpkinToken()

func newTurnIntoAPumpkinToken() *game.CardDef {
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
