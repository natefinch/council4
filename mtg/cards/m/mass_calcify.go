package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MassCalcify is the card definition for Mass Calcify.
//
// Type: Sorcery
// Cost: {5}{W}{W}
//
// Oracle text:
//
//	Destroy all nonwhite creatures.
var MassCalcify = newMassCalcify

func newMassCalcify() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Mass Calcify",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.W,
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Destroy{
							Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, ExcludedColors: []color.Color{color.White}}),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Destroy all nonwhite creatures.
		`,
		},
	}
}
