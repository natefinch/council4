package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SakashimaOfAThousandFaces is the card definition for Sakashima of a Thousand Faces.
//
// Type: Legendary Creature — Human Rogue
// Cost: {3}{U}
//
// Oracle text:
//
//	You may have Sakashima enter as a copy of another creature you control, except it has Sakashima's other abilities.
//	The "legend rule" doesn't apply to permanents you control.
//	Partner (You can have two commanders if both have partner.)
var SakashimaOfAThousandFaces = newSakashimaOfAThousandFaces

func newSakashimaOfAThousandFaces() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Sakashima of a Thousand Faces",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.U,
			}),
			Colors:     []color.Color{color.Blue},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Rogue},
			Power:      opt.Val(game.PT{Value: 3}),
			Toughness:  opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.LegendRuleDoesNotApplyStaticBody,
				game.PartnerStaticBody,
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersAsCopyWithOtherAbilities(game.EntersAsCopyWithRetainedName(game.EntersAsCopyReplacement("You may have Sakashima enter as a copy of another creature you control, except it has Sakashima's other abilities.", &game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou, ExcludeSource: true}, true, false, nil, false, nil, nil))),
			},
			OracleText: `
			You may have Sakashima enter as a copy of another creature you control, except it has Sakashima's other abilities.
			The "legend rule" doesn't apply to permanents you control.
			Partner (You can have two commanders if both have partner.)
		`,
		},
	}
}
