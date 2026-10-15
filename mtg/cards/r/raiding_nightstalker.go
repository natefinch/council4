package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RaidingNightstalker is the card definition for Raiding Nightstalker.
//
// Type: Creature — Nightstalker
// Cost: {2}{B}
//
// Oracle text:
//
//	Swampwalk (This creature can't be blocked as long as defending player controls a Swamp.)
var RaidingNightstalker = newRaidingNightstalker

func newRaidingNightstalker() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Raiding Nightstalker",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Nightstalker},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.SwampwalkStaticBody,
			},
			OracleText: `
			Swampwalk (This creature can't be blocked as long as defending player controls a Swamp.)
		`,
		},
	}
}
