package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// HarvestGwyllion is the card definition for Harvest Gwyllion.
//
// Type: Creature — Hag
// Cost: {2}{W/B}{W/B}
//
// Oracle text:
//
//	Wither (This deals damage to creatures in the form of -1/-1 counters.)
var HarvestGwyllion = newHarvestGwyllion

func newHarvestGwyllion() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Black),
		CardFace: game.CardFace{
			Name: "Harvest Gwyllion",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.HybridMana(mana.W, mana.B),
				cost.HybridMana(mana.W, mana.B),
			}),
			Colors:    []color.Color{color.Black, color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Hag},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.WitherStaticBody,
			},
			OracleText: `
			Wither (This deals damage to creatures in the form of -1/-1 counters.)
		`,
		},
	}
}
