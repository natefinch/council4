package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// ShelteredThicket is the card definition for Sheltered Thicket.
//
// Type: Land — Mountain Forest
//
// Oracle text:
//
//	({T}: Add {R} or {G}.)
//	This land enters tapped.
//	Cycling {2} ({2}, Discard this card: Draw a card.)
var ShelteredThicket = newShelteredThicket

func newShelteredThicket() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red, color.Green),
		CardFace: game.CardFace{
			Name:     "Sheltered Thicket",
			Types:    []types.Card{types.Land},
			Subtypes: []types.Sub{types.Mountain, types.Forest},
			ActivatedAbilities: []game.ActivatedAbility{
				game.CyclingActivatedAbility(cost.Mana{cost.O(2)}),
			},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.R, mana.G),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedReplacement("This land enters tapped."),
			},
			OracleText: `
			({T}: Add {R} or {G}.)
			This land enters tapped.
			Cycling {2} ({2}, Discard this card: Draw a card.)
		`,
		},
	}
}
