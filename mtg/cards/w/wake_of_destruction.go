package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// WakeOfDestruction is the card definition for Wake of Destruction.
//
// Type: Sorcery
// Cost: {3}{R}{R}{R}
//
// Oracle text:
//
//	Destroy target land and all other lands with the same name as that land.
var WakeOfDestruction = newWakeOfDestruction()

func newWakeOfDestruction() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Wake of Destruction",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.R,
				cost.R,
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target land and all other lands with the same name as that land",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Land}}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Destroy{
							Group: game.SameNamePermanentGroup(game.TargetPermanentReference(0), game.Selection{RequiredTypes: []types.Card{types.Land}}),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Destroy target land and all other lands with the same name as that land.
		`,
		},
	}
}
