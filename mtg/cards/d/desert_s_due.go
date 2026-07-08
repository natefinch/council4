package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DesertSDue is the card definition for Desert's Due.
//
// Type: Instant
// Cost: {1}{B}
//
// Oracle text:
//
//	Target creature gets -2/-2 until end of turn. It gets an additional -1/-1 until end of turn for each Desert you control.
var DesertSDue = newDesertSDue

func newDesertSDue() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Desert's Due",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
			}),
			Colors: []color.Color{color.Black},
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
							Object:         game.TargetPermanentReference(0),
							PowerDelta:     game.Fixed(-2),
							ToughnessDelta: game.Fixed(-2),
							Duration:       game.DurationUntilEndOfTurn,
						},
					},
					{
						Primitive: game.ModifyPT{
							Object: game.TargetPermanentReference(0),
							PowerDelta: game.Dynamic(game.DynamicAmount{
								Kind:       game.DynamicAmountCountSelector,
								Multiplier: -1,
								Group:      game.BattlefieldGroup(game.Selection{SubtypesAny: []types.Sub{types.Sub("Desert")}, Controller: game.ControllerYou}),
							}),
							ToughnessDelta: game.Dynamic(game.DynamicAmount{
								Kind:       game.DynamicAmountCountSelector,
								Multiplier: -1,
								Group:      game.BattlefieldGroup(game.Selection{SubtypesAny: []types.Sub{types.Sub("Desert")}, Controller: game.ControllerYou}),
							}),
							Duration: game.DurationUntilEndOfTurn,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Target creature gets -2/-2 until end of turn. It gets an additional -1/-1 until end of turn for each Desert you control.
		`,
		},
	}
}
