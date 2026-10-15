package o

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// OpalineUnicorn is the card definition for Opaline Unicorn.
//
// Type: Artifact Creature — Unicorn
// Cost: {3}
//
// Oracle text:
//
//	{T}: Add one mana of any color.
var OpalineUnicorn = newOpalineUnicorn

func newOpalineUnicorn() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Opaline Unicorn",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Unicorn},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 2}),
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.W, mana.U, mana.B, mana.R, mana.G),
			},
			OracleText: `
			{T}: Add one mana of any color.
		`,
		},
	}
}
