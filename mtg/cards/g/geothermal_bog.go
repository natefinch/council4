package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// GeothermalBog is the card definition for Geothermal Bog.
//
// Type: Land — Swamp Mountain
//
// Oracle text:
//
//	({T}: Add {B} or {R}.)
//	This land enters tapped.
var GeothermalBog = newGeothermalBog

func newGeothermalBog() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Red),
		CardFace: game.CardFace{
			Name:     "Geothermal Bog",
			Types:    []types.Card{types.Land},
			Subtypes: []types.Sub{types.Swamp, types.Mountain},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.B, mana.R),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedReplacement("This land enters tapped."),
			},
			OracleText: `
			({T}: Add {B} or {R}.)
			This land enters tapped.
		`,
		},
	}
}
