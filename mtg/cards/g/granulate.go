package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Granulate is the card definition for Granulate.
//
// Type: Sorcery
// Cost: {2}{R}{R}
//
// Oracle text:
//
//	Destroy each nonland artifact with mana value 4 or less.
var Granulate = newGranulate

func newGranulate() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Granulate",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Destroy{
							Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Artifact}, ExcludedTypes: []types.Card{types.Land}, ManaValue: opt.Val(compare.Int{Op: compare.LessOrEqual, Value: 4})}),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Destroy each nonland artifact with mana value 4 or less.
		`,
		},
	}
}
