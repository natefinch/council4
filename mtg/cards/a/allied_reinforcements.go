package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AlliedReinforcements is the card definition for Allied Reinforcements.
//
// Type: Sorcery
// Cost: {3}{W}
//
// Oracle text:
//
//	Create two 2/2 white Knight Ally creature tokens.
var AlliedReinforcements = newAlliedReinforcements

func newAlliedReinforcements() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Allied Reinforcements",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.CreateToken{
							Amount: game.Fixed(2),
							Source: game.TokenDef(alliedReinforcementsToken),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Create two 2/2 white Knight Ally creature tokens.
		`,
		},
	}
}

var alliedReinforcementsToken = newAlliedReinforcementsToken()

func newAlliedReinforcementsToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Knight Ally",
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Knight, types.Ally},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
		},
	}
}
