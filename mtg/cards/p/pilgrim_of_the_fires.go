package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// PilgrimOfTheFires is the card definition for Pilgrim of the Fires.
//
// Type: Artifact Creature — Golem
// Cost: {7}
//
// Oracle text:
//
//	First strike, trample
var PilgrimOfTheFires = newPilgrimOfTheFires

func newPilgrimOfTheFires() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Pilgrim of the Fires",
			ManaCost: opt.Val(cost.Mana{
				cost.O(7),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Golem},
			Power:     opt.Val(game.PT{Value: 6}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.FirstStrikeStaticBody,
				game.TrampleStaticBody,
			},
			OracleText: `
			First strike, trample
		`,
		},
	}
}
