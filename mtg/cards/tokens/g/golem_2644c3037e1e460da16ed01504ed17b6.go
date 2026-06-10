package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Golem
//
// Type: Token Artifact Creature — Golem
//
// Oracle text:
//   Flying

// GolemToken2644c3037e1e460da16ed01504ed17b6 is the card definition for Golem.
var GolemToken2644c3037e1e460da16ed01504ed17b6 = &game.CardDef{
	CardFace: game.CardFace{
		Name:      "Golem",
		Types:     []types.Card{types.Artifact, types.Creature},
		Subtypes:  []types.Sub{types.Golem},
		Power:     opt.Val(game.PT{Value: 3}),
		Toughness: opt.Val(game.PT{Value: 3}),
		StaticAbilities: []game.StaticAbility{
			game.FlyingStaticBody,
		},
		OracleText: `
			Flying
		`,
	},
}
