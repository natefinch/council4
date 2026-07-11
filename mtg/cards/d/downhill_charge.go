package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DownhillCharge is the card definition for Downhill Charge.
//
// Type: Instant
// Cost: {2}{R}
//
// Oracle text:
//
//	You may sacrifice a Mountain rather than pay this spell's mana cost.
//	Target creature gets +X/+0 until end of turn, where X is the number of Mountains you control.
var DownhillCharge = newDownhillCharge

func newDownhillCharge() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Downhill Charge",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Instant},
			AlternativeCosts: []cost.Alternative{
				cost.Alternative{
					Label: "Sacrifice a Mountain",
					AdditionalCosts: []cost.Additional{
						{
							Kind:        cost.AdditionalSacrifice,
							Text:        "sacrifice a Mountain",
							Amount:      1,
							SubtypesAny: cost.SubtypeSet{types.Mountain},
						},
					},
				},
			},
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
								Kind:       game.DynamicAmountCountSelector,
								Multiplier: 1,
								Group:      game.BattlefieldGroup(game.Selection{SubtypesAny: []types.Sub{types.Sub("Mountain")}, Controller: game.ControllerYou}),
							}),
							ToughnessDelta: game.Fixed(0),
							Duration:       game.DurationUntilEndOfTurn,
						},
					},
				},
			}.Ability()),
			OracleText: `
			You may sacrifice a Mountain rather than pay this spell's mana cost.
			Target creature gets +X/+0 until end of turn, where X is the number of Mountains you control.
		`,
		},
	}
}
