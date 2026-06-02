package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
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
	Name: "Malamet Battle Glyph",
	ManaCost: opt.Val(mana.Cost{
		mana.ColoredMana(mana.Green),
	}),
	Colors:        []mana.Color{mana.Green},
	ColorIdentity: mana.NewColorIdentity(mana.Green),
	Types:         []types.Card{types.Sorcery},
	OracleText:    "Choose target creature you control and target creature you don't control. If the creature you control entered this turn, put a +1/+1 counter on it. Then those creatures fight each other.",
	Abilities: []game.AbilityDef{
		{
			Kind: game.SpellAbility,
			Text: "Choose target creature you control and target creature you don't control. If the creature you control entered this turn, put a +1/+1 counter on it. Then those creatures fight each other.",
			Targets: []game.TargetSpec{
				{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "creature you control",
					Allow:      game.TargetAllowPermanent,
					Predicate: game.TargetPredicate{
						PermanentTypes: []types.Card{types.Creature},
						Controller:     game.ControllerYou,
					},
				},
				{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "creature you don't control",
					Allow:      game.TargetAllowPermanent,
					Predicate: game.TargetPredicate{
						PermanentTypes: []types.Card{types.Creature},
						Controller:     game.ControllerOpponent,
					},
				},
			},
			Effects: []game.Effect{
				{
					Type:        game.EffectAddCounter,
					Amount:      1,
					CounterKind: counter.PlusOnePlusOne,
					TargetIndex: 0,
					Condition: opt.Val(game.EffectCondition{
						Condition: opt.Val(game.Condition{TargetEnteredThisTurn: opt.Val(0)}),
					}),
					Description: "if the creature you control entered this turn, put a +1/+1 counter on it",
				},
				{Type: game.EffectFight},
			},
		},
	},
}
