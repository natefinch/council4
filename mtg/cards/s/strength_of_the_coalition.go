package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// StrengthOfTheCoalition is the card definition for Strength of the Coalition.
//
// Type: Instant
// Cost: {G}
//
// Oracle text:
//
//	Kicker {2}{W} (You may pay an additional {2}{W} as you cast this spell.)
//	Target creature you control gets +2/+2 until end of turn. If this spell was kicked, put a +1/+1 counter on each creature you control.
var StrengthOfTheCoalition = newStrengthOfTheCoalition

func newStrengthOfTheCoalition() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Green),
		CardFace: game.CardFace{
			Name: "Strength of the Coalition",
			ManaCost: opt.Val(cost.Mana{
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Instant},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					KeywordAbilities: []game.KeywordAbility{
						game.KickerKeyword{Cost: cost.Mana{cost.O(2), cost.W}},
					},
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target creature you control",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Controller: game.ControllerYou}),
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
					},
					{
						Primitive: game.AddCounter{
							Amount:      game.Fixed(1),
							Group:       game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
							CounterKind: counter.PlusOnePlusOne,
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								SpellWasKicked: true,
							}),
						}),
					},
				},
			}.Ability()),
			OracleText: `
			Kicker {2}{W} (You may pay an additional {2}{W} as you cast this spell.)
			Target creature you control gets +2/+2 until end of turn. If this spell was kicked, put a +1/+1 counter on each creature you control.
		`,
		},
	}
}
