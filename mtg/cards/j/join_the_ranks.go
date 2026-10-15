package j

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// JoinTheRanks is the card definition for Join the Ranks.
//
// Type: Instant
// Cost: {3}{W}
//
// Oracle text:
//
//	Create two 1/1 white Soldier Ally creature tokens.
var JoinTheRanks = newJoinTheRanks

func newJoinTheRanks() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Join the Ranks",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.CreateToken{
							Amount: game.Fixed(2),
							Source: game.TokenDef(joinTheRanksToken),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Create two 1/1 white Soldier Ally creature tokens.
		`,
		},
	}
}

var joinTheRanksToken = newJoinTheRanksToken()

func newJoinTheRanksToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Soldier Ally",
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Soldier, types.Ally},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
