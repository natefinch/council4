package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// HordelingOutburst is the card definition for Hordeling Outburst.
//
// Type: Sorcery
// Cost: {1}{R}{R}
//
// Oracle text:
//
//	Create three 1/1 red Goblin creature tokens.
var HordelingOutburst = newHordelingOutburst

func newHordelingOutburst() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Hordeling Outburst",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.CreateToken{
							Amount: game.Fixed(3),
							Source: game.TokenDef(hordelingOutburstToken),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Create three 1/1 red Goblin creature tokens.
		`,
		},
	}
}

var hordelingOutburstToken = newHordelingOutburstToken()

func newHordelingOutburstToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Goblin",
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Goblin},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
