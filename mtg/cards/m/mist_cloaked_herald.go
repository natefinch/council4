package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MistCloakedHerald is the card definition for Mist-Cloaked Herald.
//
// Type: Creature — Merfolk Warrior
// Cost: {U}
//
// Oracle text:
//
//	This creature can't be blocked.
var MistCloakedHerald = newMistCloakedHerald

func newMistCloakedHerald() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Mist-Cloaked Herald",
			ManaCost: opt.Val(cost.Mana{
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Merfolk, types.Warrior},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.CantBeBlockedStaticBody,
			},
			OracleText: `
			This creature can't be blocked.
		`,
		},
	}
}
