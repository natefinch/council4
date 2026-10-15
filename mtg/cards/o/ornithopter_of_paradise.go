package o

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// OrnithopterOfParadise is the card definition for Ornithopter of Paradise.
//
// Type: Artifact Creature — Thopter
// Cost: {2}
//
// Oracle text:
//
//	Flying
//	{T}: Add one mana of any color.
var OrnithopterOfParadise = newOrnithopterOfParadise

func newOrnithopterOfParadise() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Ornithopter of Paradise",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Thopter},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.W, mana.U, mana.B, mana.R, mana.G),
			},
			OracleText: `
			Flying
			{T}: Add one mana of any color.
		`,
		},
	}
}
