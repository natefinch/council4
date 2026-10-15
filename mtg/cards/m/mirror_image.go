package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MirrorImage is the card definition for Mirror Image.
//
// Type: Creature — Shapeshifter
// Cost: {2}{U}
//
// Oracle text:
//
//	You may have this creature enter as a copy of a creature you control.
var MirrorImage = newMirrorImage

func newMirrorImage() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Mirror Image",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Shapeshifter},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 0}),
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersAsCopyReplacement("You may have this creature enter as a copy of a creature you control.", &game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}, true, false, nil, false, nil, nil),
			},
			OracleText: `
			You may have this creature enter as a copy of a creature you control.
		`,
		},
	}
}
