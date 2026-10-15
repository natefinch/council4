package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AlhammarretSArchive is the card definition for Alhammarret's Archive.
//
// Type: Legendary Artifact
// Cost: {5}
//
// Oracle text:
//
//	If you would gain life, you gain twice that much life instead.
//	If you would draw a card except the first one you draw in each of your draw steps, draw two cards instead.
var AlhammarretSArchive = newAlhammarretSArchive

func newAlhammarretSArchive() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Alhammarret's Archive",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
			}),
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Artifact},
			ReplacementAbilities: []game.ReplacementAbility{
				game.LifeGainReplacement("If you would gain life, you gain twice that much life instead.", 2, 0),
				game.DrawCardMultiplierReplacement("If you would draw a card except the first one you draw in each of your draw steps, draw two cards instead.", 2, true),
			},
			OracleText: `
			If you would gain life, you gain twice that much life instead.
			If you would draw a card except the first one you draw in each of your draw steps, draw two cards instead.
		`,
		},
	}
}
