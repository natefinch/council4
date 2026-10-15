package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AcceleratedMutation is the card definition for Accelerated Mutation.
//
// Type: Instant
// Cost: {3}{G}{G}
//
// Oracle text:
//
//	Target creature gets +X/+X until end of turn, where X is the greatest mana value among permanents you control.
var AcceleratedMutation = newAcceleratedMutation

func newAcceleratedMutation() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Accelerated Mutation",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.G,
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target creature",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.ModifyPT{
							Object: game.TargetPermanentReference(0),
							PowerDelta: game.Dynamic(game.DynamicAmount{
								Kind:       game.DynamicAmountGreatestManaValueInGroup,
								Multiplier: 1,
								Group:      game.BattlefieldGroup(game.Selection{Controller: game.ControllerYou}),
							}),
							ToughnessDelta: game.Dynamic(game.DynamicAmount{
								Kind:       game.DynamicAmountGreatestManaValueInGroup,
								Multiplier: 1,
								Group:      game.BattlefieldGroup(game.Selection{Controller: game.ControllerYou}),
							}),
							Duration: game.DurationUntilEndOfTurn,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Target creature gets +X/+X until end of turn, where X is the greatest mana value among permanents you control.
		`,
		},
	}
}
