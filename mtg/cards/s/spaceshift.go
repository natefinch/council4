package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Spaceshift is the card definition for Spaceshift.
//
// Type: Instant
// Cost: {1}{U}
//
// Oracle text:
//
//	Exile target artifact or creature, then return that card to the battlefield under its owner's control with a +1/+1 counter on it.
var Spaceshift = newSpaceshift

func newSpaceshift() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Spaceshift",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target artifact or creature",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Creature}}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.MovePermanent{
							Object:        game.TargetPermanentReference(0),
							Destination:   zone.Exile,
							PublishLinked: game.LinkedKey("blink-1"),
						},
					},
					{
						Primitive: game.PutOnBattlefield{
							Source: game.LinkedBattlefieldSource(game.LinkedKey("blink-1")),
							EntryCounters: []game.CounterPlacement{
								game.CounterPlacement{
									Kind:   counter.PlusOnePlusOne,
									Amount: 1,
								},
							},
						},
					},
				},
			}.Ability()),
			OracleText: `
			Exile target artifact or creature, then return that card to the battlefield under its owner's control with a +1/+1 counter on it.
		`,
		},
	}
}
