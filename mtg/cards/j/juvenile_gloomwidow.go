package j

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// JuvenileGloomwidow is the card definition for Juvenile Gloomwidow.
//
// Type: Creature — Spider
// Cost: {G}{G}
//
// Oracle text:
//
//	Reach (This creature can block creatures with flying.)
//	Wither (This deals damage to creatures in the form of -1/-1 counters.)
var JuvenileGloomwidow = newJuvenileGloomwidow

func newJuvenileGloomwidow() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Juvenile Gloomwidow",
			ManaCost: opt.Val(cost.Mana{
				cost.G,
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Spider},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.ReachStaticBody,
				game.WitherStaticBody,
			},
			OracleText: `
			Reach (This creature can block creatures with flying.)
			Wither (This deals damage to creatures in the form of -1/-1 counters.)
		`,
		},
	}
}
