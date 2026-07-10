package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BoggartRAMGang is the card definition for Boggart Ram-Gang.
//
// Type: Creature — Goblin Warrior
// Cost: {R/G}{R/G}{R/G}
//
// Oracle text:
//
//	Haste
//	Wither (This deals damage to creatures in the form of -1/-1 counters.)
var BoggartRAMGang = newBoggartRAMGang

func newBoggartRAMGang() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red, color.Green),
		CardFace: game.CardFace{
			Name: "Boggart Ram-Gang",
			ManaCost: opt.Val(cost.Mana{
				cost.HybridMana(mana.R, mana.G),
				cost.HybridMana(mana.R, mana.G),
				cost.HybridMana(mana.R, mana.G),
			}),
			Colors:    []color.Color{color.Green, color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Goblin, types.Warrior},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.HasteStaticBody,
				game.WitherStaticBody,
			},
			OracleText: `
			Haste
			Wither (This deals damage to creatures in the form of -1/-1 counters.)
		`,
		},
	}
}
