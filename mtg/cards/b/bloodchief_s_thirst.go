package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BloodchiefSThirst is the card definition for Bloodchief's Thirst.
//
// Type: Sorcery
// Cost: {B}
//
// Oracle text:
//
//	Kicker {2}{B} (You may pay an additional {2}{B} as you cast this spell.)
//	Destroy target creature or planeswalker with mana value 2 or less. If this spell was kicked, instead destroy target creature or planeswalker.
var BloodchiefSThirst = newBloodchiefSThirst

func newBloodchiefSThirst() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Bloodchief's Thirst",
			ManaCost: opt.Val(cost.Mana{
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Sorcery},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					KeywordAbilities: []game.KeywordAbility{
						game.KickerKeyword{Cost: cost.Mana{cost.O(2), cost.B}},
					},
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target creature or planeswalker with mana value 2 or less",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature, types.Planeswalker}, ManaValue: opt.Val(compare.Int{Op: compare.LessOrEqual, Value: 2})}),
						Gate:       game.TargetGateSpellNotKicked,
					},
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target creature or planeswalker",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature, types.Planeswalker}}),
						Gate:       game.TargetGateSpellKicked,
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Destroy{
							Object: game.TargetPermanentReference(0),
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								Negate:         true,
								SpellWasKicked: true,
							}),
						}),
					},
					{
						Primitive: game.Destroy{
							Object: game.TargetPermanentReference(1),
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
			Kicker {2}{B} (You may pay an additional {2}{B} as you cast this spell.)
			Destroy target creature or planeswalker with mana value 2 or less. If this spell was kicked, instead destroy target creature or planeswalker.
		`,
		},
	}
}
