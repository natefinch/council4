package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ShadowSummoning is the card definition for Shadow Summoning.
//
// Type: Sorcery
// Cost: {W}{B}
//
// Oracle text:
//
//	Create two tapped 1/1 white Spirit creature tokens with flying.
var ShadowSummoning = newShadowSummoning

func newShadowSummoning() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Black),
		CardFace: game.CardFace{
			Name: "Shadow Summoning",
			ManaCost: opt.Val(cost.Mana{
				cost.W,
				cost.B,
			}),
			Colors: []color.Color{color.Black, color.White},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.CreateToken{
							Amount:      game.Fixed(2),
							Source:      game.TokenDef(shadowSummoningToken),
							EntryTapped: true,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Create two tapped 1/1 white Spirit creature tokens with flying.
		`,
		},
	}
}

var shadowSummoningToken = newShadowSummoningToken()

func newShadowSummoningToken() *game.CardDef {
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
