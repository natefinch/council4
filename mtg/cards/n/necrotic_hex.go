package n

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// NecroticHex is the card definition for Necrotic Hex.
//
// Type: Sorcery
// Cost: {6}{B}
//
// Oracle text:
//
//	Each player sacrifices six creatures of their choice. You create six tapped 2/2 black Zombie creature tokens.
var NecroticHex = newNecroticHex

func newNecroticHex() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Necrotic Hex",
			ManaCost: opt.Val(cost.Mana{
				cost.O(6),
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.SacrificePermanents{
							PlayerGroup: game.AllPlayersReference(),
							Amount:      game.Fixed(6),
							Selection:   game.Selection{RequiredTypes: []types.Card{types.Creature}},
						},
					},
					{
						Primitive: game.CreateToken{
							Amount:      game.Fixed(6),
							Source:      game.TokenDef(necroticHexToken),
							EntryTapped: true,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Each player sacrifices six creatures of their choice. You create six tapped 2/2 black Zombie creature tokens.
		`,
		},
	}
}

var necroticHexToken = newNecroticHexToken()

func newNecroticHexToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Zombie",
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Zombie},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
		},
	}
}
