// Legacy regression fixture preserved when the Oracle compiler was retired.
package fixtures

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RTCounteredSpellCondition is the card definition for RT Countered Spell Condition.
//
// Type: Instant
//
// Oracle text:
//
//	Counter target spell. If that spell's mana value was 3 or less, proliferate.
var RTCounteredSpellCondition = newRTCounteredSpellCondition

func newRTCounteredSpellCondition() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:  "RT Countered Spell Condition",
			Types: []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target spell",
						Allow:      game.TargetAllowStackObject,
						Predicate: game.TargetPredicate{
							StackObjectKinds: []game.StackObjectKind{game.StackSpell},
						},
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.CounterObject{
							Object: game.TargetStackObjectReference(0),
						},
					},
					{
						Primitive: game.Proliferate{
							Amount: game.Fixed(1),
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								Object:                     opt.Val(game.TargetStackObjectReference(0)),
								ObjectMatches:              opt.Val(game.Selection{ManaValue: opt.Val(compare.Int{Op: compare.LessOrEqual, Value: 3})}),
								UseCounteredSpellManaValue: true,
							}),
						}),
					},
				},
			}.Ability()),
			OracleText: `
			Counter target spell. If that spell's mana value was 3 or less, proliferate.
		`,
		},
	}
}
