package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// The Hollow Sentinel
//
// Type: Token Legendary Artifact Creature — Phyrexian Golem
//
// Oracle text:

// TheHollowSentinelTokenf74ad52e364d4c91983135851301585c is the card definition for The Hollow Sentinel.
var TheHollowSentinelTokenf74ad52e364d4c91983135851301585c = &game.CardDef{
	CardFace: game.CardFace{
		Name:       "The Hollow Sentinel",
		Supertypes: []types.Super{types.Legendary},
		Types:      []types.Card{types.Artifact, types.Creature},
		Subtypes:   []types.Sub{types.Phyrexian, types.Golem},
		Power:      opt.Val(game.PT{Value: 3}),
		Toughness:  opt.Val(game.PT{Value: 3}),
	},
}
