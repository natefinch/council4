package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Dalek
//
// Type: Token Artifact Creature — Dalek
//
// Oracle text:
//   Menace

// DalekToken4bbce60c74374fc490bb0ead3cccb024 is the card definition for Dalek.
var DalekToken4bbce60c74374fc490bb0ead3cccb024 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Black),
	CardFace: game.CardFace{
		Name:      "Dalek",
		Colors:    []color.Color{color.Black},
		Types:     []types.Card{types.Artifact, types.Creature},
		Subtypes:  []types.Sub{types.Dalek},
		Power:     opt.Val(game.PT{Value: 3}),
		Toughness: opt.Val(game.PT{Value: 3}),
		StaticAbilities: []game.StaticAbility{
			game.MenaceStaticBody,
		},
		OracleText: `
			Menace
		`,
	},
}
