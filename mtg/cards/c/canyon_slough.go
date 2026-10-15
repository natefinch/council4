package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// CanyonSlough is the card definition for Canyon Slough.
//
// Type: Land — Swamp Mountain
//
// Oracle text:
//
//	({T}: Add {B} or {R}.)
//	This land enters tapped.
//	Cycling {2} ({2}, Discard this card: Draw a card.)
var CanyonSlough = newCanyonSlough

func newCanyonSlough() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Red),
		CardFace: game.CardFace{
			Name:     "Canyon Slough",
			Types:    []types.Card{types.Land},
			Subtypes: []types.Sub{types.Swamp, types.Mountain},
			ActivatedAbilities: []game.ActivatedAbility{
				game.CyclingActivatedAbility(cost.Mana{cost.O(2)}),
			},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.B, mana.R),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedReplacement("This land enters tapped."),
			},
			OracleText: `
			({T}: Add {B} or {R}.)
			This land enters tapped.
			Cycling {2} ({2}, Discard this card: Draw a card.)
		`,
		},
	}
}
