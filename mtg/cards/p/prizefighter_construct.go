package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// PrizefighterConstruct is the card definition for Prizefighter Construct.
//
// Type: Artifact Creature — Construct
// Cost: {5}
//
// Oracle text:
var PrizefighterConstruct = newPrizefighterConstruct

func newPrizefighterConstruct() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Prizefighter Construct",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Construct},
			Power:     opt.Val(game.PT{Value: 6}),
			Toughness: opt.Val(game.PT{Value: 2}),
		},
	}
}
