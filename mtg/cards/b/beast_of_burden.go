package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BeastOfBurden is the card definition for Beast of Burden.
//
// Type: Artifact Creature — Golem
// Cost: {6}
//
// Oracle text:
//
//	Beast of Burden's power and toughness are each equal to the number of creatures on the battlefield.
var BeastOfBurden = newBeastOfBurden

func newBeastOfBurden() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Beast of Burden",
			ManaCost: opt.Val(cost.Mana{
				cost.O(6),
			}),
			Types:            []types.Card{types.Artifact, types.Creature},
			Subtypes:         []types.Sub{types.Golem},
			Power:            opt.Val(game.PT{IsStar: true}),
			Toughness:        opt.Val(game.PT{IsStar: true}),
			DynamicPower:     opt.Val(game.DynamicValue{Kind: game.DynamicValueAllBattlefieldCreatureCount}),
			DynamicToughness: opt.Val(game.DynamicValue{Kind: game.DynamicValueAllBattlefieldCreatureCount}),
			OracleText: `
			Beast of Burden's power and toughness are each equal to the number of creatures on the battlefield.
		`,
		},
	}
}
