package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// SkycatSovereign is the card definition for Skycat Sovereign.
//
// Type: Creature — Elemental Cat
// Cost: {W}{U}
//
// Oracle text:
//
//	Flying
//	This creature gets +1/+1 for each other creature you control with flying.
//	{2}{W}{U}: Create a 1/1 white Cat Bird creature token with flying.
var SkycatSovereign = newSkycatSovereign

func newSkycatSovereign() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue),
		CardFace: game.CardFace{
			Name: "Skycat Sovereign",
			ManaCost: opt.Val(cost.Mana{
				cost.W,
				cost.U,
			}),
			Colors:    []color.Color{color.Blue, color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elemental, types.Cat},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.StaticAbility{
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer:          game.LayerPowerToughnessModify,
							AffectedSource: true,
							PowerDeltaDynamic: opt.Val(game.DynamicAmount{
								Kind:       game.DynamicAmountCountSelector,
								Multiplier: 1,
								Group:      game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou, Keyword: game.Flying, ExcludeSource: true}),
							}),
							ToughnessDeltaDynamic: opt.Val(game.DynamicAmount{
								Kind:       game.DynamicAmountCountSelector,
								Multiplier: 1,
								Group:      game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou, Keyword: game.Flying, ExcludeSource: true}),
							}),
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:           "{2}{W}{U}: Create a 1/1 white Cat Bird creature token with flying.",
					ManaCost:       opt.Val(cost.Mana{cost.O(2), cost.W, cost.U}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(1),
									Source: game.TokenDef(skycatSovereignToken),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Flying
			This creature gets +1/+1 for each other creature you control with flying.
			{2}{W}{U}: Create a 1/1 white Cat Bird creature token with flying.
		`,
		},
	}
}

var skycatSovereignToken = newSkycatSovereignToken()

func newSkycatSovereignToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Cat Bird",
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Cat, types.Bird},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
		},
	}
}
