package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// InescapableBrute is the card definition for Inescapable Brute.
//
// Type: Creature — Giant Warrior
// Cost: {5}{R}
//
// Oracle text:
//
//	Wither (This deals damage to creatures in the form of -1/-1 counters.)
//	This creature must be blocked if able.
var InescapableBrute = newInescapableBrute

func newInescapableBrute() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Inescapable Brute",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Giant, types.Warrior},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.WitherStaticBody,
				game.MustBeBlockedStaticBody,
			},
			OracleText: `
			Wither (This deals damage to creatures in the form of -1/-1 counters.)
			This creature must be blocked if able.
		`,
		},
	}
}
