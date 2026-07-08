package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// FilterOut is the card definition for Filter Out.
//
// Type: Instant
// Cost: {1}{U}{U}
//
// Oracle text:
//
//	Return all noncreature, nonland permanents to their owners' hands.
var FilterOut = newFilterOut

func newFilterOut() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Filter Out",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Bounce{
							Group: game.BattlefieldGroup(game.Selection{ExcludedTypes: []types.Card{types.Creature, types.Land}}),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Return all noncreature, nonland permanents to their owners' hands.
		`,
		},
	}
}
