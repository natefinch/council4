package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// HearthfireHobgoblin is the card definition for Hearthfire Hobgoblin.
//
// Type: Creature — Goblin Soldier
// Cost: {R/W}{R/W}{R/W}
//
// Oracle text:
//
//	Double strike
var HearthfireHobgoblin = newHearthfireHobgoblin

func newHearthfireHobgoblin() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Red),
		CardFace: game.CardFace{
			Name: "Hearthfire Hobgoblin",
			ManaCost: opt.Val(cost.Mana{
				cost.HybridMana(mana.R, mana.W),
				cost.HybridMana(mana.R, mana.W),
				cost.HybridMana(mana.R, mana.W),
			}),
			Colors:    []color.Color{color.Red, color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Goblin, types.Soldier},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.DoubleStrikeStaticBody,
			},
			OracleText: `
			Double strike
		`,
		},
	}
}
