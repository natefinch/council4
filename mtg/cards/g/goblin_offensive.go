package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GoblinOffensive is the card definition for Goblin Offensive.
//
// Type: Sorcery
// Cost: {X}{1}{R}{R}
//
// Oracle text:
//
//	Create X 1/1 red Goblin creature tokens.
var GoblinOffensive = newGoblinOffensive

func newGoblinOffensive() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Goblin Offensive",
			ManaCost: opt.Val(cost.Mana{
				cost.X,
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
							Amount: game.Dynamic(game.DynamicAmount{
								Kind: game.DynamicAmountX,
							}),
							Source: game.TokenDef(goblinOffensiveToken),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Create X 1/1 red Goblin creature tokens.
		`,
		},
	}
}

var goblinOffensiveToken = newGoblinOffensiveToken()

func newGoblinOffensiveToken() *game.CardDef {
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
