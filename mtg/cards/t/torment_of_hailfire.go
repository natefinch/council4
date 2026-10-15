package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TormentOfHailfire is the card definition for Torment of Hailfire.
//
// Type: Sorcery
// Cost: {X}{B}{B}
//
// Oracle text:
//
//	Repeat the following process X times. Each opponent loses 3 life unless that player sacrifices a nonland permanent of their choice or discards a card.
var TormentOfHailfire = newTormentOfHailfire

func newTormentOfHailfire() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Torment of Hailfire",
			ManaCost: opt.Val(cost.Mana{
				cost.X,
				cost.B,
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.RepeatProcess{
							Times: game.Dynamic(game.DynamicAmount{
								Kind: game.DynamicAmountX,
							}),
							Body: game.AbilityContent{
								Modes: []game.Mode{
									game.Mode{
										Sequence: []game.Instruction{
											game.Instruction{
												Primitive: game.PunisherEachLoseLife{
													PlayerGroup:        game.OpponentsReference(),
													Amount:             game.Fixed(3),
													AllowSacrifice:     true,
													SacrificeSelection: game.Selection{ExcludedTypes: []types.Card{types.Land}},
													AllowDiscard:       true,
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
				},
			}.Ability()),
			OracleText: `
			Repeat the following process X times. Each opponent loses 3 life unless that player sacrifices a nonland permanent of their choice or discards a card.
		`,
		},
	}
}
