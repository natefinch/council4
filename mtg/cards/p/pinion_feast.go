package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// PinionFeast is the card definition for Pinion Feast.
//
// Type: Instant
// Cost: {4}{G}
//
// Oracle text:
//
//	Destroy target creature with flying. Bolster 2. (Choose a creature with the least toughness among creatures you control and put two +1/+1 counters on it.)
var PinionFeast = newPinionFeast

func newPinionFeast() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Pinion Feast",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target creature with flying",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Keyword: game.Flying}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Destroy{
							Object: game.TargetPermanentReference(0),
						},
					},
					{
						Primitive: game.Bolster{
							Amount: game.Fixed(2),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Destroy target creature with flying. Bolster 2. (Choose a creature with the least toughness among creatures you control and put two +1/+1 counters on it.)
		`,
		},
	}
}
