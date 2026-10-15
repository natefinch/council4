package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SickleRipper is the card definition for Sickle Ripper.
//
// Type: Creature — Elemental Warrior
// Cost: {1}{B}
//
// Oracle text:
//
//	Wither (This deals damage to creatures in the form of -1/-1 counters.)
var SickleRipper = newSickleRipper

func newSickleRipper() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Sickle Ripper",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elemental, types.Warrior},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.WitherStaticBody,
			},
			OracleText: `
			Wither (This deals damage to creatures in the form of -1/-1 counters.)
		`,
		},
	}
}
