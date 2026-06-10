package u

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Urzan Automaton
//
// Type: Token Creature — Urzan Automaton
//
// Oracle text:

// UrzanAutomatonToken8bf396aed9ac4711af60d97f439ab76c is the card definition for Urzan Automaton.
var UrzanAutomatonToken8bf396aed9ac4711af60d97f439ab76c = &game.CardDef{
	CardFace: game.CardFace{
		Name:      "Urzan Automaton",
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Sub("Urzan"), types.Sub("Automaton")},
		Power:     opt.Val(game.PT{IsStar: true}),
		Toughness: opt.Val(game.PT{IsStar: true}),
	},
}
