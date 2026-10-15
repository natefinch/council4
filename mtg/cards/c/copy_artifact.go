package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CopyArtifact is the card definition for Copy Artifact.
//
// Type: Enchantment
// Cost: {1}{U}
//
// Oracle text:
//
//	You may have this enchantment enter as a copy of any artifact on the battlefield, except it's an enchantment in addition to its other types.
var CopyArtifact = newCopyArtifact

func newCopyArtifact() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Copy Artifact",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Enchantment},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersAsCopyReplacement("You may have this enchantment enter as a copy of any artifact on the battlefield, except it's an enchantment in addition to its other types.", &game.Selection{RequiredTypes: []types.Card{types.Artifact}}, true, false, nil, false, nil, nil, types.Enchantment),
			},
			OracleText: `
			You may have this enchantment enter as a copy of any artifact on the battlefield, except it's an enchantment in addition to its other types.
		`,
		},
	}
}
