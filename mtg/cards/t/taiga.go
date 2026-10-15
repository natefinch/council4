package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// Taiga is the card definition for Taiga.
//
// Type: Land — Mountain Forest
//
// Oracle text:
//
//	({T}: Add {R} or {G}.)
var Taiga = newTaiga

func newTaiga() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red, color.Green),
		CardFace: game.CardFace{
			Name:     "Taiga",
			Types:    []types.Card{types.Land},
			Subtypes: []types.Sub{types.Mountain, types.Forest},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.R, mana.G),
			},
			OracleText: `
			({T}: Add {R} or {G}.)
		`,
		},
	}
}
