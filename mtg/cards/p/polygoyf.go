package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Polygoyf is the card definition for Polygoyf.
//
// Type: Creature — Lhurgoyf
// Cost: {2}{G}
//
// Oracle text:
//
//	Trample, myriad (Whenever this creature attacks, for each opponent other than defending player, you may create a token copy that's tapped and attacking that player or a planeswalker they control. Exile the tokens at end of combat.)
//	Polygoyf's power is equal to the number of card types among cards in all graveyards and its toughness is equal to that number plus 1.
var Polygoyf = newPolygoyf

func newPolygoyf() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Polygoyf",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
			}),
			Colors:           []color.Color{color.Green},
			Types:            []types.Card{types.Creature},
			Subtypes:         []types.Sub{types.Lhurgoyf},
			Power:            opt.Val(game.PT{IsStar: true}),
			Toughness:        opt.Val(game.PT{IsStar: true}),
			DynamicPower:     opt.Val(game.DynamicValue{Kind: game.DynamicValueCardTypesAmongAllGraveyards}),
			DynamicToughness: opt.Val(game.DynamicValue{Kind: game.DynamicValueCardTypesAmongAllGraveyards, Offset: 1}),
			StaticAbilities: []game.StaticAbility{
				game.TrampleStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.MyriadTriggeredBody,
			},
			OracleText: `
			Trample, myriad (Whenever this creature attacks, for each opponent other than defending player, you may create a token copy that's tapped and attacking that player or a planeswalker they control. Exile the tokens at end of combat.)
			Polygoyf's power is equal to the number of card types among cards in all graveyards and its toughness is equal to that number plus 1.
		`,
		},
	}
}
