package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// SnakeBasket is the card definition for Snake Basket.
//
// Type: Artifact
// Cost: {4}
//
// Oracle text:
//
//	{X}, Sacrifice this artifact: Create X 1/1 green Snake creature tokens. Activate only as a sorcery.
var SnakeBasket = newSnakeBasket

func newSnakeBasket() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Snake Basket",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
			}),
			Types: []types.Card{types.Artifact},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{X}, Sacrifice this artifact: Create X 1/1 green Snake creature tokens. Activate only as a sorcery.",
					ManaCost: opt.Val(cost.Mana{
						cost.X,
					}),
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalSacrificeSource,
							Text:        "Sacrifice this artifact",
							Amount:      1,
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Timing:         game.SorceryOnly,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateToken{
											Amount: game.Dynamic(game.DynamicAmount{
												Kind: game.DynamicAmountX,
											}),
											Source: game.TokenDef(snakeBasketToken),
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
			{X}, Sacrifice this artifact: Create X 1/1 green Snake creature tokens. Activate only as a sorcery.
		`,
		},
	}
}

var snakeBasketToken = newSnakeBasketToken()

func newSnakeBasketToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Snake",
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Snake},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
