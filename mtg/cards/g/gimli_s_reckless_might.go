package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"

	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GimliSRecklessMight is the card definition for Gimli's Reckless Might.
//
// Type: Enchantment
// Cost: {3}{R}
//
// Oracle text:
//
//	Creatures you control have haste.
//	Formidable — Whenever you attack, if creatures you control have total power 8 or greater, target attacking creature you control fights up to one target creature you don't control.
var GimliSRecklessMight = &game.CardDef{
	Name: "Gimli's Reckless Might",
	ManaCost: opt.Val(cost.Mana{
		cost.O(3),
		cost.R,
	}),
	Colors:        []color.Color{color.Red},
	ColorIdentity: color.NewIdentity(color.Red),
	Types:         []types.Card{types.Enchantment},
	OracleText:    "Creatures you control have haste.\nFormidable — Whenever you attack, if creatures you control have total power 8 or greater, target attacking creature you control fights up to one target creature you don't control.",
	Abilities: []game.AbilityDef{
		{
			Kind: game.StaticAbility,
			Text: "Creatures you control have haste.",
			Effects: []game.Effect{
				{
					Type:        game.EffectApplyContinuous,
					TargetIndex: game.TargetIndexController,
					ContinuousEffects: []game.ContinuousEffect{
						{
							Layer:       game.LayerAbility,
							Selector:    game.EffectSelectorCreaturesYouControl,
							AddKeywords: []game.Keyword{game.Haste},
						},
					},
				},
			},
		},
		{
			Kind: game.TriggeredAbility,
			Text: "Formidable — Whenever you attack, if creatures you control have total power 8 or greater, target attacking creature you control fights up to one target creature you don't control.",
			Trigger: opt.Val(game.TriggerCondition{
				Type: game.TriggerWhenever,
				Pattern: game.TriggerPattern{
					Event:      game.EventAttackerDeclared,
					Controller: game.TriggerControllerYou,
				},
				InterveningIf: "creatures you control have total power 8 or greater",
				InterveningCondition: opt.Val(game.Condition{
					Text: "creatures you control have total power 8 or greater",
					ControllerControls: game.PermanentFilter{
						Types:      []types.Card{types.Creature},
						TotalPower: opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 8}),
					},
				}),
			}),
			Targets: []game.TargetSpec{
				{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "attacking creature you control",
					Allow:      game.TargetAllowPermanent,
					Predicate: game.TargetPredicate{
						PermanentTypes: []types.Card{types.Creature},
						Controller:     game.ControllerYou,
						CombatState:    game.CombatStateAttacking,
					},
				},
				{
					MinTargets: 0,
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
				{Type: game.EffectFight},
			},
		},
	},
}
