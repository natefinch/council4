package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// VenserSSliver is the card definition for Venser's Sliver.
//
// Type: Artifact Creature — Sliver
// Cost: {5}
//
// Oracle text:
var VenserSSliver = newVenserSSliver

func newVenserSSliver() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Venser's Sliver",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Sliver},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
		},
	}
}
