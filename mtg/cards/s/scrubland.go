package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// Scrubland is the card definition for Scrubland.
//
// Type: Land — Plains Swamp
//
// Oracle text:
//
//	({T}: Add {W} or {B}.)
var Scrubland = newScrubland

func newScrubland() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Black),
		CardFace: game.CardFace{
			Name:     "Scrubland",
			Types:    []types.Card{types.Land},
			Subtypes: []types.Sub{types.Plains, types.Swamp},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.W, mana.B),
			},
			OracleText: `
			({T}: Add {W} or {B}.)
		`,
		},
	}
}
