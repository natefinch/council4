package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// PhobianPhantasm is the card definition for Phobian Phantasm.
//
// Type: Creature — Illusion
// Cost: {1}{B}{B}
//
// Oracle text:
//
//	Flying; fear (This creature can't be blocked except by artifact creatures and/or black creatures.)
//	Cumulative upkeep {B} (At the beginning of your upkeep, put an age counter on this permanent, then sacrifice it unless you pay its upkeep cost for each age counter on it.)
var PhobianPhantasm = newPhobianPhantasm

func newPhobianPhantasm() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Phobian Phantasm",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Illusion},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.FearStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.CumulativeUpkeepTriggeredAbility(cost.Mana{cost.B}),
			},
			OracleText: `
			Flying; fear (This creature can't be blocked except by artifact creatures and/or black creatures.)
			Cumulative upkeep {B} (At the beginning of your upkeep, put an age counter on this permanent, then sacrifice it unless you pay its upkeep cost for each age counter on it.)
		`,
		},
	}
}
