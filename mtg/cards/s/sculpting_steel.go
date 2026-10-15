package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SculptingSteel is the card definition for Sculpting Steel.
//
// Type: Artifact
// Cost: {3}
//
// Oracle text:
//
//	You may have this artifact enter as a copy of any artifact on the battlefield.
var SculptingSteel = newSculptingSteel

func newSculptingSteel() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Sculpting Steel",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types: []types.Card{types.Artifact},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersAsCopyReplacement("You may have this artifact enter as a copy of any artifact on the battlefield.", &game.Selection{RequiredTypes: []types.Card{types.Artifact}}, true, false, nil, false, nil, nil),
			},
			OracleText: `
			You may have this artifact enter as a copy of any artifact on the battlefield.
		`,
		},
	}
}
