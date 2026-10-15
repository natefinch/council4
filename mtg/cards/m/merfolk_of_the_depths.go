package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MerfolkOfTheDepths is the card definition for Merfolk of the Depths.
//
// Type: Creature — Merfolk Soldier
// Cost: {4}{G/U}{G/U}
//
// Oracle text:
//
//	Flash (You may cast this spell any time you could cast an instant.)
var MerfolkOfTheDepths = newMerfolkOfTheDepths

func newMerfolkOfTheDepths() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Green),
		CardFace: game.CardFace{
			Name: "Merfolk of the Depths",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.HybridMana(mana.G, mana.U),
				cost.HybridMana(mana.G, mana.U),
			}),
			Colors:    []color.Color{color.Green, color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Merfolk, types.Soldier},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.FlashStaticBody,
			},
			OracleText: `
			Flash (You may cast this spell any time you could cast an instant.)
		`,
		},
	}
}
