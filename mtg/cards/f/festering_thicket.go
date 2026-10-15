package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// FesteringThicket is the card definition for Festering Thicket.
//
// Type: Land — Swamp Forest
//
// Oracle text:
//
//	({T}: Add {B} or {G}.)
//	This land enters tapped.
//	Cycling {2} ({2}, Discard this card: Draw a card.)
var FesteringThicket = newFesteringThicket

func newFesteringThicket() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Green),
		CardFace: game.CardFace{
			Name:     "Festering Thicket",
			Types:    []types.Card{types.Land},
			Subtypes: []types.Sub{types.Swamp, types.Forest},
			ActivatedAbilities: []game.ActivatedAbility{
				game.CyclingActivatedAbility(cost.Mana{cost.O(2)}),
			},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.B, mana.G),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedReplacement("This land enters tapped."),
			},
			OracleText: `
			({T}: Add {B} or {G}.)
			This land enters tapped.
			Cycling {2} ({2}, Discard this card: Draw a card.)
		`,
		},
	}
}
