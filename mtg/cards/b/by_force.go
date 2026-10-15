package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ByForce is the card definition for By Force.
//
// Type: Sorcery
// Cost: {X}{R}
//
// Oracle text:
//
//	Destroy X target artifacts.
var ByForce = newByForce

func newByForce() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "By Force",
			ManaCost: opt.Val(cost.Mana{
				cost.X,
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets:   0,
						MaxTargets:   20,
						Constraint:   "x target artifacts",
						Allow:        game.TargetAllowPermanent,
						Selection:    opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Artifact}}),
						CountEqualsX: true,
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Destroy{
							Object: game.AllTargetPermanentsReference(0),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Destroy X target artifacts.
		`,
		},
	}
}
