package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// PhyrexianMetamorph is the card definition for Phyrexian Metamorph.
//
// Type: Artifact Creature — Phyrexian Shapeshifter
// Cost: {3}{U/P}
//
// Oracle text:
//
//	({U/P} can be paid with either {U} or 2 life.)
//	You may have this creature enter as a copy of any artifact or creature on the battlefield, except it's an artifact in addition to its other types.
var PhyrexianMetamorph = newPhyrexianMetamorph

func newPhyrexianMetamorph() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Phyrexian Metamorph",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.PhyrexianMana(mana.U),
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Phyrexian, types.Shapeshifter},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 0}),
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersAsCopyReplacement("You may have this creature enter as a copy of any artifact or creature on the battlefield, except it's an artifact in addition to its other types.", &game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Creature}}, true, false, nil, false, nil, nil, types.Artifact),
			},
			OracleText: `
			({U/P} can be paid with either {U} or 2 life.)
			You may have this creature enter as a copy of any artifact or creature on the battlefield, except it's an artifact in addition to its other types.
		`,
		},
	}
}
