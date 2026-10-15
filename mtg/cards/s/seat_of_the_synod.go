package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// SeatOfTheSynod is the card definition for Seat of the Synod.
//
// Type: Artifact Land
//
// Oracle text:
//
//	{T}: Add {U}.
var SeatOfTheSynod = newSeatOfTheSynod

func newSeatOfTheSynod() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name:  "Seat of the Synod",
			Types: []types.Card{types.Artifact, types.Land},
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.U),
			},
			OracleText: `
			{T}: Add {U}.
		`,
		},
	}
}
