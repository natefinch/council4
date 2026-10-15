package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// Vesuva is the card definition for Vesuva.
//
// Type: Land
//
// Oracle text:
//
//	You may have this land enter tapped as a copy of any land on the battlefield.
var Vesuva = newVesuva

func newVesuva() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:  "Vesuva",
			Types: []types.Card{types.Land},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedAsCopy(game.EntersAsCopyReplacement("You may have this land enter tapped as a copy of any land on the battlefield.", &game.Selection{RequiredTypes: []types.Card{types.Land}}, true, false, nil, false, nil, nil)),
			},
			OracleText: `
			You may have this land enter tapped as a copy of any land on the battlefield.
		`,
		},
	}
}
