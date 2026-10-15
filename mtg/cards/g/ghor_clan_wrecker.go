package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GhorClanWrecker is the card definition for Ghor-Clan Wrecker.
//
// Type: Creature — Human Warrior
// Cost: {3}{R}
//
// Oracle text:
//
//	Riot (This creature enters with your choice of a +1/+1 counter or haste.)
//	Menace (This creature can't be blocked except by two or more creatures.)
var GhorClanWrecker = newGhorClanWrecker

func newGhorClanWrecker() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Ghor-Clan Wrecker",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Warrior},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.RiotStaticBody,
				game.MenaceStaticBody,
			},
			OracleText: `
			Riot (This creature enters with your choice of a +1/+1 counter or haste.)
			Menace (This creature can't be blocked except by two or more creatures.)
		`,
		},
	}
}
