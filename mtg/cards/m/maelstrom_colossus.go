package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MaelstromColossus is the card definition for Maelstrom Colossus.
//
// Type: Artifact Creature — Golem
// Cost: {8}
//
// Oracle text:
//
//	Cascade (When you cast this spell, exile cards from the top of your library until you exile a nonland card that costs less. You may cast it without paying its mana cost. Put the exiled cards on the bottom in a random order.)
var MaelstromColossus = newMaelstromColossus

func newMaelstromColossus() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Maelstrom Colossus",
			ManaCost: opt.Val(cost.Mana{
				cost.O(8),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Golem},
			Power:     opt.Val(game.PT{Value: 7}),
			Toughness: opt.Val(game.PT{Value: 7}),
			StaticAbilities: []game.StaticAbility{
				game.CascadeStaticBody,
			},
			OracleText: `
			Cascade (When you cast this spell, exile cards from the top of your library until you exile a nonland card that costs less. You may cast it without paying its mana cost. Put the exiled cards on the bottom in a random order.)
		`,
		},
	}
}
