package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BarrageOfBoulders is the card definition for Barrage of Boulders.
//
// Type: Sorcery
// Cost: {2}{R}
//
// Oracle text:
//
//	Barrage of Boulders deals 1 damage to each creature you don't control.
//	Ferocious — If you control a creature with power 4 or greater, creatures can't block this turn.
var BarrageOfBoulders = newBarrageOfBoulders

func newBarrageOfBoulders() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Barrage of Boulders",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Damage{
							Amount:    game.Fixed(1),
							Recipient: game.GroupDamageRecipient(game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerNotYou})),
						},
					},
					{
						Primitive: game.ApplyRule{
							RuleEffects: []game.RuleEffect{
								game.RuleEffect{
									Kind:           game.RuleEffectCantBlock,
									PermanentTypes: []types.Card{types.Creature},
								},
							},
							Duration: game.DurationThisTurn,
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								ControlsMatching: opt.Val(game.SelectionCount{
									Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}, Power: opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 4})},
								}),
							}),
						}),
					},
				},
			}.Ability()),
			OracleText: `
			Barrage of Boulders deals 1 damage to each creature you don't control.
			Ferocious — If you control a creature with power 4 or greater, creatures can't block this turn.
		`,
		},
	}
}
