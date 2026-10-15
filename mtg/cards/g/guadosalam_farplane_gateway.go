package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// GuadosalamFarplaneGateway is the card definition for Guadosalam, Farplane Gateway.
//
// Type: Land — Town
//
// Oracle text:
//
//	This land enters tapped.
//	{T}: Add {G} or {U}.
var GuadosalamFarplaneGateway = newGuadosalamFarplaneGateway

func newGuadosalamFarplaneGateway() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Green),
		CardFace: game.CardFace{
			Name:     "Guadosalam, Farplane Gateway",
			Types:    []types.Card{types.Land},
			Subtypes: []types.Sub{types.Town},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.G, mana.U),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedReplacement("This land enters tapped."),
			},
			OracleText: `
			This land enters tapped.
			{T}: Add {G} or {U}.
		`,
		},
	}
}
