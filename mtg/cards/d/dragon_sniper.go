package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DragonSniper is the card definition for Dragon Sniper.
//
// Type: Creature — Human Archer
// Cost: {G}
//
// Oracle text:
//
//	Reach, vigilance, deathtouch
var DragonSniper = newDragonSniper

func newDragonSniper() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Dragon Sniper",
			ManaCost: opt.Val(cost.Mana{
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Archer},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.ReachStaticBody,
				game.VigilanceStaticBody,
				game.DeathtouchStaticBody,
			},
			OracleText: `
			Reach, vigilance, deathtouch
		`,
		},
	}
}
