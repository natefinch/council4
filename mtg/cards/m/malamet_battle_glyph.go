package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MalametBattleGlyph is the card definition for Malamet Battle Glyph.
//
// Type: Sorcery
// Cost: {G}
//
// Oracle text:
//
//	Choose target creature you control and target creature you don't control. If the creature you control entered this turn, put a +1/+1 counter on it. Then those creatures fight each other.
var MalametBattleGlyph = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Green),
	CardFace: game.CardFace{
		Name: "Malamet Battle Glyph",
		ManaCost: opt.Val(cost.Mana{
			cost.G,
		}),
		Colors: []color.Color{color.Green},
		Types:  []types.Card{types.Sorcery},
		OracleText: `
			Choose target creature you control and target creature you don't control. If the creature you control entered this turn, put a +1/+1 counter on it. Then those creatures fight each other.
		`,
		SpellAbility: opt.Val(game.Mode{
			Targets: []game.TargetSpec{
				{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "creature you control",
					Allow:      game.TargetAllowPermanent,
					Predicate: game.TargetPredicate{
						PermanentTypes: []types.Card{
							types.Creature,
						},
						Controller: game.ControllerYou,
					},
				},
				{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "creature you don't control",
					Allow:      game.TargetAllowPermanent,
					Predicate: game.TargetPredicate{
						PermanentTypes: []types.Card{
							types.Creature,
						},
						Controller: game.ControllerOpponent,
					},
				},
			},
			Sequence: []game.Instruction{
				{
					Primitive: game.AddCounter{
						Amount:      game.Fixed(1),
						TargetIndex: 0,
						CounterKind: counter.PlusOnePlusOne,
					},
					Condition: opt.Val(game.EffectCondition{
						Condition: opt.Val(game.Condition{
							TargetEnteredThisTurn: opt.Val(0),
						}),
					}),
					Description: "if the creature you control entered this turn, put a +1/+1 counter on it",
				},
				{
					Primitive: game.Fight{},
				},
			},
		}.Ability()),
	},
}
