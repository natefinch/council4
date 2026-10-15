package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// RaffineSTower is the card definition for Raffine's Tower.
//
// Type: Land — Plains Island Swamp
//
// Oracle text:
//
//	({T}: Add {W}, {U}, or {B}.)
//	This land enters tapped.
//	Cycling {3} ({3}, Discard this card: Draw a card.)
var RaffineSTower = newRaffineSTower

func newRaffineSTower() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue, color.Black),
		CardFace: game.CardFace{
			Name:     "Raffine's Tower",
			Types:    []types.Card{types.Land},
			Subtypes: []types.Sub{types.Plains, types.Island, types.Swamp},
			ActivatedAbilities: []game.ActivatedAbility{
				game.CyclingActivatedAbility(cost.Mana{cost.O(3)}),
			},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.W, mana.U, mana.B),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedReplacement("This land enters tapped."),
			},
			OracleText: `
			({T}: Add {W}, {U}, or {B}.)
			This land enters tapped.
			Cycling {3} ({3}, Discard this card: Draw a card.)
		`,
		},
	}
}
