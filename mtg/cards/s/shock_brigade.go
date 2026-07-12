package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ShockBrigade is the card definition for Shock Brigade.
//
// Type: Creature — Goblin Soldier
// Cost: {1}{R}
//
// Oracle text:
//
//	Menace (This creature can't be blocked except by two or more creatures.)
//	Mobilize 1 (Whenever this creature attacks, create a tapped and attacking 1/1 red Warrior creature token. Sacrifice it at the beginning of the next end step.)
var ShockBrigade = newShockBrigade

func newShockBrigade() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Shock Brigade",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Goblin, types.Soldier},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.MenaceStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.MobilizeTriggeredBody(game.MobilizeAmount{Fixed: 1}),
			},
			OracleText: `
			Menace (This creature can't be blocked except by two or more creatures.)
			Mobilize 1 (Whenever this creature attacks, create a tapped and attacking 1/1 red Warrior creature token. Sacrifice it at the beginning of the next end step.)
		`,
		},
	}
}
