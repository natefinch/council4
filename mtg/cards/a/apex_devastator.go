package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ApexDevastator is the card definition for Apex Devastator.
//
// Type: Creature — Chimera Hydra
// Cost: {8}{G}{G}
//
// Oracle text:
//
//	Cascade, cascade, cascade, cascade (When you cast this spell, exile cards from the top of your library until you exile a nonland card that costs less. You may cast it without paying its mana cost. Put the exiled cards on the bottom in a random order. Multiple instances of cascade each trigger separately.)
var ApexDevastator = newApexDevastator

func newApexDevastator() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Apex Devastator",
			ManaCost: opt.Val(cost.Mana{
				cost.O(8),
				cost.G,
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Chimera, types.Hydra},
			Power:     opt.Val(game.PT{Value: 10}),
			Toughness: opt.Val(game.PT{Value: 10}),
			StaticAbilities: []game.StaticAbility{
				game.CascadeStaticBody,
				game.CascadeStaticBody,
				game.CascadeStaticBody,
				game.CascadeStaticBody,
			},
			OracleText: `
			Cascade, cascade, cascade, cascade (When you cast this spell, exile cards from the top of your library until you exile a nonland card that costs less. You may cast it without paying its mana cost. Put the exiled cards on the bottom in a random order. Multiple instances of cascade each trigger separately.)
		`,
		},
	}
}
