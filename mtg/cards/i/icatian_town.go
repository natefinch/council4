package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// IcatianTown is the card definition for Icatian Town.
//
// Type: Sorcery
// Cost: {5}{W}
//
// Oracle text:
//
//	Create four 1/1 white Citizen creature tokens.
var IcatianTown = newIcatianTown

func newIcatianTown() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Icatian Town",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.CreateToken{
							Amount: game.Fixed(4),
							Source: game.TokenDef(icatianTownToken),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Create four 1/1 white Citizen creature tokens.
		`,
		},
	}
}

var icatianTownToken = newIcatianTownToken()

func newIcatianTownToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Citizen",
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Citizen},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
