package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// IntoTheFray is the card definition for Into the Fray.
//
// Type: Instant — Arcane
// Cost: {R}
//
// Oracle text:
//
//	Target creature attacks this turn if able.
//	Splice onto Arcane {R} (As you cast an Arcane spell, you may reveal this card from your hand and pay its splice cost. If you do, add this card's effects to that spell.)
var IntoTheFray = newIntoTheFray

func newIntoTheFray() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Into the Fray",
			ManaCost: opt.Val(cost.Mana{
				cost.R,
			}),
			Colors:   []color.Color{color.Red},
			Types:    []types.Card{types.Instant},
			Subtypes: []types.Sub{types.Arcane},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					KeywordAbilities: []game.KeywordAbility{
						game.SpliceKeyword{Cost: cost.Mana{cost.R}},
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
						Primitive: game.ApplyRule{
							Object: opt.Val(game.TargetPermanentReference(0)),
							RuleEffects: []game.RuleEffect{
								game.RuleEffect{
									Kind: game.RuleEffectMustAttack,
								},
							},
							Duration: game.DurationThisTurn,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Target creature attacks this turn if able.
			Splice onto Arcane {R} (As you cast an Arcane spell, you may reveal this card from your hand and pay its splice cost. If you do, add this card's effects to that spell.)
		`,
		},
	}
}
