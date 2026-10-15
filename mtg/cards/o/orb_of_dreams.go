package o

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// OrbOfDreams is the card definition for Orb of Dreams.
//
// Type: Artifact
// Cost: {3}
//
// Oracle text:
//
//	Permanents enter tapped.
var OrbOfDreams = newOrbOfDreams

func newOrbOfDreams() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Orb of Dreams",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types: []types.Card{types.Artifact},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedGroupReplacement("Permanents enter tapped.", game.TriggerControllerAny),
			},
			OracleText: `
			Permanents enter tapped.
		`,
		},
	}
}
