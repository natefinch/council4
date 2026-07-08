package o

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// OviyaPashiriSageLifecrafter is the card definition for Oviya Pashiri, Sage Lifecrafter.
//
// Type: Legendary Creature — Human Artificer
// Cost: {G}
//
// Oracle text:
//
//	{2}{G}, {T}: Create a 1/1 colorless Servo artifact creature token.
//	{4}{G}, {T}: Create an X/X colorless Construct artifact creature token, where X is the number of creatures you control.
var OviyaPashiriSageLifecrafter = newOviyaPashiriSageLifecrafter

func newOviyaPashiriSageLifecrafter() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Oviya Pashiri, Sage Lifecrafter",
			ManaCost: opt.Val(cost.Mana{
				cost.G,
			}),
			Colors:     []color.Color{color.Green},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Artificer},
			Power:      opt.Val(game.PT{Value: 1}),
			Toughness:  opt.Val(game.PT{Value: 2}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:            "{2}{G}, {T}: Create a 1/1 colorless Servo artifact creature token.",
					ManaCost:        opt.Val(cost.Mana{cost.O(2), cost.G}),
					AdditionalCosts: cost.Tap,
					ZoneOfFunction:  zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(1),
									Source: game.TokenDef(oviyaPashiriSageLifecrafterToken),
								},
							},
						},
					}.Ability(),
				},
				game.ActivatedAbility{
					Text:            "{4}{G}, {T}: Create an X/X colorless Construct artifact creature token, where X is the number of creatures you control.",
					ManaCost:        opt.Val(cost.Mana{cost.O(4), cost.G}),
					AdditionalCosts: cost.Tap,
					ZoneOfFunction:  zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(1),
									Source: game.TokenDef(oviyaPashiriSageLifecrafterToken2),
									Power: opt.Val(game.Dynamic(game.DynamicAmount{
										Kind:       game.DynamicAmountCountSelector,
										Multiplier: 1,
										Group:      game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
									})),
									Toughness: opt.Val(game.Dynamic(game.DynamicAmount{
										Kind:       game.DynamicAmountCountSelector,
										Multiplier: 1,
										Group:      game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
									})),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			{2}{G}, {T}: Create a 1/1 colorless Servo artifact creature token.
			{4}{G}, {T}: Create an X/X colorless Construct artifact creature token, where X is the number of creatures you control.
		`,
		},
	}
}

var oviyaPashiriSageLifecrafterToken = newOviyaPashiriSageLifecrafterToken()

func newOviyaPashiriSageLifecrafterToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Servo",
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Servo},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}

var oviyaPashiriSageLifecrafterToken2 = newOviyaPashiriSageLifecrafterToken2()

func newOviyaPashiriSageLifecrafterToken2() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:     "Construct",
			Types:    []types.Card{types.Artifact, types.Creature},
			Subtypes: []types.Sub{types.Construct},
		},
	}
}
