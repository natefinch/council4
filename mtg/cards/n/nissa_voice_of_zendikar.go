package n

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// NissaVoiceOfZendikar is the card definition for Nissa, Voice of Zendikar.
//
// Type: Legendary Planeswalker — Nissa
// Cost: {1}{G}{G}
//
// Oracle text:
//
//	+1: Create a 0/1 green Plant creature token.
//	−2: Put a +1/+1 counter on each creature you control.
//	−7: You gain X life and draw X cards, where X is the number of lands you control.
var NissaVoiceOfZendikar = newNissaVoiceOfZendikar

func newNissaVoiceOfZendikar() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Nissa, Voice of Zendikar",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
				cost.G,
			}),
			Colors:     []color.Color{color.Green},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Planeswalker},
			Subtypes:   []types.Sub{types.Nissa},
			Loyalty:    opt.Val(3),
			LoyaltyAbilities: []game.LoyaltyAbility{
				game.LoyaltyAbility{
					Text:        "+1: Create a 0/1 green Plant creature token.",
					LoyaltyCost: 1,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateToken{
											Amount: game.Fixed(1),
											Source: game.TokenDef(nissaVoiceOfZendikarToken),
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
					Text:        "−2: Put a +1/+1 counter on each creature you control.",
					LoyaltyCost: -2,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.AddCounter{
											Amount:      game.Fixed(1),
											Group:       game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
											CounterKind: counter.PlusOnePlusOne,
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
					Text:        "−7: You gain X life and draw X cards, where X is the number of lands you control.",
					LoyaltyCost: -7,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.GainLife{
											Amount: game.Dynamic(game.DynamicAmount{
												Kind: game.DynamicAmountX,
											}),
											Player: game.ControllerReference(),
										},
									},
									game.Instruction{
										Primitive: game.Draw{
											Amount: game.Dynamic(game.DynamicAmount{
												Kind:       game.DynamicAmountCountSelector,
												Multiplier: 1,
												Group:      game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Land}, Controller: game.ControllerYou}),
											}),
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
			OracleText: `
			+1: Create a 0/1 green Plant creature token.
			−2: Put a +1/+1 counter on each creature you control.
			−7: You gain X life and draw X cards, where X is the number of lands you control.
		`,
		},
	}
}

var nissaVoiceOfZendikarToken = newNissaVoiceOfZendikarToken()

func newNissaVoiceOfZendikarToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Plant",
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Plant},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
