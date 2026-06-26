package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CruxOfFate is the card definition for Crux of Fate.
//
// Type: Sorcery
// Cost: {3}{B}{B}
//
// Oracle text:
//
//	Choose one —
//	• Destroy all Dragon creatures.
//	• Destroy all non-Dragon creatures.
var CruxOfFate = newCruxOfFate()

func newCruxOfFate() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Crux of Fate",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.B,
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					game.Mode{
						Text: "Destroy all Dragon creatures.",
						Sequence: []game.Instruction{
							{
								Primitive: game.Destroy{
									Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, SubtypesAny: []types.Sub{types.Sub("Dragon")}}),
								},
							},
						},
					},
					game.Mode{
						Text: "Destroy all non-Dragon creatures.",
						Sequence: []game.Instruction{
							{
								Primitive: game.Destroy{
									Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, ExcludedSubtype: types.Sub("Dragon")}),
								},
							},
						},
					},
				},
				MinModes: 1,
				MaxModes: 1,
			}),
			OracleText: `
			Choose one —
			• Destroy all Dragon creatures.
			• Destroy all non-Dragon creatures.
		`,
		},
	}
}
