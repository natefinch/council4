package k

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// KessigRecluse is the card definition for Kessig Recluse.
//
// Type: Creature — Spider
// Cost: {2}{G}{G}
//
// Oracle text:
//
//	Reach (This creature can block creatures with flying.)
//	Deathtouch (Any amount of damage this deals to a creature is enough to destroy it.)
var KessigRecluse = newKessigRecluse

func newKessigRecluse() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Kessig Recluse",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Spider},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.ReachStaticBody,
				game.DeathtouchStaticBody,
			},
			OracleText: `
			Reach (This creature can block creatures with flying.)
			Deathtouch (Any amount of damage this deals to a creature is enough to destroy it.)
		`,
		},
	}
}
