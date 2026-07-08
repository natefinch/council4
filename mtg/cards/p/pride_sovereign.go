package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// PrideSovereign is the card definition for Pride Sovereign.
//
// Type: Creature — Cat
// Cost: {2}{G}
//
// Oracle text:
//
//	This creature gets +1/+1 for each other Cat you control.
//	{W}, {T}, Exert this creature: Create two 1/1 white Cat creature tokens with lifelink. (An exerted creature won't untap during your next untap step.)
var PrideSovereign = newPrideSovereign

func newPrideSovereign() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Green),
		CardFace: game.CardFace{
			Name: "Pride Sovereign",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Cat},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer:          game.LayerPowerToughnessModify,
							AffectedSource: true,
							PowerDeltaDynamic: opt.Val(game.DynamicAmount{
								Kind:       game.DynamicAmountCountSelector,
								Multiplier: 1,
								Group:      game.BattlefieldGroup(game.Selection{SubtypesAny: []types.Sub{types.Sub("Cat")}, Controller: game.ControllerYou, ExcludeSource: true}),
							}),
							ToughnessDeltaDynamic: opt.Val(game.DynamicAmount{
								Kind:       game.DynamicAmountCountSelector,
								Multiplier: 1,
								Group:      game.BattlefieldGroup(game.Selection{SubtypesAny: []types.Sub{types.Sub("Cat")}, Controller: game.ControllerYou, ExcludeSource: true}),
							}),
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:     "{W}, {T}, Exert this creature: Create two 1/1 white Cat creature tokens with lifelink. (An exerted creature won't untap during your next untap step.)",
					ManaCost: opt.Val(cost.Mana{cost.W}),
					AdditionalCosts: []cost.Additional{
						{
							Kind: cost.AdditionalTap,
						},
						{
							Kind: cost.AdditionalExert,
							Text: "Exert this creature",
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(2),
									Source: game.TokenDef(prideSovereignToken),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			This creature gets +1/+1 for each other Cat you control.
			{W}, {T}, Exert this creature: Create two 1/1 white Cat creature tokens with lifelink. (An exerted creature won't untap during your next untap step.)
		`,
		},
	}
}

var prideSovereignToken = newPrideSovereignToken()

func newPrideSovereignToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Cat",
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Cat},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.LifelinkStaticBody,
			},
		},
	}
}
