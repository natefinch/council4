package n

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// NocturnalHunger is the card definition for Nocturnal Hunger.
//
// Type: Instant
// Cost: {2}{B}
//
// Oracle text:
//
//	Gift a Food (You may promise an opponent a gift as you cast this spell. If you do, they create a Food token before its other effects. It's an artifact with "{2}, {T}, Sacrifice this token: You gain 3 life.")
//	Destroy target creature. If the gift wasn't promised, you lose 2 life.
var NocturnalHunger = newNocturnalHunger

func newNocturnalHunger() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Nocturnal Hunger",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Instant},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					KeywordAbilities: []game.KeywordAbility{
						game.GiftKeyword{Delivery: game.Mode{
							Sequence: []game.Instruction{
								{
									Primitive: game.CreateToken{
										Amount:    game.Fixed(1),
										Source:    game.TokenDef(nocturnalHungerToken),
										Recipient: opt.Val(game.GiftRecipientReference()),
									},
								},
							},
						}.Ability()},
					},
				},
			},
			SpellAbility: opt.Val(game.Mode{
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
					{
						Primitive: game.Destroy{
							Object: game.TargetPermanentReference(0),
						},
					},
					{
						Primitive: game.LoseLife{
							Amount: game.Fixed(2),
							Player: game.ControllerReference(),
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								Negate:       true,
								GiftPromised: true,
							}),
						}),
					},
				},
			}.Ability()),
			OracleText: `
			Gift a Food (You may promise an opponent a gift as you cast this spell. If you do, they create a Food token before its other effects. It's an artifact with "{2}, {T}, Sacrifice this token: You gain 3 life.")
			Destroy target creature. If the gift wasn't promised, you lose 2 life.
		`,
		},
	}
}

var nocturnalHungerToken = newNocturnalHungerToken()

func newNocturnalHungerToken() *game.CardDef {
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
