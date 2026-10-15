package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// SparaSHeadquarters is the card definition for Spara's Headquarters.
//
// Type: Land — Forest Plains Island
//
// Oracle text:
//
//	({T}: Add {G}, {W}, or {U}.)
//	This land enters tapped.
//	Cycling {3} ({3}, Discard this card: Draw a card.)
var SparaSHeadquarters = newSparaSHeadquarters

func newSparaSHeadquarters() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue, color.Green),
		CardFace: game.CardFace{
			Name:     "Spara's Headquarters",
			Types:    []types.Card{types.Land},
			Subtypes: []types.Sub{types.Forest, types.Plains, types.Island},
			ActivatedAbilities: []game.ActivatedAbility{
				game.CyclingActivatedAbility(cost.Mana{cost.O(3)}),
			},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.G, mana.W, mana.U),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedReplacement("This land enters tapped."),
			},
			OracleText: `
			({T}: Add {G}, {W}, or {U}.)
			This land enters tapped.
			Cycling {3} ({3}, Discard this card: Draw a card.)
		`,
		},
	}
}
