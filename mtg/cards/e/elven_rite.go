package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ElvenRite is the card definition for Elven Rite.
//
// Type: Sorcery
// Cost: {1}{G}
//
// Oracle text:
//
//	Distribute two +1/+1 counters among one or two target creatures.
var ElvenRite = newElvenRite

func newElvenRite() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Elven Rite",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 2,
						Constraint: "one or two target creatures",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.AddCounter{
							Amount:      game.Fixed(2),
							Object:      game.AllTargetPermanentsReference(0),
							CounterKind: counter.PlusOnePlusOne,
							Distribute:  true,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Distribute two +1/+1 counters among one or two target creatures.
		`,
		},
	}
}
