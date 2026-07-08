package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// HeronBlessedGeist is the card definition for Heron-Blessed Geist.
//
// Type: Creature — Spirit
// Cost: {4}{W}
//
// Oracle text:
//
//	Flying
//	{3}{W}, Exile this card from your graveyard: Create two 1/1 white Spirit creature tokens with flying. Activate only if you control an enchantment and only as a sorcery.
var HeronBlessedGeist = newHeronBlessedGeist

func newHeronBlessedGeist() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Heron-Blessed Geist",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Spirit},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:     "{3}{W}, Exile this card from your graveyard: Create two 1/1 white Spirit creature tokens with flying. Activate only if you control an enchantment and only as a sorcery.",
					ManaCost: opt.Val(cost.Mana{cost.O(3), cost.W}),
					AdditionalCosts: []cost.Additional{
						{
							Kind:   cost.AdditionalExileSource,
							Text:   "Exile this card from your graveyard",
							Amount: 1,
							Source: zone.Graveyard,
						},
					},
					ZoneOfFunction: zone.Graveyard,
					Timing:         game.SorceryOnly,
					ActivationCondition: opt.Val(game.Condition{
						ControlsMatching: opt.Val(game.SelectionCount{
							Selection: game.Selection{RequiredTypes: []types.Card{types.Enchantment}},
						}),
					}),
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(2),
									Source: game.TokenDef(heronBlessedGeistToken),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Flying
			{3}{W}, Exile this card from your graveyard: Create two 1/1 white Spirit creature tokens with flying. Activate only if you control an enchantment and only as a sorcery.
		`,
		},
	}
}

var heronBlessedGeistToken = newHeronBlessedGeistToken()

func newHeronBlessedGeistToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Spirit",
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Spirit},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
		},
	}
}
