package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MightOfOldKrosa is the card definition for Might of Old Krosa.
//
// Type: Instant
// Cost: {G}
//
// Oracle text:
//
//	Target creature gets +2/+2 until end of turn. If you cast this spell during your main phase, that creature gets +4/+4 until end of turn instead.
var MightOfOldKrosa = newMightOfOldKrosa

func newMightOfOldKrosa() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Might of Old Krosa",
			ManaCost: opt.Val(cost.Mana{
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
							Object:         game.TargetPermanentReference(0),
							PowerDelta:     game.Fixed(2),
							ToughnessDelta: game.Fixed(2),
							Duration:       game.DurationUntilEndOfTurn,
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								Negate:                        true,
								CastDuringControllerMainPhase: true,
							}),
						}),
					},
					{
						Primitive: game.ModifyPT{
							Object:         game.TargetPermanentReference(0),
							PowerDelta:     game.Fixed(4),
							ToughnessDelta: game.Fixed(4),
							Duration:       game.DurationUntilEndOfTurn,
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								CastDuringControllerMainPhase: true,
							}),
						}),
					},
				},
			}.Ability()),
			OracleText: `
			Target creature gets +2/+2 until end of turn. If you cast this spell during your main phase, that creature gets +4/+4 until end of turn instead.
		`,
		},
	}
}
