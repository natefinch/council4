package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BogDown is the card definition for Bog Down.
//
// Type: Sorcery
// Cost: {2}{B}
//
// Oracle text:
//
//	Kicker—Sacrifice two lands. (You may sacrifice two lands in addition to any other costs as you cast this spell.)
//	Target player discards two cards. If this spell was kicked, that player discards three cards instead.
var BogDown = newBogDown

func newBogDown() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Bog Down",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Sorcery},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Kicker ",
					KeywordAbilities: []game.KeywordAbility{
						game.KickerKeyword{
							AdditionalCosts: []cost.Additional{
								cost.Additional{
									Kind:               cost.AdditionalSacrifice,
									Text:               "Sacrifice two lands",
									Amount:             2,
									MatchPermanentType: true,
									PermanentType:      types.Land,
									CounterKind:        counter.PlusOnePlusOne,
								},
							},
						},
					},
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "Target player",
						Allow:      game.TargetAllowPlayer,
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Discard{
							Amount: game.Fixed(2),
							Player: game.TargetPlayerReference(0),
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								Negate:         true,
								SpellWasKicked: true,
							}),
						}),
						PublishCondition: "condition-0",
					},
					{
						Primitive: game.Discard{
							Amount: game.Fixed(3),
							Player: game.TargetPlayerReference(0),
						},
						ConditionGate:       "condition-0",
						ConditionGateNegate: true,
					},
				},
			}.Ability()),
			OracleText: `
			Kicker—Sacrifice two lands. (You may sacrifice two lands in addition to any other costs as you cast this spell.)
			Target player discards two cards. If this spell was kicked, that player discards three cards instead.
		`,
		},
	}
}
