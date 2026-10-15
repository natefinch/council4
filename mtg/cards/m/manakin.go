package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Manakin is the card definition for Manakin.
//
// Type: Artifact Creature — Construct
// Cost: {2}
//
// Oracle text:
//
//	{T}: Add {C}.
var Manakin = newManakin

func newManakin() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Manakin",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Construct},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.C),
			},
			OracleText: `
			{T}: Add {C}.
		`,
		},
	}
}
