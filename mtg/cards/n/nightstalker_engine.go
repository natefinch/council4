package n

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// NightstalkerEngine is the card definition for Nightstalker Engine.
//
// Type: Creature — Nightstalker
// Cost: {4}{B}
//
// Oracle text:
//
//	Nightstalker Engine's power is equal to the number of creature cards in your graveyard.
var NightstalkerEngine = newNightstalkerEngine

func newNightstalkerEngine() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Nightstalker Engine",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.B,
			}),
			Colors:       []color.Color{color.Black},
			Types:        []types.Card{types.Creature},
			Subtypes:     []types.Sub{types.Nightstalker},
			Power:        opt.Val(game.PT{IsStar: true}),
			Toughness:    opt.Val(game.PT{Value: 3}),
			DynamicPower: opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerCreatureCardsInGraveyard}),
			OracleText: `
			Nightstalker Engine's power is equal to the number of creature cards in your graveyard.
		`,
		},
	}
}
