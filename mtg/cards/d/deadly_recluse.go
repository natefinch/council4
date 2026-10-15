package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DeadlyRecluse is the card definition for Deadly Recluse.
//
// Type: Creature — Spider
// Cost: {1}{G}
//
// Oracle text:
//
//	Reach (This creature can block creatures with flying.)
//	Deathtouch (Any amount of damage this deals to a creature is enough to destroy it.)
var DeadlyRecluse = newDeadlyRecluse

func newDeadlyRecluse() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Deadly Recluse",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Spider},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 2}),
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
