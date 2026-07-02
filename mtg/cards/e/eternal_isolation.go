package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// EternalIsolation is the card definition for Eternal Isolation.
var EternalIsolation = newEternalIsolation()

func newEternalIsolation() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Eternal Isolation",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target creature with power 4 or greater",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Power: opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 4})}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.PutPermanentOnLibrary{
							Object: game.TargetPermanentReference(0),
							Bottom: true,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Put target creature with power 4 or greater on the bottom of its owner's library.
		`,
		},
	}
}
