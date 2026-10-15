package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CursedMirror is the card definition for Cursed Mirror.
//
// Type: Artifact
// Cost: {2}{R}
//
// Oracle text:
//
//	{T}: Add {R}.
//	As this artifact enters, you may have it become a copy of any creature on the battlefield until end of turn, except it has haste.
var CursedMirror = newCursedMirror

func newCursedMirror() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Cursed Mirror",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Artifact},
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.R),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersAsCopyReplacement("As this artifact enters, you may have it become a copy of any creature on the battlefield until end of turn, except it has haste.", &game.Selection{RequiredTypes: []types.Card{types.Creature}}, true, false, nil, true, []game.Keyword{game.Haste}, nil),
			},
			OracleText: `
			{T}: Add {R}.
			As this artifact enters, you may have it become a copy of any creature on the battlefield until end of turn, except it has haste.
		`,
		},
	}
}
