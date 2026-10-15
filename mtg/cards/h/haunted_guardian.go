package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// HauntedGuardian is the card definition for Haunted Guardian.
//
// Type: Artifact Creature — Construct
// Cost: {2}
//
// Oracle text:
//
//	Defender, first strike
var HauntedGuardian = newHauntedGuardian

func newHauntedGuardian() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Haunted Guardian",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Construct},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.DefenderStaticBody,
				game.FirstStrikeStaticBody,
			},
			OracleText: `
			Defender, first strike
		`,
		},
	}
}
