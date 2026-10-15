package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MonsterMashup is the card definition for Monster Mashup.
//
// Type: Creature — Werewolf Fish Zombie Vampire
// Cost: {3}{R}
//
// Oracle text:
//
//	Reach (This creature can block creatures with flying.)
//	Menace (This creature can't be blocked except by two or more creatures.)
var MonsterMashup = newMonsterMashup

func newMonsterMashup() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Monster Mashup",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Werewolf, types.Fish, types.Zombie, types.Vampire},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.ReachStaticBody,
				game.MenaceStaticBody,
			},
			OracleText: `
			Reach (This creature can block creatures with flying.)
			Menace (This creature can't be blocked except by two or more creatures.)
		`,
		},
	}
}
