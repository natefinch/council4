package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Shambleshark is the card definition for Shambleshark.
//
// Type: Creature — Shark Crab
// Cost: {G}{U}
//
// Oracle text:
//
//	Flash (You may cast this spell any time you could cast an instant.)
//	Evolve (Whenever a creature you control enters, if that creature has greater power or toughness than this creature, put a +1/+1 counter on this creature.)
var Shambleshark = newShambleshark

func newShambleshark() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Green),
		CardFace: game.CardFace{
			Name: "Shambleshark",
			ManaCost: opt.Val(cost.Mana{
				cost.G,
				cost.U,
			}),
			Colors:    []color.Color{color.Green, color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Shark, types.Crab},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.FlashStaticBody,
				game.EvolveStaticBody,
			},
			OracleText: `
			Flash (You may cast this spell any time you could cast an instant.)
			Evolve (Whenever a creature you control enters, if that creature has greater power or toughness than this creature, put a +1/+1 counter on this creature.)
		`,
		},
	}
}
