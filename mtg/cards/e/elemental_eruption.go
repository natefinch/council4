package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ElementalEruption is the card definition for Elemental Eruption.
//
// Type: Sorcery
// Cost: {4}{R}{R}
//
// Oracle text:
//
//	Create a 4/4 red Dragon Elemental creature token with flying and prowess.
//	Storm (When you cast this spell, copy it for each spell cast before it this turn.)
var ElementalEruption = newElementalEruption

func newElementalEruption() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Elemental Eruption",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.R,
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Sorcery},
			StaticAbilities: []game.StaticAbility{
				game.StormStaticBody,
			},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.CreateToken{
							Amount: game.Fixed(1),
							Source: game.TokenDef(elementalEruptionToken),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Create a 4/4 red Dragon Elemental creature token with flying and prowess.
			Storm (When you cast this spell, copy it for each spell cast before it this turn.)
		`,
		},
	}
}

var elementalEruptionToken = newElementalEruptionToken()

func newElementalEruptionToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Dragon Elemental",
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Dragon, types.Elemental},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.ProwessStaticBody,
			},
		},
	}
}
