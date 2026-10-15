package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GuardiansOfMeletis is the card definition for Guardians of Meletis.
//
// Type: Artifact Creature — Golem
// Cost: {3}
//
// Oracle text:
//
//	Defender (This creature can't attack.)
var GuardiansOfMeletis = newGuardiansOfMeletis

func newGuardiansOfMeletis() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Guardians of Meletis",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Golem},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 6}),
			StaticAbilities: []game.StaticAbility{
				game.DefenderStaticBody,
			},
			OracleText: `
			Defender (This creature can't attack.)
		`,
		},
	}
}
