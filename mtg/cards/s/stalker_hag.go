package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// StalkerHag is the card definition for Stalker Hag.
//
// Type: Creature — Hag
// Cost: {B/G}{B/G}{B/G}
//
// Oracle text:
//
//	Swampwalk, forestwalk (This creature can't be blocked as long as defending player controls a Swamp or a Forest.)
var StalkerHag = newStalkerHag

func newStalkerHag() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Green),
		CardFace: game.CardFace{
			Name: "Stalker Hag",
			ManaCost: opt.Val(cost.Mana{
				cost.HybridMana(mana.B, mana.G),
				cost.HybridMana(mana.B, mana.G),
				cost.HybridMana(mana.B, mana.G),
			}),
			Colors:    []color.Color{color.Black, color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Hag},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.SwampwalkStaticBody,
				game.ForestwalkStaticBody,
			},
			OracleText: `
			Swampwalk, forestwalk (This creature can't be blocked as long as defending player controls a Swamp or a Forest.)
		`,
		},
	}
}
