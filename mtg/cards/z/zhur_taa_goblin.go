package z

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ZhurTaaGoblin is the card definition for Zhur-Taa Goblin.
//
// Type: Creature — Goblin Berserker
// Cost: {R}{G}
//
// Oracle text:
//
//	Riot (This creature enters with your choice of a +1/+1 counter or haste.)
var ZhurTaaGoblin = newZhurTaaGoblin

func newZhurTaaGoblin() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red, color.Green),
		CardFace: game.CardFace{
			Name: "Zhur-Taa Goblin",
			ManaCost: opt.Val(cost.Mana{
				cost.R,
				cost.G,
			}),
			Colors:    []color.Color{color.Green, color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Goblin, types.Berserker},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.RiotStaticBody,
			},
			OracleText: `
			Riot (This creature enters with your choice of a +1/+1 counter or haste.)
		`,
		},
	}
}
