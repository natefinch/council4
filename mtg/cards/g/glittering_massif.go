package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// GlitteringMassif is the card definition for Glittering Massif.
//
// Type: Land — Mountain Plains
//
// Oracle text:
//
//	({T}: Add {R} or {W}.)
//	This land enters tapped.
//	Cycling {2} ({2}, Discard this card: Draw a card.)
var GlitteringMassif = newGlitteringMassif

func newGlitteringMassif() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Red),
		CardFace: game.CardFace{
			Name:     "Glittering Massif",
			Types:    []types.Card{types.Land},
			Subtypes: []types.Sub{types.Mountain, types.Plains},
			ActivatedAbilities: []game.ActivatedAbility{
				game.CyclingActivatedAbility(cost.Mana{cost.O(2)}),
			},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.R, mana.W),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedReplacement("This land enters tapped."),
			},
			OracleText: `
			({T}: Add {R} or {W}.)
			This land enters tapped.
			Cycling {2} ({2}, Discard this card: Draw a card.)
		`,
		},
	}
}
