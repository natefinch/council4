package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// FlurryOfHorns is the card definition for Flurry of Horns.
//
// Type: Sorcery
// Cost: {4}{R}
//
// Oracle text:
//
//	Create two 2/3 red Minotaur creature tokens with haste.
var FlurryOfHorns = newFlurryOfHorns

func newFlurryOfHorns() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Flurry of Horns",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.CreateToken{
							Amount: game.Fixed(2),
							Source: game.TokenDef(flurryOfHornsToken),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Create two 2/3 red Minotaur creature tokens with haste.
		`,
		},
	}
}

var flurryOfHornsToken = newFlurryOfHornsToken()

func newFlurryOfHornsToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Minotaur",
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Minotaur},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.HasteStaticBody,
			},
		},
	}
}
