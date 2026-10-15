package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// Bayou is the card definition for Bayou.
//
// Type: Land — Swamp Forest
//
// Oracle text:
//
//	({T}: Add {B} or {G}.)
var Bayou = newBayou

func newBayou() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Green),
		CardFace: game.CardFace{
			Name:     "Bayou",
			Types:    []types.Card{types.Land},
			Subtypes: []types.Sub{types.Swamp, types.Forest},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.B, mana.G),
			},
			OracleText: `
			({T}: Add {B} or {G}.)
		`,
		},
	}
}
