package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// PridefulFeastling is the card definition for Prideful Feastling.
//
// Type: Creature — Shapeshifter
// Cost: {2}{W/B}
//
// Oracle text:
//
//	Changeling (This card is every creature type.)
//	Lifelink
var PridefulFeastling = newPridefulFeastling

func newPridefulFeastling() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Black),
		CardFace: game.CardFace{
			Name: "Prideful Feastling",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.HybridMana(mana.W, mana.B),
			}),
			Colors:    []color.Color{color.Black, color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Shapeshifter},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.ChangelingStaticBody,
				game.LifelinkStaticBody,
			},
			OracleText: `
			Changeling (This card is every creature type.)
			Lifelink
		`,
		},
	}
}
