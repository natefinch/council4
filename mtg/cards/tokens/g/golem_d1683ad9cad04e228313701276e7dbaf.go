package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Golem
//
// Type: Token Enchantment Artifact Creature — Golem
//
// Oracle text:

// GolemTokend1683ad9cad04e228313701276e7dbaf is the card definition for Golem.
var GolemTokend1683ad9cad04e228313701276e7dbaf = &game.CardDef{
	CardFace: game.CardFace{
		Name:      "Golem",
		Types:     []types.Card{types.Enchantment, types.Artifact, types.Creature},
		Subtypes:  []types.Sub{types.Golem},
		Power:     opt.Val(game.PT{Value: 3}),
		Toughness: opt.Val(game.PT{Value: 3}),
	},
}
