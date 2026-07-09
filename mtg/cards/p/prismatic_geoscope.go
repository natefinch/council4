package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// PrismaticGeoscope is the card definition for Prismatic Geoscope.
//
// Type: Artifact
// Cost: {5}
//
// Oracle text:
//
//	This artifact enters tapped.
//	Domain — {T}: Add X mana in any combination of colors, where X is the number of basic land types among lands you control.
var PrismaticGeoscope = newPrismaticGeoscope

func newPrismaticGeoscope() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Prismatic Geoscope",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
			}),
			Types: []types.Card{types.Artifact},
			ManaAbilities: []game.ManaAbility{
				game.ManaAbility{
					AdditionalCosts: cost.Tap,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.AddMana{
									Amount: game.Dynamic(game.DynamicAmount{
										Kind:       game.DynamicAmountControllerBasicLandTypeCount,
										Multiplier: 1,
									}),
									CombinationColors: []mana.Color{mana.W, mana.U, mana.B, mana.R, mana.G},
								},
							},
						},
					}.Ability(),
				},
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedReplacement("This artifact enters tapped."),
			},
			OracleText: `
			This artifact enters tapped.
			Domain — {T}: Add X mana in any combination of colors, where X is the number of basic land types among lands you control.
		`,
		},
	}
}
