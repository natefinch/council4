package o

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// OversoulOfDusk is the card definition for Oversoul of Dusk.
//
// Type: Creature — Spirit Avatar
// Cost: {G/W}{G/W}{G/W}{G/W}{G/W}
//
// Oracle text:
//
//	Protection from blue, from black, and from red
var OversoulOfDusk = newOversoulOfDusk

func newOversoulOfDusk() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Green),
		CardFace: game.CardFace{
			Name: "Oversoul of Dusk",
			ManaCost: opt.Val(cost.Mana{
				cost.HybridMana(mana.G, mana.W),
				cost.HybridMana(mana.G, mana.W),
				cost.HybridMana(mana.G, mana.W),
				cost.HybridMana(mana.G, mana.W),
				cost.HybridMana(mana.G, mana.W),
			}),
			Colors:    []color.Color{color.Green, color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Spirit, types.Avatar},
			Power:     opt.Val(game.PT{Value: 5}),
			Toughness: opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.ProtectionFromColorsStaticAbility(color.Blue, color.Black, color.Red),
			},
			OracleText: `
			Protection from blue, from black, and from red
		`,
		},
	}
}
