package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ReturnToDust is the card definition for Return to Dust.
//
// Type: Instant
// Cost: {2}{W}{W}
//
// Oracle text:
//
//	Exile target artifact or enchantment. If you cast this spell during your main phase, you may exile up to one other target artifact or enchantment.
var ReturnToDust = newReturnToDust

func newReturnToDust() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Return to Dust",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target artifact or enchantment",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Enchantment}}),
					},
					game.TargetSpec{
						MinTargets:               0,
						MaxTargets:               1,
						Constraint:               "up to one other target artifact or enchantment",
						Allow:                    game.TargetAllowPermanent,
						Selection:                opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Enchantment}}),
						DistinctFromPriorTargets: true,
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.MovePermanent{
							Object:      game.TargetPermanentReference(0),
							Destination: zone.Exile,
						},
					},
					{
						Primitive: game.MovePermanent{
							Object:      game.TargetPermanentReference(1),
							Destination: zone.Exile,
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								CastDuringControllerMainPhase: true,
							}),
						}),
						Optional: true,
					},
				},
			}.Ability()),
			OracleText: `
			Exile target artifact or enchantment. If you cast this spell during your main phase, you may exile up to one other target artifact or enchantment.
		`,
		},
	}
}
