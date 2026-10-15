package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SpectralProcession is the card definition for Spectral Procession.
//
// Type: Sorcery
// Cost: {2/W}{2/W}{2/W}
//
// Oracle text:
//
//	Create three 1/1 white Spirit creature tokens with flying.
var SpectralProcession = newSpectralProcession

func newSpectralProcession() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Spectral Procession",
			ManaCost: opt.Val(cost.Mana{
				cost.Twobrid(mana.W),
				cost.Twobrid(mana.W),
				cost.Twobrid(mana.W),
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.CreateToken{
							Amount: game.Fixed(3),
							Source: game.TokenDef(spectralProcessionToken),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Create three 1/1 white Spirit creature tokens with flying.
		`,
		},
	}
}

var spectralProcessionToken = newSpectralProcessionToken()

func newSpectralProcessionToken() *game.CardDef {
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
