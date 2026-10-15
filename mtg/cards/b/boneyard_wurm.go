package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BoneyardWurm is the card definition for Boneyard Wurm.
//
// Type: Creature — Wurm
// Cost: {1}{G}
//
// Oracle text:
//
//	Boneyard Wurm's power and toughness are each equal to the number of creature cards in your graveyard.
var BoneyardWurm = newBoneyardWurm

func newBoneyardWurm() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Boneyard Wurm",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
			}),
			Colors:           []color.Color{color.Green},
			Types:            []types.Card{types.Creature},
			Subtypes:         []types.Sub{types.Wurm},
			Power:            opt.Val(game.PT{IsStar: true}),
			Toughness:        opt.Val(game.PT{IsStar: true}),
			DynamicPower:     opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerCreatureCardsInGraveyard}),
			DynamicToughness: opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerCreatureCardsInGraveyard}),
			OracleText: `
			Boneyard Wurm's power and toughness are each equal to the number of creature cards in your graveyard.
		`,
		},
	}
}
