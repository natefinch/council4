package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Spellbook is the card definition for Spellbook.
//
// Type: Artifact — Book
// Cost: {0}
//
// Oracle text:
//
//	You have no maximum hand size.
var Spellbook = newSpellbook

func newSpellbook() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Spellbook",
			ManaCost: opt.Val(cost.Mana{
				cost.O(0),
			}),
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Book},
			StaticAbilities: []game.StaticAbility{
				game.NoMaximumHandSizeStaticBody,
			},
			OracleText: `
			You have no maximum hand size.
		`,
		},
	}
}
