package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ChainOfAcid is the card definition for Chain of Acid.
//
// Type: Sorcery
// Cost: {3}{G}
//
// Oracle text:
//
//	Destroy target noncreature permanent. Then that permanent's controller may copy this spell and may choose a new target for that copy.
var ChainOfAcid = newChainOfAcid

func newChainOfAcid() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Chain of Acid",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target noncreature permanent",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{ExcludedTypes: []types.Card{types.Creature}}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Destroy{
							Object: game.TargetPermanentReference(0),
						},
					},
					{
						Primitive: game.CopyStackObject{
							Object:              game.ResolvingStackObjectReference(),
							MayChooseNewTargets: true,
							Chooser:             opt.Val(game.AffectedTargetControllerReference(0)),
						},
						Optional:      true,
						OptionalActor: opt.Val(game.AffectedTargetControllerReference(0)),
					},
				},
			}.Ability()),
			OracleText: `
			Destroy target noncreature permanent. Then that permanent's controller may copy this spell and may choose a new target for that copy.
		`,
		},
	}
}
