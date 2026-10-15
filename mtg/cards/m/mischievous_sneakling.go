package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MischievousSneakling is the card definition for Mischievous Sneakling.
//
// Type: Creature — Shapeshifter
// Cost: {1}{U/B}
//
// Oracle text:
//
//	Changeling (This card is every creature type.)
//	Flash
var MischievousSneakling = newMischievousSneakling

func newMischievousSneakling() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Black),
		CardFace: game.CardFace{
			Name: "Mischievous Sneakling",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.HybridMana(mana.U, mana.B),
			}),
			Colors:    []color.Color{color.Black, color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Shapeshifter},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.ChangelingStaticBody,
				game.FlashStaticBody,
			},
			OracleText: `
			Changeling (This card is every creature type.)
			Flash
		`,
		},
	}
}
