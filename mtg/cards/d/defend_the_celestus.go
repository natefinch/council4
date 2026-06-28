package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DefendTheCelestus is the card definition for Defend the Celestus.
//
// Type: Instant
// Cost: {2}{G}{G}
//
// Oracle text:
//
//	Distribute three +1/+1 counters among one, two, or three target creatures you control.
var DefendTheCelestus = newDefendTheCelestus()

func newDefendTheCelestus() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Defend the Celestus",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 3,
						Constraint: "one, two, or three target creatures you control",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Controller: game.ControllerYou}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.AddCounter{
							Amount:      game.Fixed(3),
							Object:      game.AllTargetPermanentsReference(0),
							CounterKind: counter.PlusOnePlusOne,
							Distribute:  true,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Distribute three +1/+1 counters among one, two, or three target creatures you control.
		`,
		},
	}
}
