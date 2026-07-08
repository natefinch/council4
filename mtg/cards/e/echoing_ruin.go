package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// EchoingRuin is the card definition for Echoing Ruin.
//
// Type: Sorcery
// Cost: {1}{R}
//
// Oracle text:
//
//	Destroy target artifact and all other artifacts with the same name as that artifact.
var EchoingRuin = newEchoingRuin

func newEchoingRuin() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Echoing Ruin",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target artifact and all other artifacts with the same name as that artifact",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Artifact}}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Destroy{
							Group: game.SameNamePermanentGroup(game.TargetPermanentReference(0), game.Selection{RequiredTypes: []types.Card{types.Artifact}}),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Destroy target artifact and all other artifacts with the same name as that artifact.
		`,
		},
	}
}
