package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// FetidPools is the card definition for Fetid Pools.
//
// Type: Land — Island Swamp
//
// Oracle text:
//
//	({T}: Add {U} or {B}.)
//	This land enters tapped.
//	Cycling {2} ({2}, Discard this card: Draw a card.)
var FetidPools = newFetidPools

func newFetidPools() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Black),
		CardFace: game.CardFace{
			Name:     "Fetid Pools",
			Types:    []types.Card{types.Land},
			Subtypes: []types.Sub{types.Island, types.Swamp},
			ActivatedAbilities: []game.ActivatedAbility{
				game.CyclingActivatedAbility(cost.Mana{cost.O(2)}),
			},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.U, mana.B),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedReplacement("This land enters tapped."),
			},
			OracleText: `
			({T}: Add {U} or {B}.)
			This land enters tapped.
			Cycling {2} ({2}, Discard this card: Draw a card.)
		`,
		},
	}
}
