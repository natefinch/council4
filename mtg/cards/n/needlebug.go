package n

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Needlebug is the card definition for Needlebug.
//
// Type: Artifact Creature — Insect
// Cost: {4}
//
// Oracle text:
//
//	Flash
//	Protection from artifacts
var Needlebug = newNeedlebug

func newNeedlebug() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Needlebug",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Insect},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.FlashStaticBody,
				game.ProtectionFromTypesStaticAbility(types.Artifact),
			},
			OracleText: `
			Flash
			Protection from artifacts
		`,
		},
	}
}
