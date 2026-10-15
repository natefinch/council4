package o

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ObsianusGolem is the card definition for Obsianus Golem.
//
// Type: Artifact Creature — Golem
// Cost: {6}
//
// Oracle text:
var ObsianusGolem = newObsianusGolem

func newObsianusGolem() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Obsianus Golem",
			ManaCost: opt.Val(cost.Mana{
				cost.O(6),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Golem},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 6}),
		},
	}
}
