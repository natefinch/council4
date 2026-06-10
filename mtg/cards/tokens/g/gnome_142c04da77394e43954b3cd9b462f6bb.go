package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Gnome
//
// Type: Token Artifact Creature — Gnome
//
// Oracle text:

// GnomeToken142c04da77394e43954b3cd9b462f6bb is the card definition for Gnome.
var GnomeToken142c04da77394e43954b3cd9b462f6bb = &game.CardDef{
	CardFace: game.CardFace{
		Name:      "Gnome",
		Types:     []types.Card{types.Artifact, types.Creature},
		Subtypes:  []types.Sub{types.Gnome},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	},
}
