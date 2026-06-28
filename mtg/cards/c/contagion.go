package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Contagion is the card definition for Contagion.
//
// Type: Instant
// Cost: {3}{B}{B}
//
// Oracle text:
//
//	You may pay 1 life and exile a black card from your hand rather than pay this spell's mana cost.
//	Distribute two -2/-1 counters among one or two target creatures.
var Contagion = newContagion()

func newContagion() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Contagion",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.B,
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Instant},
			AlternativeCosts: []cost.Alternative{
				cost.Alternative{
					Label: "Exile a black card",
					AdditionalCosts: []cost.Additional{
						{
							Kind:   cost.AdditionalPayLife,
							Amount: 1,
						},
						{
							Kind:           cost.AdditionalExile,
							Amount:         1,
							Source:         zone.Hand,
							MatchCardColor: true,
							CardColor:      color.Black,
						},
					},
				},
			},
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
							CounterKind: counter.MinusTwoMinusOne,
							Distribute:  true,
						},
					},
				},
			}.Ability()),
			OracleText: `
			You may pay 1 life and exile a black card from your hand rather than pay this spell's mana cost.
			Distribute two -2/-1 counters among one or two target creatures.
		`,
		},
	}
}
